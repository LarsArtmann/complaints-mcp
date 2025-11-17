package domain

import (
	"encoding/json"
	"fmt"
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

	// Validate: cannot exceed maximum length
	if len(name) > MaxAgentNameLength {
		return AgentName{}, fmt.Errorf("agent name exceeds maximum length of %d characters (got %d)", MaxAgentNameLength, len(name))
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
// This should never be true for a properly constructed AgentName
func (a AgentName) IsEmpty() bool {
	return a.value == ""
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
