package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	// TokenVersion is embedded in every access token; logging out increments it,
	// which invalidates all tokens issued before.
	TokenVersion int
	CreatedAt    time.Time
}

type UserRepository interface {
	// CreateWithCategories stores the user and their categories atomically.
	// It returns ErrEmailTaken when the email is already registered.
	CreateWithCategories(ctx context.Context, user User, categories []Category) (User, error)
	// FindByEmail returns ErrNotFound when no user has the email.
	FindByEmail(ctx context.Context, email string) (User, error)
	// FindByID returns ErrNotFound when no user has the ID.
	FindByID(ctx context.Context, id string) (User, error)
	// TokenVersion returns ErrNotFound when no user has the ID.
	TokenVersion(ctx context.Context, id string) (int, error)
	// IncrementTokenVersion returns ErrNotFound when no user has the ID.
	IncrementTokenVersion(ctx context.Context, id string) error
}
