package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/larsartmann/go-composable-business-types/id"
)

// ProjectBrand is the phantom type brand for ProjectID.
type ProjectBrand struct{}

// ProjectID represents a project identifier using branded ID type.
type ProjectID = id.ID[ProjectBrand, string]

// Project ID validation pattern.
var projectIDPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_\s\.]{1,100}$`)

// NewProjectID creates a new valid ProjectID.
func NewProjectID(name string) (ProjectID, error) {
	trimmed := strings.TrimSpace(name)
	if err := validateProjectID(trimmed); err != nil {
		return id.NewID[ProjectBrand](""), fmt.Errorf("invalid ProjectID: %w", err)
	}
	return id.NewID[ProjectBrand](trimmed), nil
}

// ParseProjectID validates and creates a ProjectID from string.
// Allow empty strings for optional project tracking.
func ParseProjectID(s string) (ProjectID, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return id.NewID[ProjectBrand](""), nil // Empty is valid for optional project
	}
	if err := validateProjectID(trimmed); err != nil {
		return id.NewID[ProjectBrand](""), fmt.Errorf("invalid ProjectID: %w", err)
	}
	return id.NewID[ProjectBrand](trimmed), nil
}

// MustParseProjectID creates a ProjectID from string, panics on invalid input.
func MustParseProjectID(s string) ProjectID {
	projectID, err := ParseProjectID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid ProjectID: %s", err))
	}
	return projectID
}

// validateProjectID validates ProjectID format.
func validateProjectID(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}
	if len(s) > 100 {
		return errors.New("cannot exceed 100 characters")
	}
	if !projectIDPattern.MatchString(s) {
		return errors.New("contains invalid characters")
	}
	return nil
}
