package domain

import (
	"time"
)

// Severity represents the severity level of a complaint
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// ResolutionState represents the resolution state of a complaint
type ResolutionState string

const (
	ResolutionStateOpen      ResolutionState = "open"
	ResolutionStateResolved  ResolutionState = "resolved"
)

// IsResolved returns true if the complaint is resolved
func (r ResolutionState) IsResolved() bool {
	return r == ResolutionStateResolved
}

// Complaint represents a structured complaint with phantom type ID
type Complaint struct {
	ID              ComplaintID     `json:"id"`              // ✅ Phantom type - flat JSON
	AgentID         string          `json:"agent_name"`       // Keep for API compatibility
	SessionID       string          `json:"session_name"`     // Keep for API compatibility
	ProjectName     string          `json:"project_name"`     // Keep for API compatibility
	TaskDescription string          `json:"task_description"`
	ContextInfo     string          `json:"context_info"`
	MissingInfo     string          `json:"missing_info"`
	ConfusedBy      string          `json:"confused_by"`
	FutureWishes    string          `json:"future_wishes"`
	Severity        Severity        `json:"severity"`
	Timestamp       time.Time       `json:"timestamp"`
	ResolutionState ResolutionState `json:"resolution_state"`
	ResolvedAt      *time.Time     `json:"resolved_at,omitempty"`
	ResolvedBy      string          `json:"resolved_by,omitempty"`
}

// Validate checks if all fields are valid
func (c *Complaint) Validate() error {
	// Validate phantom type ID
	if err := c.ID.Validate(); err != nil {
		return fmt.Errorf("invalid ComplaintID: %w", err)
	}
	
	// Validate severity
	switch c.Severity {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		// Valid severity
	default:
		return fmt.Errorf("invalid severity: %s", c.Severity)
	}
	
	// Validate required fields
	if c.AgentID == "" {
		return fmt.Errorf("agent_name cannot be empty")
	}
	if c.TaskDescription == "" {
		return fmt.Errorf("task_description cannot be empty")
	}
	
	return nil
}

// IsValid returns true if the complaint is valid
func (c *Complaint) IsValid() bool {
	return c.Validate() == nil
}