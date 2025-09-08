package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/Franklyne-kibet/aster-scheduler/internal/common"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db/store"
)

// RunHandler handles run-related HTTP requests
type RunHandler struct {
	runStore *store.RunStore
	logger   *zap.Logger
}

// NewRunHandler creates a new run handler
func NewRunHandler(runStore *store.RunStore, logger *zap.Logger) *RunHandler {
	return &RunHandler{
		runStore: runStore,
		logger:   logger,
	}
}

// GetRun handles GET /api/v1/runs/{id}
func (h *RunHandler) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := common.ParseUUID(idStr)
	if err != nil {
		common.WriteValidationError(w, "Invalid run ID format", h.logger)
		return
	}

	run, err := h.runStore.GetRun(r.Context(), id)
	if err != nil {
		if err.Error() == "run not found" {
			common.WriteNotFoundError(w, "Run", h.logger)
		} else {
			h.logger.Error("Failed to get run", zap.Error(err))
			common.WriteInternalError(w, h.logger)
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, run, h.logger)
}

// ListRuns handles GET /api/v1/runs
func (h *RunHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	jobIDStr := r.URL.Query().Get("job_id")

	limit := common.ParsePositiveIntWithDefault(limitStr, 50)
	offset := common.ParseIntWithDefault(offsetStr, 0)

	var jobID *uuid.UUID
	if jobIDStr != "" {
		if parsed, err := common.ParseUUID(jobIDStr); err == nil {
			jobID = &parsed
		} else {
			common.WriteValidationError(w, "Invalid job_id format", h.logger)
			return
		}
	}

	runs, err := h.runStore.ListRuns(r.Context(), jobID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list runs", zap.Error(err))
		common.WriteInternalError(w, h.logger)
		return
	}

	common.WriteJSON(w, http.StatusOK, runs, h.logger)
}
