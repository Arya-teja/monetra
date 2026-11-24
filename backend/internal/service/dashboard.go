package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

const (
	topCategoryCount    = 5
	recentCount         = 5
	defaultTrendMonths  = 6
	maxTrendMonths      = 24
	jakartaOffsetSecond = 7 * 60 * 60
)

// jakarta is Asia/Jakarta (WIB). Indonesia has no daylight saving time, so a fixed
// UTC+7 zone is exact and does not depend on the server's time zone database.
var jakarta = time.FixedZone("WIB", jakartaOffsetSecond)

type DashboardSummary struct {
	Month        time.Time
	TotalBalance int64
	Income       int64
	Expense      int64
	// UsedPercentage is expense / income * 100 rounded down; nil when income is 0.
	UsedPercentage     *int64
	TopCategories      []domain.CategoryTotal
	BudgetAlerts       []BudgetSummary // budgets with status warning or over
	RecentTransactions []domain.Transaction
}

type DashboardService struct {
	dashboard    domain.DashboardRepository
	transactions domain.TransactionRepository
	budgets      domain.BudgetRepository
	now          func() time.Time
}

func NewDashboardService(
	dashboard domain.DashboardRepository,
	transactions domain.TransactionRepository,
	budgets domain.BudgetRepository,
	now func() time.Time,
) *DashboardService {
	return &DashboardService{dashboard: dashboard, transactions: transactions, budgets: budgets, now: now}
}

// Summary returns the dashboard of a month (YYYY-MM); an empty month means the
// current month in Asia/Jakarta.
func (s *DashboardService) Summary(ctx context.Context, userID, month string) (DashboardSummary, error) {
	start := s.currentMonth()
	if month != "" {
		var err error
		if start, err = parseMonth(month); err != nil {
			return DashboardSummary{}, err
		}
	}
	end := start.AddDate(0, 1, 0)

	balance, err := s.dashboard.TotalActiveBalance(ctx, userID)
	if err != nil {
		return DashboardSummary{}, fmt.Errorf("total balance: %w", err)
	}
	totals, err := s.dashboard.MonthlyTotals(ctx, userID, start, end)
	if err != nil {
		return DashboardSummary{}, fmt.Errorf("month totals: %w", err)
	}
	top, err := s.dashboard.TopExpenseCategories(ctx, userID, start, end, topCategoryCount)
	if err != nil {
		return DashboardSummary{}, fmt.Errorf("top categories: %w", err)
	}
	budgets, err := s.budgets.List(ctx, userID, start)
	if err != nil {
		return DashboardSummary{}, fmt.Errorf("budgets: %w", err)
	}
	recent, _, err := s.transactions.List(ctx, domain.TransactionFilter{UserID: userID, Limit: recentCount})
	if err != nil {
		return DashboardSummary{}, fmt.Errorf("recent transactions: %w", err)
	}

	summary := DashboardSummary{
		Month:              start,
		TotalBalance:       balance,
		TopCategories:      top,
		BudgetAlerts:       []BudgetSummary{},
		RecentTransactions: recent,
	}
	if len(totals) > 0 {
		summary.Income, summary.Expense = totals[0].Income, totals[0].Expense
	}
	summary.UsedPercentage = usedPercentage(summary.Income, summary.Expense)
	for _, b := range budgets {
		if bs := summarize(b); bs.Status != domain.BudgetSafe {
			summary.BudgetAlerts = append(summary.BudgetAlerts, bs)
		}
	}
	return summary, nil
}

// Trend returns income and expense for the last months (default 6) up to and
// including the current month, oldest first. Months without transactions are 0.
func (s *DashboardService) Trend(ctx context.Context, userID string, months int) ([]domain.MonthTotal, error) {
	if months == 0 {
		months = defaultTrendMonths
	}
	if months < 1 || months > maxTrendMonths {
		return nil, &domain.ValidationError{Fields: map[string]string{
			"months": fmt.Sprintf("must be 1-%d", maxTrendMonths),
		}}
	}

	end := s.currentMonth().AddDate(0, 1, 0)
	start := end.AddDate(0, -months, 0)

	totals, err := s.dashboard.MonthlyTotals(ctx, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("monthly totals: %w", err)
	}

	byMonth := make(map[string]domain.MonthTotal, len(totals))
	for _, t := range totals {
		byMonth[t.Month.Format(monthLayout)] = t
	}
	trend := make([]domain.MonthTotal, 0, months)
	for m := start; m.Before(end); m = m.AddDate(0, 1, 0) {
		t := byMonth[m.Format(monthLayout)]
		t.Month = m
		trend = append(trend, t)
	}
	return trend, nil
}

// currentMonth is the first day of the current month in Asia/Jakarta, as a UTC
// date so it compares equal to the dates read from the database.
func (s *DashboardService) currentMonth() time.Time {
	now := s.now().In(jakarta)
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func usedPercentage(income, expense int64) *int64 {
	if income == 0 {
		return nil
	}
	p := expense * 100 / income
	return &p
}
