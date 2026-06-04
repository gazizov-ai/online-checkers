package identity

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/service"
	githubjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RSAIssuer struct {
	privateKey     *rsa.PrivateKey
	keyID          string
	issuer         string
	audience       string
	accessTokenTTL time.Duration
	idTokenTTL     time.Duration
}

func NewRSAIssuer(
	privateKey *rsa.PrivateKey,
	keyID string,
	issuer string,
	audience string,
	accessTokenTTL time.Duration,
	idTokenTTL time.Duration,
) *RSAIssuer {
	return &RSAIssuer{
		privateKey:     privateKey,
		keyID:          keyID,
		issuer:         issuer,
		audience:       audience,
		accessTokenTTL: accessTokenTTL,
		idTokenTTL:     idTokenTTL,
	}
}

func (i *RSAIssuer) IssueTokens(ctx context.Context, subject service.TokenSubject) (*service.TokenPair, error) {
	accessToken, err := i.issueToken(subject, "access", i.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	idToken, err := i.issueToken(subject, "id", i.idTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("issue id token: %w", err)
	}

	return &service.TokenPair{
		AccessToken: accessToken,
		IDToken:     idToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(i.accessTokenTTL.Seconds()),
	}, nil
}

func (i *RSAIssuer) issueToken(subject service.TokenSubject, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)
	jti := uuid.NewString()
	claims := appjwt.Claims{
		TokenUse:          tokenType,
		PreferredUsername: subject.Username,
		Email:             subject.Email,
		RegisteredClaims: githubjwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   subject.UserID.String(),
			Audience:  githubjwt.ClaimStrings{i.audience},
			ExpiresAt: githubjwt.NewNumericDate(expiresAt),
			IssuedAt:  githubjwt.NewNumericDate(now),
			NotBefore: githubjwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := githubjwt.NewWithClaims(githubjwt.SigningMethodRS256, claims)
	token.Header["kid"] = i.keyID
	token.Header["typ"] = "JWT"
	signed, err := token.SignedString(i.privateKey)
	return signed, err
}
