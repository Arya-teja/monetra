package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

// fakeTransactionRepository is an in-memory domain.TransactionRepository that applies
// balance changes to the accounts of a fakeAccountRepository, like the real one does.
type fakeTransactionRepository struct {
	transactions map[string]domain.Transaction
	accounts     *fakeAccountRepository
	lastFilter   domain.TransactionFilter
	nextID       int
}

func (r *fakeTransactionRepository) List(_ context.Context, f domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	r.lastFilter = f
	return nil, 0, nil
}

func (r *fakeTransactionRepository) Get(_ context.Context, userID, id string) (domain.Transaction, error) {
	t, ok := r.transactions[id]
	if !ok || t.UserID != userID {
		return domain.Transaction{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *fakeTransactionRepository) Create(_ context.Context, t domain.Transaction, changes []domain.BalanceChange) (domain.Transaction, error) {
	r.nextID++
	t.ID = testUUID('t', r.nextID)
	t.UpdatedAt = time.Unix(int64(r.nextID), 0)
	r.transactions[t.ID] = t
	r.apply(changes)
	return t, nil
}

func (r *fakeTransactionRepository) Update(_ context.Context, t domain.Transaction, expectedUpdatedAt time.Time, changes []domain.BalanceChange) (domain.Transaction, error) {
	if !r.transactions[t.ID].UpdatedAt.Equal(expectedUpdatedAt) {
		return domain.Transaction{}, domain.ErrStaleData
	}
	r.nextID++
	t.UpdatedAt = time.Unix(int64(r.nextID), 0)
	r.transactions[t.ID] = t
	r.apply(changes)
	return t, nil
}

func (r *fakeTransactionRepository) Delete(_ context.Context, t domain.Transaction, changes []domain.BalanceChange) error {
	delete(r.transactions, t.ID)
	r.apply(changes)
	return nil
}

func (r *fakeTransactionRepository) apply(changes []domain.BalanceChange) {
	for _, c := range changes {
		a := r.accounts.accounts[c.AccountID]
		a.Balance += c.Delta
		r.accounts.accounts[c.AccountID] = a
	}
}

type transactionFixture struct {
	svc        *TransactionService
	accounts   *fakeAccountRepository
	wallet     string // active account, balance 1_000_000
	cash       string // active account, balance 100_000
	archived   string // archived account, balance 0
	salary     string // income category
	food       string // expense category
	otherUsers string // expense category of another user
}

func newTransactionFixture(t *testing.T) transactionFixture {
	t.Helper()
	ctx := context.Background()
	accounts := newFakeAccountRepository()
	categories := newFakeCategoryRepository()

	addAccount := func(n int, balance int64, archived bool) string {
		id := testUUID('a', n)
		accounts.accounts[id] = domain.Account{ID: id, UserID: "user-1", Name: id, Type: domain.AccountBank, Balance: balance, IsArchived: archived}
		return id
	}
	addCategory := func(userID, name string, typ domain.CashflowType) string {
		c, err := categories.Create(ctx, domain.Category{UserID: userID, Name: name, Type: typ})
		if err != nil {
			t.Fatalf("create category: %v", err)
		}
		return c.ID
	}

	f := transactionFixture{
		accounts:   accounts,
		wallet:     addAccount(1, 1_000_000, false),
		cash:       addAccount(2, 100_000, false),
		archived:   addAccount(3, 0, true),
		salary:     addCategory("user-1", "Gaji", domain.CashflowIncome),
		food:       addCategory("user-1", "Makan", domain.CashflowExpense),
		otherUsers: addCategory("user-2", "Makan", domain.CashflowExpense),
	}
	transactions := &fakeTransactionRepository{transactions: map[string]domain.Transaction{}, accounts: accounts}
	f.svc = NewTransactionService(transactions, accounts, categories)
	return f
}

func (f transactionFixture) balance(accountID string) int64 {
	return f.accounts.accounts[accountID].Balance
}

func (f transactionFixture) createExpense(t *testing.T, accountID string, amount int64) domain.Transaction {
	t.Helper()
	tx, err := f.svc.Create(context.Background(), "user-1", CreateTransactionInput{
		AccountID: accountID, CategoryID: f.food, Type: domain.CashflowExpense,
		Amount: amount, OccurredAt: "2026-11-24",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	return tx
}

func TestTransactionCreate_AdjustsBalanceByType(t *testing.T) {
	f := newTransactionFixture(t)

	f.createExpense(t, f.wallet, 50_000)
	if got := f.balance(f.wallet); got != 950_000 {
		t.Errorf("after expense balance = %d, want 950000", got)
	}

	_, err := f.svc.Create(context.Background(), "user-1", CreateTransactionInput{
		AccountID: f.wallet, CategoryID: f.salary, Type: domain.CashflowIncome,
		Amount: 2_000_000, OccurredAt: "2026-11-25", Note: "  gaji November  ",
	})
	if err != nil {
		t.Fatalf("Create() income error = %v", err)
	}
	if got := f.balance(f.wallet); got != 2_950_000 {
		t.Errorf("after income balance = %d, want 2950000", got)
	}
}

func TestTransactionCreate_AllowsNegativeBalance(t *testing.T) {
	f := newTransactionFixture(t)

	f.createExpense(t, f.cash, 150_000)
	if got := f.balance(f.cash); got != -50_000 {
		t.Errorf("balance = %d, want -50000", got)
	}
}

func TestTransactionCreate_RejectsInvalidInput(t *testing.T) {
	f := newTransactionFixture(t)

	tests := []struct {
		name  string
		in    CreateTransactionInput
		field string
	}{
		{"zero amount", CreateTransactionInput{AccountID: f.wallet, CategoryID: f.food, Type: domain.CashflowExpense, Amount: 0, OccurredAt: "2026-11-24"}, "amount"},
		{"bad date", CreateTransactionInput{AccountID: f.wallet, CategoryID: f.food, Type: domain.CashflowExpense, Amount: 1, OccurredAt: "24-11-2026"}, "occurredAt"},
		{"category type mismatch", CreateTransactionInput{AccountID: f.wallet, CategoryID: f.salary, Type: domain.CashflowExpense, Amount: 1, OccurredAt: "2026-11-24"}, "categoryId"},
		{"category of another user", CreateTransactionInput{AccountID: f.wallet, CategoryID: f.otherUsers, Type: domain.CashflowExpense, Amount: 1, OccurredAt: "2026-11-24"}, "categoryId"},
		{"archived account", CreateTransactionInput{AccountID: f.archived, CategoryID: f.food, Type: domain.CashflowExpense, Amount: 1, OccurredAt: "2026-11-24"}, "accountId"},
		{"unknown account", CreateTransactionInput{AccountID: testUUID('a', 9), CategoryID: f.food, Type: domain.CashflowExpense, Amount: 1, OccurredAt: "2026-11-24"}, "accountId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.svc.Create(context.Background(), "user-1", tt.in)
			var validationErr *domain.ValidationError
			if !errors.As(err, &validationErr) || validationErr.Fields[tt.field] == "" {
				t.Fatalf("Create() error = %v, want validation error on %q", err, tt.field)
			}
		})
	}

	if got := f.balance(f.wallet); got != 1_000_000 {
		t.Errorf("rejected transactions changed the balance to %d", got)
	}
}

func TestTransactionUpdate_CorrectsBalanceOnSameAccount(t *testing.T) {
	f := newTransactionFixture(t)
	tx := f.createExpense(t, f.wallet, 50_000)

	amount := int64(80_000)
	if _, err := f.svc.Update(context.Background(), "user-1", tx.ID, UpdateTransactionInput{Amount: &amount}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if got := f.balance(f.wallet); got != 920_000 {
		t.Errorf("balance = %d, want 920000", got)
	}
}

func TestTransactionUpdate_MovesAmountToAnotherAccount(t *testing.T) {
	f := newTransactionFixture(t)
	tx := f.createExpense(t, f.wallet, 50_000)

	if _, err := f.svc.Update(context.Background(), "user-1", tx.ID, UpdateTransactionInput{AccountID: &f.cash}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if got := f.balance(f.wallet); got != 1_000_000 {
		t.Errorf("old account balance = %d, want 1000000 (restored)", got)
	}
	if got := f.balance(f.cash); got != 50_000 {
		t.Errorf("new account balance = %d, want 50000", got)
	}
}

func TestTransactionUpdate_SwitchingTypeReversesDirection(t *testing.T) {
	f := newTransactionFixture(t)
	tx := f.createExpense(t, f.wallet, 50_000)

	income := domain.CashflowIncome
	if _, err := f.svc.Update(context.Background(), "user-1", tx.ID, UpdateTransactionInput{Type: &income, CategoryID: &f.salary}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if got := f.balance(f.wallet); got != 1_050_000 {
		t.Errorf("balance = %d, want 1050000", got)
	}
}

func TestTransactionUpdate_ArchivedAccountRules(t *testing.T) {
	f := newTransactionFixture(t)
	tx := f.createExpense(t, f.wallet, 50_000)

	_, err := f.svc.Update(context.Background(), "user-1", tx.ID, UpdateTransactionInput{AccountID: &f.archived})
	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) || validationErr.Fields["accountId"] == "" {
		t.Fatalf("moving to archived account: error = %v, want accountId validation error", err)
	}

	// A transaction already on an account that gets archived stays editable.
	a := f.accounts.accounts[f.wallet]
	a.IsArchived = true
	f.accounts.accounts[f.wallet] = a
	note := "tetap bisa diedit"
	if _, err := f.svc.Update(context.Background(), "user-1", tx.ID, UpdateTransactionInput{Note: &note}); err != nil {
		t.Fatalf("editing transaction on archived account: error = %v", err)
	}
}

func TestTransactionDelete_RestoresBalance(t *testing.T) {
	f := newTransactionFixture(t)
	tx := f.createExpense(t, f.wallet, 50_000)

	if err := f.svc.Delete(context.Background(), "user-1", tx.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if got := f.balance(f.wallet); got != 1_000_000 {
		t.Errorf("balance = %d, want 1000000", got)
	}
	if err := f.svc.Delete(context.Background(), "user-2", tx.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete() of missing transaction error = %v, want ErrNotFound", err)
	}
}

func TestBalanceChanges(t *testing.T) {
	expense := func(account string, amount int64) *domain.Transaction {
		return &domain.Transaction{AccountID: account, Type: domain.CashflowExpense, Amount: amount}
	}

	tests := []struct {
		name          string
		before, after *domain.Transaction
		want          []domain.BalanceChange
	}{
		{"create", nil, expense("a", 100), []domain.BalanceChange{{AccountID: "a", Delta: -100}}},
		{"delete", expense("a", 100), nil, []domain.BalanceChange{{AccountID: "a", Delta: 100}}},
		{"same account", expense("a", 100), expense("a", 300), []domain.BalanceChange{{AccountID: "a", Delta: -200}}},
		{"no balance effect", expense("a", 100), expense("a", 100), nil},
		{"other account", expense("a", 100), expense("b", 100), []domain.BalanceChange{{AccountID: "a", Delta: 100}, {AccountID: "b", Delta: -100}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := balanceChanges(tt.before, tt.after)
			if len(got) != len(tt.want) {
				t.Fatalf("balanceChanges() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("balanceChanges()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTransactionList_BuildsFilter(t *testing.T) {
	f := newTransactionFixture(t)
	repo := f.svc.transactions.(*fakeTransactionRepository)

	result, err := f.svc.List(context.Background(), "user-1", ListTransactionsInput{
		Month: "2026-12", Type: "expense", AccountID: f.wallet, Page: 3, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	got := repo.lastFilter
	if got.From.Format(dateLayout) != "2026-12-01" || got.To.Format(dateLayout) != "2027-01-01" {
		t.Errorf("month range = %v..%v, want 2026-12-01..2027-01-01", got.From, got.To)
	}
	if got.Limit != 10 || got.Offset != 20 || result.Page != 3 {
		t.Errorf("limit/offset/page = %d/%d/%d, want 10/20/3", got.Limit, got.Offset, result.Page)
	}
	if *got.Type != domain.CashflowExpense || *got.AccountID != f.wallet || got.CategoryID != nil {
		t.Errorf("filter = %+v, want type=expense, account=wallet, no category", got)
	}
}

func TestTransactionList_RejectsInvalidFilters(t *testing.T) {
	f := newTransactionFixture(t)

	_, err := f.svc.List(context.Background(), "user-1", ListTransactionsInput{
		Month: "2026-13", AccountID: "abc", Type: "transfer", Page: -1, PageSize: 500,
	})
	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("List() error = %v, want *domain.ValidationError", err)
	}
	for _, field := range []string{"month", "accountId", "type", "page", "pageSize"} {
		if validationErr.Fields[field] == "" {
			t.Errorf("missing validation error for %q", field)
		}
	}
}
