package common

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// ErrorResponse represents a standard error response structure
type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// WriteJSON writes a JSON response with the given status code and data
func WriteJSON(w http.ResponseWriter, status int, data any, logger *zap.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		if logger != nil {
			logger.Error("Failed to encode JSON response", zap.Error(err))
		}
		// Fallback error response
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// WriteError writes a standardized JSON error response
func WriteError(w http.ResponseWriter, status int, message string, logger *zap.Logger) {
	errorResponse := ErrorResponse{
		Error:   true,
		Message: message,
		Status:  status,
	}

	WriteJSON(w, status, errorResponse, logger)
}

// WriteValidationError writes a validation error response with 400 status
func WriteValidationError(w http.ResponseWriter, message string, logger *zap.Logger) {
	WriteError(w, http.StatusBadRequest, message, logger)
}

// WriteNotFoundError writes a not found error response
func WriteNotFoundError(w http.ResponseWriter, resource string, logger *zap.Logger) {
	message := resource + " not found"
	WriteError(w, http.StatusNotFound, message, logger)
}

// WriteInternalError writes an internal server error response
func WriteInternalError(w http.ResponseWriter, logger *zap.Logger) {
	WriteError(w, http.StatusInternalServerError, "Internal server error", logger)
}

// WriteNoContent writes a 204 No Content response
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
