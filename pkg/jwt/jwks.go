package jwt

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func NewJWKSResponse(publicKey rsa.PublicKey, keyID string) JWKSResponse {
	return JWKSResponse{
		Keys: []JWK{
			{
				Kty: "RSA",
				Use: "sig",
				Kid: keyID,
				Alg: "RS256",
				N:   encodeBase64URL(publicKey.N.Bytes()),
				E:   encodeExponent(publicKey.E),
			},
		},
	}
}

func LoadRSAPublicKeyFromJWKS(ctx context.Context, jwksURL string, keyID string) (*rsa.PublicKey, error) {
	return LoadRSAPublicKeyFromJWKSWithClient(ctx, http.DefaultClient, jwksURL, keyID)
}

func LoadRSAPublicKeyFromJWKSWithClient(
	ctx context.Context,
	client *http.Client,
	jwksURL string,
	keyID string,
) (*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create jwks request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch jwks: unexpected status %d", resp.StatusCode)
	}

	var jwks JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}

	return FindRSAPublicKey(jwks, keyID)
}

func FindRSAPublicKey(jwks JWKSResponse, keyID string) (*rsa.PublicKey, error) {
	for _, key := range jwks.Keys {
		if key.Kid != keyID {
			continue
		}

		if key.Kty != "RSA" || key.Use != "sig" || key.Alg != "RS256" {
			return nil, ErrInvalidToken
		}

		return rsaPublicKeyFromJWK(key)
	}

	return nil, ErrInvalidToken
}

func rsaPublicKeyFromJWK(key JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, ErrInvalidToken
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, ErrInvalidToken
	}

	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	if n.Sign() <= 0 || e <= 0 {
		return nil, ErrInvalidToken
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

func encodeBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func encodeExponent(e int) string {
	return encodeBase64URL(big.NewInt(int64(e)).Bytes())
}
