package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/Franklyne-kibet/aster-scheduler/internal/config"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db/store"
	"github.com/Franklyne-kibet/aster-scheduler/internal/executor"
	"github.com/Franklyne-kibet/aster-scheduler/internal/worker"
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

	// Create worker ID
	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("worker-%s-%d", hostname, os.Getpid())

	logger.Info("Starting Aster Worker",
		zap.String("worker_id", workerID),
		zap.String("log_level", cfg.LogLevel))

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := db.NewConnection(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	// Create stores and executor
	jobStore := store.NewJobStore(database.Pool())
	runStore := store.NewRunStore(database.Pool())
	exec := executor.NewExecutor(logger)

	// Create worker
	w := worker.NewWorker(workerID, jobStore, runStore, exec, logger)

	// Channel to capture worker errors
	workerErrCh := make(chan error, 1)

	// Start worker in goroutine
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := w.Run(ctx); err != nil {
			workerErrCh <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("Worker shutting down...", zap.String("signal", sig.String()))
	case err := <-workerErrCh:
		logger.Fatal("Worker failed", zap.Error(err))
	}

	// Cancel worker context for graceful shutdown
	cancel()

	// Give worker some time to cleanup
	time.Sleep(2 * time.Second)

	logger.Info("Worker exited")
}
