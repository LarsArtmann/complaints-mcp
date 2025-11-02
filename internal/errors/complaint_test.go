package errors

import (
	"errors"
	"testing"
)

func TestComplaintError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ComplaintError
		want string
	}{
		{
			name: "error with field",
			err: &ComplaintError{
				Code:    "VALIDATION_FAILED",
				Message: "validation error",
				Field:   "name",
			},
			want: "VALIDATION_FAILED: validation error (field: name)",
		},
		{
			name: "error without field",
			err: &ComplaintError{
				Code:    "NOT_FOUND",
				Message: "not found error",
			},
			want: "NOT_FOUND: not found error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("ComplaintError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplaintError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &ComplaintError{
		Code:    "STORAGE_ERROR",
		Message: "storage error",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("ComplaintError.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestNewComplaintError(t *testing.T) {
	code := "TEST_CODE"
	message := "test message"
	field := "test_field"

	err := NewComplaintError(code, message, field)

	if err.Code != code {
		t.Errorf("NewComplaintError().Code = %v, want %v", err.Code, code)
	}

	if err.Message != message {
		t.Errorf("NewComplaintError().Message = %v, want %v", err.Message, message)
	}

	if err.Field != field {
		t.Errorf("NewComplaintError().Field = %v, want %v", err.Field, field)
	}

	if err.Details == nil {
		t.Error("NewComplaintError().Details should be initialized")
	}
}

func TestNewValidationError(t *testing.T) {
	message := "validation failed"
	field := "email"

	err := NewValidationError(message, field)

	if err.Code != ErrCodeValidationFailed {
		t.Errorf("NewValidationError().Code = %v, want %v", err.Code, ErrCodeValidationFailed)
	}

	if err.Message != message {
		t.Errorf("NewValidationError().Message = %v, want %v", err.Message, message)
	}

	if err.Field != field {
		t.Errorf("NewValidationError().Field = %v, want %v", err.Field, field)
	}
}

func TestNewNotFoundError(t *testing.T) {
	message := "not found"

	err := NewNotFoundError(message)

	if err.Code != ErrCodeNotFound {
		t.Errorf("NewNotFoundError().Code = %v, want %v", err.Code, ErrCodeNotFound)
	}

	if err.Message != message {
		t.Errorf("NewNotFoundError().Message = %v, want %v", err.Message, message)
	}

	if err.Field != "" {
		t.Errorf("NewNotFoundError().Field = %v, want empty string", err.Field)
	}
}

func TestNewDuplicateError(t *testing.T) {
	message := "duplicate entry"

	err := NewDuplicateError(message)

	if err.Code != ErrCodeDuplicate {
		t.Errorf("NewDuplicateError().Code = %v, want %v", err.Code, ErrCodeDuplicate)
	}

	if err.Message != message {
		t.Errorf("NewDuplicateError().Message = %v, want %v", err.Message, message)
	}
}

func TestNewStorageError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		cause    []error
		hasCause bool
	}{
		{
			name:     "storage error without cause",
			message:  "file write failed",
			cause:    nil,
			hasCause: false,
		},
		{
			name:     "storage error with cause",
			message:  "file write failed",
			cause:    []error{errors.New("disk full")},
			hasCause: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewStorageError(tt.message, tt.cause...)

			if err.Code != ErrCodeStorageError {
				t.Errorf("NewStorageError().Code = %v, want %v", err.Code, ErrCodeStorageError)
			}

			if err.Message != tt.message {
				t.Errorf("NewStorageError().Message = %v, want %v", err.Message, tt.message)
			}

			if (err.Cause != nil) != tt.hasCause {
				t.Errorf("NewStorageError().Cause presence = %v, want %v", err.Cause != nil, tt.hasCause)
			}
		})
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	message := "unauthorized access"

	err := NewUnauthorizedError(message)

	if err.Code != ErrCodeUnauthorized {
		t.Errorf("NewUnauthorizedError().Code = %v, want %v", err.Code, ErrCodeUnauthorized)
	}

	if err.Message != message {
		t.Errorf("NewUnauthorizedError().Message = %v, want %v", err.Message, message)
	}
}

func TestNewInvalidFormatError(t *testing.T) {
	message := "invalid format"
	field := "date"

	err := NewInvalidFormatError(message, field)

	if err.Code != ErrCodeInvalidFormat {
		t.Errorf("NewInvalidFormatError().Code = %v, want %v", err.Code, ErrCodeInvalidFormat)
	}

	if err.Message != message {
		t.Errorf("NewInvalidFormatError().Message = %v, want %v", err.Message, message)
	}

	if err.Field != field {
		t.Errorf("NewInvalidFormatError().Field = %v, want %v", err.Field, field)
	}
}

func TestComplaintError_WithDetails(t *testing.T) {
	err := NewValidationError("invalid input", "email")

	// Test adding details
	updatedErr := err.WithDetails("min_length", 5)
	updatedErr = updatedErr.WithDetails("required", true)

	if updatedErr != err {
		t.Error("WithDetails() should return the same error instance")
	}

	if updatedErr.Details["min_length"] != 5 {
		t.Errorf("WithDetails() Details[min_length] = %v, want %v", updatedErr.Details["min_length"], 5)
	}

	if updatedErr.Details["required"] != true {
		t.Errorf("WithDetails() Details[required] = %v, want %v", updatedErr.Details["required"], true)
	}
}

func TestComplaintError_WithCause(t *testing.T) {
	err := NewValidationError("invalid input", "email")
	cause := errors.New("validation failed")

	// Test adding cause
	updatedErr := err.WithCause(cause)

	if updatedErr != err {
		t.Error("WithCause() should return the same error instance")
	}

	if updatedErr.Cause != cause {
		t.Errorf("WithCause() Cause = %v, want %v", updatedErr.Cause, cause)
	}
}
