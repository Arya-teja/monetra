package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrInvalidToken means the access token is malformed, expired, or revoked by a logout.
	ErrInvalidToken   = errors.New("invalid or expired token")
	ErrAccountInUse   = errors.New("account has transactions")
	ErrCategoryExists = errors.New("category already exists")
	ErrCategoryInUse  = errors.New("category is used by transactions")
	ErrBudgetExists   = errors.New("budget already exists for this category and month")
	// ErrStaleData means the record changed after it was read; the caller should reload.
	ErrStaleData = errors.New("data was changed by another request")
)

// ValidationError lists invalid input fields with a message per field.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

// TooManyAttemptsError means login for this email is locked after too many failed
// attempts; RetryAfter is how long until it unlocks.
type TooManyAttemptsError struct {
	RetryAfter time.Duration
}

func (e *TooManyAttemptsError) Error() string {
	return "too many failed login attempts"
}
