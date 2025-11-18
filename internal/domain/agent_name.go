package domain

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// AgentName is a value object representing an AI agent's name
// It enforces validation rules at construction time, making invalid states unrepresentable
type AgentName struct {
	value string
}

// NewAgentName creates a new AgentName with validation
// Returns error if name is empty or exceeds maximum length
func NewAgentName(name string) (AgentName, error) {
	// Validate: cannot be empty
	if name == "" {
		return AgentName{}, fmt.Errorf("agent name cannot be empty")
	}

	// Validate: cannot exceed maximum length in Unicode characters
	if utf8.RuneCountInString(name) > MaxAgentNameLength {
		return AgentName{}, fmt.Errorf("agent name exceeds maximum length of %d characters (got %d)", MaxAgentNameLength, utf8.RuneCountInString(name))
	}

	return AgentName{value: name}, nil
}

// MustNewAgentName creates a new AgentName or panics on validation failure
// Use only in tests or when you're certain the input is valid
func MustNewAgentName(name string) AgentName {
	agentName, err := NewAgentName(name)
	if err != nil {
		panic(fmt.Sprintf("invalid agent name: %v", err))
	}
	return agentName
}

// String returns the string representation of the agent name
func (a AgentName) String() string {
	return a.value
}

// IsEmpty returns true if the agent name is empty
// Returns true only for the zero-value AgentName and is intended as a defensive check
// for deserialization, reflection, or other cases where zero-values might occur
func (a AgentName) IsEmpty() bool {
	return a.value == ""
}

// Validate checks if the agent name is valid
// For AgentName, this should never fail if constructed via NewAgentName()
// Use IsEmpty() to check for zero-values from deserialization/reflection
func (a AgentName) Validate() error {
	// Check for zero-value case (shouldn't happen with proper construction)
	if a.IsEmpty() {
		return fmt.Errorf("agent name cannot be empty")
	}

	// Length validation is handled during construction
	return nil
}

// MarshalJSON implements json.Marshaler for JSON serialization
func (a AgentName) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.value)
}

// UnmarshalJSON implements json.Unmarshaler for JSON deserialization with validation
func (a *AgentName) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	agentName, err := NewAgentName(s)
	if err != nil {
		return err
	}

	*a = agentName
	return nil
}
