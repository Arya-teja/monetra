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

type CategoryRepository struct {
	queries *db.Queries
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{queries: db.New(pool)}
}

func (r *CategoryRepository) List(ctx context.Context, userID string, typ *domain.CashflowType) ([]domain.Category, error) {
	params := db.ListCategoriesParams{UserID: userID}
	if typ != nil {
		params.Type = db.NullCashflowType{CashflowType: db.CashflowType(*typ), Valid: true}
	}

	rows, err := r.queries.ListCategories(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("select categories: %w", err)
	}

	categories := make([]domain.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, toDomainCategory(row))
	}
	return categories, nil
}

func (r *CategoryRepository) Get(ctx context.Context, userID, id string) (domain.Category, error) {
	row, err := r.queries.GetCategory(ctx, db.GetCategoryParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) || hasCode(err, codeInvalidText) {
		return domain.Category{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Category{}, fmt.Errorf("select category: %w", err)
	}
	return toDomainCategory(row), nil
}

func (r *CategoryRepository) Create(ctx context.Context, c domain.Category) (domain.Category, error) {
	row, err := r.queries.CreateCategory(ctx, db.CreateCategoryParams{
		UserID:    c.UserID,
		Name:      c.Name,
		Type:      db.CashflowType(c.Type),
		IsDefault: c.IsDefault,
	})
	if hasCode(err, codeUniqueViolation) {
		return domain.Category{}, domain.ErrCategoryExists
	}
	if err != nil {
		return domain.Category{}, fmt.Errorf("insert category: %w", err)
	}
	return toDomainCategory(row), nil
}

func (r *CategoryRepository) Rename(ctx context.Context, userID, id, name string) (domain.Category, error) {
	row, err := r.queries.RenameCategory(ctx, db.RenameCategoryParams{ID: id, UserID: userID, Name: name})
	switch {
	case errors.Is(err, pgx.ErrNoRows) || hasCode(err, codeInvalidText):
		return domain.Category{}, domain.ErrNotFound
	case hasCode(err, codeUniqueViolation):
		return domain.Category{}, domain.ErrCategoryExists
	case err != nil:
		return domain.Category{}, fmt.Errorf("rename category: %w", err)
	}
	return toDomainCategory(row), nil
}

func (r *CategoryRepository) Delete(ctx context.Context, userID, id string) error {
	deleted, err := r.queries.DeleteCategory(ctx, db.DeleteCategoryParams{ID: id, UserID: userID})
	switch {
	case hasCode(err, codeForeignKeyViolation):
		return domain.ErrCategoryInUse
	case hasCode(err, codeInvalidText):
		return domain.ErrNotFound
	case err != nil:
		return fmt.Errorf("delete category: %w", err)
	case deleted == 0:
		return domain.ErrNotFound
	}
	return nil
}

func toDomainCategory(row db.Category) domain.Category {
	return domain.Category{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		Type:      domain.CashflowType(row.Type),
		IsDefault: row.IsDefault,
		CreatedAt: row.CreatedAt,
	}
}
