package domain

import (
	"context"
	"testing"
)

func TestNewComplaintID(t *testing.T) {
	id, err := NewComplaintID()
	if err != nil {
		t.Errorf("NewComplaintID() returned error: %v", err)
	}
	if id.IsEmpty() {
		t.Error("NewComplaintID() returned empty ID")
	}
	if id.String() == "" {
		t.Error("NewComplaintID() returned empty string")
	}
}

func TestComplaintID_String(t *testing.T) {
	id := ComplaintID{Value: "test-id"}
	want := "test-id"
	got := id.String()

	if got != want {
		t.Errorf("ComplaintID.String() = %v, want %v", got, want)
	}
}

func TestComplaintID_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		id   ComplaintID
		want bool
	}{
		{
			name: "non-empty ID",
			id:   ComplaintID{Value: "test-id"},
			want: false,
		},
		{
			name: "empty ID",
			id:   ComplaintID{Value: ""},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.IsEmpty()
			if got != tt.want {
				t.Errorf("ComplaintID.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewComplaint(t *testing.T) {
	ctx := context.Background()
	complaint, err := NewComplaint(
		ctx, // ✅ Added context parameter
		"test-agent",
		"test-session",
		"test task",
		"test context",
		"missing info",
		"confused by",
		"future wishes",
		SeverityHigh, // ✅ Use domain.Severity type instead of string
		"test-project",
	)
	if err != nil {
		t.Errorf("NewComplaint() returned error: %v", err)
	}

	if complaint == nil {
		t.Fatal("NewComplaint() returned nil")
	}

	if complaint.AgentName != "test-agent" {
		t.Errorf("NewComplaint().AgentName = %v, want %v", complaint.AgentName, "test-agent")
	}

	if complaint.Severity != "high" {
		t.Errorf("NewComplaint().Severity = %v, want %v", complaint.Severity, "high")
	}

	if complaint.Resolved {
		t.Error("NewComplaint().Resolved should be false")
	}

	if complaint.Timestamp.IsZero() {
		t.Error("NewComplaint().Timestamp should not be zero")
	}
}

func TestComplaint_Resolve(t *testing.T) {
	id, _ := NewComplaintID()
	complaint := &Complaint{
		ID:       id,
		Resolved: false,
	}

	complaint.Resolve("test-agent") // Pure domain method - no context

	if !complaint.Resolved {
		t.Error("Complaint.Resolve() did not set Resolved to true")
	}

	// ✅ Verify ResolvedAt timestamp is set (prevents split-brain)
	if complaint.ResolvedAt == nil {
		t.Error("Complaint.Resolve() did not set ResolvedAt timestamp")
	}

	// ✅ NEW: Verify ResolvedBy field is set
	if complaint.ResolvedBy != "test-agent" {
		t.Errorf("Complaint.Resolve() did not set ResolvedBy correctly. Expected 'test-agent', got '%s'", complaint.ResolvedBy)
	}
}

func TestComplaint_IsResolved(t *testing.T) {
	tests := []struct {
		name      string
		complaint *Complaint
		want      bool
	}{
		{
			name: "resolved complaint",
			complaint: &Complaint{
				Resolved: true,
			},
			want: true,
		},
		{
			name: "unresolved complaint",
			complaint: &Complaint{
				Resolved: false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.complaint.IsResolved()
			if got != tt.want {
				t.Errorf("Complaint.IsResolved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplaint_Validate(t *testing.T) {
	tests := []struct {
		name      string
		complaint *Complaint
		wantErr   bool
	}{
		{
			name: "valid complaint",
			complaint: &Complaint{
				AgentName:       "test-agent",
				TaskDescription: "test task",
				Severity:        "high",
			},
			wantErr: false,
		},
		{
			name: "missing agent name",
			complaint: &Complaint{
				TaskDescription: "test task",
				Severity:        "high",
			},
			wantErr: true,
		},
		{
			name: "missing task description",
			complaint: &Complaint{
				AgentName: "test-agent",
				Severity:  "high",
			},
			wantErr: true,
		},
		{
			name: "missing severity",
			complaint: &Complaint{
				AgentName:       "test-agent",
				TaskDescription: "test task",
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			complaint: &Complaint{
				AgentName:       "test-agent",
				TaskDescription: "test task",
				Severity:        "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.complaint.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Complaint.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
