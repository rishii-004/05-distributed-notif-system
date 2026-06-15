package config

import "os"

type Config struct {
	Port    string
	Broker  string
	DB      string
}

func Load() *Config {
	return &Config{
		Port:   getEnv("PORT", "8080"),
		Broker: getEnv("BROKER_URL", "amqp://localhost:5672"),
		DB:     getEnv("DATABASE_URL", "postgres://localhost:5432/notif"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
