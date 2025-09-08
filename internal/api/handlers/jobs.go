package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/Franklyne-kibet/aster-scheduler/internal/common"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db/store"
	"github.com/Franklyne-kibet/aster-scheduler/internal/scheduler"
	"github.com/Franklyne-kibet/aster-scheduler/internal/types"
)

// JobHandler handles job-related HTTP requests
type JobHandler struct {
	jobStore   *store.JobStore
	cronParser *scheduler.CronParser
	logger     *zap.Logger
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobStore *store.JobStore, logger *zap.Logger) *JobHandler {
	return &JobHandler{
		jobStore:   jobStore,
		cronParser: scheduler.NewCronParser(),
		logger:     logger,
	}
}

// CreateJob handles POST /api/v1/jobs
func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var job types.Job

	// Parse JSON request body
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		common.WriteValidationError(w, "Invalid JSON: "+err.Error(), h.logger)
		return
	}

	// Validate required fields
	requiredFields := map[string]string{
		"name":      job.Name,
		"cron_expr": job.CronExpr,
		"command":   job.Command,
	}
	
	if err := common.ValidateRequiredFields(requiredFields); err != nil {
		common.WriteValidationError(w, err.Error(), h.logger)
		return
	}

	// Validate cron expression
	if err := h.cronParser.Validate(job.CronExpr); err != nil {
		common.WriteValidationError(w, "Invalid cron expression: "+err.Error(), h.logger)
		return
	}

	// Set defaults
	if job.Status == "" {
		job.Status = types.JobStatusActive
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}
	if job.Args == nil {
		job.Args = []string{}
	}
	if job.Env == nil {
		job.Env = make(map[string]string)
	}

	// Create job in database
	if err := h.jobStore.CreateJob(r.Context(), &job); err != nil {
		h.logger.Error("Failed to create job", zap.Error(err))
		common.WriteInternalError(w, h.logger)
		return
	}

	h.logger.Info("Job created",
		zap.String("job_id", job.ID.String()),
		zap.String("job_name", job.Name))

	// Return created job
	common.WriteJSON(w, http.StatusCreated, job, h.logger)
}

// GetJob handles GET /api/v1/jobs/{id}
func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := common.ParseUUID(idStr)
	if err != nil {
		common.WriteValidationError(w, "Invalid job ID format", h.logger)
		return
	}

	// Get job from database
	job, err := h.jobStore.GetJob(r.Context(), id)
	if err != nil {
		if err.Error() == "job not found" {
			common.WriteNotFoundError(w, "Job", h.logger)
		} else {
			h.logger.Error("Failed to get job", zap.Error(err))
			common.WriteInternalError(w, h.logger)
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, job, h.logger)
}

// ListJobs handles GET /api/v1/jobs
func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := common.ParsePositiveIntWithDefault(limitStr, 50)
	offset := common.ParseIntWithDefault(offsetStr, 0)

	// Get jobs from database
	jobs, err := h.jobStore.ListJobs(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list jobs", zap.Error(err))
		common.WriteInternalError(w, h.logger)
		return
	}

	// Return jobs as JSON array
	common.WriteJSON(w, http.StatusOK, jobs, h.logger)
}

// UpdateJob handles PUT /api/v1/jobs/{id}
func (h *JobHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := common.ParseUUID(idStr)
	if err != nil {
		common.WriteValidationError(w, "Invalid job ID format", h.logger)
		return
	}

	// Check if job exists
	existingJob, err := h.jobStore.GetJob(r.Context(), id)
	if err != nil {
		if err.Error() == "job not found" {
			common.WriteNotFoundError(w, "Job", h.logger)
		} else {
			h.logger.Error("Failed to get job for update", zap.Error(err))
			common.WriteInternalError(w, h.logger)
		}
		return
	}

	// Parse updated job data
	var updatedJob types.Job
	if err := json.NewDecoder(r.Body).Decode(&updatedJob); err != nil {
		common.WriteValidationError(w, "Invalid JSON: "+err.Error(), h.logger)
		return
	}

	// Preserve ID and timestamps
	updatedJob.ID = existingJob.ID
	updatedJob.CreatedAt = existingJob.CreatedAt

	// Validate cron expression if it was changed
	if updatedJob.CronExpr != "" && updatedJob.CronExpr != existingJob.CronExpr {
		if err := h.cronParser.Validate(updatedJob.CronExpr); err != nil {
			common.WriteValidationError(w, "Invalid cron expression: "+err.Error(), h.logger)
			return
		}
	}

	// Set defaults for required fields if empty
	if updatedJob.Name == "" {
		updatedJob.Name = existingJob.Name
	}
	if updatedJob.CronExpr == "" {
		updatedJob.CronExpr = existingJob.CronExpr
	}
	if updatedJob.Command == "" {
		updatedJob.Command = existingJob.Command
	}
	if updatedJob.Status == "" {
		updatedJob.Status = existingJob.Status
	}

	// Update job in database
	if err := h.jobStore.UpdateJob(r.Context(), &updatedJob); err != nil {
		h.logger.Error("Failed to update job", zap.Error(err))
		common.WriteInternalError(w, h.logger)
		return
	}

	h.logger.Info("Job updated",
		zap.String("job_id", updatedJob.ID.String()),
		zap.String("job_name", updatedJob.Name))

	common.WriteJSON(w, http.StatusOK, updatedJob, h.logger)
}

// DeleteJob handles DELETE /api/v1/jobs/{id}
func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := common.ParseUUID(idStr)
	if err != nil {
		common.WriteValidationError(w, "Invalid job ID format", h.logger)
		return
	}

	// Delete job from database
	if err := h.jobStore.DeleteJob(r.Context(), id); err != nil {
		if err.Error() == "job not found" {
			common.WriteNotFoundError(w, "Job", h.logger)
		} else {
			h.logger.Error("Failed to delete job", zap.Error(err))
			common.WriteInternalError(w, h.logger)
		}
		return
	}

	h.logger.Info("Job deleted", zap.String("job_id", id.String()))

	// Return 204 No Content
	common.WriteNoContent(w)
}
