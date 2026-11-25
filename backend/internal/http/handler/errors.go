package handler

import (
	"errors"
	"log/slog"
	"math"
	"net/http"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/http/response"
)

// writeError maps a service error to the standard error response.
func writeError(w http.ResponseWriter, err error) {
	var validationErr *domain.ValidationError
	var tooManyErr *domain.TooManyAttemptsError
	switch {
	case errors.As(err, &validationErr):
		response.ValidationError(w, validationErr.Fields)
	case errors.As(err, &tooManyErr):
		// Round up so the client never retries a second too early.
		seconds := int(math.Ceil(tooManyErr.RetryAfter.Seconds()))
		response.TooManyRequests(w, "too many failed login attempts", seconds)
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, response.CodeUnauthorized, "invalid email or password")
	case errors.Is(err, domain.ErrEmailTaken):
		response.Error(w, http.StatusConflict, response.CodeConflict, "email already registered")
	case errors.Is(err, domain.ErrAccountInUse):
		response.Error(w, http.StatusConflict, response.CodeConflict, "account has transactions; archive it instead")
	case errors.Is(err, domain.ErrCategoryExists):
		response.Error(w, http.StatusConflict, response.CodeConflict, "category already exists")
	case errors.Is(err, domain.ErrBudgetExists):
		response.Error(w, http.StatusConflict, response.CodeConflict, "budget already exists for this category and month")
	case errors.Is(err, domain.ErrCategoryInUse):
		response.Error(w, http.StatusConflict, response.CodeConflict, "category is used by transactions")
	case errors.Is(err, domain.ErrStaleData):
		response.Error(w, http.StatusConflict, response.CodeConflict, "data was changed; reload and try again")
	case errors.Is(err, domain.ErrNotFound):
		response.Error(w, http.StatusNotFound, response.CodeNotFound, "resource not found")
	default:
		slog.Error("unhandled error", "error", err)
		response.Error(w, http.StatusInternalServerError, response.CodeInternal, "internal server error")
	}
}
