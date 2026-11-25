package handler

import (
	"net/http"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/http/middleware"
	"github.com/Arya-teja/monetra/backend/internal/http/response"
	"github.com/Arya-teja/monetra/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type userResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func toUserResponse(u domain.User) userResponse {
	return userResponse{ID: u.ID, Name: u.Name, Email: u.Email, CreatedAt: u.CreatedAt}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	user, err := h.auth.Register(r.Context(), service.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, toUserResponse(user))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	user, token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, struct {
		AccessToken string       `json:"accessToken"`
		ExpiresAt   time.Time    `json:"expiresAt"`
		User        userResponse `json:"user"`
	}{AccessToken: token.Value, ExpiresAt: token.ExpiresAt, User: toUserResponse(user)})
}

// Logout revokes all of the user's access tokens (every device).
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.Logout(r.Context(), middleware.UserID(r.Context())); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.auth.Me(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toUserResponse(user))
}
