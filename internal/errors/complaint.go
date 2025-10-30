package errors

import (
	"fmt"
)

// ComplaintError represents a complaint-related error with structured information
type ComplaintError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Field   string                 `json:"field,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
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

// Predefined error codes
const (
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeDuplicate       = "DUPLICATE"
	ErrCodeStorageError    = "STORAGE_ERROR"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeInvalidFormat   = "INVALID_FORMAT"
)

// NewComplaintError creates a new complaint error
func NewComplaintError(code, message, field string) *ComplaintError {
	return &ComplaintError{
		Code:    code,
		Message: message,
		Field:   field,
		Details: make(map[string]interface{}),
	}
}

// NewValidationError creates a validation error
func NewValidationError(message, field string) *ComplaintError {
	return NewComplaintError(ErrCodeValidationFailed, message, field)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string) *ComplaintError {
	return NewComplaintError(ErrCodeNotFound, message, "")
}

// NewDuplicateError creates a duplicate error
func NewDuplicateError(message string) *ComplaintError {
	return NewComplaintError(ErrCodeDuplicate, message, "")
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
	return NewComplaintError(ErrCodeUnauthorized, message, "")
}

// NewInvalidFormatError creates an invalid format error
func NewInvalidFormatError(message, field string) *ComplaintError {
	return NewComplaintError(ErrCodeInvalidFormat, message, field)
}

// WithDetails adds details to an error
func (e *ComplaintError) WithDetails(key string, value interface{}) *ComplaintError {
	e.Details[key] = value
	return e
}

// WithCause adds a cause to an error
func (e *ComplaintError) WithCause(cause error) *ComplaintError {
	e.Cause = cause
	return e
}