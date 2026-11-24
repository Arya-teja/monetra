package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

const (
	warningPercent = 80
	overPercent    = 100
)

// BudgetSummary is a budget together with how much of it has been used.
type BudgetSummary struct {
	domain.Budget
	Percentage int64 // spent / amount * 100, rounded down; for display only
	Status     domain.BudgetStatus
}

type BudgetService struct {
	budgets    domain.BudgetRepository
	categories domain.CategoryRepository
}

func NewBudgetService(budgets domain.BudgetRepository, categories domain.CategoryRepository) *BudgetService {
	return &BudgetService{budgets: budgets, categories: categories}
}

// List returns the budgets of a month (YYYY-MM) with their spending status.
func (s *BudgetService) List(ctx context.Context, userID, month string) ([]BudgetSummary, error) {
	start, err := parseMonth(month)
	if err != nil {
		return nil, err
	}

	budgets, err := s.budgets.List(ctx, userID, start)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}

	summaries := make([]BudgetSummary, 0, len(budgets))
	for _, b := range budgets {
		summaries = append(summaries, summarize(b))
	}
	return summaries, nil
}

func (s *BudgetService) Create(ctx context.Context, userID, categoryID, month string, amount int64) (BudgetSummary, error) {
	fields := map[string]string{}
	start, err := time.Parse(monthLayout, month)
	if err != nil {
		fields["month"] = "must be YYYY-MM"
	}
	if amount <= 0 {
		fields["amount"] = "must be greater than 0"
	}

	category, err := s.categories.Get(ctx, userID, categoryID)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		fields["categoryId"] = msgNotFound
	case err != nil:
		return BudgetSummary{}, fmt.Errorf("get category: %w", err)
	case category.Type != domain.CashflowExpense:
		fields["categoryId"] = "must be an expense category"
	}
	if len(fields) > 0 {
		return BudgetSummary{}, &domain.ValidationError{Fields: fields}
	}

	created, err := s.budgets.Create(ctx, domain.Budget{
		UserID:     userID,
		CategoryID: categoryID,
		Month:      start,
		Amount:     amount,
	})
	if err != nil {
		return BudgetSummary{}, fmt.Errorf("create budget: %w", err)
	}
	return summarize(created), nil
}

func (s *BudgetService) UpdateAmount(ctx context.Context, userID, id string, amount int64) (BudgetSummary, error) {
	if amount <= 0 {
		return BudgetSummary{}, &domain.ValidationError{Fields: map[string]string{"amount": "must be greater than 0"}}
	}

	updated, err := s.budgets.UpdateAmount(ctx, userID, id, amount)
	if err != nil {
		return BudgetSummary{}, fmt.Errorf("update budget: %w", err)
	}
	return summarize(updated), nil
}

func (s *BudgetService) Delete(ctx context.Context, userID, id string) error {
	if err := s.budgets.Delete(ctx, userID, id); err != nil {
		return fmt.Errorf("delete budget: %w", err)
	}
	return nil
}

// CopyFromPreviousMonth copies last month's budgets into month (YYYY-MM) and returns
// how many were created. Categories that already have a budget in month are kept.
func (s *BudgetService) CopyFromPreviousMonth(ctx context.Context, userID, month string) (int64, error) {
	start, err := parseMonth(month)
	if err != nil {
		return 0, err
	}

	copied, err := s.budgets.CopyMonth(ctx, userID, start.AddDate(0, -1, 0), start)
	if err != nil {
		return 0, fmt.Errorf("copy budgets: %w", err)
	}
	return copied, nil
}

func summarize(b domain.Budget) BudgetSummary {
	return BudgetSummary{
		Budget:     b,
		Percentage: b.Spent * 100 / b.Amount,
		Status:     budgetStatus(b.Spent, b.Amount),
	}
}

// budgetStatus compares with integer math so that e.g. 79.9% is still safe.
func budgetStatus(spent, amount int64) domain.BudgetStatus {
	switch {
	case spent*100 >= amount*overPercent:
		return domain.BudgetOver
	case spent*100 >= amount*warningPercent:
		return domain.BudgetWarning
	default:
		return domain.BudgetSafe
	}
}

// parseMonth turns "YYYY-MM" into the first day of that month.
func parseMonth(month string) (time.Time, error) {
	start, err := time.Parse(monthLayout, month)
	if err != nil {
		return time.Time{}, &domain.ValidationError{Fields: map[string]string{"month": "must be YYYY-MM"}}
	}
	return start, nil
}
