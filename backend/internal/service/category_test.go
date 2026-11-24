package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

// fakeCategoryRepository is an in-memory domain.CategoryRepository.
type fakeCategoryRepository struct {
	categories map[string]domain.Category // keyed by ID
	inUse      map[string]bool            // category IDs used by transactions
	nextID     int
}

func newFakeCategoryRepository() *fakeCategoryRepository {
	return &fakeCategoryRepository{categories: map[string]domain.Category{}, inUse: map[string]bool{}}
}

func (r *fakeCategoryRepository) List(_ context.Context, userID string, typ *domain.CashflowType) ([]domain.Category, error) {
	var out []domain.Category
	for _, c := range r.categories {
		if c.UserID == userID && (typ == nil || c.Type == *typ) {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *fakeCategoryRepository) Get(_ context.Context, userID, id string) (domain.Category, error) {
	c, ok := r.categories[id]
	if !ok || c.UserID != userID {
		return domain.Category{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *fakeCategoryRepository) Create(_ context.Context, c domain.Category) (domain.Category, error) {
	for _, existing := range r.categories {
		if existing.UserID == c.UserID && existing.Type == c.Type && existing.Name == c.Name {
			return domain.Category{}, domain.ErrCategoryExists
		}
	}
	r.nextID++
	c.ID = testUUID('c', r.nextID)
	r.categories[c.ID] = c
	return c, nil
}

func (r *fakeCategoryRepository) Rename(ctx context.Context, userID, id, name string) (domain.Category, error) {
	c, err := r.Get(ctx, userID, id)
	if err != nil {
		return domain.Category{}, err
	}
	c.Name = name
	r.categories[id] = c
	return c, nil
}

func (r *fakeCategoryRepository) Delete(ctx context.Context, userID, id string) error {
	if _, err := r.Get(ctx, userID, id); err != nil {
		return err
	}
	if r.inUse[id] {
		return domain.ErrCategoryInUse
	}
	delete(r.categories, id)
	return nil
}

// testUUID builds a valid, readable UUID such as "cccccccc-0000-0000-0000-000000000001".
func testUUID(prefix byte, n int) string {
	return strings.Repeat(string(prefix), 8) + "-0000-0000-0000-" + strings.Repeat("0", 11) + string(rune('0'+n%10))
}

func TestCategoryCreate_TrimsNameAndValidates(t *testing.T) {
	svc := NewCategoryService(newFakeCategoryRepository())

	created, err := svc.Create(context.Background(), "user-1", "  Kopi  ", domain.CashflowExpense)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Name != "Kopi" {
		t.Errorf("name = %q, want %q", created.Name, "Kopi")
	}

	_, err = svc.Create(context.Background(), "user-1", " ", "kripto")
	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) || validationErr.Fields["name"] == "" || validationErr.Fields["type"] == "" {
		t.Fatalf("Create() error = %v, want name and type validation errors", err)
	}
}

func TestCategoryCreate_RejectsDuplicateNameForSameType(t *testing.T) {
	svc := NewCategoryService(newFakeCategoryRepository())

	if _, err := svc.Create(context.Background(), "user-1", "Lainnya", domain.CashflowExpense); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if _, err := svc.Create(context.Background(), "user-1", "Lainnya", domain.CashflowExpense); !errors.Is(err, domain.ErrCategoryExists) {
		t.Fatalf("duplicate Create() error = %v, want ErrCategoryExists", err)
	}
	if _, err := svc.Create(context.Background(), "user-1", "Lainnya", domain.CashflowIncome); err != nil {
		t.Fatalf("same name with other type: Create() error = %v", err)
	}
}

func TestCategoryList_RejectsUnknownType(t *testing.T) {
	svc := NewCategoryService(newFakeCategoryRepository())

	var validationErr *domain.ValidationError
	if _, err := svc.List(context.Background(), "user-1", "kripto"); !errors.As(err, &validationErr) {
		t.Fatalf("List() error = %v, want *domain.ValidationError", err)
	}
}

func TestCategoryDelete_RejectsCategoryInUse(t *testing.T) {
	repo := newFakeCategoryRepository()
	svc := NewCategoryService(repo)
	created, err := svc.Create(context.Background(), "user-1", "Kopi", domain.CashflowExpense)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	repo.inUse[created.ID] = true

	if err := svc.Delete(context.Background(), "user-1", created.ID); !errors.Is(err, domain.ErrCategoryInUse) {
		t.Fatalf("Delete() error = %v, want ErrCategoryInUse", err)
	}
}
