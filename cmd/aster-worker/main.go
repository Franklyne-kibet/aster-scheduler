package main

import (
	"context"
	"log"
	"os"
	"fmt"
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

	// Start worker in goroutine
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := w.Run(ctx); err != nil {
			logger.Error("Worker error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Worker is shutting down...")
	cancel() // Cancel context to stop worker

	logger.Info("Worker exited")
}