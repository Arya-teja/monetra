package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountInUse       = errors.New("account has transactions")
	ErrCategoryExists     = errors.New("category already exists")
	ErrCategoryInUse      = errors.New("category is used by transactions")
	ErrBudgetExists       = errors.New("budget already exists for this category and month")
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
