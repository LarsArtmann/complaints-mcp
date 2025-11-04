package domain

import (
	"testing"
)

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Severity
		expectError bool
	}{
		{
			name:     "valid low",
			input:    "low",
			expected: SeverityLow,
		},
		{
			name:     "valid medium",
			input:    "medium",
			expected: SeverityMedium,
		},
		{
			name:     "valid high",
			input:    "high",
			expected: SeverityHigh,
		},
		{
			name:     "valid critical",
			input:    "critical",
			expected: SeverityCritical,
		},
		{
			name:        "invalid severity",
			input:       "invalid",
			expectError: true,
		},
		{
			name:        "empty severity",
			input:       "",
			expectError: true,
		},
		{
			name:        "case sensitive",
			input:       "Low",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSeverity(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseSeverity(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSeverity(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("ParseSeverity(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMustParseSeverity(t *testing.T) {
	// Valid case should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustParseSeverity(valid) panicked unexpectedly: %v", r)
		}
	}()
	result := MustParseSeverity("high")
	if result != SeverityHigh {
		t.Errorf("MustParseSeverity(high) = %v, want %v", result, SeverityHigh)
	}

	// Invalid case should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseSeverity(invalid) should have panicked")
		}
	}()
	MustParseSeverity("invalid")
}
