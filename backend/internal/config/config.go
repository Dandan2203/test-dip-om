package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort         string
	DatabaseURL     string
	JWTSecret       string
	AIServiceURL    string
	AIInternalToken string
	EncryptionKey   string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:         getEnv("APP_PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AIServiceURL:    getEnv("AI_SERVICE_URL", "http://localhost:8000"),
		AIInternalToken: os.Getenv("AI_INTERNAL_TOKEN"),
		EncryptionKey:   os.Getenv("ENCRYPTION_KEY"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: не задано обов'язкову змінну DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("config: не задано обов'язкову змінну JWT_SECRET")
	}
	if cfg.AIInternalToken == "" {
		return nil, fmt.Errorf("config: не задано обов'язкову змінну AI_INTERNAL_TOKEN")
	}
	if cfg.EncryptionKey == "" {
		return nil, fmt.Errorf("config: не задано обов'язкову змінну ENCRYPTION_KEY")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
