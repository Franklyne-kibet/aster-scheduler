package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// CronParser handles parsing cron expressions and calculating next run times
type CronParser struct {
	parser cron.Parser
}

// Create a new cron parser
func NewCronParser() *CronParser {
	// Create parser that supports seconds (optional), minutes, hours, day of month, month, day of week
	parser := cron.NewParser(
		cron.SecondOptional | // Allow optional seconds field
		cron.Minute | // Required minutes field
		cron.Hour | // Required hours field  
		cron.Dom | // Day of month
		cron.Month | // Month
		cron.Dow | // Day of week
		cron.Descriptor, // Allow @yearly, @monthly,
	)

	return &CronParser{
		parser: parser,
	}
}

// ParserAndNext parsers a cron expression and returns the next run time
func (cp *CronParser) ParserAndNext(cronExpr string, fromTime time.Time) (time.Time, error) {
	// Parse the cron expression
	schedule, err := cp.parser.Parse(cronExpr)

	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression '%s': %w", cronExpr, err)
	}

	// Calculate next run time from the given time
	nextRun := schedule.Next(fromTime)

	// Check if we got a valid next run
	if nextRun.IsZero() {
		return time.Time{}, fmt.Errorf("cron expression '%s' will never trigger", cronExpr)
	}

	return nextRun, nil
}

// Validate checks if a cron expression is valid without calculating next run
func (cp *CronParser) Validate(cronExpr string) error {
	_, err := cp.parser.Parse(cronExpr)
	if err != nil {
		return fmt.Errorf("invalid cron expression '%s': %w", cronExpr, err)
	}
	return nil
}

// GetNextRuns returns the next N run times for a cron expression
func (cp *CronParser) GetNextNRuns(cronExpr string, fromTime time.Time, n int) ([]time.Time, error) {
	schedule, err := cp.parser.Parse(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression '%s': %w", cronExpr, err)
	}

	var nextRuns []time.Time
	currentTime := fromTime

	for range n {
		nextRun := schedule.Next(currentTime)
		if nextRun.IsZero(){
			break // No more runs
		}
		nextRuns = append(nextRuns, nextRun)
		currentTime = nextRun
	}
	return nextRuns, nil
}