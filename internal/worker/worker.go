package worker

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Franklyne-kibet/aster-scheduler/internal/db/store"
	"github.com/Franklyne-kibet/aster-scheduler/internal/executor"
	"github.com/Franklyne-kibet/aster-scheduler/internal/types"
)

// Worker polls for scheduled runs and executes them
type Worker struct {
	id        string // Unique worker identifier
	jobStore  *store.JobStore
	runStore  *store.RunStore
	executor  *executor.Executor
	logger    *zap.Logger

	// Configuration
	pollInterval time.Duration
	maxJobs      int // Maximum concurrent jobs
}

// NewWorker creates a new worker instance
func NewWorker(id string, jobStore *store.JobStore, runStore *store.RunStore, executor *executor.Executor, logger *zap.Logger) *Worker {
	return &Worker{
		id:           id,
		jobStore:     jobStore,
		runStore:     runStore,
		executor:     executor,
		logger:       logger,
		pollInterval: 5 * time.Second, // Poll every 5 seconds
		maxJobs:      1,               // Simple worker - one job at a time
	}
}

// SetPollInterval configures how often to check for new runs
func (w *Worker) SetPollInterval(interval time.Duration) {
	w.pollInterval = interval
}

// Run starts the worker (blocking operation)
func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info("Starting worker",
		zap.String("worker_id", w.id),
		zap.Duration("poll_interval", w.pollInterval),
		zap.Int("max_concurrent_jobs", w.maxJobs))

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Do an initial check immediately
	if err := w.checkAndExecuteRuns(ctx); err != nil {
		w.logger.Error("Error in initial run check", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker stopping due to context cancellation")
			return ctx.Err()

		case <-ticker.C:
			if err := w.checkAndExecuteRuns(ctx); err != nil {
				w.logger.Error("Error checking for runs", zap.Error(err))
				// Don't stop worker on errors
			}
		}
	}
}

// checkAndExecuteRuns looks for scheduled runs and executes them
func (w *Worker) checkAndExecuteRuns(ctx context.Context) error {
	// Get scheduled runs (limit to maxJobs for simplicity)
	runs, err := w.runStore.GetRunsByStatus(ctx, types.RunStatusScheduled, w.maxJobs)
	if err != nil {
		return fmt.Errorf("failed to get scheduled runs: %w", err)
	}

	if len(runs) == 0 {
		w.logger.Debug("No scheduled runs found")
		return nil
	}

	w.logger.Info("Found scheduled runs", zap.Int("count", len(runs)))

	for _, run := range runs {
		if err := w.executeRun(ctx, run); err != nil {
			w.logger.Error("Failed to execute run",
				zap.String("run_id", run.ID.String()),
				zap.String("job_id", run.JobID.String()),
				zap.Error(err))
			// Continue with other runs
		}
	}

	return nil
}

// executeRun executes a single run
func (w *Worker) executeRun(ctx context.Context, run *types.Run) error {
	// First, get the job details
	job, err := w.jobStore.GetJob(ctx, run.JobID)
	if err != nil {
		return fmt.Errorf("failed to get job for run: %w", err)
	}

	w.logger.Info("Executing run",
		zap.String("run_id", run.ID.String()),
		zap.String("job_id", job.ID.String()),
		zap.String("job_name", job.Name))

	// Mark run as started
	if err := w.runStore.MarkRunStarted(ctx, run.ID); err != nil {
		return fmt.Errorf("failed to mark run as started: %w", err)
	}

	// Execute the job
	result := w.executor.Execute(ctx, job)

	// Prepare error message for database
	var errorMsg *string
	if result.Error != nil {
		errStr := result.Error.Error()
		errorMsg = &errStr
	}

	// Mark run as finished with results
	if err := w.runStore.MarkRunFinished(ctx, run.ID, result.Status, result.Output, errorMsg); err != nil {
		w.logger.Error("Failed to mark run as finished",
			zap.String("run_id", run.ID.String()),
			zap.Error(err))
		// This is a problem but don't fail the execution
	}

	// Log execution summary
	w.logger.Info("Run execution completed",
		zap.String("run_id", run.ID.String()),
		zap.String("job_name", job.Name),
		zap.String("status", string(result.Status)),
		zap.Duration("duration", result.Duration))

	return nil
}