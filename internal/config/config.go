package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL string
	KafkaBroker string
	ServerPort  string
	Environment string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/product_views?sslmode=disable"),
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
		ServerPort:  getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv gets an environment variable with a fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GetIntEnv gets an integer environment variable with a fallback
func GetIntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
