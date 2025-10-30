package domain

import (
	"fmt"
	"time"
	"context"
	
	"github.com/gofrs/uuid"
	"github.com/go-playground/validator/v10"
	"github.com/charmbracelet/log"
)

// Severity represents the severity level of a complaint
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// AllSeverities returns all valid severity levels
func AllSeverities() []Severity {
	return []Severity{SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical}
}

// IsValid checks if the severity is valid
func (s Severity) IsValid() bool {
	switch s {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		return true
	default:
		return false
	}
}

// ComplaintID represents a unique identifier for a complaint
type ComplaintID struct {
	Value string
}

// NewComplaintID creates a new complaint ID
func NewComplaintID() (ComplaintID, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return ComplaintID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	return ComplaintID{Value: id.String()}, nil
}

// String returns the string representation of the complaint ID
func (id ComplaintID) String() string {
	return id.Value
}

// IsEmpty checks if the complaint ID is empty
func (id ComplaintID) IsEmpty() bool {
	return id.Value == ""
}

// Complaint represents a complaint filed by an AI agent
type Complaint struct {
	ID              ComplaintID   `json:"id" validate:"required"`
	AgentName       string        `json:"agent_name" validate:"required,min=1,max=100"`
	SessionName     string        `json:"session_name" validate:"max=100"`
	TaskDescription  string        `json:"task_description" validate:"required,min=1,max=1000"`
	ContextInfo     string        `json:"context_info" validate:"max=500"`
	MissingInfo     string        `json:"missing_info" validate:"max=500"`
	ConfusedBy      string        `json:"confused_by" validate:"max=500"`
	FutureWishes    string        `json:"future_wishes" validate:"max=500"`
	Severity        Severity      `json:"severity" validate:"required,oneof=low medium high critical"`
	Timestamp       time.Time     `json:"timestamp"`
	ProjectName     string        `json:"project_name" validate:"max=100"`
	Resolved        bool          `json:"resolved"`
}

// NewComplaint creates a new complaint with the given parameters
func NewComplaint(ctx context.Context, agentName, sessionName, taskDescription, contextInfo, missingInfo, confusedBy, futureWishes string, severity Severity, projectName string) (*Complaint, error) {
	logger := zerolog.Ctx(ctx)
	
	id, err := NewComplaintID()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate complaint ID")
		return nil, fmt.Errorf("failed to generate complaint ID: %w", err)
	}
	
	if !severity.IsValid() {
		return nil, fmt.Errorf("invalid severity: %s", severity)
	}
	
	now := time.Now()
	complaint := &Complaint{
		ID:              id,
		AgentName:       agentName,
		SessionName:     sessionName,
		TaskDescription:  taskDescription,
		ContextInfo:     contextInfo,
		MissingInfo:     missingInfo,
		ConfusedBy:      confusedBy,
		FutureWishes:    futureWishes,
		Severity:        severity,
		Timestamp:       now,
		ProjectName:     projectName,
		Resolved:        false,
	}
	
	// Validate the complaint
	if err := complaint.Validate(); err != nil {
		logger.Error().Err(err).Interface("complaint", complaint).Msg("Invalid complaint data")
		return nil, err
	}
	
	logger.Info().
		Str("complaint_id", id.String()).
		Str("agent_name", agentName).
		Str("severity", string(severity)).
		Msg("Created new complaint")
	
	return complaint, nil
}

// Resolve marks a complaint as resolved
func (c *Complaint) Resolve(ctx context.Context) {
	logger := zerolog.Ctx(ctx)
	c.Resolved = true
	logger.Info().
		Str("complaint_id", c.ID.String()).
		Msg("Marked complaint as resolved")
}

// IsResolved checks if the complaint is resolved
func (c *Complaint) IsResolved() bool {
	return c.Resolved
}

// Validate checks if the complaint data is valid using validator
func (c *Complaint) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// ValidateLegacy performs legacy validation (kept for backward compatibility)
func (c *Complaint) ValidateLegacy() error {
	if c.AgentName == "" {
		return fmt.Errorf("agent name is required")
	}
	
	if c.TaskDescription == "" {
		return fmt.Errorf("task description is required")
	}
	
	if c.Severity == "" {
		return fmt.Errorf("severity is required")
	}
	
	if !c.Severity.IsValid() {
		return fmt.Errorf("invalid severity: %s", c.Severity)
	}
	
	return nil
}

// GetSummary returns a summary of the complaint
func (c *Complaint) GetSummary() string {
	return fmt.Sprintf("[%s] %s - %s", c.Severity, c.AgentName, c.TaskDescription)
}

// GetDuration returns how long the complaint has been open
func (c *Complaint) GetDuration() time.Duration {
	return time.Since(c.Timestamp)
}