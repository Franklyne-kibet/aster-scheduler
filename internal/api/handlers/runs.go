package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/Franklyne-kibet/aster-scheduler/internal/db/store"
)

// RunHandler handles run-related HTTP requests
type RunHandler struct {
	runStore	*store.RunStore
	logger 		*zap.Logger
}

// NewRunHandler creates a new run handler
func NewRunHandler(runStore *store.RunStore, logger *zap.Logger) *RunHandler {
	return &RunHandler{
		runStore: runStore,
		logger: 	logger,
	}
}


// GetRun handles GET /api/v1/runs/{id}
func (h *RunHandler) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid run ID format")
		return
	}

	run, err := h.runStore.GetRun(r.Context(), id)
	if err != nil {
		if err.Error() == "run not found" {
			h.writeError(w, http.StatusNotFound, "Run not found")
		} else {
			h.logger.Error("Failed to get run", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "Failed to get run")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, run)
}

// ListRuns handles GET /api/v1/runs
func (h *RunHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	jobIDStr := r.URL.Query().Get("job_id")

	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var jobID *uuid.UUID
	if jobIDStr != "" {
		if parsed, err := uuid.Parse(jobIDStr); err == nil {
			jobID = &parsed
		} else {
			h.writeError(w, http.StatusBadRequest, "Invalid job_id format")
			return
		}
	}

	runs, err := h.runStore.ListRuns(r.Context(), jobID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list runs", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to list runs")
		return
	}

	h.writeJSON(w, http.StatusOK, runs)
}

// Helper methods (same as JobHandler)
func (h *RunHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *RunHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	errorResponse := map[string]interface{}{
		"error":   true,
		"message": message,
		"status":  status,
	}
	
	json.NewEncoder(w).Encode(errorResponse)
}