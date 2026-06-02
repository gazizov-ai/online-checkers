package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/identity"
	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

type TokenVerifier interface {
	VerifyAccessToken(ctx context.Context, rawToken string) (*identity.TokenClaims, error)
}

func AuthMiddleware(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken, ok := bearerToken(r)
			if !ok {
				_ = httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}

			claims, err := verifier.VerifyAccessToken(r.Context(), rawToken)
			if err != nil {
				_ = httpx.WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid bearer token")
				return
			}

			ctx := withUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}
	return token, true
}

func userIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	value, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	return value, true
}

func withUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}
