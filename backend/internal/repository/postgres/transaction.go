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

type TransactionRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool, queries: db.New(pool)}
}

func (r *TransactionRepository) List(ctx context.Context, f domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	var typ db.NullCashflowType
	if f.Type != nil {
		typ = db.NullCashflowType{CashflowType: db.CashflowType(*f.Type), Valid: true}
	}

	rows, err := r.queries.ListTransactions(ctx, db.ListTransactionsParams{
		UserID:     f.UserID,
		FromDate:   f.From,
		ToDate:     f.To,
		AccountID:  f.AccountID,
		CategoryID: f.CategoryID,
		Type:       typ,
		RowLimit:   int32(f.Limit),
		RowOffset:  int32(f.Offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("select transactions: %w", err)
	}

	total, err := r.queries.CountTransactions(ctx, db.CountTransactionsParams{
		UserID:     f.UserID,
		FromDate:   f.From,
		ToDate:     f.To,
		AccountID:  f.AccountID,
		CategoryID: f.CategoryID,
		Type:       typ,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	transactions := make([]domain.Transaction, 0, len(rows))
	for _, row := range rows {
		transactions = append(transactions, toDomainTransaction(db.GetTransactionRow(row)))
	}
	return transactions, total, nil
}

func (r *TransactionRepository) Get(ctx context.Context, userID, id string) (domain.Transaction, error) {
	return getTransaction(ctx, r.queries, userID, id)
}

func (r *TransactionRepository) Create(ctx context.Context, t domain.Transaction, changes []domain.BalanceChange) (domain.Transaction, error) {
	var created domain.Transaction
	err := r.inTx(ctx, func(q *db.Queries) error {
		id, err := q.CreateTransaction(ctx, db.CreateTransactionParams{
			UserID:     t.UserID,
			AccountID:  t.AccountID,
			CategoryID: t.CategoryID,
			Type:       db.CashflowType(t.Type),
			Amount:     t.Amount,
			OccurredAt: t.OccurredAt,
			Note:       t.Note,
		})
		if err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}
		if err := applyBalanceChanges(ctx, q, t.UserID, changes); err != nil {
			return err
		}
		created, err = getTransaction(ctx, q, t.UserID, id)
		return err
	})
	return created, err
}

func (r *TransactionRepository) Update(ctx context.Context, t domain.Transaction, expectedUpdatedAt time.Time, changes []domain.BalanceChange) (domain.Transaction, error) {
	var updated domain.Transaction
	err := r.inTx(ctx, func(q *db.Queries) error {
		affected, err := q.UpdateTransaction(ctx, db.UpdateTransactionParams{
			ID:                t.ID,
			UserID:            t.UserID,
			AccountID:         t.AccountID,
			CategoryID:        t.CategoryID,
			Type:              db.CashflowType(t.Type),
			Amount:            t.Amount,
			OccurredAt:        t.OccurredAt,
			Note:              t.Note,
			ExpectedUpdatedAt: expectedUpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("update transaction: %w", err)
		}
		if affected == 0 {
			return domain.ErrStaleData
		}
		if err := applyBalanceChanges(ctx, q, t.UserID, changes); err != nil {
			return err
		}
		updated, err = getTransaction(ctx, q, t.UserID, t.ID)
		return err
	})
	return updated, err
}

func (r *TransactionRepository) Delete(ctx context.Context, t domain.Transaction, changes []domain.BalanceChange) error {
	return r.inTx(ctx, func(q *db.Queries) error {
		affected, err := q.DeleteTransaction(ctx, db.DeleteTransactionParams{
			ID:                t.ID,
			UserID:            t.UserID,
			ExpectedUpdatedAt: t.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("delete transaction: %w", err)
		}
		if affected == 0 {
			return domain.ErrStaleData
		}
		return applyBalanceChanges(ctx, q, t.UserID, changes)
	})
}

// inTx runs fn in one database transaction; any error rolls everything back.
func (r *TransactionRepository) inTx(ctx context.Context, fn func(q *db.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// Rollback is a no-op once the transaction is committed.
	defer tx.Rollback(ctx)

	if err := fn(r.queries.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func applyBalanceChanges(ctx context.Context, q *db.Queries, userID string, changes []domain.BalanceChange) error {
	for _, c := range changes {
		affected, err := q.AdjustAccountBalance(ctx, db.AdjustAccountBalanceParams{
			ID:     c.AccountID,
			UserID: userID,
			Delta:  c.Delta,
		})
		if err != nil {
			return fmt.Errorf("adjust balance of account %s: %w", c.AccountID, err)
		}
		if affected == 0 {
			return fmt.Errorf("adjust balance of account %s: %w", c.AccountID, domain.ErrNotFound)
		}
	}
	return nil
}

func getTransaction(ctx context.Context, q *db.Queries, userID, id string) (domain.Transaction, error) {
	row, err := q.GetTransaction(ctx, db.GetTransactionParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) || hasCode(err, codeInvalidText) {
		return domain.Transaction{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("select transaction: %w", err)
	}
	return toDomainTransaction(row), nil
}

func toDomainTransaction(row db.GetTransactionRow) domain.Transaction {
	return domain.Transaction{
		ID:           row.ID,
		UserID:       row.UserID,
		AccountID:    row.AccountID,
		CategoryID:   row.CategoryID,
		Type:         domain.CashflowType(row.Type),
		Amount:       row.Amount,
		OccurredAt:   row.OccurredAt,
		Note:         row.Note,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		AccountName:  row.AccountName,
		CategoryName: row.CategoryName,
	}
}
