package config

import (
	"os"
	"testing"
)

// TestLoad tests the Load function of the config package
func TestLoad(t *testing.T) {
	// Test with default values
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check default values
	if cfg.APIPort != 8080 {
		t.Errorf("Expected APIPort 8080, got %d", cfg.APIPort)
	}

	if cfg.WorkerPoolSize != 5 {
		t.Errorf("Expected WorkerPoolSize 5, got %d", cfg.WorkerPoolSize)
	}
}

// TestLoadWithEnvironmentVariables tests loading with custom env vars
func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("API_PORT", "9090")
	os.Setenv("WORKER_POOL_SIZE", "10")
	os.Setenv(("LOG_LEVEL"), "debug")

	// Clean up after test
	defer func() {
		os.Unsetenv("API_PORT")
		os.Unsetenv("WORKER_POOL_SIZE")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify our environment variables were used
	if cfg.APIPort != 9090 {
		t.Errorf("Expected APIPort 9090, got %d", cfg.APIPort)
	}

	if cfg.WorkerPoolSize != 10 {
		t.Errorf("Expected WorkerPoolSize 10, got %d", cfg.WorkerPoolSize)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel debug, got %s", cfg.LogLevel)
	}
}
