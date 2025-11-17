package domain

import (
	"encoding/json"
	"fmt"
)

// ProjectName is a value object representing a project identifier
// It enforces validation rules at construction time, making invalid states unrepresentable
// Note: ProjectName can be empty (optional field)
type ProjectName struct {
	value string
}

// NewProjectName creates a new ProjectName with validation
// Returns error if name exceeds maximum length
// Empty project names are allowed (optional field)
func NewProjectName(name string) (ProjectName, error) {
	// Empty is allowed for project names (optional field)
	if name == "" {
		return ProjectName{value: ""}, nil
	}

	// Validate: cannot exceed maximum length
	if len(name) > MaxProjectNameLength {
		return ProjectName{}, fmt.Errorf("project name exceeds maximum length of %d characters (got %d)", MaxProjectNameLength, len(name))
	}

	return ProjectName{value: name}, nil
}

// MustNewProjectName creates a new ProjectName or panics on validation failure
// Use only in tests or when you're certain the input is valid
func MustNewProjectName(name string) ProjectName {
	projectName, err := NewProjectName(name)
	if err != nil {
		panic(fmt.Sprintf("invalid project name: %v", err))
	}
	return projectName
}

// String returns the string representation of the project name
func (p ProjectName) String() string {
	return p.value
}

// IsEmpty returns true if the project name is empty
func (p ProjectName) IsEmpty() bool {
	return p.value == ""
}

// MarshalJSON implements json.Marshaler for JSON serialization
func (p ProjectName) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.value)
}

// UnmarshalJSON implements json.Unmarshaler for JSON deserialization with validation
func (p *ProjectName) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	projectName, err := NewProjectName(str)
	if err != nil {
		return err
	}

	*p = projectName
	return nil
}
