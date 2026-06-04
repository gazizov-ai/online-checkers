package identity

import (
	"context"

	appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"
)

func (i *RSAIssuer) VerifyAccessToken(ctx context.Context, rawToken string) (*appjwt.TokenClaims, error) {
	verifier := appjwt.NewRS256Verifier(
		&i.privateKey.PublicKey,
		i.keyID,
		i.issuer,
		i.audience,
	)

	return verifier.VerifyAccessToken(ctx, rawToken)
}
