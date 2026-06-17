// Package config loads application settings from environment variables.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the notification system.
type Config struct {
	Port    string // HTTP server listen port (e.g. "8080")
	Broker  string // RabbitMQ connection URL
	Redis   string // Redis server address (host:port)
	DB      string // PostgreSQL connection URL (Phase 4)
}

// Load reads .env (if present), then reads environment variables with defaults.
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:   getEnv("PORT", "8080"),
		Broker: getEnv("BROKER_URL", "amqp://guest:guest@localhost:5672"),
		Redis:  getEnv("REDIS_URL", "localhost:6379"),
		DB:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/notif"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
