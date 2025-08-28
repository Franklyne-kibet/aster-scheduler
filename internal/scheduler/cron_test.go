package scheduler

import (
	"testing"
	"time"
)

func TestCronParser_ParserAndNext(t *testing.T) {
	parser := NewCronParser()

	// Test time: January 1, 2024, 12:00:00 UTC (Monday)
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		cronExpr string
		want     time.Time
		wantErr  bool
	}{
		{
			name:     "every minute",
			cronExpr: "* * * * *",
			want:     time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "daily at midnight",
			cronExpr: "0 0 * * *",
			want:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "every 5 minutes",
			cronExpr: "*/5 * * * *",
			want:     time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "hourly descriptor",
			cronExpr: "@hourly",
			want:     time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "invalid expression",
			cronExpr: "invalid",
			wantErr:  true,
		},
		{
			name:     "too many fields",
			cronExpr: "0 0 0 0 0 0 0",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParserAndNext(tt.cronExpr, testTime)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for cron expression '%s', got nil", tt.cronExpr)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !got.Equal(tt.want) {
				t.Errorf("ParseAndNext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCronParser_Validate(t *testing.T) {
	parser := NewCronParser()

	validExpressions := []string{
		"* * * * *",
		"0 0 * * *",
		"*/5 * * * *",
		"@hourly",
		"@daily",
		"30 2 * * 1", // Every Monday at 2:30 AM
	}

	for _, expr := range validExpressions {
		t.Run("valid_"+expr, func(t *testing.T) {
			if err := parser.Validate(expr); err != nil {
				t.Errorf("Expected '%s' to be valid, got error: %v", expr, err)
			}
		})
	}

	invalidExpressions := []string{
		"invalid",
		"60 * * * *", // Invalid minute (60)
		"* 25 * * *", // Invalid hour (25)
		"* * 32 * *", // Invalid day (32)
		"* * * 13 *", // Invalid month (13)
		"* * * * 8",  // Invalid day of week (8)
	}

	for _, expr := range invalidExpressions {
		t.Run("invalid_"+expr, func(t *testing.T) {
			if err := parser.Validate(expr); err == nil {
				t.Errorf("Expected '%s' to be invalid, but validation passed", expr)
			}
		})
	}
}

func TestCronParser_GetNextNRuns(t *testing.T) {
	parser := NewCronParser()
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Test "every 5 minutes" - get next 3 runs
	runs, err := parser.GetNextNRuns("*/5 * * * *", testTime, 3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []time.Time{
		time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 12, 10, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 12, 15, 0, 0, time.UTC),
	}

	if len(runs) != len(expected) {
		t.Fatalf("Expected %d runs, got %d", len(expected), len(runs))
	}

	for i, run := range runs {
		if !run.Equal(expected[i]) {
			t.Errorf("Run %d: expected %v, got %v", i, expected[i], run)
		}
	}
}
