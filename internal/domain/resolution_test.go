package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewResolution(t *testing.T) {
	tests := []struct {
		name        string
		resolvedBy  AgentName
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid resolution",
			resolvedBy: MustNewAgentName("AI Assistant"),
			wantErr:    false,
		},
		{
			name:        "empty resolver name",
			resolvedBy:  AgentName{}, // Empty value object
			wantErr:     true,
			errContains: "resolver name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewResolution(tt.resolvedBy)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewResolution() expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("NewResolution() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewResolution() unexpected error = %v", err)
				return
			}

			// Verify timestamp is recent (within last second)
			if time.Since(got.Timestamp()) > time.Second {
				t.Errorf("NewResolution() timestamp is not recent: %v", got.Timestamp())
			}

			// Verify resolver is set correctly
			if got.ResolvedBy().String() != tt.resolvedBy.String() {
				t.Errorf("NewResolution() resolvedBy = %v, want %v", got.ResolvedBy(), tt.resolvedBy)
			}
		})
	}
}

func TestNewResolutionWithTime(t *testing.T) {
	validTime := time.Now().Add(-1 * time.Hour)
	validAgent := MustNewAgentName("AI Assistant")

	tests := []struct {
		name        string
		timestamp   time.Time
		resolvedBy  AgentName
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid resolution with time",
			timestamp:  validTime,
			resolvedBy: validAgent,
			wantErr:    false,
		},
		{
			name:        "zero timestamp",
			timestamp:   time.Time{},
			resolvedBy:  validAgent,
			wantErr:     true,
			errContains: "resolution timestamp cannot be zero",
		},
		{
			name:        "empty resolver",
			timestamp:   validTime,
			resolvedBy:  AgentName{},
			wantErr:     true,
			errContains: "resolver name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewResolutionWithTime(tt.timestamp, tt.resolvedBy)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewResolutionWithTime() expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("NewResolutionWithTime() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewResolutionWithTime() unexpected error = %v", err)
				return
			}

			if !got.Timestamp().Equal(tt.timestamp) {
				t.Errorf("NewResolutionWithTime() timestamp = %v, want %v", got.Timestamp(), tt.timestamp)
			}

			if got.ResolvedBy().String() != tt.resolvedBy.String() {
				t.Errorf("NewResolutionWithTime() resolvedBy = %v, want %v", got.ResolvedBy(), tt.resolvedBy)
			}
		})
	}
}

func TestResolution_Getters(t *testing.T) {
	timestamp := time.Date(2025, 11, 19, 10, 30, 0, 0, time.UTC)
	agent := MustNewAgentName("Test Agent")

	resolution, err := NewResolutionWithTime(timestamp, agent)
	if err != nil {
		t.Fatalf("NewResolutionWithTime() failed: %v", err)
	}

	// Test Timestamp()
	if !resolution.Timestamp().Equal(timestamp) {
		t.Errorf("Timestamp() = %v, want %v", resolution.Timestamp(), timestamp)
	}

	// Test ResolvedBy()
	if resolution.ResolvedBy().String() != agent.String() {
		t.Errorf("ResolvedBy() = %v, want %v", resolution.ResolvedBy(), agent)
	}

	// Test String()
	expected := "Resolved by Test Agent at 2025-11-19T10:30:00Z"
	if resolution.String() != expected {
		t.Errorf("String() = %v, want %v", resolution.String(), expected)
	}
}

func TestResolution_IsZero(t *testing.T) {
	tests := []struct {
		name       string
		resolution Resolution
		want       bool
	}{
		{
			name:       "zero value resolution",
			resolution: Resolution{},
			want:       true,
		},
		{
			name: "valid resolution",
			resolution: Resolution{
				timestamp:  time.Now(),
				resolvedBy: MustNewAgentName("Agent"),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resolution.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolution_JSON(t *testing.T) {
	timestamp := time.Date(2025, 11, 19, 10, 30, 0, 0, time.UTC)
	agent := MustNewAgentName("Test Agent")

	original, err := NewResolutionWithTime(timestamp, agent)
	if err != nil {
		t.Fatalf("NewResolutionWithTime() failed: %v", err)
	}

	// Test MarshalJSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	// Verify JSON structure
	expectedJSON := `{"timestamp":"2025-11-19T10:30:00Z","resolved_by":"Test Agent"}`
	if string(jsonData) != expectedJSON {
		t.Errorf("MarshalJSON() = %s, want %s", string(jsonData), expectedJSON)
	}

	// Test UnmarshalJSON
	var unmarshaled Resolution
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	// Verify round-trip preservation
	if !unmarshaled.Equal(original) {
		t.Errorf("JSON round-trip failed: got %v, want %v", unmarshaled, original)
	}
}

func TestResolution_JSON_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid JSON",
			jsonData: `{"timestamp":"2025-11-19T10:30:00Z","resolved_by":"Agent"}`,
			wantErr:  false,
		},
		{
			name:        "invalid JSON syntax",
			jsonData:    `{invalid}`,
			wantErr:     true,
			errContains: "", // Any error is acceptable for malformed JSON
		},
		{
			name:        "empty resolver name",
			jsonData:    `{"timestamp":"2025-11-19T10:30:00Z","resolved_by":""}`,
			wantErr:     true,
			errContains: "invalid resolver name",
		},
		{
			name:        "zero timestamp",
			jsonData:    `{"timestamp":"0001-01-01T00:00:00Z","resolved_by":"Agent"}`,
			wantErr:     true,
			errContains: "invalid resolution data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resolution Resolution
			err := json.Unmarshal([]byte(tt.jsonData), &resolution)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UnmarshalJSON() expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("UnmarshalJSON() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalJSON() unexpected error = %v", err)
			}
		})
	}
}

func TestResolution_Equal(t *testing.T) {
	time1 := time.Date(2025, 11, 19, 10, 30, 0, 0, time.UTC)
	time2 := time.Date(2025, 11, 19, 10, 31, 0, 0, time.UTC)
	agent1 := MustNewAgentName("Agent1")
	agent2 := MustNewAgentName("Agent2")

	res1, _ := NewResolutionWithTime(time1, agent1)
	res2, _ := NewResolutionWithTime(time1, agent1) // Same as res1
	res3, _ := NewResolutionWithTime(time2, agent1) // Different time
	res4, _ := NewResolutionWithTime(time1, agent2) // Different agent

	tests := []struct {
		name  string
		res1  Resolution
		res2  Resolution
		equal bool
	}{
		{
			name:  "equal resolutions",
			res1:  res1,
			res2:  res2,
			equal: true,
		},
		{
			name:  "different timestamps",
			res1:  res1,
			res2:  res3,
			equal: false,
		},
		{
			name:  "different agents",
			res1:  res1,
			res2:  res4,
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.res1.Equal(tt.res2); got != tt.equal {
				t.Errorf("Equal() = %v, want %v", got, tt.equal)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
