package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServiceName  string
	HTTPPort     string
	DatabaseURL  string
	KafkaBrokers string
	JWTSecret    string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServiceName:  getEnv("SERVICE_NAME", "unknown-service"),
		HTTPPort:     getEnv("HTTP_PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		KafkaBrokers: getEnv("KAFKA_BROKERS", ""),
		JWTSecret:    getEnv("JWT_SECRET", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port, err := strconv.Atoi(cfg.HTTPPort)

	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}

	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("HTTP_PORT must be between 0 and 65535")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
