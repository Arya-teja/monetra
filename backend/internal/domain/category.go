package domain

import (
	"context"
	"time"
)

type CashflowType string

const (
	CashflowIncome  CashflowType = "income"
	CashflowExpense CashflowType = "expense"
)

func (t CashflowType) Valid() bool {
	return t == CashflowIncome || t == CashflowExpense
}

type Category struct {
	ID        string
	UserID    string
	Name      string
	Type      CashflowType
	IsDefault bool
	CreatedAt time.Time
}

// CategoryRepository only touches categories owned by the given user;
// a category of another user behaves as if it does not exist (ErrNotFound).
type CategoryRepository interface {
	// List returns all categories, or only those of one type when typ is not nil.
	List(ctx context.Context, userID string, typ *CashflowType) ([]Category, error)
	Get(ctx context.Context, userID, id string) (Category, error)
	// Create returns ErrCategoryExists when the user already has the name for that type.
	Create(ctx context.Context, category Category) (Category, error)
	// Rename returns ErrCategoryExists when the new name is already taken for that type.
	Rename(ctx context.Context, userID, id, name string) (Category, error)
	// Delete returns ErrCategoryInUse when transactions still use the category.
	Delete(ctx context.Context, userID, id string) error
}
