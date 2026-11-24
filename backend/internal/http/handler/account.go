package handler

import (
	"net/http"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/http/middleware"
	"github.com/Arya-teja/monetra/backend/internal/http/response"
	"github.com/Arya-teja/monetra/backend/internal/service"
)

type AccountHandler struct {
	accounts *service.AccountService
}

func NewAccountHandler(accounts *service.AccountService) *AccountHandler {
	return &AccountHandler{accounts: accounts}
}

type accountResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	InstitutionName *string   `json:"institutionName"`
	AccountNumber   *string   `json:"accountNumber"`
	Balance         int64     `json:"balance"`
	IsArchived      bool      `json:"isArchived"`
	CreatedAt       time.Time `json:"createdAt"`
}

func toAccountResponse(a domain.Account) accountResponse {
	return accountResponse{
		ID:              a.ID,
		Name:            a.Name,
		Type:            string(a.Type),
		InstitutionName: a.InstitutionName,
		AccountNumber:   a.AccountNumber,
		Balance:         a.Balance,
		IsArchived:      a.IsArchived,
		CreatedAt:       a.CreatedAt,
	}
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	includeArchived := r.URL.Query().Get("includeArchived") == "true"

	accounts, err := h.accounts.List(r.Context(), middleware.UserID(r.Context()), includeArchived)
	if err != nil {
		writeError(w, err)
		return
	}

	body := make([]accountResponse, 0, len(accounts))
	for _, a := range accounts {
		body = append(body, toAccountResponse(a))
	}
	response.JSON(w, http.StatusOK, body)
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string `json:"name"`
		Type            string `json:"type"`
		InstitutionName string `json:"institutionName"`
		AccountNumber   string `json:"accountNumber"`
		Balance         int64  `json:"balance"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	account, err := h.accounts.Create(r.Context(), middleware.UserID(r.Context()), service.CreateAccountInput{
		Name:            req.Name,
		Type:            domain.AccountType(req.Type),
		InstitutionName: req.InstitutionName,
		AccountNumber:   req.AccountNumber,
		Balance:         req.Balance,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, toAccountResponse(account))
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            *string             `json:"name"`
		Type            *domain.AccountType `json:"type"`
		InstitutionName *string             `json:"institutionName"`
		AccountNumber   *string             `json:"accountNumber"`
		IsArchived      *bool               `json:"isArchived"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	account, err := h.accounts.Update(r.Context(), middleware.UserID(r.Context()), r.PathValue("id"), service.UpdateAccountInput{
		Name:            req.Name,
		Type:            req.Type,
		InstitutionName: req.InstitutionName,
		AccountNumber:   req.AccountNumber,
		IsArchived:      req.IsArchived,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toAccountResponse(account))
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.accounts.Delete(r.Context(), middleware.UserID(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
