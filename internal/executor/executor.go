package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Franklyne-kibet/aster-scheduler/internal/types"
)

type ExecutionResult struct {
	Status types.RunStatus
	Output    string
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// Executor handles running jobs locally
type Executor struct {
	logger *zap.Logger
}

// NewExecutor creates a new job executor
func NewExecutor(logger *zap.Logger) *Executor {
	return &Executor{
		logger: logger,
	}
}

// Execute runs a job and returns the result
func (e *Executor) Execute(ctx context.Context, job *types.Job) *ExecutionResult {
	result := &ExecutionResult{
		StartTime: time.Now(),
		Status:    types.RunStatusRunning,
	}

	e.logger.Info("Starting job execution",
		zap.String("job_id", job.ID.String()),
		zap.String("job_name", job.Name),
		zap.String("command", job.Command),
		zap.Strings("args", job.Args))

	// Create command context with timeout if specified
	cmdCtx := ctx
	if job.Timeout != nil {
		var cancel context.CancelFunc
		cmdCtx, cancel = context.WithTimeout(ctx, *job.Timeout)
		defer cancel()
	}

	// Create the command
	cmd := exec.CommandContext(cmdCtx, job.Command, job.Args...)

	// Set environment variables
	if len(job.Env) > 0 {
		cmd.Env = make([]string, 0, len(job.Env))
		for key, value := range job.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Run the command and capture output
	output, err := cmd.CombinedOutput()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = string(output)

	// Determine the result status
	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			result.Status = types.RunStatusTimedOut
			result.Error = fmt.Errorf("job timed out after %s", job.Timeout)
		} else if cmdCtx.Err() == context.Canceled {
			result.Status = types.RunStatusCancelled
			result.Error = fmt.Errorf("job was cancelled")
		} else {
			result.Status = types.RunStatusFailed
			result.Error = fmt.Errorf("command failed: %w", err)
		}
	} else {
		result.Status = types.RunStatusSucceeded
	}

	e.logger.Info("Job execution completed",
		zap.String("job_id", job.ID.String()),
		zap.String("job_name", job.Name),
		zap.String("status", string(result.Status)),
		zap.Duration("duration", result.Duration),
		zap.String("output_preview", e.truncateOutput(result.Output, 200)))

	return result
}

// truncateOutput limits output length for logging
func (e *Executor) truncateOutput(output string, maxLen int) string {
	// Remove leading/trailing whitespace and newlines
	output = strings.TrimSpace(output)

	if len(output) <= maxLen {
		return output
	}

	return output[:maxLen] + "... (truncated)"
}

// ValidateCommand checks if a command is likely to work before executing it
func (e *Executor) ValidateCommand(command string) error {
	// Check if command exists in PATH
	_, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("command '%s' not found in PATH: %w", command, err)
	}
	return nil
}
