package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

const maxCategoryNameLength = 50

type CategoryService struct {
	categories domain.CategoryRepository
}

func NewCategoryService(categories domain.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

// List returns the user's categories; typ filters by type when it is not empty.
func (s *CategoryService) List(ctx context.Context, userID string, typ string) ([]domain.Category, error) {
	var filter *domain.CashflowType
	if typ != "" {
		t := domain.CashflowType(typ)
		if !t.Valid() {
			return nil, &domain.ValidationError{Fields: map[string]string{"type": "must be income or expense"}}
		}
		filter = &t
	}

	categories, err := s.categories.List(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return categories, nil
}

func (s *CategoryService) Create(ctx context.Context, userID, name string, typ domain.CashflowType) (domain.Category, error) {
	name = strings.TrimSpace(name)

	fields := map[string]string{}
	if msg := validateCategoryName(name); msg != "" {
		fields["name"] = msg
	}
	if !typ.Valid() {
		fields["type"] = "must be income or expense"
	}
	if len(fields) > 0 {
		return domain.Category{}, &domain.ValidationError{Fields: fields}
	}

	category, err := s.categories.Create(ctx, domain.Category{UserID: userID, Name: name, Type: typ})
	if err != nil {
		return domain.Category{}, fmt.Errorf("create category: %w", err)
	}
	return category, nil
}

func (s *CategoryService) Rename(ctx context.Context, userID, id, name string) (domain.Category, error) {
	name = strings.TrimSpace(name)
	if msg := validateCategoryName(name); msg != "" {
		return domain.Category{}, &domain.ValidationError{Fields: map[string]string{"name": msg}}
	}

	category, err := s.categories.Rename(ctx, userID, id, name)
	if err != nil {
		return domain.Category{}, fmt.Errorf("rename category: %w", err)
	}
	return category, nil
}

func (s *CategoryService) Delete(ctx context.Context, userID, id string) error {
	if err := s.categories.Delete(ctx, userID, id); err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}

func validateCategoryName(name string) string {
	if n := utf8.RuneCountInString(name); n == 0 || n > maxCategoryNameLength {
		return fmt.Sprintf("must be 1-%d characters", maxCategoryNameLength)
	}
	return ""
}
