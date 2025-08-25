package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgx connection pool
type DB struct {
	pool *pgxpool.Pool
}

func NewConnection(ctx context.Context, databaseURL string) (*DB, error) {
	// Parse db URL and create configuration
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = 10 // Maximum 10 connections in pool
	config.MinConns = 2  // Keep at least 2 connections open
	config.MaxConnIdleTime = time.Hour // Recycle connections every hour
	config.MaxConnLifetime = 30 * time.Hour // Close idle connections after 30 min

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Ping the database to verify connection
	if err := pool.Ping(ctx); err != nil {
		// If ping fails, close the pool and return error
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{pool: pool}, nil
}

// Pool returns the underlying connection pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.pool.Close()
}

// Ping tests if the database is reachable
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
