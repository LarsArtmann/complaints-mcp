package domain

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// ComplaintID represents a unique complaint identifier using phantom type pattern
type ComplaintID string

// UUID v4 pattern for validation
var complaintIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// NewComplaintID creates a new valid ComplaintID with UUID v4 format
func NewComplaintID() (ComplaintID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return ComplaintID(""), fmt.Errorf("failed to generate ComplaintID: %w", err)
	}
	return ComplaintID(id.String()), nil
}

// ParseComplaintID validates and creates a ComplaintID from string
func ParseComplaintID(s string) (ComplaintID, error) {
	if err := validateComplaintID(s); err != nil {
		return ComplaintID(""), fmt.Errorf("invalid ComplaintID: %w", err)
	}
	return ComplaintID(s), nil
}

// MustParseComplaintID creates a ComplaintID from string, panics on invalid input
func MustParseComplaintID(s string) ComplaintID {
	id, err := ParseComplaintID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid ComplaintID: %s", err))
	}
	return id
}

// Validate checks if ComplaintID is valid
func (id ComplaintID) Validate() error {
	return validateComplaintID(string(id))
}

// IsValid returns true if ComplaintID is valid
func (id ComplaintID) IsValid() bool {
	return id.Validate() == nil
}

// IsEmpty returns true if ComplaintID is empty
func (id ComplaintID) IsEmpty() bool {
	return string(id) == ""
}

// String returns the string representation of ComplaintID
func (id ComplaintID) String() string {
	return string(id)
}

// UUID returns the underlying uuid.UUID value
func (id ComplaintID) UUID() (uuid.UUID, error) {
	return uuid.Parse(string(id))
}

// MarshalJSON implements json.Marshaler for flat JSON structure
func (id ComplaintID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON implements json.Unmarshaler for flat JSON structure
func (id *ComplaintID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseComplaintID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// validateComplaintID validates ComplaintID format
func validateComplaintID(s string) error {
	if s == "" {
		return fmt.Errorf("cannot be empty")
	}
	if !complaintIDPattern.MatchString(s) {
		return fmt.Errorf("must be valid UUID v4 format")
	}
	return nil
}