package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorCode string

const (
	CodeValidation   ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
)

type errorBody struct {
	Code    ErrorCode         `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *errorBody `json:"error,omitempty"`
}

// JSON writes a success response: { "data": ... }.
func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, envelope{Data: data})
}

// Error writes an error response: { "error": { "code": ..., "message": ... } }.
func Error(w http.ResponseWriter, status int, code ErrorCode, message string) {
	write(w, status, envelope{Error: &errorBody{Code: code, Message: message}})
}

// ValidationError writes a 400 response with a message per invalid field.
func ValidationError(w http.ResponseWriter, details map[string]string) {
	write(w, http.StatusBadRequest, envelope{Error: &errorBody{
		Code:    CodeValidation,
		Message: "invalid input",
		Details: details,
	}})
}

func write(w http.ResponseWriter, status int, body envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write response", "error", err)
	}
}
