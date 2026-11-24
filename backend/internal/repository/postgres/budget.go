package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/repository/postgres/db"
)

type BudgetRepository struct {
	queries *db.Queries
}

func NewBudgetRepository(pool *pgxpool.Pool) *BudgetRepository {
	return &BudgetRepository{queries: db.New(pool)}
}

func (r *BudgetRepository) List(ctx context.Context, userID string, month time.Time) ([]domain.Budget, error) {
	rows, err := r.queries.ListBudgets(ctx, db.ListBudgetsParams{UserID: userID, Month: month})
	if err != nil {
		return nil, fmt.Errorf("select budgets: %w", err)
	}

	budgets := make([]domain.Budget, 0, len(rows))
	for _, row := range rows {
		budgets = append(budgets, toDomainBudget(db.GetBudgetRow(row)))
	}
	return budgets, nil
}

func (r *BudgetRepository) Get(ctx context.Context, userID, id string) (domain.Budget, error) {
	row, err := r.queries.GetBudget(ctx, db.GetBudgetParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) || hasCode(err, codeInvalidText) {
		return domain.Budget{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Budget{}, fmt.Errorf("select budget: %w", err)
	}
	return toDomainBudget(row), nil
}

func (r *BudgetRepository) Create(ctx context.Context, b domain.Budget) (domain.Budget, error) {
	id, err := r.queries.CreateBudget(ctx, db.CreateBudgetParams{
		UserID:     b.UserID,
		CategoryID: b.CategoryID,
		Month:      b.Month,
		Amount:     b.Amount,
	})
	if hasCode(err, codeUniqueViolation) {
		return domain.Budget{}, domain.ErrBudgetExists
	}
	if err != nil {
		return domain.Budget{}, fmt.Errorf("insert budget: %w", err)
	}
	return r.Get(ctx, b.UserID, id)
}

func (r *BudgetRepository) UpdateAmount(ctx context.Context, userID, id string, amount int64) (domain.Budget, error) {
	affected, err := r.queries.UpdateBudgetAmount(ctx, db.UpdateBudgetAmountParams{ID: id, UserID: userID, Amount: amount})
	switch {
	case hasCode(err, codeInvalidText):
		return domain.Budget{}, domain.ErrNotFound
	case err != nil:
		return domain.Budget{}, fmt.Errorf("update budget: %w", err)
	case affected == 0:
		return domain.Budget{}, domain.ErrNotFound
	}
	return r.Get(ctx, userID, id)
}

func (r *BudgetRepository) Delete(ctx context.Context, userID, id string) error {
	deleted, err := r.queries.DeleteBudget(ctx, db.DeleteBudgetParams{ID: id, UserID: userID})
	switch {
	case hasCode(err, codeInvalidText):
		return domain.ErrNotFound
	case err != nil:
		return fmt.Errorf("delete budget: %w", err)
	case deleted == 0:
		return domain.ErrNotFound
	}
	return nil
}

func (r *BudgetRepository) CopyMonth(ctx context.Context, userID string, from, to time.Time) (int64, error) {
	copied, err := r.queries.CopyBudgets(ctx, db.CopyBudgetsParams{UserID: userID, FromMonth: from, ToMonth: to})
	if err != nil {
		return 0, fmt.Errorf("copy budgets: %w", err)
	}
	return copied, nil
}

func toDomainBudget(row db.GetBudgetRow) domain.Budget {
	return domain.Budget{
		ID:           row.ID,
		UserID:       row.UserID,
		CategoryID:   row.CategoryID,
		CategoryName: row.CategoryName,
		Month:        row.Month,
		Amount:       row.Amount,
		Spent:        row.Spent,
	}
}
