package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/repository/postgres/db"
)

const uniqueViolation = "23505"

type UserRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool, queries: db.New(pool)}
}

func (r *UserRepository) CreateWithCategories(ctx context.Context, user domain.User, categories []domain.Category) (domain.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.User{}, fmt.Errorf("begin transaction: %w", err)
	}
	// Rollback is a no-op once the transaction is committed.
	defer tx.Rollback(ctx)

	q := r.queries.WithTx(tx)

	row, err := q.CreateUser(ctx, db.CreateUserParams{
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	for _, c := range categories {
		err := q.CreateCategory(ctx, db.CreateCategoryParams{
			UserID:    row.ID,
			Name:      c.Name,
			Type:      db.CashflowType(c.Type),
			IsDefault: c.IsDefault,
		})
		if err != nil {
			return domain.User{}, fmt.Errorf("insert category %q: %w", c.Name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, fmt.Errorf("commit transaction: %w", err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("select user by email: %w", err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("select user by id: %w", err)
	}
	return toDomainUser(row), nil
}

func toDomainUser(row db.User) domain.User {
	return domain.User{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}
}
