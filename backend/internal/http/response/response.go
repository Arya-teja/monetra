package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

type ErrorCode string

const (
	CodeValidation   ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeTooMany      ErrorCode = "TOO_MANY_REQUESTS"
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
)

type errorBody struct {
	Code    ErrorCode         `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Meta describes one page of a paginated list.
type Meta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}

type envelope struct {
	Data  any        `json:"data,omitempty"`
	Meta  *Meta      `json:"meta,omitempty"`
	Error *errorBody `json:"error,omitempty"`
}

// JSON writes a success response: { "data": ... }.
func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, envelope{Data: data})
}

// Page writes a paginated success response: { "data": [...], "meta": {...} }.
func Page(w http.ResponseWriter, data any, meta Meta) {
	write(w, http.StatusOK, envelope{Data: data, Meta: &meta})
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

// TooManyRequests writes a 429 response. The wait is sent both as the standard
// Retry-After header and in details, which the frontend reads from the body.
func TooManyRequests(w http.ResponseWriter, message string, retryAfterSeconds int) {
	seconds := strconv.Itoa(retryAfterSeconds)
	w.Header().Set("Retry-After", seconds)
	write(w, http.StatusTooManyRequests, envelope{Error: &errorBody{
		Code:    CodeTooMany,
		Message: message,
		Details: map[string]string{"retryAfterSeconds": seconds},
	}})
}

func write(w http.ResponseWriter, status int, body envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write response", "error", err)
	}
}
