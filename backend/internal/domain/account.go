package domain

import (
	"context"
	"time"
)

type AccountType string

const (
	AccountBank    AccountType = "bank"
	AccountEwallet AccountType = "ewallet"
	AccountCash    AccountType = "cash"
)

func (t AccountType) Valid() bool {
	switch t {
	case AccountBank, AccountEwallet, AccountCash:
		return true
	}
	return false
}

type Account struct {
	ID              string
	UserID          string
	Name            string
	Type            AccountType
	InstitutionName *string
	AccountNumber   *string
	Balance         int64
	IsArchived      bool
	CreatedAt       time.Time
}

// AccountRepository only touches accounts owned by the given user;
// an account of another user behaves as if it does not exist (ErrNotFound).
type AccountRepository interface {
	List(ctx context.Context, userID string, includeArchived bool) ([]Account, error)
	Get(ctx context.Context, userID, id string) (Account, error)
	Create(ctx context.Context, account Account) (Account, error)
	// Update saves name, type, institution, account number, and archive flag; never the balance.
	Update(ctx context.Context, account Account) (Account, error)
	// Delete returns ErrAccountInUse when the account still has transactions.
	Delete(ctx context.Context, userID, id string) error
}
