package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/repository/postgres/db"
)

type AccountRepository struct {
	queries *db.Queries
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{queries: db.New(pool)}
}

func (r *AccountRepository) List(ctx context.Context, userID string, includeArchived bool) ([]domain.Account, error) {
	rows, err := r.queries.ListAccounts(ctx, db.ListAccountsParams{
		UserID:          userID,
		IncludeArchived: includeArchived,
	})
	if err != nil {
		return nil, fmt.Errorf("select accounts: %w", err)
	}

	accounts := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, toDomainAccount(row))
	}
	return accounts, nil
}

func (r *AccountRepository) Get(ctx context.Context, userID, id string) (domain.Account, error) {
	row, err := r.queries.GetAccount(ctx, db.GetAccountParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) || hasCode(err, codeInvalidText) {
		return domain.Account{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("select account: %w", err)
	}
	return toDomainAccount(row), nil
}

func (r *AccountRepository) Create(ctx context.Context, a domain.Account) (domain.Account, error) {
	row, err := r.queries.CreateAccount(ctx, db.CreateAccountParams{
		UserID:          a.UserID,
		Name:            a.Name,
		Type:            db.AccountType(a.Type),
		InstitutionName: a.InstitutionName,
		AccountNumber:   a.AccountNumber,
		Balance:         a.Balance,
	})
	if err != nil {
		return domain.Account{}, fmt.Errorf("insert account: %w", err)
	}
	return toDomainAccount(row), nil
}

func (r *AccountRepository) Update(ctx context.Context, a domain.Account) (domain.Account, error) {
	row, err := r.queries.UpdateAccount(ctx, db.UpdateAccountParams{
		ID:              a.ID,
		UserID:          a.UserID,
		Name:            a.Name,
		Type:            db.AccountType(a.Type),
		InstitutionName: a.InstitutionName,
		AccountNumber:   a.AccountNumber,
		IsArchived:      a.IsArchived,
	})
	if errors.Is(err, pgx.ErrNoRows) || hasCode(err, codeInvalidText) {
		return domain.Account{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("update account: %w", err)
	}
	return toDomainAccount(row), nil
}

func (r *AccountRepository) Delete(ctx context.Context, userID, id string) error {
	deleted, err := r.queries.DeleteAccount(ctx, db.DeleteAccountParams{ID: id, UserID: userID})
	switch {
	case hasCode(err, codeForeignKeyViolation):
		return domain.ErrAccountInUse
	case hasCode(err, codeInvalidText):
		return domain.ErrNotFound
	case err != nil:
		return fmt.Errorf("delete account: %w", err)
	case deleted == 0:
		return domain.ErrNotFound
	}
	return nil
}

func toDomainAccount(row db.Account) domain.Account {
	return domain.Account{
		ID:              row.ID,
		UserID:          row.UserID,
		Name:            row.Name,
		Type:            domain.AccountType(row.Type),
		InstitutionName: row.InstitutionName,
		AccountNumber:   row.AccountNumber,
		Balance:         row.Balance,
		IsArchived:      row.IsArchived,
		CreatedAt:       row.CreatedAt,
	}
}
