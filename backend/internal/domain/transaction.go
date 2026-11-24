package domain

import (
	"context"
	"time"
)

type Transaction struct {
	ID         string
	UserID     string
	AccountID  string
	CategoryID string
	Type       CashflowType
	Amount     int64     // always positive; Type decides the direction
	OccurredAt time.Time // calendar date (UTC midnight)
	Note       *string
	CreatedAt  time.Time
	UpdatedAt  time.Time

	// Filled when reading, for display only.
	AccountName  string
	CategoryName string
}

// BalanceChange adds Delta (may be negative) to the balance of one account.
type BalanceChange struct {
	AccountID string
	Delta     int64
}

// TransactionFilter selects transactions of one user. Nil fields are not filtered on.
type TransactionFilter struct {
	UserID     string
	From       *time.Time // inclusive
	To         *time.Time // exclusive
	AccountID  *string
	CategoryID *string
	Type       *CashflowType
	Limit      int
	Offset     int
}

// TransactionRepository stores a transaction change and its balance changes in one
// database transaction: either everything is saved or nothing is.
type TransactionRepository interface {
	// List returns one page of matching transactions (newest first) and the total match count.
	List(ctx context.Context, filter TransactionFilter) ([]Transaction, int64, error)
	Get(ctx context.Context, userID, id string) (Transaction, error)
	Create(ctx context.Context, t Transaction, changes []BalanceChange) (Transaction, error)
	// Update saves t only while the stored row still has the updated_at the caller read
	// (expectedUpdatedAt); otherwise it returns ErrStaleData and changes nothing.
	Update(ctx context.Context, t Transaction, expectedUpdatedAt time.Time, changes []BalanceChange) (Transaction, error)
	// Delete removes t under the same rule as Update.
	Delete(ctx context.Context, t Transaction, changes []BalanceChange) error
}
