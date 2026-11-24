package domain

import (
	"context"
	"time"
)

// MonthTotal is the income and expense of one calendar month.
type MonthTotal struct {
	Month   time.Time // first day of the month
	Income  int64
	Expense int64
}

// CategoryTotal is the expense total of one category.
type CategoryTotal struct {
	CategoryID   string
	CategoryName string
	Total        int64
}

// DashboardRepository reads aggregates of one user's data; [from, to) are dates.
type DashboardRepository interface {
	// TotalActiveBalance sums the balances of all non-archived accounts.
	TotalActiveBalance(ctx context.Context, userID string) (int64, error)
	// MonthlyTotals returns only months that have transactions, oldest first.
	MonthlyTotals(ctx context.Context, userID string, from, to time.Time) ([]MonthTotal, error)
	// TopExpenseCategories returns the categories with the highest expenses, highest first.
	TopExpenseCategories(ctx context.Context, userID string, from, to time.Time, limit int) ([]CategoryTotal, error)
}
