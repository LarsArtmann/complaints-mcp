package domain

import (
	"fmt"
)

// ValidationError represents domain-level validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// ParseSeverity safely converts string to domain Severity with validation
func ParseSeverity(s string) (Severity, error) {
	switch s {
	case "low":
		return SeverityLow, nil
	case "medium":
		return SeverityMedium, nil
	case "high":
		return SeverityHigh, nil
	case "critical":
		return SeverityCritical, nil
	default:
		return "", ValidationError{Field: "severity", Message: "invalid severity: " + s}
	}
}

// MustParseSeverity converts string to Severity, panics on invalid input (for tests)
func MustParseSeverity(s string) Severity {
	severity, err := ParseSeverity(s)
	if err != nil {
		panic(err.Error())
	}
	return severity
}