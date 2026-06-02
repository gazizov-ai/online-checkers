package identity

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID   uuid.UUID
	Username string
	Email    *string
	TokenUse string
}

func (i *RSAIssuer) VerifyAccessToken(ctx context.Context, rawToken string) (*TokenClaims, error) {
	_ = ctx

	keyFunc := func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, ErrInvalidToken
		}

		kid, _ := token.Header["kid"].(string)
		if kid != i.keyID {
			return nil, ErrInvalidToken
		}

		return &i.privateKey.PublicKey, nil
	}

	parsed, err := jwt.ParseWithClaims(
		rawToken,
		&Claims{},
		keyFunc,
		jwt.WithIssuer(i.issuer),
		jwt.WithAudience(i.audience),
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenUse != "access" {
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
