package domain

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// SessionName is a value object representing a session identifier
// It enforces validation rules at construction time, making invalid states unrepresentable
// Note: SessionName can be empty (optional field), unlike AgentName
type SessionName struct {
	value string
}

// NewSessionName creates a new SessionName with validation
// Returns error if name exceeds maximum length
// Empty session names are allowed (optional field)
func NewSessionName(name string) (SessionName, error) {
	// Empty is allowed for session names (optional field)
	if name == "" {
		return SessionName{value: ""}, nil
	}

	// Validate: cannot exceed maximum length in Unicode characters
	if utf8.RuneCountInString(name) > MaxSessionNameLength {
		return SessionName{}, fmt.Errorf("session name exceeds maximum length of %d characters (got %d)", MaxSessionNameLength, utf8.RuneCountInString(name))
	}

	return SessionName{value: name}, nil
}

// MustNewSessionName creates a new SessionName or panics on validation failure
// Use only in tests or when you're certain the input is valid
func MustNewSessionName(name string) SessionName {
	sessionName, err := NewSessionName(name)
	if err != nil {
		panic(fmt.Sprintf("invalid session name: %v", err))
	}
	return sessionName
}

// String returns the string representation of the session name
func (s SessionName) String() string {
	return s.value
}

// IsEmpty returns true if the session name is empty
func (s SessionName) IsEmpty() bool {
	return s.value == ""
}

// Validate checks if the session name is valid
// For SessionName, empty values are allowed, so this always returns nil
// Use IsEmpty() to check if a session name is set
func (s SessionName) Validate() error {
	// SessionName allows empty values (optional field)
	// Length validation is handled during construction
	return nil
}

// MarshalJSON implements json.Marshaler for JSON serialization
func (s SessionName) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}

// UnmarshalJSON implements json.Unmarshaler for JSON deserialization with validation
func (s *SessionName) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	sessionName, err := NewSessionName(str)
	if err != nil {
		return err
	}

	*s = sessionName
	return nil
}
