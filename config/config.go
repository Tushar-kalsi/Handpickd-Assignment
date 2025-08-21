package config

import "os"

// Config holds all configuration for the application
type Config struct {
    DB     DBConfig
    Server ServerConfig
    Kafka  KafkaConfig
}

// DBConfig holds database configuration
type DBConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

// ServerConfig holds server configuration
type ServerConfig struct {
    Port string
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
    Brokers  string
    Topic    string
    GroupID  string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
    return &Config{
        DB: DBConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", "postgres"),
            DBName:   getEnv("DB_NAME", "product_views"),
            SSLMode:  getEnv("DB_SSLMODE", "disable"),
        },
        Server: ServerConfig{
            Port: getEnv("SERVER_PORT", "8080"),
        },
        Kafka: KafkaConfig{
            Brokers:  getEnv("KAFKA_BROKERS", "localhost:9092"),
            Topic:    getEnv("KAFKA_TOPIC", "product-views"),
            GroupID:  getEnv("KAFKA_GROUP_ID", "product-views-group"),
        },
    }
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}
