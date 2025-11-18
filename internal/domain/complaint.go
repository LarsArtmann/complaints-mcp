package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/gofrs/uuid"
)

// ComplaintID represents a unique complaint identifier using phantom type pattern (string-based)
type ComplaintID string

// UUID v4 pattern for validation
var complaintIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// NewComplaintID creates a new valid ComplaintID with UUID v4 format
func NewComplaintID() (ComplaintID, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return ComplaintID(""), fmt.Errorf("failed to generate ComplaintID: %w", err)
	}
	return ComplaintID(id.String()), nil
}

// String returns string representation of ComplaintID
func (id ComplaintID) String() string {
	return string(id)
}

// Validate checks if ComplaintID is valid
func (id ComplaintID) Validate() error {
	s := string(id)
	if s == "" {
		return fmt.Errorf("cannot be empty")
	}
	if !complaintIDPattern.MatchString(s) {
		return fmt.Errorf("must be valid UUID v4 format")
	}
	return nil
}

// IsValid returns true if ComplaintID is valid
func (id ComplaintID) IsValid() bool {
	return id.Validate() == nil
}

// MarshalJSON implements json.Marshaler for flat JSON structure
func (id ComplaintID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON implements json.Unmarshaler for flat JSON structure
func (id *ComplaintID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if err := validateComplaintID(s); err != nil {
		return err
	}
	*id = ComplaintID(s)
	return nil
}

// validateComplaintID validates ComplaintID format
func validateComplaintID(s string) error {
	if s == "" {
		return fmt.Errorf("cannot be empty")
	}
	if !complaintIDPattern.MatchString(s) {
		return fmt.Errorf("must be valid UUID v4 format")
	}
	return nil
}

// ResolutionState represents the resolution state of a complaint
type ResolutionState string

const (
	ResolutionStateOpen     ResolutionState = "open"
	ResolutionStateResolved  ResolutionState = "resolved"
)

// IsResolved returns true if the complaint is resolved
func (r ResolutionState) IsResolved() bool {
	return r == ResolutionStateResolved
}

// Severity represents the severity level of a complaint
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

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

// Resolve marks a complaint as resolved
func (c *Complaint) Resolve(resolvedBy string) error {
	if c.ResolutionState.IsResolved() {
		return fmt.Errorf("complaint is already resolved")
	}
	if resolvedBy == "" {
		return fmt.Errorf("resolver name cannot be empty")
	}
	
	now := time.Now()
	c.ResolvedAt = &now
	c.ResolvedBy = resolvedBy
	c.ResolutionState = ResolutionStateResolved
	
	return nil
}

// IsResolved returns true if the complaint is resolved
func (c *Complaint) IsResolved() bool {
	return c.ResolutionState.IsResolved()
}