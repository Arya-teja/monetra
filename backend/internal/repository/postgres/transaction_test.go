package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/repository/postgres"
	"github.com/Arya-teja/monetra/backend/internal/testdb"
)

// Two requests read the same transaction, then both try to save it. Only the first
// may win; the second must get ErrStaleData and leave balances untouched, otherwise
// the balance correction would be applied twice.
func TestTransactionUpdateRejectsStaleData(t *testing.T) {
	pool := testdb.Open(t)
	ctx := context.Background()

	users := postgres.NewUserRepository(pool)
	user, err := users.CreateWithCategories(ctx,
		domain.User{Name: "Lock", Email: "lock@example.com", PasswordHash: "x"},
		[]domain.Category{{Name: "Makan", Type: domain.CashflowExpense}})
	if err != nil {
		t.Fatal(err)
	}
	var categoryID string
	if err := pool.QueryRow(ctx, "SELECT id FROM categories WHERE user_id = $1", user.ID).Scan(&categoryID); err != nil {
		t.Fatal(err)
	}
	account, err := postgres.NewAccountRepository(pool).Create(ctx, domain.Account{UserID: user.ID, Name: "BSI", Type: domain.AccountBank, Balance: 1_000})
	if err != nil {
		t.Fatal(err)
	}

	repo := postgres.NewTransactionRepository(pool)
	tx, err := repo.Create(ctx, domain.Transaction{
		UserID: user.ID, AccountID: account.ID, CategoryID: categoryID,
		Type: domain.CashflowExpense, Amount: 100, OccurredAt: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	}, []domain.BalanceChange{{AccountID: account.ID, Delta: -100}})
	if err != nil {
		t.Fatal(err)
	}
	readByBoth := tx.UpdatedAt

	first := tx
	first.Amount = 200
	if _, err := repo.Update(ctx, first, readByBoth, []domain.BalanceChange{{AccountID: account.ID, Delta: -100}}); err != nil {
		t.Fatalf("first update: %v", err)
	}

	second := tx
	second.Amount = 300
	_, err = repo.Update(ctx, second, readByBoth, []domain.BalanceChange{{AccountID: account.ID, Delta: -200}})
	if !errors.Is(err, domain.ErrStaleData) {
		t.Fatalf("second update error = %v, want ErrStaleData", err)
	}
	if err := repo.Delete(ctx, tx, []domain.BalanceChange{{AccountID: account.ID, Delta: 100}}); !errors.Is(err, domain.ErrStaleData) {
		t.Fatalf("delete with stale data error = %v, want ErrStaleData", err)
	}

	var balance, amount int64
	if err := pool.QueryRow(ctx, "SELECT a.balance, t.amount FROM accounts a JOIN transactions t ON t.account_id = a.id WHERE a.id = $1", account.ID).Scan(&balance, &amount); err != nil {
		t.Fatal(err)
	}
	if balance != 800 || amount != 200 {
		t.Errorf("balance/amount = %d/%d, want 800/200 (only the first update applied)", balance, amount)
	}
}
