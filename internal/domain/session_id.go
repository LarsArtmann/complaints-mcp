package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// SessionID represents a development session identifier using phantom type pattern.
type SessionID string

// Session ID validation pattern.
var sessionIDPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_\s]{1,100}$`)

// NewSessionID creates a new valid SessionID.
func NewSessionID(name string) (SessionID, error) {
	trimmed := strings.TrimSpace(name)
	if err := validateSessionID(trimmed); err != nil {
		return SessionID(""), fmt.Errorf("invalid SessionID: %w", err)
	}
	return SessionID(trimmed), nil
}

// ParseSessionID validates and creates a SessionID from string.
func ParseSessionID(s string) (SessionID, error) {
	// Allow empty strings for optional session tracking
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return SessionID(""), nil // Empty is valid for optional session
	}
	if err := validateSessionID(trimmed); err != nil {
		return SessionID(""), fmt.Errorf("invalid SessionID: %w", err)
	}
	return SessionID(trimmed), nil
}

// MustParseSessionID creates a SessionID from string, panics on invalid input.
func MustParseSessionID(s string) SessionID {
	id, err := ParseSessionID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid SessionID: %s", err))
	}
	return id
}

// Validate checks if SessionID is valid.
func (id SessionID) Validate() error {
	// Allow empty strings for optional session tracking
	trimmed := strings.TrimSpace(string(id))
	if trimmed == "" {
		return nil // Empty is valid for optional session
	}
	if len(trimmed) > 100 {
		return errors.New("cannot exceed 100 characters")
	}
	if !sessionIDPattern.MatchString(trimmed) {
		return errors.New("contains invalid characters")
	}
	return nil
}

// IsValid returns true if SessionID is valid.
func (id SessionID) IsValid() bool {
	return id.Validate() == nil
}

// IsEmpty returns true if SessionID is empty.
func (id SessionID) IsEmpty() bool {
	return strings.TrimSpace(string(id)) == ""
}

// String returns string representation of SessionID.
func (id SessionID) String() string {
	return string(id)
}

// MarshalJSON implements json.Marshaler for flat JSON structure.
func (id SessionID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON implements json.Unmarshaler for flat JSON structure.
func (id *SessionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseSessionID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// validateSessionID validates SessionID format.
func validateSessionID(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return errors.New("cannot be empty")
	}
	if len(trimmed) > 100 {
		return errors.New("cannot exceed 100 characters")
	}
	if !sessionIDPattern.MatchString(trimmed) {
		return errors.New("contains invalid characters")
	}
	return nil
}
