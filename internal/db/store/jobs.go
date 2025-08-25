package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Franklyne-kibet/aster-scheduler/internal/types"
)

// JobStore handles all job-related database operations
type JobStore struct { 
	pool *pgxpool.Pool
}

// NewJobStore creates a new JobStore
func NewJobStore(pool *pgxpool.Pool) *JobStore {
	return &JobStore{pool: pool}
}

// CreateJob inserts a new job into the database
func (s *JobStore) CreateJob(ctx context.Context, job *types.Job) error {
	// Convert Go slices/maps to JSON for storage
	argsJSON, err := json.Marshal(job.Args)
	if err != nil {
		return fmt.Errorf("failed to marshal args: %w", err)
	}

	envJSON, err := json.Marshal(job.Env)
	if err != nil {
		return fmt.Errorf("failed to marshal env: %w", err)
	}

	// SQL query to insert job
	query := `
		INSERT INTO jobs (id, name, description, cron_expr, command, args, env, status, max_retries, timeout)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	// Generate UUID if not provided
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}

	// Execute the query
	_, err = s.pool.Exec(ctx, query,
		job.ID,
		job.Name,
		job.Description,
		job.CronExpr,
		job.Command,
		argsJSON,
		envJSON,
		job.Status,
		job.MaxRetries,
		job.Timeout,
	)

	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}
	return nil
}

// GetJob retrieves a job by ID
func (s *JobStore) GetJob(ctx context.Context, id uuid.UUID) (*types.Job, error) {
	query := `
		SELECT id, name, description, cron_expr, command, args, env, 
		status, max_retries, timeout, created_at, updated_at, next_run_at
		FROM jobs 
		WHERE id = $1
	`

	var job types.Job
	var argsJSON, envJSON []byte

	// QueryRow returns at most one row
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.Name,
		&job.Description,
		&job.CronExpr,
		&job.Command,
		&argsJSON,    // Scan JSON as bytes
		&envJSON,    // Scan JSON as bytes
		&job.Status,
		&job.MaxRetries,
		&job.Timeout,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.NextRunAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// No job found with this ID
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Convert JSON back to Go types
	if err := json.Unmarshal((argsJSON), &job.Args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal args: %w", err)
	}

	if err := json.Unmarshal((envJSON), &job.Env); err != nil {
		return nil, fmt.Errorf("failed to unmarshal env: %w", err)
	}

	return &job, nil
}

// ListJobs returns a paginated list of jobs
func (s *JobStore) ListJobs(ctx context.Context, limit, offset int) ([]*types.Job, error) {
	query := `
		SELECT id, name, description, cron_expr, command, args, env,
		status, max_retries, timeout, created_at, updated_at, next_run_at
		FROM jobs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	// Query returns multiple rows
	rows, err := s.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close() // close rows when done

	var jobs []*types.Job

	// Iterate through all rows
	for rows.Next(){
		var job types.Job
		var argsJSON, envJSON []byte

		err := rows.Scan(
			&job.ID,
			&job.Name,
			&job.Description,
			&job.CronExpr,
			&job.Command,
			&argsJSON,
			&envJSON,
			&job.Status,
			&job.MaxRetries,
			&job.Timeout,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.NextRunAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		// Convert JSON to Go types
		if err := json.Unmarshal(argsJSON, &job.Args); err != nil {
			return nil, fmt.Errorf("failed to unmarshal args: %w", err)
		}

		if err := json.Unmarshal(envJSON, &job.Env); err != nil {
			return nil, fmt.Errorf("failed to unmarshal env: %w", err)
		}

		jobs = append(jobs, &job)
	}

	// Check for errors that occurred during iteration
	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating rows: %w", rows.Err())
	}

	return jobs, nil
}

// UpdateJob updates existing job
func (s *JobStore) UpdateJob(ctx context.Context, job *types.Job) error {
	argsJSON, err := json.Marshal(job.Args)
	if err != nil {
		return fmt.Errorf("failed to marshal args: %w", err)
	}

	envJSON, err := json.Marshal(job.Env)
	if err != nil {
		return fmt.Errorf("failed to marshal env: %w", err)
	}

		query := `
		UPDATE jobs 
		SET name = $2, description = $3, cron_expr = $4, command = $5,
		args = $6, env = $7, status = $8, max_retries = $9, 
		timeout = $10, updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.pool.Exec(ctx, query,
		job.ID,
		job.Name,
		job.Description,
		job.CronExpr,
		job.Command,
		argsJSON,
		envJSON,
		job.Status,
		job.MaxRetries,
		job.Timeout,
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Check if any row was actually updated
	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}


// DeleteJob removes a job from the database
func (s *JobStore) DeleteJob(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM jobs WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found")
	}
	return nil
}