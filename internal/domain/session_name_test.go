package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewSessionName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{"valid session name", "session-123", false},
		{"valid single char", "s", false},
		{"valid max length", strings.Repeat("s", MaxSessionNameLength), false},
		{"empty is allowed", "", false}, // SessionName can be empty (optional)
		{"exceeds max length", strings.Repeat("s", MaxSessionNameLength+1), true},
		{"valid with spaces", "Session Name", false},
		{"valid with numbers", "session-456", false},
		{"valid with special chars", "session_name-v1.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionName, err := NewSessionName(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if sessionName.String() != tt.input {
					t.Errorf("expected %q, got %q", tt.input, sessionName.String())
				}
			}
		})
	}
}

func TestMustNewSessionName(t *testing.T) {
	t.Run("valid name does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewSessionName panicked unexpectedly: %v", r)
			}
		}()
		sessionName := MustNewSessionName("test-session")
		if sessionName.String() != "test-session" {
			t.Errorf("expected 'test-session', got %q", sessionName.String())
		}
	})

	t.Run("empty name does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewSessionName should not panic for empty name (it's optional): %v", r)
			}
		}()
		sessionName := MustNewSessionName("")
		if !sessionName.IsEmpty() {
			t.Errorf("expected empty session name")
		}
	})

	t.Run("too long name panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustNewSessionName should have panicked for name exceeding max length")
			}
		}()
		MustNewSessionName(strings.Repeat("s", MaxSessionNameLength+1))
	})
}

func TestSessionName_String(t *testing.T) {
	sessionName := MustNewSessionName("test-session")
	if sessionName.String() != "test-session" {
		t.Errorf("expected 'test-session', got %q", sessionName.String())
	}
}

func TestSessionName_IsEmpty(t *testing.T) {
	t.Run("valid session name is not empty", func(t *testing.T) {
		sessionName := MustNewSessionName("test-session")
		if sessionName.IsEmpty() {
			t.Errorf("expected non-empty session name")
		}
	})

	t.Run("empty session name is empty", func(t *testing.T) {
		sessionName := MustNewSessionName("")
		if !sessionName.IsEmpty() {
			t.Errorf("expected empty session name")
		}
	})

	t.Run("zero value is empty", func(t *testing.T) {
		var sessionName SessionName
		if !sessionName.IsEmpty() {
			t.Errorf("expected empty session name for zero value")
		}
	})
}

func TestSessionName_JSON(t *testing.T) {
	t.Run("marshal valid session name", func(t *testing.T) {
		sessionName := MustNewSessionName("test-session")
		data, err := json.Marshal(sessionName)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		expected := `"test-session"`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("marshal empty session name", func(t *testing.T) {
		sessionName := MustNewSessionName("")
		data, err := json.Marshal(sessionName)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		expected := `""`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("unmarshal valid session name", func(t *testing.T) {
		data := []byte(`"test-session"`)
		var sessionName SessionName
		if err := json.Unmarshal(data, &sessionName); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if sessionName.String() != "test-session" {
			t.Errorf("expected 'test-session', got %q", sessionName.String())
		}
	})

	t.Run("unmarshal empty name succeeds", func(t *testing.T) {
		data := []byte(`""`)
		var sessionName SessionName
		if err := json.Unmarshal(data, &sessionName); err != nil {
			t.Errorf("unexpected error when unmarshaling empty name: %v", err)
		}
		if !sessionName.IsEmpty() {
			t.Errorf("expected empty session name")
		}
	})

	t.Run("unmarshal too long name fails", func(t *testing.T) {
		longName := strings.Repeat("s", MaxSessionNameLength+1)
		data, _ := json.Marshal(longName)
		var sessionName SessionName
		if err := json.Unmarshal(data, &sessionName); err == nil {
			t.Errorf("expected error when unmarshaling name exceeding max length")
		}
	})

	t.Run("round-trip marshal/unmarshal", func(t *testing.T) {
		original := MustNewSessionName("session-xyz-123")

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var unmarshaled SessionName
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if unmarshaled.String() != original.String() {
			t.Errorf("expected %q, got %q", original.String(), unmarshaled.String())
		}
	})
}

func TestSessionName_Immutability(t *testing.T) {
	sessionName := MustNewSessionName("original-session")
	originalValue := sessionName.String()

	// Attempt to modify by creating a new one
	_ = MustNewSessionName("modified-session")

	// Original should remain unchanged
	if sessionName.String() != originalValue {
		t.Errorf("session name was modified unexpectedly")
	}
}
