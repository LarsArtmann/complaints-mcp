package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewProjectName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{"valid project name", "my-project", false},
		{"valid single char", "p", false},
		{"valid max length", strings.Repeat("p", MaxProjectNameLength), false},
		{"empty is allowed", "", false}, // ProjectName can be empty (optional)
		{"exceeds max length", strings.Repeat("p", MaxProjectNameLength+1), true},
		{"valid with spaces", "Project Name", false},
		{"valid with numbers", "project-123", false},
		{"valid with special chars", "project_name-v1.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectName, err := NewProjectName(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if projectName.String() != tt.input {
					t.Errorf("expected %q, got %q", tt.input, projectName.String())
				}
			}
		})
	}
}

func TestMustNewProjectName(t *testing.T) {
	t.Run("valid name does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewProjectName panicked unexpectedly: %v", r)
			}
		}()
		projectName := MustNewProjectName("test-project")
		if projectName.String() != "test-project" {
			t.Errorf("expected 'test-project', got %q", projectName.String())
		}
	})

	t.Run("empty name does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewProjectName should not panic for empty name (it's optional): %v", r)
			}
		}()
		projectName := MustNewProjectName("")
		if !projectName.IsEmpty() {
			t.Errorf("expected empty project name")
		}
	})

	t.Run("too long name panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustNewProjectName should have panicked for name exceeding max length")
			}
		}()
		MustNewProjectName(strings.Repeat("p", MaxProjectNameLength+1))
	})
}

func TestProjectName_String(t *testing.T) {
	projectName := MustNewProjectName("test-project")
	if projectName.String() != "test-project" {
		t.Errorf("expected 'test-project', got %q", projectName.String())
	}
}

func TestProjectName_IsEmpty(t *testing.T) {
	t.Run("valid project name is not empty", func(t *testing.T) {
		projectName := MustNewProjectName("test-project")
		if projectName.IsEmpty() {
			t.Errorf("expected non-empty project name")
		}
	})

	t.Run("empty project name is empty", func(t *testing.T) {
		projectName := MustNewProjectName("")
		if !projectName.IsEmpty() {
			t.Errorf("expected empty project name")
		}
	})

	t.Run("zero value is empty", func(t *testing.T) {
		var projectName ProjectName
		if !projectName.IsEmpty() {
			t.Errorf("expected empty project name for zero value")
		}
	})
}

func TestProjectName_JSON(t *testing.T) {
	t.Run("marshal valid project name", func(t *testing.T) {
		projectName := MustNewProjectName("complaints-mcp")
		data, err := json.Marshal(projectName)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		expected := `"complaints-mcp"`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("marshal empty project name", func(t *testing.T) {
		projectName := MustNewProjectName("")
		data, err := json.Marshal(projectName)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		expected := `""`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("unmarshal valid project name", func(t *testing.T) {
		data := []byte(`"complaints-mcp"`)
		var projectName ProjectName
		if err := json.Unmarshal(data, &projectName); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if projectName.String() != "complaints-mcp" {
			t.Errorf("expected 'complaints-mcp', got %q", projectName.String())
		}
	})

	t.Run("unmarshal empty name succeeds", func(t *testing.T) {
		data := []byte(`""`)
		var projectName ProjectName
		if err := json.Unmarshal(data, &projectName); err != nil {
			t.Errorf("unexpected error when unmarshaling empty name: %v", err)
		}
		if !projectName.IsEmpty() {
			t.Errorf("expected empty project name")
		}
	})

	t.Run("unmarshal too long name fails", func(t *testing.T) {
		longName := strings.Repeat("p", MaxProjectNameLength+1)
		data, _ := json.Marshal(longName)
		var projectName ProjectName
		if err := json.Unmarshal(data, &projectName); err == nil {
			t.Errorf("expected error when unmarshaling name exceeding max length")
		}
	})

	t.Run("round-trip marshal/unmarshal", func(t *testing.T) {
		original := MustNewProjectName("my-awesome-project")

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var unmarshaled ProjectName
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if unmarshaled.String() != original.String() {
			t.Errorf("expected %q, got %q", original.String(), unmarshaled.String())
		}
	})
}

func TestProjectName_Immutability(t *testing.T) {
	projectName := MustNewProjectName("original-project")
	originalValue := projectName.String()

	// Attempt to modify by creating a new one
	_ = MustNewProjectName("modified-project")

	// Original should remain unchanged
	if projectName.String() != originalValue {
		t.Errorf("project name was modified unexpectedly")
	}
}
