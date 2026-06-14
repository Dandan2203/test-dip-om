// Package config — завантаження та валідація конфігурації застосунку.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort       string
	DatabaseURL   string
	JWTSecret     string
	AIServiceURL  string
	EncryptionKey string
}

func Load() (*Config, error) {
	// .env необов'язковий — у проді змінні задає оточення.
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:       getEnv("APP_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		AIServiceURL:  getEnv("AI_SERVICE_URL", "http://localhost:8000"),
		EncryptionKey: getEnv("ENCRYPTION_KEY", "dev-encryption-key-change-in-production"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: не задано обов'язкову змінну DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("config: не задано обов'язкову змінну JWT_SECRET")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
