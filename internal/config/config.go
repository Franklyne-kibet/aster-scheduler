// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Database connection string
	DatabaseURL string

	// API server port
	APIPort int

	// API server configuration
	AllowedOrigins []string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration

	// Jobs running at once
	WorkerPoolSize int

	// Logging level (debug, info, warn, error)
	LogLevel string

	// Scheduler leader election TTL
	LeaderElectionTTL time.Duration
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	user := getEnv("POSTGRES_USER", "postgres")
	pass := getEnv("POSTGRES_PASSWORD", "postgres")
	host := getEnv("POSTGRES_HOST", "postgres")
	port := getEnv("POSTGRES_PORT", "5432")
	db := getEnv("POSTGRES_DB", "aster")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, db)

	// Create a new Config with default values
	cfg := &Config{
		DatabaseURL:       dbURL,
		APIPort:           getEnvInt("API_PORT", 8080),
		AllowedOrigins:    getEnvStringSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		ReadTimeout:       getEnvDuration("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:      getEnvDuration("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:       getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
		WorkerPoolSize:    getEnvInt("WORKER_POOL_SIZE", 5),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LeaderElectionTTL: getEnvDuration("LEADER_ELECTION_TTL", 30*time.Second),
	}
	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as an integer
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// Convert to int
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
		// use default if conversion fails
	}
	return defaultValue
}

// getEnvDuration gets an environment variable as a time.Duration
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		// Convert to time.Duration
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvStringSlice gets an environment variable as a string slice (comma-separated)
func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
