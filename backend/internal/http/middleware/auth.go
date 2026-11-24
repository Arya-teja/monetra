package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Arya-teja/monetra/backend/internal/http/response"
	"github.com/Arya-teja/monetra/backend/internal/service"
)

type userIDKey struct{}

// Auth rejects requests without a valid Bearer token and stores the user ID in the request context.
func Auth(tokens *service.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				response.Error(w, http.StatusUnauthorized, response.CodeUnauthorized, "missing bearer token")
				return
			}

			userID, err := tokens.Verify(token)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, response.CodeUnauthorized, "invalid or expired token")
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
