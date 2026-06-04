package identity

import appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"

func (i *RSAIssuer) JWKS() appjwt.JWKSResponse {
	return appjwt.NewJWKSResponse(i.privateKey.PublicKey, i.keyID)
}
