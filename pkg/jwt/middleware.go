package jwt

import (
	"net/http"

	"github.com/gazizov-ai/online-checkers/pkg/httpx"
)

func Middleware(verifier Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken, ok := TokenFromRequest(r)
			if !ok {
				_ = httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}

			claims, err := verifier.VerifyAccessToken(r.Context(), rawToken)
			if err != nil {
				_ = httpx.WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid bearer token")
				return
			}

			ctx := WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
