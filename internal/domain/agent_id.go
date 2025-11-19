package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// AgentID represents an AI agent identifier using phantom type pattern
type AgentID string

// Agent ID validation pattern - support Unicode characters including emojis
var agentIDPattern = regexp.MustCompile(`^.{1,100}$`)

// NewAgentID creates a new valid AgentID
func NewAgentID(name string) (AgentID, error) {
	trimmed := strings.TrimSpace(name)
	if err := validateAgentID(trimmed); err != nil {
		return AgentID(""), fmt.Errorf("invalid AgentID: %w", err)
	}
	return AgentID(trimmed), nil
}

// ParseAgentID validates and creates an AgentID from string
func ParseAgentID(s string) (AgentID, error) {
	// Allow empty strings for optional agent tracking  
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return AgentID(""), nil // Empty is valid for optional agent
	}
	if err := validateAgentID(trimmed); err != nil {
		return AgentID(""), fmt.Errorf("invalid AgentID: %w", err)
	}
	return AgentID(trimmed), nil
}

// MustParseAgentID creates an AgentID from string, panics on invalid input
func MustParseAgentID(s string) AgentID {
	id, err := ParseAgentID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid AgentID: %s", err))
	}
	return id
}

// Validate checks if AgentID is valid
func (id AgentID) Validate() error {
	trimmed := strings.TrimSpace(string(id))
	if trimmed == "" {
		return fmt.Errorf("cannot be empty")
	}
	if len(trimmed) > 100 {
		return fmt.Errorf("cannot exceed 100 characters")
	}
	if !agentIDPattern.MatchString(trimmed) {
		return fmt.Errorf("contains invalid characters")
	}
	return nil
}

// IsValid returns true if AgentID is valid
func (id AgentID) IsValid() bool {
	return id.Validate() == nil
}

// IsEmpty returns true if AgentID is empty
func (id AgentID) IsEmpty() bool {
	return strings.TrimSpace(string(id)) == ""
}

// String returns string representation of AgentID
func (id AgentID) String() string {
	return string(id)
}

// MarshalJSON implements json.Marshaler for flat JSON structure
func (id AgentID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON implements json.Unmarshaler for flat JSON structure
func (id *AgentID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseAgentID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// validateAgentID validates AgentID format
func validateAgentID(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return fmt.Errorf("cannot be empty")
	}
	if len(trimmed) > 100 {
		return fmt.Errorf("cannot exceed 100 characters")
	}
	if !agentIDPattern.MatchString(trimmed) {
		return fmt.Errorf("contains invalid characters")
	}
	return nil
}