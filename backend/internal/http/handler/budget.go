package handler

import (
	"net/http"

	"github.com/Arya-teja/monetra/backend/internal/http/middleware"
	"github.com/Arya-teja/monetra/backend/internal/http/response"
	"github.com/Arya-teja/monetra/backend/internal/service"
)

const monthLayout = "2006-01"

type BudgetHandler struct {
	budgets *service.BudgetService
}

func NewBudgetHandler(budgets *service.BudgetService) *BudgetHandler {
	return &BudgetHandler{budgets: budgets}
}

type budgetResponse struct {
	ID           string `json:"id"`
	CategoryID   string `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Month        string `json:"month"`
	Amount       int64  `json:"amount"`
	Spent        int64  `json:"spent"`
	Percentage   int64  `json:"percentage"`
	Status       string `json:"status"`
}

func toBudgetResponse(b service.BudgetSummary) budgetResponse {
	return budgetResponse{
		ID:           b.ID,
		CategoryID:   b.CategoryID,
		CategoryName: b.CategoryName,
		Month:        b.Month.Format(monthLayout),
		Amount:       b.Amount,
		Spent:        b.Spent,
		Percentage:   b.Percentage,
		Status:       string(b.Status),
	}
}

func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	budgets, err := h.budgets.List(r.Context(), middleware.UserID(r.Context()), r.URL.Query().Get("month"))
	if err != nil {
		writeError(w, err)
		return
	}

	body := make([]budgetResponse, 0, len(budgets))
	for _, b := range budgets {
		body = append(body, toBudgetResponse(b))
	}
	response.JSON(w, http.StatusOK, body)
}

func (h *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CategoryID string `json:"categoryId"`
		Month      string `json:"month"`
		Amount     int64  `json:"amount"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	budget, err := h.budgets.Create(r.Context(), middleware.UserID(r.Context()), req.CategoryID, req.Month, req.Amount)
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, toBudgetResponse(budget))
}

func (h *BudgetHandler) Copy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Month string `json:"month"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	copied, err := h.budgets.CopyFromPreviousMonth(r.Context(), middleware.UserID(r.Context()), req.Month)
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]int64{"copied": copied})
}

func (h *BudgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Amount int64 `json:"amount"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	budget, err := h.budgets.UpdateAmount(r.Context(), middleware.UserID(r.Context()), r.PathValue("id"), req.Amount)
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toBudgetResponse(budget))
}

func (h *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.budgets.Delete(r.Context(), middleware.UserID(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
