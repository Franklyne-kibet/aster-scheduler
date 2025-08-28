package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

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
		h.writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Validate required fields
	if job.Name == "" {
		h.writeError(w, http.StatusBadRequest, "Job name is required")
		return
	}

	if job.CronExpr == "" {
		h.writeError(w, http.StatusBadRequest, "Cron expression is required")
		return
	}

	if job.Command == "" {
		h.writeError(w, http.StatusBadRequest, "Command is required")
		return
	}

	// Validate cron expression
	if err := h.cronParser.Validate(job.CronExpr); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid cron expression: "+err.Error())
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
		h.writeError(w, http.StatusInternalServerError, "Failed to create job")
		return
	}

	h.logger.Info("Job created",
		zap.String("job_id", job.ID.String()),
		zap.String("job_name", job.Name))

	// Return created job
	h.writeJSON(w, http.StatusCreated, job)
}

// GetJob handles GET /api/v1/jobs/{id}
func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid job ID format")
		return
	}

	// Get job from database
	job, err := h.jobStore.GetJob(r.Context(), id)
	if err != nil {
		if err.Error() == "job not found" {
			h.writeError(w, http.StatusNotFound, "Job not found")
		} else {
			h.logger.Error("Failed to get job", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "Failed to get job")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, job)
}

// ListJobs handles GET /api/v1/jobs
func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get jobs from database
	jobs, err := h.jobStore.ListJobs(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list jobs", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to list jobs")
		return
	}

	// Return jobs as JSON array
	h.writeJSON(w, http.StatusOK, jobs)
}

// UpdateJob handles PUT /api/v1/jobs/{id}
func (h *JobHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid job ID format")
		return
	}

	// Check if job exists
	existingJob, err := h.jobStore.GetJob(r.Context(), id)
	if err != nil {
		if err.Error() == "job not found" {
			h.writeError(w, http.StatusNotFound, "Job not found")
		} else {
			h.logger.Error("Failed to get job for update", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "Failed to get job")
		}
		return
	}

	// Parse updated job data
	var updatedJob types.Job
	if err := json.NewDecoder(r.Body).Decode(&updatedJob); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Preserve ID and timestamps
	updatedJob.ID = existingJob.ID
	updatedJob.CreatedAt = existingJob.CreatedAt

	// Validate cron expression if it was changed
	if updatedJob.CronExpr != "" && updatedJob.CronExpr != existingJob.CronExpr {
		if err := h.cronParser.Validate(updatedJob.CronExpr); err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid cron expression: "+err.Error())
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
		h.writeError(w, http.StatusInternalServerError, "Failed to update job")
		return
	}

	h.logger.Info("Job updated",
		zap.String("job_id", updatedJob.ID.String()),
		zap.String("job_name", updatedJob.Name))

	h.writeJSON(w, http.StatusOK, updatedJob)
}

// DeleteJob handles DELETE /api/v1/jobs/{id}
func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid job ID format")
		return
	}

	// Delete job from database
	if err := h.jobStore.DeleteJob(r.Context(), id); err != nil {
		if err.Error() == "job not found" {
			h.writeError(w, http.StatusNotFound, "Job not found")
		} else {
			h.logger.Error("Failed to delete job", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "Failed to delete job")
		}
		return
	}

	h.logger.Info("Job deleted", zap.String("job_id", id.String()))

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

// writeJSON writes a JSON response
func (h *JobHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// writeError writes a JSON error response
func (h *JobHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := map[string]interface{}{
		"error":   true,
		"message": message,
		"status":  status,
	}

	json.NewEncoder(w).Encode(errorResponse)
}
