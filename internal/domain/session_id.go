package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/larsartmann/go-composable-business-types/id"
)

// SessionBrand is the phantom type brand for SessionID.
type SessionBrand struct{}

// SessionID represents a development session identifier using branded ID type.
type SessionID = id.ID[SessionBrand, string]

// Session ID validation pattern.
var sessionIDPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_\s]{1,100}$`)

// NewSessionID creates a new valid SessionID.
func NewSessionID(name string) (SessionID, error) {
	trimmed := strings.TrimSpace(name)
	err := validateSessionID(trimmed)
	if err != nil {
		return id.NewID[SessionBrand](""), fmt.Errorf("invalid SessionID: %w", err)
	}

	return id.NewID[SessionBrand](trimmed), nil
}

// ParseSessionID validates and creates a SessionID from string.
// Allow empty strings for optional session tracking.
func ParseSessionID(s string) (SessionID, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return id.NewID[SessionBrand](""), nil // Empty is valid for optional session
	}

	err := validateSessionID(trimmed)
	if err != nil {
		return id.NewID[SessionBrand](""), fmt.Errorf("invalid SessionID: %w", err)
	}

	return id.NewID[SessionBrand](trimmed), nil
}

// MustParseSessionID creates a SessionID from string, panics on invalid input.
func MustParseSessionID(s string) SessionID {
	sessionID, err := ParseSessionID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid SessionID: %s", err))
	}

	return sessionID
}

// validateSessionID validates SessionID format.
func validateSessionID(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}

	if len(s) > 100 {
		return errors.New("cannot exceed 100 characters")
	}

	if !sessionIDPattern.MatchString(s) {
		return errors.New("contains invalid characters")
	}

	return nil
}
