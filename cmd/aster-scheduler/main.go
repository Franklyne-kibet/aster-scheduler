package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/Franklyne-kibet/aster-scheduler/internal/config"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db/store"
	"github.com/Franklyne-kibet/aster-scheduler/internal/scheduler"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Aster Scheduler",
		zap.String("log_level", cfg.LogLevel))

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := db.NewConnection(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	// Create stores
	jobStore := store.NewJobStore(database.Pool())
	runStore := store.NewRunStore(database.Pool())

	// Create scheduler
	sched := scheduler.NewScheduler(jobStore, runStore, logger)

	// Channel to capture scheduler errors
	schedErrCh := make(chan error, 1)

	// Start scheduler in goroutine
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := sched.Run(ctx); err != nil {
			schedErrCh <- err
		}
	}()

	// Wait for interrupt signal or scheduler error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("Scheduler shutting down...", zap.String("signal", sig.String()))
	case err := <-schedErrCh:
		logger.Fatal("Scheduler failed", zap.Error(err))
	}

	// Cancel scheduler context for graceful shutdown
	cancel()

	// Give scheduler some time to cleanup
	time.Sleep(2 * time.Second)

	logger.Info("Scheduler exited")
}
