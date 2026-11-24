package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

// fakeAccountRepository is an in-memory domain.AccountRepository.
type fakeAccountRepository struct {
	accounts map[string]domain.Account // keyed by ID
	inUse    map[string]bool           // account IDs that have transactions
	nextID   int
}

func newFakeAccountRepository() *fakeAccountRepository {
	return &fakeAccountRepository{accounts: map[string]domain.Account{}, inUse: map[string]bool{}}
}

func (r *fakeAccountRepository) List(_ context.Context, userID string, includeArchived bool) ([]domain.Account, error) {
	var out []domain.Account
	for _, a := range r.accounts {
		if a.UserID == userID && (includeArchived || !a.IsArchived) {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *fakeAccountRepository) Get(_ context.Context, userID, id string) (domain.Account, error) {
	a, ok := r.accounts[id]
	if !ok || a.UserID != userID {
		return domain.Account{}, domain.ErrNotFound
	}
	return a, nil
}

func (r *fakeAccountRepository) Create(_ context.Context, a domain.Account) (domain.Account, error) {
	r.nextID++
	a.ID = "account-" + string(rune('0'+r.nextID))
	r.accounts[a.ID] = a
	return a, nil
}

func (r *fakeAccountRepository) Update(_ context.Context, a domain.Account) (domain.Account, error) {
	existing, ok := r.accounts[a.ID]
	if !ok || existing.UserID != a.UserID {
		return domain.Account{}, domain.ErrNotFound
	}
	a.Balance = existing.Balance // the real query never writes the balance
	r.accounts[a.ID] = a
	return a, nil
}

func (r *fakeAccountRepository) Delete(_ context.Context, userID, id string) error {
	a, ok := r.accounts[id]
	if !ok || a.UserID != userID {
		return domain.ErrNotFound
	}
	if r.inUse[id] {
		return domain.ErrAccountInUse
	}
	delete(r.accounts, id)
	return nil
}

func ptr[T any](v T) *T { return &v }

func createTestAccount(t *testing.T, svc *AccountService, userID string) domain.Account {
	t.Helper()
	account, err := svc.Create(context.Background(), userID, CreateAccountInput{
		Name:            "Tabungan",
		Type:            domain.AccountBank,
		InstitutionName: "BSI",
		AccountNumber:   "1234567890",
		Balance:         1_500_000,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	return account
}

func TestAccountCreate_TrimsInputAndStoresInitialBalance(t *testing.T) {
	svc := NewAccountService(newFakeAccountRepository())

	account, err := svc.Create(context.Background(), "user-1", CreateAccountInput{
		Name:            "  Dompet  ",
		Type:            domain.AccountCash,
		InstitutionName: "   ",
		Balance:         250_000,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if account.Name != "Dompet" {
		t.Errorf("name = %q, want %q", account.Name, "Dompet")
	}
	if account.InstitutionName != nil || account.AccountNumber != nil {
		t.Errorf("blank optional fields should be nil, got %v / %v", account.InstitutionName, account.AccountNumber)
	}
	if account.Balance != 250_000 || account.UserID != "user-1" {
		t.Errorf("balance/user = %d/%q, want 250000/user-1", account.Balance, account.UserID)
	}
}

func TestAccountCreate_RejectsInvalidInput(t *testing.T) {
	svc := NewAccountService(newFakeAccountRepository())

	_, err := svc.Create(context.Background(), "user-1", CreateAccountInput{
		Name:          " ",
		Type:          "kripto",
		AccountNumber: strings.Repeat("9", maxAccountTextLength+1),
		Balance:       -1,
	})

	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Create() error = %v, want *domain.ValidationError", err)
	}
	for _, field := range []string{"name", "type", "accountNumber", "balance"} {
		if _, ok := validationErr.Fields[field]; !ok {
			t.Errorf("missing validation error for %q", field)
		}
	}
}

func TestAccountUpdate_ChangesOnlyGivenFields(t *testing.T) {
	svc := NewAccountService(newFakeAccountRepository())
	created := createTestAccount(t, svc, "user-1")

	updated, err := svc.Update(context.Background(), "user-1", created.ID, UpdateAccountInput{
		Name:          ptr("Tabungan Utama"),
		AccountNumber: ptr(""),
		IsArchived:    ptr(true),
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Name != "Tabungan Utama" || !updated.IsArchived {
		t.Errorf("name/archived = %q/%v, want %q/true", updated.Name, updated.IsArchived, "Tabungan Utama")
	}
	if updated.AccountNumber != nil {
		t.Errorf("account number = %v, want nil after clearing", *updated.AccountNumber)
	}
	if updated.InstitutionName == nil || *updated.InstitutionName != "BSI" {
		t.Errorf("institution changed although it was not given: %v", updated.InstitutionName)
	}
	if updated.Type != domain.AccountBank || updated.Balance != 1_500_000 {
		t.Errorf("type/balance = %q/%d, want unchanged bank/1500000", updated.Type, updated.Balance)
	}
}

func TestAccountUpdate_RejectsInvalidType(t *testing.T) {
	svc := NewAccountService(newFakeAccountRepository())
	created := createTestAccount(t, svc, "user-1")

	invalid := domain.AccountType("kripto")
	_, err := svc.Update(context.Background(), "user-1", created.ID, UpdateAccountInput{Type: &invalid})

	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) || validationErr.Fields["type"] == "" {
		t.Fatalf("Update() error = %v, want type validation error", err)
	}
}

func TestAccount_OtherUsersAccountIsNotFound(t *testing.T) {
	svc := NewAccountService(newFakeAccountRepository())
	created := createTestAccount(t, svc, "user-1")

	if _, err := svc.Update(context.Background(), "user-2", created.ID, UpdateAccountInput{Name: ptr("Hack")}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Update() by another user error = %v, want ErrNotFound", err)
	}
	if err := svc.Delete(context.Background(), "user-2", created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete() by another user error = %v, want ErrNotFound", err)
	}
}

func TestAccountDelete_RejectsAccountWithTransactions(t *testing.T) {
	repo := newFakeAccountRepository()
	svc := NewAccountService(repo)
	created := createTestAccount(t, svc, "user-1")
	repo.inUse[created.ID] = true

	if err := svc.Delete(context.Background(), "user-1", created.ID); !errors.Is(err, domain.ErrAccountInUse) {
		t.Fatalf("Delete() error = %v, want ErrAccountInUse", err)
	}
}

func TestAccountList_HidesArchivedUnlessRequested(t *testing.T) {
	svc := NewAccountService(newFakeAccountRepository())
	active := createTestAccount(t, svc, "user-1")
	archived := createTestAccount(t, svc, "user-1")
	if _, err := svc.Update(context.Background(), "user-1", archived.ID, UpdateAccountInput{IsArchived: ptr(true)}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	visible, err := svc.List(context.Background(), "user-1", false)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(visible) != 1 || visible[0].ID != active.ID {
		t.Errorf("List(includeArchived=false) = %v, want only the active account", visible)
	}

	all, err := svc.List(context.Background(), "user-1", true)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(all) != 2 {
		t.Errorf("List(includeArchived=true) returned %d accounts, want 2", len(all))
	}
}
