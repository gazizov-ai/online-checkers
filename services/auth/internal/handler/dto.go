package handler

import (
	"time"

	"github.com/gazizov-ai/online-checkers/services/auth/internal/domain"
)

type RegisterRequest struct {
	Username string  `json:"username"`
	Email    *string `json:"email,omitempty"`
	Password string  `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	Email     *string `json:"email,omitempty"`
	CreatedAt string  `json:"created_at"`
}

type RegisterResponse struct {
	User UserResponse `json:"user"`
}

type MeResponse struct {
	User UserResponse `json:"user"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	IDToken     string       `json:"id_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	User        UserResponse `json:"user"`
}

func userToResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}
