package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName         string
	HTTPPort            string
	GRPCPort            string
	GameServiceGRPCAddr string
	DatabaseURL         string
	KafkaBrokers        string
	JWTSecret           string
	OIDCIssuer          string
	OIDCAudience        string
	JWKSURL             string
	JWTPrivateKeyPath   string
	JWTKeyID            string
	AccessTokenTTL      time.Duration
	IDTokenTTL          time.Duration
}

func validatePort(name, value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("%s must be a number", name)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", name)
	}

	return nil
}

func validateHostPort(name, value string) error {
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return fmt.Errorf("%s must be in host:port format", name)
	}

	if host == "" {
		return fmt.Errorf("%s host must not be empty", name)
	}

	if err := validatePort(name, port); err != nil {
		return err
	}

	return nil
}

func Load() (*Config, error) {
	cfg := &Config{
		ServiceName:         getEnv("SERVICE_NAME", "unknown-service"),
		HTTPPort:            getEnv("HTTP_PORT", "8080"),
		GRPCPort:            getEnv("GRPC_PORT", "9090"),
		GameServiceGRPCAddr: getEnv("GAME_SERVICE_GRPC_ADDR", "localhost:9093"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		KafkaBrokers:        getEnv("KAFKA_BROKERS", ""),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		OIDCIssuer:          getEnv("OIDC_ISSUER", "http://localhost:8081"),
		OIDCAudience:        getEnv("OIDC_AUDIENCE", "online-checkers"),
		JWKSURL:             getEnv("JWKS_URL", "http://localhost:8081/.well-known/jwks.json"),
		JWTPrivateKeyPath:   getEnv("JWT_PRIVATE_KEY_PATH", ""),
		JWTKeyID:            getEnv("JWT_KEY_ID", "dev-key-1"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	httpPort := getEnv("HTTP_PORT", "8080")

	if err := validatePort("HTTP_PORT", httpPort); err != nil {
		return nil, err
	}

	grpcPort := getEnv("GRPC_PORT", "9090")

	if err := validatePort("GRPC_PORT", grpcPort); err != nil {
		return nil, err
	}

	gameServiceGRPCAddr := getEnv("GAME_SERVICE_GRPC_ADDR", "localhost:9093")
	if err := validateHostPort("GAME_SERVICE_GRPC_ADDR", gameServiceGRPCAddr); err != nil {
		return nil, err
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
