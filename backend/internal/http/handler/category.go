package handler

import (
	"net/http"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/http/middleware"
	"github.com/Arya-teja/monetra/backend/internal/http/response"
	"github.com/Arya-teja/monetra/backend/internal/service"
)

type CategoryHandler struct {
	categories *service.CategoryService
}

func NewCategoryHandler(categories *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

type categoryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsDefault bool   `json:"isDefault"`
}

func toCategoryResponse(c domain.Category) categoryResponse {
	return categoryResponse{ID: c.ID, Name: c.Name, Type: string(c.Type), IsDefault: c.IsDefault}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categories.List(r.Context(), middleware.UserID(r.Context()), r.URL.Query().Get("type"))
	if err != nil {
		writeError(w, err)
		return
	}

	body := make([]categoryResponse, 0, len(categories))
	for _, c := range categories {
		body = append(body, toCategoryResponse(c))
	}
	response.JSON(w, http.StatusOK, body)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	category, err := h.categories.Create(r.Context(), middleware.UserID(r.Context()), req.Name, domain.CashflowType(req.Type))
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, toCategoryResponse(category))
}

func (h *CategoryHandler) Rename(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	category, err := h.categories.Rename(r.Context(), middleware.UserID(r.Context()), r.PathValue("id"), req.Name)
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toCategoryResponse(category))
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.categories.Delete(r.Context(), middleware.UserID(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
