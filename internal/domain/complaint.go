package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/gofrs/uuid"
	"github.com/larsartmann/go-composable-business-types/id"
)

// ComplaintBrand is the phantom type brand for ComplaintID.
type ComplaintBrand struct{}

// ComplaintID represents a unique complaint identifier using branded ID type.
type ComplaintID = id.ID[ComplaintBrand, string]

// UUID v4 pattern for validation.
var complaintIDPattern = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

// NewComplaintID creates a new valid ComplaintID with UUID v4 format.
func NewComplaintID() (ComplaintID, error) {
	uuidValue, err := uuid.NewV4()
	if err != nil {
		return id.NewID[ComplaintBrand](""), fmt.Errorf("failed to generate ComplaintID: %w", err)
	}
	return id.NewID[ComplaintBrand](uuidValue.String()), nil
}

// ParseComplaintID validates and creates a ComplaintID from string.
func ParseComplaintID(s string) (ComplaintID, error) {
	if err := validateComplaintID(s); err != nil {
		return id.NewID[ComplaintBrand](""), fmt.Errorf("invalid ComplaintID: %w", err)
	}
	return id.NewID[ComplaintBrand](s), nil
}

// MustParseComplaintID parses a ComplaintID from string, panicking on error.
func MustParseComplaintID(s string) ComplaintID {
	complaintID, err := ParseComplaintID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid ComplaintID: %v", err))
	}
	return complaintID
}

// validateComplaintID validates ComplaintID format.
func validateComplaintID(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}
	if !complaintIDPattern.MatchString(s) {
		return errors.New("must be valid UUID v4 format")
	}
	return nil
}

// IsValidComplaintID returns true if the string is a valid ComplaintID.
func IsValidComplaintID(s string) bool {
	return validateComplaintID(s) == nil
}

// ResolutionState represents the resolution state of a complaint.
type ResolutionState string

const (
	ResolutionStateOpen     ResolutionState = "open"
	ResolutionStateResolved ResolutionState = "resolved"
)

// IsResolved returns true if the complaint is resolved.
func (r ResolutionState) IsResolved() bool {
	return r == ResolutionStateResolved
}

// Severity represents the severity level of a complaint.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Complaint represents a structured complaint with branded ID.
type Complaint struct {
	ID              ComplaintID     `json:"id"`
	AgentID         AgentID         `json:"agent_id"`
	SessionID       SessionID       `json:"session_id"`
	ProjectName     ProjectID       `json:"project_id"`
	TaskDescription string          `json:"task_description"`
	ContextInfo     string          `json:"context_info"`
	MissingInfo     string          `json:"missing_info"`
	ConfusedBy      string          `json:"confused_by"`
	FutureWishes    string          `json:"future_wishes"`
	Severity        Severity        `json:"severity"`
	Timestamp       time.Time       `json:"timestamp"`
	ResolutionState ResolutionState `json:"resolution_state"`
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty"`
	ResolvedBy      string          `json:"resolved_by,omitempty"`
}

// Validate checks if all fields are valid.
func (c *Complaint) Validate() error {
	// Validate branded ID
	if c.ID.IsZero() {
		return errors.New("complaint ID cannot be empty")
	}
	if err := validateComplaintID(c.ID.Get()); err != nil {
		return fmt.Errorf("invalid ComplaintID: %w", err)
	}

	// Validate severity
	switch c.Severity {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		// Valid severity
	default:
		return fmt.Errorf("invalid severity: %s", c.Severity)
	}

	// Validate optional branded types (only validate format if not empty)
	if !c.AgentID.IsZero() {
		if err := validateAgentID(c.AgentID.Get()); err != nil {
			return fmt.Errorf("invalid AgentID: %w", err)
		}
	}

	if !c.SessionID.IsZero() {
		if err := validateSessionID(c.SessionID.Get()); err != nil {
			return fmt.Errorf("invalid SessionID: %w", err)
		}
	}

	if !c.ProjectName.IsZero() {
		if err := validateProjectID(c.ProjectName.Get()); err != nil {
			return fmt.Errorf("invalid ProjectID: %w", err)
		}
	}

	if c.TaskDescription == "" {
		return errors.New("task description is required")
	}

	return nil
}

// IsValid returns true if the complaint is valid.
func (c *Complaint) IsValid() bool {
	return c.Validate() == nil
}

// Resolve marks a complaint as resolved.
func (c *Complaint) Resolve(resolvedBy string) error {
	if resolvedBy == "" {
		return errors.New("resolver name cannot be empty")
	}

	// Idempotent: if already resolved, just update resolvedBy if different
	if c.ResolutionState.IsResolved() {
		if c.ResolvedBy != resolvedBy {
			c.ResolvedBy = resolvedBy
			return nil
		}
		return nil // Already resolved with same resolver
	}

	now := time.Now()
	c.ResolvedAt = &now
	c.ResolvedBy = resolvedBy
	c.ResolutionState = ResolutionStateResolved

	return nil
}

// IsResolved returns true if the complaint is resolved.
func (c *Complaint) IsResolved() bool {
	return c.ResolutionState.IsResolved()
}
