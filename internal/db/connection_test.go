package db

import (
	"context"
	"testing"
	"time"
)

func TestNewConnection(t *testing.T) {
	// Skip this test if we don't have a database running
	// You can run: docker-compose up -d postgres
	// Then: DATABASE_URL=postgres://postgres:password@localhost:5432/aster?sslmode=disable go test ./internal/db

	databaseURL := "postgres://postgres:password@localhost:5432/aster?sslmode=disable"

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Cancel the context to avoid leaks

	// Try to connect
	db, err := NewConnection(ctx, databaseURL)
	if err != nil {
		t.Skipf("Skipping test, could not connect to database: %v", err)
		return
	}
	defer db.Close()

	// Test ping
	if err := db.Ping(ctx); err != nil {
		t.Errorf("Expected ping to succeed, got error: %v", err)
	}
}

func TestNewConnection_InvalidURL(t *testing.T) {
	ctx := context.Background()

	// Test with an invalid database URL
	_, err := NewConnection(ctx, "invalid-url")
	if err == nil {
		t.Error("Expected error for invalid database URL, got nil")
	}
}
