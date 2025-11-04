package errors

// Complaint-specific error codes extending base ErrorCode
const (
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeStorageError     ErrorCode = "STORAGE_ERROR"
	ErrCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"
)

// NewComplaintValidationError creates a validation error for complaint fields
func NewComplaintValidationError(message string, field string) *AppError {
	return NewValidationError(message)
}

// NewComplaintStorageError creates a storage error for complaint operations
func NewComplaintStorageError(message string) *AppError {
	return NewAppError(ErrCodeStorageError, message)
}

// NewComplaintFormatError creates a format error for complaint data
func NewComplaintFormatError(message string) *AppError {
	return NewAppError(ErrCodeInvalidFormat, message)
}
