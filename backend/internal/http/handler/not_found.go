package handler

import (
	"net/http"

	"github.com/Arya-teja/monetra/backend/internal/http/response"
)

// NotFound answers unknown routes with the standard error envelope.
func NotFound(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotFound, response.CodeNotFound, "route not found")
}
