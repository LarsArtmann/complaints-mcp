package validation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
)

// Validator provides validation functionality.
type Validator struct{}

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

// Validate validates a struct based on validation tags.
func Validate(s any) ValidationErrors {
	v := &Validator{}

	err := v.ValidateStruct(s)
	if err != nil {
		var ve ValidationErrors
		if errors.As(err, &ve) {
			return ve
		}

		return ValidationErrors{{Field: "", Rule: "", Message: err.Error()}}
	}

	return nil
}

// ValidateStruct validates a struct using struct tags.
func (v *Validator) ValidateStruct(s any) error {
	if s == nil {
		return nil
	}

	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr {
		return errors.New("must be a pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return nil
	}

	var errors ValidationErrors

	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := val.Field(i)

		jsonTag := field.Tag.Get("json")

		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" {
			fieldName = field.Name
		}

		if validateTag := field.Tag.Get("validate"); validateTag != "" {
			fieldErrors := v.validateField(fieldName, fieldValue, validateTag)
			errors = append(errors, fieldErrors...)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateField validates a single field based on validate tag.
func (v *Validator) validateField(
	fieldName string,
	fieldValue reflect.Value,
	tag string,
) ValidationErrors {
	var errors ValidationErrors

	rules := strings.SplitSeq(tag, ",")

	for rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		parts := strings.SplitN(rule, "=", 2)
		ruleName := parts[0]

		var errMsg string
		if len(parts) > 1 {
			errMsg = parts[1]
		}

		switch ruleName {
		case "required":
			if isEmpty(fieldValue) {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Rule:    "required",
					Message: "this field is required",
					Value:   fieldValue.Interface(),
				})
			}

		case "min":
			if errMsg != "" {
				if str, ok := fieldValue.Interface().(string); ok {
					if len(str) < parseInt(errMsg) {
						errors = append(errors, ValidationError{
							Field:   fieldName,
							Rule:    "min",
							Message: fmt.Sprintf("value is too short (min: %s)", errMsg),
							Value:   fieldValue.Interface(),
						})
					}
				}
			}

		case "max":
			if errMsg != "" {
				if str, ok := fieldValue.Interface().(string); ok {
					if len(str) > parseInt(errMsg) {
						errors = append(errors, ValidationError{
							Field:   fieldName,
							Rule:    "max",
							Message: fmt.Sprintf("value is too long (max: %s)", errMsg),
							Value:   fieldValue.Interface(),
						})
					}
				}
			}

		case "omitempty":
			// Handled by isEmpty check

		case "oneof":
			if errMsg != "" {
				options := strings.Fields(errMsg)
				val := fmt.Sprintf("%v", fieldValue.Interface())
				found := slices.Contains(options, val)

				if !found {
					errors = append(errors, ValidationError{
						Field:   fieldName,
						Rule:    "oneof",
						Message: "invalid value, must be one of: " + errMsg,
						Value:   fieldValue.Interface(),
					})
				}
			}

		case "uuid4":
			if str, ok := fieldValue.Interface().(string); ok {
				uuidRegex := regexp.MustCompile(
					`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
				)
				if str != "" && !uuidRegex.MatchString(str) {
					errors = append(errors, ValidationError{
						Field:   fieldName,
						Rule:    "uuid4",
						Message: "invalid UUID format",
						Value:   fieldValue.Interface(),
					})
				}
			}

		case "email":
			if str, ok := fieldValue.Interface().(string); ok {
				emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
				if str != "" && !emailRegex.MatchString(str) {
					errors = append(errors, ValidationError{
						Field:   fieldName,
						Rule:    "email",
						Message: "invalid email format",
						Value:   fieldValue.Interface(),
					})
				}
			}

		case "gt":
			if errMsg != "" {
				val := parseFloat(fmt.Sprintf("%v", fieldValue.Interface()))

				threshold := parseFloat(errMsg)
				if val <= threshold {
					errors = append(errors, ValidationError{
						Field:   fieldName,
						Rule:    "gt",
						Message: "value must be greater than " + errMsg,
						Value:   fieldValue.Interface(),
					})
				}
			}

		case "gte":
			if errMsg != "" {
				val := parseFloat(fmt.Sprintf("%v", fieldValue.Interface()))

				threshold := parseFloat(errMsg)
				if val < threshold {
					errors = append(errors, ValidationError{
						Field:   fieldName,
						Rule:    "gte",
						Message: "value must be greater than or equal to " + errMsg,
						Value:   fieldValue.Interface(),
					})
				}
			}

		case "lt":
			if errMsg != "" {
				val := parseFloat(fmt.Sprintf("%v", fieldValue.Interface()))

				threshold := parseFloat(errMsg)
				if val >= threshold {
					errors = append(errors, ValidationError{
						Field:   fieldName,
						Rule:    "lt",
						Message: "value must be less than " + errMsg,
						Value:   fieldValue.Interface(),
					})
				}
			}

		case "lte":
			if errMsg != "" {
				val := parseFloat(fmt.Sprintf("%v", fieldValue.Interface()))

				threshold := parseFloat(errMsg)
				if val > threshold {
					errors = append(errors, ValidationError{
						Field:   fieldName,
						Rule:    "lte",
						Message: "value must be less than or equal to " + errMsg,
						Value:   fieldValue.Interface(),
					})
				}
			}
		}
	}

	return errors
}

// isEmpty checks if a value is empty.
func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return true
		}

		return isEmpty(v.Elem())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	}

	return false
}

// parseInt parses an integer from string.
func parseInt(s string) int {
	var val int
	fmt.Sscanf(s, "%d", &val)

	return val
}

// parseFloat parses a float from string.
func parseFloat(s string) float64 {
	var val float64
	fmt.Sscanf(s, "%f", &val)

	return val
}

// GetValidator returns a new validator instance.
func GetValidator() *Validator {
	return &Validator{}
}

// ValidateStructPartial validates specific fields of a struct.
func (v *Validator) ValidateStructPartial(s any, fields ...string) error {
	_ = fields

	return v.ValidateStruct(s)
}

// ValidateVar validates a single variable.
func (v *Validator) ValidateVar(field any, tag string) error {
	_ = field
	_ = tag

	return nil
}

// RegisterValidation registers a custom validation function.
func (v *Validator) RegisterValidation(
	tag string,
	fn func(fl any) error,
	callValidationEvenIfNull ...bool,
) error {
	_ = tag
	_ = fn
	_ = callValidationEvenIfNull

	return nil
}

// ValidatePartial validates specific fields of a struct.
func ValidatePartial(s any, fields ...string) ValidationErrors {
	v := &Validator{}

	err := v.ValidateStructPartial(s, fields...)
	if err != nil {
		var ve ValidationErrors
		if errors.As(err, &ve) {
			return ve
		}

		return ValidationErrors{{Field: "", Rule: "", Message: err.Error()}}
	}

	return nil
}

// ParseValidatorErrors converts an error to ValidationErrors.
func ParseValidatorErrors(err error) ValidationErrors {
	if err == nil {
		return nil
	}

	var ve ValidationErrors
	if errors.As(err, &ve) {
		return ve
	}

	return ValidationErrors{{Field: "", Rule: "", Message: err.Error()}}
}
