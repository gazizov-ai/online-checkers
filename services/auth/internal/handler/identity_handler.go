package handler

import (
	"net/http"

	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/identity"
)

type IdentityProvider interface {
	JWKS() appjwt.JWKSResponse
	DiscoveryDocument() identity.DiscoveryDocument
}

type IdentityHandler struct {
	idp IdentityProvider
}

func NewIdentityHandler(idp IdentityProvider) *IdentityHandler {
	return &IdentityHandler{
		idp: idp,
	}
}

func (h *IdentityHandler) JWKS(w http.ResponseWriter, r *http.Request) {
	_ = httpx.WriteJSON(w, http.StatusOK, h.idp.JWKS())
}

func (h *IdentityHandler) Discovery(w http.ResponseWriter, r *http.Request) {
	_ = httpx.WriteJSON(w, http.StatusOK, h.idp.DiscoveryDocument())
}
