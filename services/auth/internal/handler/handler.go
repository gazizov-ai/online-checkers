package handler

import (
	"errors"
	"net/http"

	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{
		auth: auth,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	user, err := h.auth.Register(r.Context(), service.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUsername):
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_username", "invalid username")
		case errors.Is(err, service.ErrInvalidPassword):
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_password", "invalid password")
		case errors.Is(err, service.ErrUsernameAlreadyTaken):
			_ = httpx.WriteError(w, http.StatusConflict, "username_taken", "username already taken")
		case errors.Is(err, service.ErrEmailAlreadyTaken):
			_ = httpx.WriteError(w, http.StatusConflict, "email_taken", "email already in use")
		default:
			_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		}
		return
	}
	_ = httpx.WriteJSON(w, http.StatusCreated, RegisterResponse{User: userToResponse(user)})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	result, err := h.auth.Login(r.Context(), service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			_ = httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "invalid credentials provided")
		} else {
			_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		}
		return
	}
	_ = httpx.WriteJSON(w, http.StatusOK, LoginResponse{
		AccessToken: result.Tokens.AccessToken,
		IDToken:     result.Tokens.IDToken,
		TokenType:   result.Tokens.TokenType,
		ExpiresIn:   result.Tokens.ExpiresIn,
		User:        userToResponse(result.User),
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	user, err := h.auth.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "user_not_found", "user not found")
		} else {
			_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		}
		return
	}
	_ = httpx.WriteJSON(w, http.StatusOK, MeResponse{User: userToResponse(user)})
}
