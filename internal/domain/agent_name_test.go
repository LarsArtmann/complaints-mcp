package domain

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
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
		{"valid non-ASCII name", "Jörn", false},
		{"valid emoji name", "AI 🤖", false},
		{"valid mixed non-ASCII", "Jörn-🤖-Agent", false},
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

func TestAgentName_RuneVsByteSemantics(t *testing.T) {
	t.Run("agent name length measured in runes not bytes", func(t *testing.T) {
		// Create a name where rune count is within limit but byte count exceeds
		// Each emoji is 4 bytes but 1 rune
		longByteName := strings.Repeat("🤖", MaxAgentNameLength) // 100 runes, 400 bytes
		
		// Should be valid because we count runes, not bytes
		agentName, err := NewAgentName(longByteName)
		if err != nil {
			t.Errorf("expected name with %d runes to be valid, got error: %v", MaxAgentNameLength, err)
		}
		if agentName.String() != longByteName {
			t.Errorf("name not preserved correctly")
		}
		
		// Now exceed the rune limit (101 runes = 404 bytes)
		tooManyRunes := strings.Repeat("🤖", MaxAgentNameLength+1)
		_, err = NewAgentName(tooManyRunes)
		if err == nil {
			t.Errorf("expected name with %d runes to be invalid", MaxAgentNameLength+1)
		}
	})
	
	t.Run("mixed ASCII and multibyte characters", func(t *testing.T) {
		// Create name that's exactly at the rune limit
		name := "Jörn-" + strings.Repeat("a", MaxAgentNameLength-5) + "🤖" // 5 runes + 95 + 1 = 101 runes
		if utf8.RuneCountInString(name) != MaxAgentNameLength+1 {
			t.Fatalf("test setup error: expected %d runes, got %d", MaxAgentNameLength+1, utf8.RuneCountInString(name))
		}
		
		_, err := NewAgentName(name)
		if err == nil {
			t.Errorf("expected name with %d runes to be invalid", MaxAgentNameLength+1)
		}
		
		// Make it valid by removing one character (rune)
		runes := []rune(name)
		validName := string(runes[:len(runes)-1]) // Remove the emoji
		if utf8.RuneCountInString(validName) != MaxAgentNameLength {
			t.Fatalf("test setup error: expected %d runes, got %d", MaxAgentNameLength, utf8.RuneCountInString(validName))
		}
		
		agentName, err := NewAgentName(validName)
		if err != nil {
			t.Errorf("expected name with %d runes to be valid, got error: %v", MaxAgentNameLength, err)
		}
		if agentName.String() != validName {
			t.Errorf("valid name not preserved correctly")
		}
	})
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
		data, err := json.Marshal(longName)
		if err != nil {
			t.Fatalf("failed to marshal long name: %v", err)
		}
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

	t.Run("round-trip with non-ASCII characters", func(t *testing.T) {
		original := MustNewAgentName("Jörn 🤖 AI")
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal non-ASCII name: %v", err)
		}

		var unmarshaled AgentName
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal non-ASCII name: %v", err)
		}

		if unmarshaled.String() != original.String() {
			t.Errorf("expected %q, got %q", original.String(), unmarshaled.String())
		}
	})
}

func TestAgentName_Immutability(t *testing.T) {
	// Test 1: Verify AgentName type has no exported fields (ensuring structural immutability)
	agentType := reflect.TypeOf(AgentName{})
	
	// Check that all fields are unexported (start with lowercase letter)
	exportedFieldCount := 0
	for i := 0; i < agentType.NumField(); i++ {
		field := agentType.Field(i)
		if field.IsExported() {
			exportedFieldCount++
			t.Errorf("AgentName has exported field: %s", field.Name)
		}
	}
	
	if exportedFieldCount > 0 {
		t.Errorf("AgentName should have no exported fields for immutability, found %d", exportedFieldCount)
	}
	
	// Test 2: Verify JSON round-trip preserves original value
	original := MustNewAgentName("test-agent")
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	
	var copy AgentName
	if err := json.Unmarshal(data, &copy); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	
	// The original should be unchanged
	if original.String() != "test-agent" {
		t.Errorf("original value changed: expected 'test-agent', got %q", original.String())
	}
	
	// The copy should have the same value
	if copy.String() != original.String() {
		t.Errorf("round-trip changed value: expected %q, got %q", original.String(), copy.String())
	}
}
