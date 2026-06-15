// Package config loads application settings from environment variables.
// It uses sensible defaults so the app can run without any configuration.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the notification system.
type Config struct {
	Port   string // HTTP server listen port (e.g. "8080")
	Broker string // RabbitMQ connection URL (e.g. "amqp://guest:guest@localhost:5672")
	DB     string // PostgreSQL connection URL (used in Phase 4)
}

// Load reads .env (if present), then reads environment variables with defaults.
func Load() *Config {
	// Try loading .env file — silently skip if not found
	_ = godotenv.Load()

	return &Config{
		Port:   getEnv("PORT", "8080"),
		Broker: getEnv("BROKER_URL", "amqp://guest:guest@localhost:5672"),
		DB:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/notif"),
	}
}

// getEnv returns the env var value, or the fallback if not set or empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
