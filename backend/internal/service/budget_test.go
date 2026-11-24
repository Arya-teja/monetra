package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

// fakeBudgetRepository is an in-memory domain.BudgetRepository.
type fakeBudgetRepository struct {
	budgets  map[string]domain.Budget
	copyFrom time.Time
	copyTo   time.Time
	nextID   int
}

func newFakeBudgetRepository() *fakeBudgetRepository {
	return &fakeBudgetRepository{budgets: map[string]domain.Budget{}}
}

func (r *fakeBudgetRepository) List(_ context.Context, userID string, month time.Time) ([]domain.Budget, error) {
	var out []domain.Budget
	for _, b := range r.budgets {
		if b.UserID == userID && b.Month.Equal(month) {
			out = append(out, b)
		}
	}
	return out, nil
}

func (r *fakeBudgetRepository) Get(_ context.Context, userID, id string) (domain.Budget, error) {
	b, ok := r.budgets[id]
	if !ok || b.UserID != userID {
		return domain.Budget{}, domain.ErrNotFound
	}
	return b, nil
}

func (r *fakeBudgetRepository) Create(_ context.Context, b domain.Budget) (domain.Budget, error) {
	for _, existing := range r.budgets {
		if existing.UserID == b.UserID && existing.CategoryID == b.CategoryID && existing.Month.Equal(b.Month) {
			return domain.Budget{}, domain.ErrBudgetExists
		}
	}
	r.nextID++
	b.ID = testUUID('b', r.nextID)
	r.budgets[b.ID] = b
	return b, nil
}

func (r *fakeBudgetRepository) UpdateAmount(ctx context.Context, userID, id string, amount int64) (domain.Budget, error) {
	b, err := r.Get(ctx, userID, id)
	if err != nil {
		return domain.Budget{}, err
	}
	b.Amount = amount
	r.budgets[id] = b
	return b, nil
}

func (r *fakeBudgetRepository) Delete(ctx context.Context, userID, id string) error {
	if _, err := r.Get(ctx, userID, id); err != nil {
		return err
	}
	delete(r.budgets, id)
	return nil
}

func (r *fakeBudgetRepository) CopyMonth(_ context.Context, _ string, from, to time.Time) (int64, error) {
	r.copyFrom, r.copyTo = from, to
	return 0, nil
}

func newBudgetFixture(t *testing.T) (*BudgetService, *fakeBudgetRepository, string, string) {
	t.Helper()
	categories := newFakeCategoryRepository()
	food, err := categories.Create(context.Background(), domain.Category{UserID: "user-1", Name: "Makan", Type: domain.CashflowExpense})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	salary, err := categories.Create(context.Background(), domain.Category{UserID: "user-1", Name: "Gaji", Type: domain.CashflowIncome})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	repo := newFakeBudgetRepository()
	return NewBudgetService(repo, categories), repo, food.ID, salary.ID
}

func TestBudgetStatus(t *testing.T) {
	tests := []struct {
		spent, amount int64
		wantStatus    domain.BudgetStatus
		wantPercent   int64
	}{
		{0, 1_000_000, domain.BudgetSafe, 0},
		{799_999, 1_000_000, domain.BudgetSafe, 79}, // 79.9999% is still safe
		{800_000, 1_000_000, domain.BudgetWarning, 80},
		{999_999, 1_000_000, domain.BudgetWarning, 99},
		{1_000_000, 1_000_000, domain.BudgetOver, 100},
		{1_500_000, 1_000_000, domain.BudgetOver, 150},
	}
	for _, tt := range tests {
		got := summarize(domain.Budget{Spent: tt.spent, Amount: tt.amount})
		if got.Status != tt.wantStatus || got.Percentage != tt.wantPercent {
			t.Errorf("spent %d of %d: status/percentage = %s/%d, want %s/%d",
				tt.spent, tt.amount, got.Status, got.Percentage, tt.wantStatus, tt.wantPercent)
		}
	}
}

func TestBudgetCreate_StoresFirstDayOfMonth(t *testing.T) {
	svc, _, food, _ := newBudgetFixture(t)

	created, err := svc.Create(context.Background(), "user-1", food, "2026-11", 1_500_000)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got := created.Month.Format(dateLayout); got != "2026-11-01" {
		t.Errorf("month = %s, want 2026-11-01", got)
	}
	if created.Status != domain.BudgetSafe {
		t.Errorf("status = %s, want safe", created.Status)
	}
}

func TestBudgetCreate_RejectsInvalidInput(t *testing.T) {
	svc, _, food, salary := newBudgetFixture(t)

	tests := []struct {
		name       string
		categoryID string
		month      string
		amount     int64
		field      string
	}{
		{"income category", salary, "2026-11", 100, "categoryId"},
		{"unknown category", testUUID('c', 9), "2026-11", 100, "categoryId"},
		{"bad month", food, "2026-13", 100, "month"},
		{"zero amount", food, "2026-11", 0, "amount"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), "user-1", tt.categoryID, tt.month, tt.amount)
			var validationErr *domain.ValidationError
			if !errors.As(err, &validationErr) || validationErr.Fields[tt.field] == "" {
				t.Fatalf("Create() error = %v, want validation error on %q", err, tt.field)
			}
		})
	}
}

func TestBudgetCreate_RejectsSecondBudgetForSameCategoryAndMonth(t *testing.T) {
	svc, _, food, _ := newBudgetFixture(t)

	if _, err := svc.Create(context.Background(), "user-1", food, "2026-11", 100); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if _, err := svc.Create(context.Background(), "user-1", food, "2026-11", 200); !errors.Is(err, domain.ErrBudgetExists) {
		t.Fatalf("second Create() error = %v, want ErrBudgetExists", err)
	}
	if _, err := svc.Create(context.Background(), "user-1", food, "2026-12", 200); err != nil {
		t.Fatalf("next month Create() error = %v", err)
	}
}

func TestBudgetCopy_UsesPreviousMonthAcrossYears(t *testing.T) {
	svc, repo, _, _ := newBudgetFixture(t)

	if _, err := svc.CopyFromPreviousMonth(context.Background(), "user-1", "2027-01"); err != nil {
		t.Fatalf("CopyFromPreviousMonth() error = %v", err)
	}
	if from, to := repo.copyFrom.Format(dateLayout), repo.copyTo.Format(dateLayout); from != "2026-12-01" || to != "2027-01-01" {
		t.Errorf("copy %s -> %s, want 2026-12-01 -> 2027-01-01", from, to)
	}
}

func TestBudgetUpdateAmount_RejectsNonPositive(t *testing.T) {
	svc, _, food, _ := newBudgetFixture(t)
	created, err := svc.Create(context.Background(), "user-1", food, "2026-11", 100)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var validationErr *domain.ValidationError
	if _, err := svc.UpdateAmount(context.Background(), "user-1", created.ID, -5); !errors.As(err, &validationErr) {
		t.Fatalf("UpdateAmount() error = %v, want *domain.ValidationError", err)
	}
	if _, err := svc.UpdateAmount(context.Background(), "user-2", created.ID, 500); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("UpdateAmount() by another user error = %v, want ErrNotFound", err)
	}
}
