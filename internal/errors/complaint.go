package errors

// Complaint represents a complaint filed by an AI agent
type Complaint struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

// Error interface for complaint-related errors
type Error interface {
	error() string
	Error() string
}

// complaintError implements Error interface
type complaintError struct {
	code    string
	message string
	field   string
}

func (e *complaintError) error() string {
	return e.code
}

func (e *complaintError) Error() string {
	return e.message
}

func (e *complaintError) ErrorField() string {
	return e.field
}

// NewComplaintError creates a new complaint error
func NewComplaintError(code, message, field string) Error {
	return &complaintError{
		code:    code,
		message: message,
		field:   field,
	}
}

// Predefined error codes
const (
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeDuplicate       = "DUPLICATE"
	ErrCodeStorageError   = "STORAGE_ERROR"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeInvalidFormat   = "INVALID_FORMAT"
)

// Convenience functions for creating errors
func NewValidationError(message, field string) Error {
	return NewComplaintError(ErrCodeValidationFailed, message, field)
}

func NewNotFoundError(message string) Error {
	return NewComplaintError(ErrCodeNotFound, message, "")
}

func NewDuplicateError(message string) Error {
	return NewComplaintError(ErrCodeDuplicate, message, "")
}

func NewStorageError(message string) Error {
	return NewComplaintError(ErrCodeStorageError, message, "")
}

func NewUnauthorizedError(message string) Error {
	return NewComplaintError(ErrCodeUnauthorized, message, "")
}

func NewInvalidFormatError(message, field string) Error {
	return NewComplaintError(ErrCodeInvalidFormat, message, field)
}