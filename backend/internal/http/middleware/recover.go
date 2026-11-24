package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/Arya-teja/monetra/backend/internal/http/response"
)

// Recover turns a panic in a handler into a 500 response instead of a dropped connection.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
			slog.Error("panic recovered", "error", rec, "stack", string(debug.Stack()))
			response.Error(w, http.StatusInternalServerError, response.CodeInternal, "internal server error")
		}()

		next.ServeHTTP(w, r)
	})
}
