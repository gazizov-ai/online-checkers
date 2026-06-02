package identity

import (
	"encoding/base64"
	"math/big"
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

func (i *RSAIssuer) JWKS() JWKSResponse {
	publicKey := i.privateKey.PublicKey

	return JWKSResponse{
		Keys: []JWK{
			{
				Kty: "RSA",
				Use: "sig",
				Kid: i.keyID,
				Alg: "RS256",
				N:   encodeBase64URL(publicKey.N.Bytes()),
				E:   encodeExponent(publicKey.E),
			},
		},
	}
}

func encodeBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func encodeExponent(e int) string {
	bytes := big.NewInt(int64(e)).Bytes()
	return encodeBase64URL(bytes)
}
