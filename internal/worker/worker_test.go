package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/Franklyne-kibet/aster-scheduler/internal/db"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db/store"
	"github.com/Franklyne-kibet/aster-scheduler/internal/executor"
	"github.com/Franklyne-kibet/aster-scheduler/internal/types"
	"github.com/google/uuid"
)

func setupWorkerTest(t *testing.T) (*Worker, *store.JobStore, *store.RunStore) {
	t.Helper()

	// Connect to test database
	databaseURL := "postgres://postgres:password@localhost:5432/aster?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.NewConnection(ctx, databaseURL)
	if err != nil {
		t.Skipf("Skipping test, could not connect to database: %v", err)
		return nil, nil, nil
	}

	// Clean up test data
	database.Pool().Exec(ctx, "DELETE FROM runs WHERE 1=1")
	database.Pool().Exec(ctx, "DELETE FROM jobs WHERE name LIKE 'test_worker_%'")

	logger := zaptest.NewLogger(t)
	jobStore := store.NewJobStore(database.Pool())
	runStore := store.NewRunStore(database.Pool())
	executor := executor.NewExecutor(logger)
	worker := NewWorker("test-worker-1", jobStore, runStore, executor, logger)

	// Speed up polling for tests
	worker.SetPollInterval(100 * time.Millisecond)

	return worker, jobStore, runStore
}

func TestWorker_ExecuteRun(t *testing.T) {
	worker, jobStore, runStore := setupWorkerTest(t)
	if worker == nil {
		return // Test was skipped
	}

	ctx := context.Background()

	// Create a test job
	job := &types.Job{
		ID:      uuid.New(),
		Name:    "test_worker_echo",
		Command: "echo",
		Args:    []string{"worker", "test"},
		Status:  types.JobStatusActive,
	}

	if err := jobStore.CreateJob(ctx, job); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Create a scheduled run
	run := &types.Run{
		ID:          uuid.New(),
		JobID:       job.ID,
		Status:      types.RunStatusScheduled,
		AttemptNum:  1,
		ScheduledAt: time.Now(),
	}

	if err := runStore.CreateRun(ctx, run); err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}

	// Execute the run
	if err := worker.executeRun(ctx, run); err != nil {
		t.Fatalf("Failed to execute run: %v", err)
	}

	// Check that run was updated
	updatedRun, err := runStore.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("Failed to get updated run: %v", err)
	}

	// Should be completed successfully
	if updatedRun.Status != types.RunStatusSucceeded {
		t.Errorf("Expected status %s, got %s", types.RunStatusSucceeded, updatedRun.Status)
	}

	// Should have output
	if len(updatedRun.Output) == 0 {
		t.Error("Expected non-empty output")
	}

	// Should have timing information
	if updatedRun.StartedAt == nil {
		t.Error("Expected StartedAt to be set")
	}

	if updatedRun.FinishedAt == nil {
		t.Error("Expected FinishedAt to be set")
	}
}

func TestWorker_CheckAndExecuteRuns(t *testing.T) {
	worker, jobStore, runStore := setupWorkerTest(t)
	if worker == nil {
		return
	}

	ctx := context.Background()

	// Create multiple test jobs and runs
	for i := 0; i < 3; i++ {
		job := &types.Job{
			ID:      uuid.New(),
			Name:    fmt.Sprintf("test_worker_job_%d", i),
			Command: "echo",
			Args:    []string{fmt.Sprintf("job_%d", i)},
			Status:  types.JobStatusActive,
		}

		if err := jobStore.CreateJob(ctx, job); err != nil {
			t.Fatalf("Failed to create job %d: %v", i, err)
		}

		run := &types.Run{
			ID:          uuid.New(),
			JobID:       job.ID,
			Status:      types.RunStatusScheduled,
			AttemptNum:  1,
			ScheduledAt: time.Now(),
		}

		if err := runStore.CreateRun(ctx, run); err != nil {
			t.Fatalf("Failed to create run %d: %v", i, err)
		}
	}

	// Execute all scheduled runs
	if err := worker.checkAndExecuteRuns(ctx); err != nil {
		t.Fatalf("Failed to check and execute runs: %v", err)
	}

	// Verify all runs completed
	runs, err := runStore.ListRuns(ctx, nil, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list runs: %v", err)
	}

	completedCount := 0
	for _, run := range runs {
		if run.Status == types.RunStatusSucceeded {
			completedCount++
		}
	}

	// We created 3 runs, all should be completed
	// (Note: might be more runs from other tests)
	if completedCount < 3 {
		t.Errorf("Expected at least 3 completed runs, got %d", completedCount)
	}
}
