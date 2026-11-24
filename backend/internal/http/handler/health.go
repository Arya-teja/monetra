package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/http/response"
)

// Pinger checks that the database is reachable.
type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db Pinger
}

func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		slog.Error("health check: ping database", "error", err)
		response.Error(w, http.StatusInternalServerError, response.CodeInternal, "database unavailable")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
