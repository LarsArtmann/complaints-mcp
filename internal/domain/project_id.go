package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ProjectID represents a project identifier using phantom type pattern.
type ProjectID string

// Project ID validation pattern.
var projectIDPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_\s\.]{1,100}$`)

// NewProjectID creates a new valid ProjectID.
func NewProjectID(name string) (ProjectID, error) {
	trimmed := strings.TrimSpace(name)
	if err := validateProjectID(trimmed); err != nil {
		return ProjectID(""), fmt.Errorf("invalid ProjectID: %w", err)
	}
	return ProjectID(trimmed), nil
}

// ParseProjectID validates and creates a ProjectID from string.
func ParseProjectID(s string) (ProjectID, error) {
	// Allow empty strings for optional project tracking
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ProjectID(""), nil // Empty is valid for optional project
	}
	if err := validateProjectID(trimmed); err != nil {
		return ProjectID(""), fmt.Errorf("invalid ProjectID: %w", err)
	}
	return ProjectID(trimmed), nil
}

// MustParseProjectID creates a ProjectID from string, panics on invalid input.
func MustParseProjectID(s string) ProjectID {
	id, err := ParseProjectID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid ProjectID: %s", err))
	}
	return id
}

// Validate checks if ProjectID is valid.
func (id ProjectID) Validate() error {
	// Allow empty strings for optional project tracking
	trimmed := strings.TrimSpace(string(id))
	if trimmed == "" {
		return nil // Empty is valid for optional project
	}
	if len(trimmed) > 100 {
		return errors.New("cannot exceed 100 characters")
	}
	if !projectIDPattern.MatchString(trimmed) {
		return errors.New("contains invalid characters")
	}
	return nil
}

// IsValid returns true if ProjectID is valid.
func (id ProjectID) IsValid() bool {
	return id.Validate() == nil
}

// IsEmpty returns true if ProjectID is empty.
func (id ProjectID) IsEmpty() bool {
	return strings.TrimSpace(string(id)) == ""
}

// String returns string representation of ProjectID.
func (id ProjectID) String() string {
	return string(id)
}

// MarshalJSON implements json.Marshaler for flat JSON structure.
func (id ProjectID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON implements json.Unmarshaler for flat JSON structure.
func (id *ProjectID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseProjectID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// validateProjectID validates ProjectID format.
func validateProjectID(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return errors.New("cannot be empty")
	}
	if len(trimmed) > 100 {
		return errors.New("cannot exceed 100 characters")
	}
	if !projectIDPattern.MatchString(trimmed) {
		return errors.New("contains invalid characters")
	}
	return nil
}
