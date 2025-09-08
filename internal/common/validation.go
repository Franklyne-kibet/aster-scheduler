package common

import (
	"strconv"

	"github.com/google/uuid"
)

// ParseUUID parses a string into a UUID and returns an error if invalid
func ParseUUID(idStr string) (uuid.UUID, error) {
	return uuid.Parse(idStr)
}

// ParseIntWithDefault parses a string to int with a default value
func ParseIntWithDefault(str string, defaultValue int) int {
	if str == "" {
		return defaultValue
	}

	if parsed, err := strconv.Atoi(str); err == nil && parsed >= 0 {
		return parsed
	}

	return defaultValue
}

// ParsePositiveIntWithDefault parses a string to positive int with a default value
func ParsePositiveIntWithDefault(str string, defaultValue int) int {
	if str == "" {
		return defaultValue
	}

	if parsed, err := strconv.Atoi(str); err == nil && parsed > 0 {
		return parsed
	}

	return defaultValue
}

// ValidateRequiredFields checks if required fields are not empty
func ValidateRequiredFields(fields map[string]string) error {
	for fieldName, value := range fields {
		if value == "" {
			return NewValidationError(fieldName + " is required")
		}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}
