package validation

import (
	"sync"

	v10 "github.com/go-playground/validator/v10"
)

// Validator provides a centralized, thread-safe validation instance.
type Validator struct {
	validate *v10.Validate
}

var (
	instance *Validator
	once     sync.Once
)

// GetValidator returns the singleton validator instance.
func GetValidator() *Validator {
	once.Do(func() {
		v := v10.New()
		instance = &Validator{validate: v}
	})

	return instance
}

// ValidateStruct validates a struct using struct tags.
func (v *Validator) ValidateStruct(s any) error {
	return v.validate.Struct(s)
}

// ValidateStructPartial validates specific fields of a struct.
func (v *Validator) ValidateStructPartial(s any, fields ...string) error {
	return v.validate.StructPartial(s, fields...)
}

// ValidateVar validates a single variable.
func (v *Validator) ValidateVar(field any, tag string) error {
	return v.validate.Var(field, tag)
}

// RegisterValidation registers a custom validation function.
func (v *Validator) RegisterValidation(
	tag string,
	fn v10.Func,
	callValidationEvenIfNull ...bool,
) error {
	return v.validate.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}

// ValidationError represents a structured validation error.
type ValidationError struct {
	Value   any    `json:"value,omitempty"`
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}

	return e[0].Message
}

// IsEmpty returns true if there are no validation errors.
func (e ValidationErrors) IsEmpty() bool {
	return len(e) == 0
}

// ToMap converts validation errors to a map for serialization.
func (e ValidationErrors) ToMap() map[string]string {
	m := make(map[string]string, len(e))
	for _, err := range e {
		m[err.Field] = err.Message
	}

	return m
}

// ParseValidatorErrors converts validator.ValidationErrors to ValidationErrors.
func ParseValidatorErrors(err error) ValidationErrors {
	if err == nil {
		return nil
	}

	validatorErrors, ok := err.(v10.ValidationErrors)
	if !ok {
		return ValidationErrors{{
			Field:   "",
			Rule:    "",
			Message: err.Error(),
		}}
	}

	result := make(ValidationErrors, 0, len(validatorErrors))
	for _, e := range validatorErrors {
		result = append(result, ValidationError{
			Field:   e.Field(),
			Rule:    e.Tag(),
			Message: formatErrorMessage(e),
			Value:   e.Value(),
		})
	}

	return result
}

// formatErrorMessage creates a human-readable error message.
func formatErrorMessage(e v10.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "min":
		return "value is too short"
	case "max":
		return "value is too long"
	case "email":
		return "invalid email format"
	case "uuid":
		return "invalid UUID format"
	case "oneof":
		return "invalid value, must be one of: " + e.Param()
	case "gt":
		return "value must be greater than " + e.Param()
	case "gte":
		return "value must be greater than or equal to " + e.Param()
	case "lt":
		return "value must be less than " + e.Param()
	case "lte":
		return "value must be less than or equal to " + e.Param()
	default:
		return "failed validation: " + e.Tag()
	}
}

// Validate validates a struct and returns structured errors.
func Validate(s any) ValidationErrors {
	v := GetValidator()
	err := v.ValidateStruct(s)

	return ParseValidatorErrors(err)
}

// ValidatePartial validates specific fields of a struct.
func ValidatePartial(s any, fields ...string) ValidationErrors {
	v := GetValidator()
	err := v.ValidateStructPartial(s, fields...)

	return ParseValidatorErrors(err)
}
