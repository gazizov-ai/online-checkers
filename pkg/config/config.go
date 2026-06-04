package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName       string
	HTTPPort          string
	DatabaseURL       string
	KafkaBrokers      string
	JWTSecret         string
	OIDCIssuer        string
	OIDCAudience      string
	JWKSURL           string
	JWTPrivateKeyPath string
	JWTKeyID          string
	AccessTokenTTL    time.Duration
	IDTokenTTL        time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		ServiceName:       getEnv("SERVICE_NAME", "unknown-service"),
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		KafkaBrokers:      getEnv("KAFKA_BROKERS", ""),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		OIDCIssuer:        getEnv("OIDC_ISSUER", "http://localhost:8081"),
		OIDCAudience:      getEnv("OIDC_AUDIENCE", "online-checkers"),
		JWKSURL:           getEnv("JWKS_URL", "http://localhost:8081/.well-known/jwks.json"),
		JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", ""),
		JWTKeyID:          getEnv("JWT_KEY_ID", "dev-key-1"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port, err := strconv.Atoi(cfg.HTTPPort)

	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}

	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("HTTP_PORT must be between 1 and 65535")
	}

	accessTokenTTL, err := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_TTL: %w", err)
	}
	cfg.AccessTokenTTL = accessTokenTTL

	idTokenTTL, err := time.ParseDuration(getEnv("ID_TOKEN_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid ID_TOKEN_TTL: %w", err)
	}
	cfg.IDTokenTTL = idTokenTTL

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
