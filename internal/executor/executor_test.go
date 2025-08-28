package executor

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/Franklyne-kibet/aster-scheduler/internal/types"
	"github.com/google/uuid"
)

func TestExecutor_Execute_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	executor := NewExecutor(logger)

	job := &types.Job{
		ID:      uuid.New(),
		Name:    "test_echo",
		Command: "echo",
		Args:    []string{"hello", "world"},
		Env:     map[string]string{"TEST": "value"},
	}

	ctx := context.Background()
	result := executor.Execute(ctx, job)

	// Check basic result properties
	if result.Status != types.RunStatusSucceeded {
		t.Errorf("Expected status %s, got %s", types.RunStatusSucceeded, result.Status)
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}

	// Check output contains expected text
	expectedOutput := "hello world"
	if !containsIgnoreWhitespace(result.Output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got: '%s'", expectedOutput, result.Output)
	}

	// Check timing
	if result.Duration <= 0 {
		t.Errorf("Expected positive duration, got: %s", result.Duration)
	}

	if result.StartTime.IsZero() || result.EndTime.IsZero() {
		t.Error("Expected start and end times to be set")
	}
}

func TestExecutor_Execute_CommandNotFound(t *testing.T) {
	logger := zaptest.NewLogger(t)
	executor := NewExecutor(logger)

	job := &types.Job{
		ID:      uuid.New(),
		Name:    "test_nonexistent",
		Command: "nonexistent_command_xyz",
		Args:    []string{},
	}

	ctx := context.Background()
	result := executor.Execute(ctx, job)

	// Should fail
	if result.Status != types.RunStatusFailed {
		t.Errorf("Expected status %s, got %s", types.RunStatusFailed, result.Status)
	}

	if result.Error == nil {
		t.Error("Expected error for nonexistent command")
	}
}

func TestExecutor_Execute_Timeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	executor := NewExecutor(logger)

	// Job that sleeps for 2 seconds but times out after 100ms
	timeout := 100 * time.Millisecond
	job := &types.Job{
		ID:      uuid.New(),
		Name:    "test_timeout",
		Command: "sleep",
		Args:    []string{"2"}, // Sleep for 2 seconds
		Timeout: &timeout,
	}

	ctx := context.Background()
	start := time.Now()
	result := executor.Execute(ctx, job)
	elapsed := time.Since(start)

	// Should timeout
	if result.Status != types.RunStatusTimedOut {
		t.Errorf("Expected status %s, got %s", types.RunStatusTimedOut, result.Status)
	}

	if result.Error == nil {
		t.Error("Expected timeout error")
	}

	// Should complete quickly (near the timeout, not the full 2 seconds)
	if elapsed > timeout*2 {
		t.Errorf("Expected execution to complete near timeout (%s), took %s", timeout, elapsed)
	}
}

func TestExecutor_Execute_WithEnvironment(t *testing.T) {
	logger := zaptest.NewLogger(t)
	executor := NewExecutor(logger)

	job := &types.Job{
		ID:      uuid.New(),
		Name:    "test_env",
		Command: "sh",
		Args:    []string{"-c", "echo $TEST_VAR"},
		Env:     map[string]string{"TEST_VAR": "test_value"},
	}

	ctx := context.Background()
	result := executor.Execute(ctx, job)

	if result.Status != types.RunStatusSucceeded {
		t.Errorf("Expected status %s, got %s", types.RunStatusSucceeded, result.Status)
	}

	if !containsIgnoreWhitespace(result.Output, "test_value") {
		t.Errorf("Expected output to contain 'test_value', got: '%s'", result.Output)
	}
}

func TestExecutor_Execute_Cancellation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	executor := NewExecutor(logger)

	job := &types.Job{
		ID:      uuid.New(),
		Name:    "test_cancellation",
		Command: "sleep",
		Args:    []string{"10"}, // Sleep for 10 seconds
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result := executor.Execute(ctx, job)
	elapsed := time.Since(start)

	// Should be cancelled
	if result.Status != types.RunStatusCancelled {
		t.Errorf("Expected status %s, got %s", types.RunStatusCancelled, result.Status)
	}

	if result.Error == nil {
		t.Error("Expected cancellation error")
	}

	// Should complete quickly
	if elapsed > time.Second {
		t.Errorf("Expected execution to complete quickly, took %s", elapsed)
	}
}

func TestExecutor_Execute_CommandFailure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	executor := NewExecutor(logger)

	job := &types.Job{
		ID:      uuid.New(),
		Name:    "test_failure",
		Command: "sh",
		Args:    []string{"-c", "exit 1"}, // Command that exits with error code 1
	}

	ctx := context.Background()
	result := executor.Execute(ctx, job)

	// Should fail
	if result.Status != types.RunStatusFailed {
		t.Errorf("Expected status %s, got %s", types.RunStatusFailed, result.Status)
	}

	if result.Error == nil {
		t.Error("Expected error for command that exits with non-zero code")
	}

	// Should still have timing information
	if result.Duration <= 0 {
		t.Errorf("Expected positive duration even for failed command, got: %s", result.Duration)
	}
}

func TestExecutor_ValidateCommand(t *testing.T) {
	logger := zaptest.NewLogger(t)
	executor := NewExecutor(logger)

	// Test with command that should exist
	if err := executor.ValidateCommand("echo"); err != nil {
		t.Errorf("Expected 'echo' command to be valid, got error: %v", err)
	}

	// Test with command that shouldn't exist
	if err := executor.ValidateCommand("nonexistent_command_xyz"); err == nil {
		t.Error("Expected error for nonexistent command, got nil")
	}
}

func TestExecutor_TruncateOutput(t *testing.T) {
	logger := zaptest.NewLogger(t)
	executor := NewExecutor(logger)

	tests := []struct {
		name     string
		output   string
		maxLen   int
		expected string
	}{
		{
			name:     "short output",
			output:   "hello",
			maxLen:   100,
			expected: "hello",
		},
		{
			name:     "exact length",
			output:   "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "needs truncation",
			output:   "this is a very long output that should be truncated",
			maxLen:   10,
			expected: "this is a ... (truncated)",
		},
		{
			name:     "with whitespace",
			output:   "\n  hello world  \n",
			maxLen:   100,
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.truncateOutput(tt.output, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestExecutor_Execute_LargeOutput(t *testing.T) {
	logger := zaptest.NewLogger(t)
	executor := NewExecutor(logger)

	// Generate large output
	job := &types.Job{
		ID:      uuid.New(),
		Name:    "test_large_output",
		Command: "sh",
		Args:    []string{"-c", "for i in $(seq 1 1000); do echo \"Line $i with some text\"; done"},
	}

	ctx := context.Background()
	result := executor.Execute(ctx, job)

	if result.Status != types.RunStatusSucceeded {
		t.Errorf("Expected status %s, got %s", types.RunStatusSucceeded, result.Status)
	}

	// Output should be captured even if large
	if len(result.Output) == 0 {
		t.Error("Expected output to be captured")
	}

	// Should contain expected content
	if !strings.Contains(result.Output, "Line 1") || !strings.Contains(result.Output, "Line 1000") {
		t.Error("Expected output to contain first and last lines")
	}
}

// Helper function for comparing output with whitespace differences
func containsIgnoreWhitespace(output, expected string) bool {
	// Simple contains check, ignoring exact whitespace
	return len(output) > 0 && len(expected) > 0 &&
		(output == expected ||
			strings.Contains(strings.TrimSpace(output), strings.TrimSpace(expected)))
}
