package domain

import (
	"context"
	"time"
)

type BudgetStatus string

const (
	BudgetSafe    BudgetStatus = "safe"    // below 80% spent
	BudgetWarning BudgetStatus = "warning" // 80-99% spent
	BudgetOver    BudgetStatus = "over"    // 100% or more spent
)

// Budget limits the expenses of one expense category in one calendar month.
type Budget struct {
	ID           string
	UserID       string
	CategoryID   string
	CategoryName string    // filled when reading
	Month        time.Time // first day of the month
	Amount       int64
	Spent        int64 // filled when reading: expenses of the category in Month
}

// BudgetRepository only touches budgets owned by the given user;
// a budget of another user behaves as if it does not exist (ErrNotFound).
type BudgetRepository interface {
	List(ctx context.Context, userID string, month time.Time) ([]Budget, error)
	Get(ctx context.Context, userID, id string) (Budget, error)
	// Create returns ErrBudgetExists when the category already has a budget that month.
	Create(ctx context.Context, budget Budget) (Budget, error)
	UpdateAmount(ctx context.Context, userID, id string, amount int64) (Budget, error)
	Delete(ctx context.Context, userID, id string) error
	// CopyMonth copies the budgets of from into to, skipping categories that already
	// have a budget in to, and returns how many budgets were created.
	CopyMonth(ctx context.Context, userID string, from, to time.Time) (int64, error)
}
