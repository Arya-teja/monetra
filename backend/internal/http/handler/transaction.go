package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/http/middleware"
	"github.com/Arya-teja/monetra/backend/internal/http/response"
	"github.com/Arya-teja/monetra/backend/internal/service"
)

const dateLayout = "2006-01-02"

type TransactionHandler struct {
	transactions *service.TransactionService
}

func NewTransactionHandler(transactions *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactions: transactions}
}

type transactionResponse struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"accountId"`
	AccountName  string    `json:"accountName"`
	CategoryID   string    `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Type         string    `json:"type"`
	Amount       int64     `json:"amount"`
	OccurredAt   string    `json:"occurredAt"`
	Note         *string   `json:"note"`
	CreatedAt    time.Time `json:"createdAt"`
}

func toTransactionResponse(t domain.Transaction) transactionResponse {
	return transactionResponse{
		ID:           t.ID,
		AccountID:    t.AccountID,
		AccountName:  t.AccountName,
		CategoryID:   t.CategoryID,
		CategoryName: t.CategoryName,
		Type:         string(t.Type),
		Amount:       t.Amount,
		OccurredAt:   t.OccurredAt.Format(dateLayout),
		Note:         t.Note,
		CreatedAt:    t.CreatedAt,
	}
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, pageErr := optionalInt(q.Get("page"))
	pageSize, pageSizeErr := optionalInt(q.Get("pageSize"))
	if pageErr || pageSizeErr {
		response.ValidationError(w, map[string]string{"page": "page and pageSize must be numbers"})
		return
	}

	result, err := h.transactions.List(r.Context(), middleware.UserID(r.Context()), service.ListTransactionsInput{
		Month:      q.Get("month"),
		AccountID:  q.Get("accountId"),
		CategoryID: q.Get("categoryId"),
		Type:       q.Get("type"),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	body := make([]transactionResponse, 0, len(result.Items))
	for _, t := range result.Items {
		body = append(body, toTransactionResponse(t))
	}
	response.Page(w, body, response.Meta{Page: result.Page, PageSize: result.PageSize, Total: result.Total})
}

func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	t, err := h.transactions.Get(r.Context(), middleware.UserID(r.Context()), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toTransactionResponse(t))
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID  string `json:"accountId"`
		CategoryID string `json:"categoryId"`
		Type       string `json:"type"`
		Amount     int64  `json:"amount"`
		OccurredAt string `json:"occurredAt"`
		Note       string `json:"note"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	t, err := h.transactions.Create(r.Context(), middleware.UserID(r.Context()), service.CreateTransactionInput{
		AccountID:  req.AccountID,
		CategoryID: req.CategoryID,
		Type:       domain.CashflowType(req.Type),
		Amount:     req.Amount,
		OccurredAt: req.OccurredAt,
		Note:       req.Note,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, toTransactionResponse(t))
}

func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID  *string              `json:"accountId"`
		CategoryID *string              `json:"categoryId"`
		Type       *domain.CashflowType `json:"type"`
		Amount     *int64               `json:"amount"`
		OccurredAt *string              `json:"occurredAt"`
		Note       *string              `json:"note"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	t, err := h.transactions.Update(r.Context(), middleware.UserID(r.Context()), r.PathValue("id"), service.UpdateTransactionInput{
		AccountID:  req.AccountID,
		CategoryID: req.CategoryID,
		Type:       req.Type,
		Amount:     req.Amount,
		OccurredAt: req.OccurredAt,
		Note:       req.Note,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toTransactionResponse(t))
}

func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.transactions.Delete(r.Context(), middleware.UserID(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// optionalInt parses s as an integer; an empty string is 0. It reports whether parsing failed.
func optionalInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	return n, err != nil
}
