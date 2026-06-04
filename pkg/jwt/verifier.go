package jwt

import (
	"context"
	"crypto/rsa"

	githubjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Verifier interface {
	VerifyAccessToken(ctx context.Context, rawToken string) (*TokenClaims, error)
}

type RS256Verifier struct {
	publicKey *rsa.PublicKey
	keyID     string
	issuer    string
	audience  string
}

func NewRS256Verifier(
	publicKey *rsa.PublicKey,
	keyID string,
	issuer string,
	audience string,
) *RS256Verifier {
	return &RS256Verifier{
		publicKey: publicKey,
		keyID:     keyID,
		issuer:    issuer,
		audience:  audience,
	}
}

func (v *RS256Verifier) VerifyAccessToken(ctx context.Context, rawToken string) (*TokenClaims, error) {
	_ = ctx

	keyFunc := func(token *githubjwt.Token) (any, error) {
		if token.Method.Alg() != githubjwt.SigningMethodRS256.Alg() {
			return nil, ErrInvalidToken
		}

		kid, _ := token.Header["kid"].(string)
		if kid != v.keyID {
			return nil, ErrInvalidToken
		}

		return v.publicKey, nil
	}

	parsed, err := githubjwt.ParseWithClaims(
		rawToken,
		&Claims{},
		keyFunc,
		githubjwt.WithIssuer(v.issuer),
		githubjwt.WithAudience(v.audience),
	)
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenUse != TokenUseAccess {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return &TokenClaims{
		UserID:   userID,
		Username: claims.PreferredUsername,
		Email:    claims.Email,
		TokenUse: claims.TokenUse,
	}, nil
}
