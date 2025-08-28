package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/Franklyne-kibet/aster-scheduler/internal/db/store"
	"github.com/Franklyne-kibet/aster-scheduler/internal/types"
	"go.uber.org/zap"
)

// Scheduler is responsible for finding due jobs and scheduling them for execution
type Scheduler struct {
	jobStore		*store.JobStore
	cronParser	*CronParser
	logger			*zap.Logger

	// Configuration
	checkInterval time.Duration
}

// NewScheduler created a new scheduler instance
func NewScheduler(jobStore *store.JobStore, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		jobStore: 			jobStore,
		cronParser: 		NewCronParser(),
		logger: 				logger,
		checkInterval: 	30 * time.Second, // Check every 30 seconds by default
	}
}

// SetCheckInterval allows customizing how often we check for due jobs
func (s *Scheduler) SetCheckInterval(interval time.Duration) {
	s.checkInterval = interval
}

// Run starts the schedular in a loop (this is a blocking operation)
func (s *Scheduler) Run(ctx context.Context) error {
	s.logger.Info("Starting scheduler", zap.Duration("check_interval", s.checkInterval))

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	// Do an initial check immediately
	if err := s.checkAndScheduleDueJobs(ctx); err != nil {
		s.logger.Error("Error in initial job check", zap.Error(err))
	}

	for {
		select {
			case <-ctx.Done():
				s.logger.Info("Scheduler stopping due to context cancellation")
				return ctx.Err()

			case <-ticker.C:
				// Time to check for due jobs
				if err := s.checkAndScheduleDueJobs(ctx); err != nil {
					s.logger.Error("Error checking for due jobs", zap.Error(err))
					// Don't return error - keep running
				}
		}
	}
}

// checkAndScheduleDueJobs finds jobs that are due and schedules them
func (s *Scheduler) checkAndScheduleDueJobs(ctx context.Context) error {
	now := time.Now()

	s.logger.Debug("Checking for due jobs", zap.Time("current_time", now))

	// Get active jobs that are due
	dueJobs, err := s.jobStore.GetActiveJobsDue(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to get due jobs: %w", err)
	}

	s.logger.Debug("Found due jobs", zap.Int("count", len(dueJobs)))

	for _, job := range dueJobs {
		if err := s.scheduleJob(ctx, job, now); err != nil {
		s.logger.Error("Failed to schedule job", 
			zap.String("job_id", job.ID.String()),
			zap.String("job_name", job.Name),
			zap.Error(err))
			// Continue with other jobs even if one fails
			continue
		}

		s.logger.Info("Job scheduled successfully",
			zap.String("job_id", job.ID.String()),
			zap.String("job_name", job.Name))
	}

	return nil
}

// scheduleJob creates a run for a job and calculates the next run time
func (s *Scheduler) scheduleJob(ctx context.Context, job *types.Job, scheduledAt time.Time) error {
	// Create a new run for this job
	// run := &types.Run{
	// 	JobID:       job.ID,
	// 	Status:      types.RunStatusScheduled,
	// 	AttemptNum:  1, // This is the first attempt
	// 	ScheduledAt: scheduledAt,
	// }

	// Create the run in the database
	// TODO: Implement Runstore, we'll need it
	// Log what we would do
	s.logger.Info("Would create run for job",
		zap.String("job_id", job.ID.String()),
		zap.String("job_name", job.Name),
		zap.Time("scheduled_at", scheduledAt))

	// Calculate next run time
	nextRunAt, err := s.cronParser.ParserAndNext(job.CronExpr, scheduledAt)
	if err != nil {
		return fmt.Errorf("failed to calculate next run time for job %s: %w", job.Name, err)
	}

	// Update the job's next_run_at field
	if err := s.jobStore.UpdateJobNextRunAt(ctx, job.ID, &nextRunAt); err != nil {
		return fmt.Errorf("failed to update next run time for job %s: %w", job.Name, err)
	}

	s.logger.Debug("Updated next run time",
		zap.String("job_name", job.Name),
		zap.Time("next_run_at", nextRunAt))

	return nil
}

// GetJobNextRuns returns the next N run times for a job
func (s *Scheduler) GetJobNextRuns(cronExpr string, fromTime time.Time, n int) ([]time.Time, error) {
	return s.cronParser.GetNextNRuns(cronExpr, fromTime, n)
}