package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewAgentName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{"valid name", "claude-agent", false},
		{"valid single char", "a", false},
		{"valid max length", strings.Repeat("a", MaxAgentNameLength), false},
		{"empty name", "", true},
		{"exceeds max length", strings.Repeat("a", MaxAgentNameLength+1), true},
		{"valid with spaces", "Agent Name", false},
		{"valid with numbers", "agent-123", false},
		{"valid with special chars", "agent_name-v1.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentName, err := NewAgentName(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if !agentName.IsEmpty() && tt.input == "" {
					t.Errorf("expected empty agent name for invalid input")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if agentName.String() != tt.input {
					t.Errorf("expected %q, got %q", tt.input, agentName.String())
				}
			}
		})
	}
}

func TestMustNewAgentName(t *testing.T) {
	t.Run("valid name does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewAgentName panicked unexpectedly: %v", r)
			}
		}()
		agentName := MustNewAgentName("claude-agent")
		if agentName.String() != "claude-agent" {
			t.Errorf("expected 'claude-agent', got %q", agentName.String())
		}
	})

	t.Run("invalid name panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustNewAgentName should have panicked for empty name")
			}
		}()
		MustNewAgentName("")
	})
}

func TestAgentName_String(t *testing.T) {
	agentName := MustNewAgentName("test-agent")
	if agentName.String() != "test-agent" {
		t.Errorf("expected 'test-agent', got %q", agentName.String())
	}
}

func TestAgentName_IsEmpty(t *testing.T) {
	t.Run("valid agent name is not empty", func(t *testing.T) {
		agentName := MustNewAgentName("claude-agent")
		if agentName.IsEmpty() {
			t.Errorf("expected non-empty agent name")
		}
	})

	t.Run("zero value is empty", func(t *testing.T) {
		var agentName AgentName
		if !agentName.IsEmpty() {
			t.Errorf("expected empty agent name for zero value")
		}
	})
}

func TestAgentName_JSON(t *testing.T) {
	t.Run("marshal valid agent name", func(t *testing.T) {
		agentName := MustNewAgentName("claude-agent")
		data, err := json.Marshal(agentName)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		expected := `"claude-agent"`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("unmarshal valid agent name", func(t *testing.T) {
		data := []byte(`"claude-agent"`)
		var agentName AgentName
		if err := json.Unmarshal(data, &agentName); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if agentName.String() != "claude-agent" {
			t.Errorf("expected 'claude-agent', got %q", agentName.String())
		}
	})

	t.Run("unmarshal empty name fails", func(t *testing.T) {
		data := []byte(`""`)
		var agentName AgentName
		if err := json.Unmarshal(data, &agentName); err == nil {
			t.Errorf("expected error when unmarshaling empty name")
		}
	})

	t.Run("unmarshal too long name fails", func(t *testing.T) {
		longName := strings.Repeat("a", MaxAgentNameLength+1)
		data, _ := json.Marshal(longName)
		var agentName AgentName
		if err := json.Unmarshal(data, &agentName); err == nil {
			t.Errorf("expected error when unmarshaling name exceeding max length")
		}
	})

	t.Run("round-trip marshal/unmarshal", func(t *testing.T) {
		original := MustNewAgentName("test-agent-123")

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var unmarshaled AgentName
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if unmarshaled.String() != original.String() {
			t.Errorf("expected %q, got %q", original.String(), unmarshaled.String())
		}
	})
}

func TestAgentName_Immutability(t *testing.T) {
	agentName := MustNewAgentName("original")
	originalValue := agentName.String()

	// Attempt to modify by creating a new one
	_ = MustNewAgentName("modified")

	// Original should remain unchanged
	if agentName.String() != originalValue {
		t.Errorf("agent name was modified unexpectedly")
	}
}
