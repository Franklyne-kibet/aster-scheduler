package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db"
	"github.com/Franklyne-kibet/aster-scheduler/internal/types"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *JobStore {
	t.Helper()

	// Connect to test database
	databaseURL := "postgres://postgres:password@localhost:5432/aster?sslmode=disable"

	ctx,  cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.NewConnection(ctx, databaseURL)
	if err != nil {
		t.Skipf("Skipping test, could not connect to database: %v", err)
		return nil
	}

	// Clean up any existing test data
		_, err = database.Pool().Exec(ctx, "DELETE FROM jobs WHERE name LIKE 'test_%'")
	if err != nil {
		t.Fatalf("Failed to clean up test data: %v", err)
	}
	
	return NewJobStore(database.Pool())
}

func TestJobStore_CreateJob(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return // Test was skipped
	}

	ctx := context.Background()

	// Create a test job
	job := &types.Job{
		ID:          uuid.New(),
		Name:        "test_hello_world",
		Description: "A simple test job",
		CronExpr:    "0 */5 * * *", // Every 5 minutes
		Command:     "echo",
		Args:        []string{"hello", "world"},
		Env:         map[string]string{"ENV": "test"},
		Status:      types.JobStatusActive,
		MaxRetries:  3,
		Timeout:     nil, // No timeout
	}

	// Test creating the job
	err := store.CreateJob(ctx, job)
	if err != nil {
		t.Fatalf("Expected no error creating job, got: %v", err)
	}

	// Verify it was created by fetching it back
	retrieved, err := store.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("Expected no error getting job, got: %v", err)
	}

		// Check that values match
	if retrieved.Name != job.Name {
		t.Errorf("Expected name %s, got %s", job.Name, retrieved.Name)
	}

	if retrieved.Command != job.Command {
		t.Errorf("Expected command %s, got %s", job.Command, retrieved.Command)
	}

	// Check that arrays/maps were stored correctly
	if len(retrieved.Args) != 2 || retrieved.Args[0] != "hello" || retrieved.Args[1] != "world" {
		t.Errorf("Expected args [hello, world], got %v", retrieved.Args)
	}

	if retrieved.Env["ENV"] != "test" {
		t.Errorf("Expected ENV=test, got %v", retrieved.Env)
	}
}

func TestJobStore_GetJob_NotFound(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}

	ctx := context.Background()

	// Try to get a job that doesn't exist
	randomID := uuid.New()
	_, err := store.GetJob(ctx, randomID)

	// We should get an error
	if err == nil {
		t.Error("Expected error for non-existent job, got nil")
	}

	// Check that error message contains "not found"
	if err.Error() != "job not found" {
		t.Errorf("Expected 'job not found' error, got: %v", err)
	}
}

func TestJobStore_ListJobs(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}

	ctx := context.Background()

	// Create multiple test jobs
	jobs := []*types.Job{
		{
			ID:       uuid.New(),
			Name:     "test_job_1",
			CronExpr: "0 0 * * *",
			Command:  "echo",
			Args:     []string{"job1"},
			Env:      map[string]string{},
			Status:   types.JobStatusActive,
		},
		{
			ID:       uuid.New(), 
			Name:     "test_job_2",
			CronExpr: "0 1 * * *",
			Command:  "echo",
			Args:     []string{"job2"},
			Env:      map[string]string{},
			Status:   types.JobStatusActive,
		},
	}

	// Create all jobs
	for _, job := range jobs {
		if err := store.CreateJob(ctx, job); err != nil {
			t.Fatalf("Failed to create job %s: %v", job.Name, err)
		}
	}

	// List jobs
	retrieved, err := store.ListJobs(ctx, 10, 0) // Limit 10, offset 0
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	// We should get at least our 2 test jobs
	if len(retrieved) < 2 {
		t.Errorf("Expected at least 2 jobs, got %d", len(retrieved))
	}

	// Check that our jobs are in the list
	foundCount := 0
	for _, job := range retrieved {
		if job.Name == "test_job_1" || job.Name == "test_job_2" {
			foundCount++
		}
	}

	if foundCount != 2 {
		t.Errorf("Expected to find 2 test jobs, found %d", foundCount)
	}
}

func TestJobStore_UpdateJob(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}

	ctx := context.Background()

	// Create a job
	job := &types.Job{
		ID:       uuid.New(),
		Name:     "test_update_job",
		CronExpr: "0 0 * * *",
		Command:  "echo",
		Args:     []string{"original"},
		Env:      map[string]string{"VERSION": "1"},
		Status:   types.JobStatusActive,
	}

	if err := store.CreateJob(ctx, job); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Update the job
	job.Command = "curl"
	job.Args = []string{"https://example.com"}
	job.Env["VERSION"] = "2"
	job.Status = types.JobStatusInactive

	if err := store.UpdateJob(ctx, job); err != nil {
		t.Fatalf("Failed to update job: %v", err)
	}

	// Fetch and verify updates
	updated, err := store.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("Failed to get updated job: %v", err)
	}

	if updated.Command != "curl" {
		t.Errorf("Expected command 'curl', got %s", updated.Command)
	}

	if len(updated.Args) != 1 || updated.Args[0] != "https://example.com" {
		t.Errorf("Expected args [https://example.com], got %v", updated.Args)
	}

	if updated.Env["VERSION"] != "2" {
		t.Errorf("Expected VERSION=2, got %s", updated.Env["VERSION"])
	}

	if updated.Status != types.JobStatusInactive {
		t.Errorf("Expected status inactive, got %s", updated.Status)
	}
}

func TestJobStore_DeleteJob(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}

	ctx := context.Background()

	// Create a job
	job := &types.Job{
		ID:       uuid.New(),
		Name:     "test_delete_job",
		CronExpr: "0 0 * * *",
		Command:  "echo",
		Args:     []string{"delete", "me"},
		Env:      map[string]string{"TEST": "delete"},
		Status:   types.JobStatusActive,
	}

	// Create the job first
	if err := store.CreateJob(ctx, job); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Verify it exists
	_, err := store.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("Job should exist before deletion: %v", err)
	}

	// Delete the job
	if err := store.DeleteJob(ctx, job.ID); err != nil {
		t.Fatalf("Failed to delete job: %v", err)
	}

	// Verify it no longer exists
	_, err = store.GetJob(ctx, job.ID)
	if err == nil {
		t.Error("Expected error when getting deleted job, got nil")
	}

	if err.Error() != "job not found" {
		t.Errorf("Expected 'job not found' error, got: %v", err)
	}
}

func TestJobStore_DeleteJob_NotFound(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}

	ctx := context.Background()

	// Try to delete a job that doesn't exist
	randomID := uuid.New()
	err := store.DeleteJob(ctx, randomID)

	// We should get an error
	if err == nil {
		t.Error("Expected error for deleting non-existent job, got nil")
	}

	// Check that error message contains "not found"
	if err.Error() != "job not found" {
		t.Errorf("Expected 'job not found' error, got: %v", err)
	}
}
