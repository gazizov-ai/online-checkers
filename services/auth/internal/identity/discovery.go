package identity

type DiscoveryDocument struct {
	Issuer                           string   `json:"issuer"`
	JWKSURI                          string   `json:"jwks_uri"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	UserInfoEndpoint                 string   `json:"userinfo_endpoint"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	ClaimsSupported                  []string `json:"claims_supported"`
}

func (i *RSAIssuer) DiscoveryDocument() DiscoveryDocument {
	return DiscoveryDocument{
		Issuer:           i.issuer,
		JWKSURI:          i.issuer + "/.well-known/jwks.json",
		TokenEndpoint:    i.issuer + "/api/v1/login",
		UserInfoEndpoint: i.issuer + "/api/v1/me",

		IDTokenSigningAlgValuesSupported: []string{"RS256"},
		ResponseTypesSupported:           []string{"token"},
		SubjectTypesSupported:            []string{"public"},
		ClaimsSupported: []string{
			"sub",
			"iss",
			"aud",
			"exp",
			"iat",
			"nbf",
			"jti",
			"preferred_username",
			"email",
			"token_use",
		},
	}
}
