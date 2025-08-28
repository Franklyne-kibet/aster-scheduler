package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Franklyne-kibet/aster-scheduler/internal/types"
)

// RunStore handles all run-related database operations
type RunStore struct {
	pool *pgxpool.Pool
}

// NewRunStore creates a new run store
func NewRunStore(pool *pgxpool.Pool) *RunStore {
	return &RunStore{pool: pool}
}

// CreateRun inserts a new run into the database
func (s *RunStore) CreateRun(ctx context.Context, run *types.Run) error {
	query := `
		INSERT INTO runs (id, job_id, status, attempt_num, scheduled_at, started_at, finished_at, output, error_msg)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Generate UUID if not provided
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}

	_, err := s.pool.Exec(ctx, query,
		run.ID,
		run.JobID,
		run.Status,
		run.AttemptNum,
		run.ScheduledAt,
		run.StartedAt,
		run.FinishedAt,
		run.Output,
		run.ErrorMsg,
	)

	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	return nil
}

// GetRun retrieves a run by ID
func (s *RunStore) GetRun(ctx context.Context, id uuid.UUID) (*types.Run, error) {
	query := `
		SELECT id, job_id, status, attempt_num, scheduled_at, started_at, 
		    finished_at, output, error_msg, created_at, updated_at
		FROM runs 
		WHERE id = $1
	`

	var run types.Run
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&run.ID,
		&run.JobID,
		&run.Status,
		&run.AttemptNum,
		&run.ScheduledAt,
		&run.StartedAt,
		&run.FinishedAt,
		&run.Output,
		&run.ErrorMsg,
		&run.CreatedAt,
		&run.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("run not found")
		}
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	return &run, nil
}

// ListRuns returns runs, optionally filtered by job ID
func (s *RunStore) ListRuns(ctx context.Context, jobID *uuid.UUID, limit, offset int) ([]*types.Run, error) {
	var query string
	var args []interface{}

	if jobID != nil {
		query = `
			SELECT id, job_id, status, attempt_num, scheduled_at, started_at,
			    finished_at, output, error_msg, created_at, updated_at
			FROM runs
			WHERE job_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{*jobID, limit, offset}
	} else {
		query = `
			SELECT id, job_id, status, attempt_num, scheduled_at, started_at,
			    finished_at, output, error_msg, created_at, updated_at
			FROM runs
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query runs: %w", err)
	}
	defer rows.Close()

	var runs []*types.Run

	for rows.Next() {
		var run types.Run
		err := rows.Scan(
			&run.ID,
			&run.JobID,
			&run.Status,
			&run.AttemptNum,
			&run.ScheduledAt,
			&run.StartedAt,
			&run.FinishedAt,
			&run.Output,
			&run.ErrorMsg,
			&run.CreatedAt,
			&run.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}

		runs = append(runs, &run)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating rows: %w", rows.Err())
	}

	return runs, nil
}

// UpdateRunStatus updates a run's status and related fields
func (s *RunStore) UpdateRunStatus(ctx context.Context, runID uuid.UUID, status types.RunStatus, output string, errorMsg *string) error {
	query := `
		UPDATE runs 
		SET status = $2, output = $3, error_msg = $4, updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.pool.Exec(ctx, query, runID, status, output, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("run not found")
	}

	return nil
}

// MarkRunStarted marks a run as started
func (s *RunStore) MarkRunStarted(ctx context.Context, runID uuid.UUID) error {
	query := `
		UPDATE runs 
		SET status = $2, started_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.pool.Exec(ctx, query, runID, types.RunStatusRunning)
	if err != nil {
		return fmt.Errorf("failed to mark run as started: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("run not found")
	}

	return nil
}

// MarkRunFinished marks a run as finished with final status
func (s *RunStore) MarkRunFinished(ctx context.Context, runID uuid.UUID, status types.RunStatus, output string, errorMsg *string) error {
	query := `
		UPDATE runs 
		SET status = $2, finished_at = NOW(), output = $3, error_msg = $4, updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.pool.Exec(ctx, query, runID, status, output, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to mark run as finished: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("run not found")
	}

	return nil
}

// GetRunsByStatus returns runs with a specific status
func (s *RunStore) GetRunsByStatus(ctx context.Context, status types.RunStatus, limit int) ([]*types.Run, error) {
	query := `
		SELECT id, job_id, status, attempt_num, scheduled_at, started_at,
		    finished_at, output, error_msg, created_at, updated_at
		FROM runs
		WHERE status = $1
		ORDER BY scheduled_at ASC
		LIMIT $2
	`

	rows, err := s.pool.Query(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query runs by status: %w", err)
	}
	defer rows.Close()

	var runs []*types.Run

	for rows.Next() {
		var run types.Run
		err := rows.Scan(
			&run.ID,
			&run.JobID,
			&run.Status,
			&run.AttemptNum,
			&run.ScheduledAt,
			&run.StartedAt,
			&run.FinishedAt,
			&run.Output,
			&run.ErrorMsg,
			&run.CreatedAt,
			&run.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}

		runs = append(runs, &run)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating rows: %w", rows.Err())
	}

	return runs, nil
}