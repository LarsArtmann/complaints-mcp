package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/larsartmann/go-composable-business-types/id"
)

// AgentBrand is the phantom type brand for AgentID.
type AgentBrand struct{}

// AgentID represents an AI agent identifier using branded ID type.
type AgentID = id.ID[AgentBrand, string]

// Agent ID validation pattern - support Unicode characters including emojis.
var agentIDPattern = regexp.MustCompile(`^.{1,100}$`)

// NewAgentID creates a new valid AgentID.
func NewAgentID(name string) (AgentID, error) {
	trimmed := strings.TrimSpace(name)
	err := validateAgentID(trimmed)
	if err != nil {
		return id.NewID[AgentBrand](""), fmt.Errorf("invalid AgentID: %w", err)
	}

	return id.NewID[AgentBrand](trimmed), nil
}

// ParseAgentID validates and creates an AgentID from string.
// Allow empty strings for optional agent tracking.
func ParseAgentID(s string) (AgentID, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return id.NewID[AgentBrand](""), nil // Empty is valid for optional agent
	}

	err := validateAgentID(trimmed)
	if err != nil {
		return id.NewID[AgentBrand](""), fmt.Errorf("invalid AgentID: %w", err)
	}

	return id.NewID[AgentBrand](trimmed), nil
}

// MustParseAgentID creates an AgentID from string, panics on invalid input.
func MustParseAgentID(s string) AgentID {
	agentID, err := ParseAgentID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid AgentID: %s", err))
	}

	return agentID
}

// validateAgentID validates AgentID format.
func validateAgentID(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}

	if len(s) > 100 {
		return errors.New("cannot exceed 100 characters")
	}

	if !agentIDPattern.MatchString(s) {
		return errors.New("contains invalid characters")
	}

	return nil
}
