package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Database connection string
	DatabaseURL string

	// API server port
	APIPort int

	// Jobs running at once
	WorkerPoolSize int

	// Logging level (debug, info, warn, error)
	LogLevel string

	// Scheduler leader election TTL
	LeaderElectionTTL time.Duration
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Create a new Config with default values
	cfg := &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/aster?sslmode=disable"),
		APIPort:           getEnvInt("API_PORT", 8080),
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
