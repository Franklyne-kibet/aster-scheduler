package types

import (
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusActive   JobStatus = "active"
	JobStatusInactive JobStatus = "inactive"
	JobStatusArchived JobStatus = "archived"
)

// Job represents a scheduled task
// struct tags to convert to/from JSON and DB
type Job struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Description string            `json:"description" db:"description"`
	CronExpr    string            `json:"cron_expr" db:"cron_expr"`
	Command     string            `json:"command" db:"command"`
	Args        []string          `json:"args" db:"args"`
	Env         map[string]string `json:"env" db:"env"`
	Status      JobStatus         `json:"status" db:"status"`
	MaxRetries  int               `json:"max_retries" db:"max_retries"`
	Timeout     *time.Duration    `json:"timeout,omitempty" db:"timeout"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	NextRunAt   *time.Time        `json:"next_run_at,omitempty" db:"next_run_at"`
}

// RunStatus represents the state of a single execution
type RunStatus string

const (
	RunStatusScheduled RunStatus = "scheduled"
	RunStatusClaimed   RunStatus = "claimed"
	RunStatusRunning   RunStatus = "running"
	RunStatusSucceeded RunStatus = "succeeded"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCancelled RunStatus = "cancelled"
	RunStatusTimedOut  RunStatus = "timed_out"
)

// Run represents a single execution of a Job
type Run struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	JobID       uuid.UUID  `json:"job_id" db:"job_id"`
	Status      RunStatus  `json:"status" db:"status"`
	AttemptNum  int        `json:"attempt_num" db:"attempt_num"`
	ScheduledAt time.Time  `json:"scheduled_at" db:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at,omitempty" db:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty" db:"finished_at"`
	Output      string     `json:"output" db:"output"`
	ErrorMsg    *string    `json:"error_msg,omitempty" db:"error_msg"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
