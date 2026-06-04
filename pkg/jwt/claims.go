package jwt

import (
	githubjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenUseAccess = "access"
	TokenUseID     = "id"
)

type Claims struct {
	TokenUse          string  `json:"token_use"`
	PreferredUsername string  `json:"preferred_username"`
	Email             *string `json:"email,omitempty"`

	githubjwt.RegisteredClaims
}

type TokenClaims struct {
	UserID   uuid.UUID
	Username string
	Email    *string
	TokenUse string
}
