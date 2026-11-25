package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/http/response"
)

// Authenticator resolves an access token to a user ID. It returns
// domain.ErrInvalidToken for tokens that must be rejected.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (string, error)
}

type userIDKey struct{}

// Auth rejects requests without a valid Bearer token and stores the user ID in the request context.
func Auth(auth Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				response.Error(w, http.StatusUnauthorized, response.CodeUnauthorized, "missing bearer token")
				return
			}

			userID, err := auth.Authenticate(r.Context(), token)
			switch {
			case errors.Is(err, domain.ErrInvalidToken):
				response.Error(w, http.StatusUnauthorized, response.CodeUnauthorized, "invalid or expired token")
				return
			case err != nil:
				slog.Error("authenticate", "error", err)
				response.Error(w, http.StatusInternalServerError, response.CodeInternal, "internal server error")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserID returns the authenticated user ID set by Auth.
func UserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey{}).(string)
	return id
}
