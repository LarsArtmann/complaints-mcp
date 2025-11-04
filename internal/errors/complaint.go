package errors

import (
	"fmt"
)

// ComplaintError represents a complaint-related error with structured information
type ComplaintError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Field   string         `json:"field,omitempty"`
	Details map[string]any `json:"details,omitempty"`
	Cause   error          `json:"-"`
}

func (e *ComplaintError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *ComplaintError) Unwrap() error {
	return e.Cause
}

// Predefined error codes specific to complaints
const (
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeStorageError     = "STORAGE_ERROR"
	ErrCodeInvalidFormat    = "INVALID_FORMAT"
)

// NewComplaintError creates a new complaint error
func NewComplaintError(code, message, field string) *ComplaintError {
	return &ComplaintError{
		Code:    code,
		Message: message,
		Field:   field,
		Details: make(map[string]any),
	}
}

// NewComplaintValidationError creates a validation error for complaints
func NewComplaintValidationError(message, field string) *ComplaintError {
	return NewComplaintError(ErrCodeValidationFailed, message, field)
}

// NewStorageError creates a storage error with optional cause
func NewStorageError(message string, cause ...error) *ComplaintError {
	err := NewComplaintError(ErrCodeStorageError, message, "")
	if len(cause) > 0 {
		err.Cause = cause[0]
	}
	return err
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *ComplaintError {
	return NewComplaintError("UNAUTHORIZED", message, "")
}

// NewInvalidFormatError creates an invalid format error
func NewInvalidFormatError(message, field string) *ComplaintError {
	return NewComplaintError(ErrCodeInvalidFormat, message, field)
}

// WithDetails adds details to an error
func (e *ComplaintError) WithDetails(key string, value any) *ComplaintError {
	e.Details[key] = value
	return e
}

// WithCause adds a cause to an error
func (e *ComplaintError) WithCause(cause error) *ComplaintError {
	e.Cause = cause
	return e
}
