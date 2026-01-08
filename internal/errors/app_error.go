package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents different types of application errors.
type ErrorCode string

const (
	// Domain errors.
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeDuplicate    ErrorCode = "DUPLICATE_ERROR"
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"

	// Repository errors.
	ErrCodeRepository ErrorCode = "REPOSITORY_ERROR"
	ErrCodeStorage    ErrorCode = "STORAGE_ERROR"
	ErrCodeFileIO     ErrorCode = "FILE_IO_ERROR"
	ErrCodeDatabase   ErrorCode = "DATABASE_ERROR"

	// Service errors.
	ErrCodeService      ErrorCode = "SERVICE_ERROR"
	ErrCodeBusiness     ErrorCode = "BUSINESS_ERROR"
	ErrCodePermission   ErrorCode = "PERMISSION_ERROR"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED_ERROR"

	// System errors.
	ErrCodeInternal    ErrorCode = "INTERNAL_ERROR"
	ErrCodeExternal    ErrorCode = "EXTERNAL_ERROR"
	ErrCodeTimeout     ErrorCode = "TIMEOUT_ERROR"
	ErrCodeUnavailable ErrorCode = "UNAVAILABLE_ERROR"
	ErrCodeRateLimit   ErrorCode = "RATE_LIMIT_ERROR"
)

// AppError represents a structured application error.
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    any       `json:"details,omitempty"`
	Cause      error     `json:"-"`
	HTTPStatus int       `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %s (details: %v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error.
func NewAppError(code ErrorCode, message string) *AppError {
	httpStatus := errorCodeToHTTPStatus(code)
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// NewAppErrorWithCause creates a new application error with underlying cause.
func NewAppErrorWithCause(code ErrorCode, message string, cause error) *AppError {
	httpStatus := errorCodeToHTTPStatus(code)
	return &AppError{
		Code:       code,
		Message:    message,
		Cause:      cause,
		HTTPStatus: httpStatus,
	}
}

// NewAppErrorWithDetails creates a new application error with details.
func NewAppErrorWithDetails(code ErrorCode, message string, details any) *AppError {
	httpStatus := errorCodeToHTTPStatus(code)
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatus,
	}
}

// Wrap wraps an existing error with application context.
func Wrap(err error, code ErrorCode, message string) *AppError {
	httpStatus := errorCodeToHTTPStatus(code)
	return &AppError{
		Code:       code,
		Message:    message,
		Cause:      err,
		HTTPStatus: httpStatus,
	}
}

// WrapDetails wraps an existing error with application context and details.
func WrapDetails(err error, code ErrorCode, message string, details any) *AppError {
	httpStatus := errorCodeToHTTPStatus(code)
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		Cause:      err,
		HTTPStatus: httpStatus,
	}
}

// IsAppError checks if an error is an AppError.
func IsAppError(err error) (*AppError, bool) {
	appErr := &AppError{}
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// GetHTTPStatus returns HTTP status for any error.
func GetHTTPStatus(err error) int {
	if appErr, ok := IsAppError(err); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// errorCodeToHTTPStatus maps error codes to HTTP status codes.
func errorCodeToHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrCodeValidation, ErrCodeInvalidInput:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodePermission:
		return http.StatusForbidden
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeDuplicate:
		return http.StatusConflict
	case ErrCodeRateLimit:
		return http.StatusTooManyRequests
	case ErrCodeTimeout:
		return http.StatusRequestTimeout
	case ErrCodeUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeExternal:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

// Predefined error constructors

// NewValidationError creates a validation error.
func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidation, message)
}

// NewValidationErrorWithDetails creates a validation error with details.
func NewValidationErrorWithDetails(message string, details any) *AppError {
	return NewAppErrorWithDetails(ErrCodeValidation, message, details)
}

// NewNotFoundError creates a not found error.
func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, resource+" not found")
}

// NewDuplicateError creates a duplicate error.
func NewDuplicateError(resource string) *AppError {
	return NewAppError(ErrCodeDuplicate, resource+" already exists")
}

// NewRepositoryError creates a repository error.
func NewRepositoryError(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeRepository, message, cause)
}

// NewFileIOError creates a file I/O error.
func NewFileIOError(operation, path string, cause error) *AppError {
	message := fmt.Sprintf("failed to %s file: %s", operation, path)
	return NewAppErrorWithCause(ErrCodeFileIO, message, cause)
}

// NewServiceError creates a service error.
func NewServiceError(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeService, message, cause)
}

// NewInternalError creates an internal server error.
func NewInternalError(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeInternal, message, cause)
}

// NewTimeoutError creates a timeout error.
func NewTimeoutError(operation string) *AppError {
	return NewAppError(ErrCodeTimeout, fmt.Sprintf("operation %s timed out", operation))
}

// NewExternalError creates an external service error.
func NewExternalError(service string, cause error) *AppError {
	message := fmt.Sprintf("external service %s error", service)
	return NewAppErrorWithCause(ErrCodeExternal, message, cause)
}
