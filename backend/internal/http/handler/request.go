package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Arya-teja/monetra/backend/internal/http/response"
)

const maxBodyBytes = 1 << 20 // 1 MB

// decodeJSON reads the request body into dst. On failure it writes a 400 response and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeValidation, "invalid JSON body")
		return false
	}
	return true
}
