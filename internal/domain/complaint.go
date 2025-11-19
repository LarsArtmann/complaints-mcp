package domain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
)

var (
	// Global validator instance (thread-safe, created once)
	validate     *validator.Validate
	validateOnce sync.Once
)

// Severity represents the severity level of a complaint
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"

	// Content validation limits
	MaxAgentNameLength       = 100
	MaxSessionNameLength     = 100
	MaxTaskDescriptionLength = 1000
	MaxContextInfoLength     = 2000000 // 2MB for testing
	MaxMissingInfoLength     = 2000000 // 2MB for testing
	MaxConfusedByLength      = 2000000 // 2MB for testing
	MaxFutureWishesLength    = 2000000 // 2MB for testing
	MaxProjectNameLength     = 100
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
// Uses value objects for type safety and validation
type Complaint struct {
	ID              ComplaintID `json:"id" validate:"required"`
	AgentName       AgentName   `json:"agent_name"`    // Value object: required, max 100 chars
	SessionName     SessionName `json:"session_name"`  // Value object: optional, max 100 chars
	TaskDescription string      `json:"task_description" validate:"required,min=1,max=1000"`
	ContextInfo     string      `json:"context_info" validate:"max=2000000"`
	MissingInfo     string      `json:"missing_info" validate:"max=2000000"`
	ConfusedBy      string      `json:"confused_by" validate:"max=2000000"`
	FutureWishes    string      `json:"future_wishes" validate:"max=2000000"`
	Severity        Severity    `json:"severity" validate:"required,oneof=low medium high critical"`
	Timestamp       time.Time   `json:"timestamp"`
	ProjectName     ProjectName `json:"project_name"` // Value object: optional, max 100 chars

	// Resolution tracking (SINGLE VALUE OBJECT - NO SPLIT BRAIN!)
	// nil = not resolved, non-nil = resolved with timestamp + who
	// This eliminates split-brain: impossible to have timestamp without resolver or vice versa
	// TODO: Make private when we refactor to immutable Complaint (HIGH-3)
	Resolution *Resolution `json:"resolution,omitempty"` // nil = not resolved, non-nil = resolved
}

// NewComplaint creates a new complaint with the given parameters
// Accepts strings and converts them to value objects for type safety
func NewComplaint(ctx context.Context, agentName, sessionName, taskDescription, contextInfo, missingInfo, confusedBy, futureWishes string, severity Severity, projectName string) (*Complaint, error) {
	// Domain layer is now pure - no external dependencies

	id, err := NewComplaintID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate complaint ID: %w", err)
	}

	if !severity.IsValid() {
		return nil, fmt.Errorf("invalid severity: %s", severity)
	}

	// Create value objects with validation
	agentNameVO, err := NewAgentName(agentName)
	if err != nil {
		return nil, fmt.Errorf("invalid agent name: %w", err)
	}

	sessionNameVO, err := NewSessionName(sessionName)
	if err != nil {
		return nil, fmt.Errorf("invalid session name: %w", err)
	}

	projectNameVO, err := NewProjectName(projectName)
	if err != nil {
		return nil, fmt.Errorf("invalid project name: %w", err)
	}

	now := time.Now()
	complaint := &Complaint{
		ID:              id,
		AgentName:       agentNameVO,
		SessionName:     sessionNameVO,
		TaskDescription: taskDescription,
		ContextInfo:     contextInfo,
		MissingInfo:     missingInfo,
		ConfusedBy:      confusedBy,
		FutureWishes:    futureWishes,
		Severity:        severity,
		Timestamp:       now,
		ProjectName:     projectNameVO,
		// ResolvedAt is nil by default (not resolved)
	}

	// Validate the complaint (pure domain logic)
	if err := complaint.Validate(); err != nil {
		return nil, err
	}

	// Complaint created - service layer should handle logging
	return complaint, nil
}

// Resolve marks a complaint as resolved
// Thread safety is handled at the service/repository layer, not in domain entity
func (c *Complaint) Resolve(resolvedBy AgentName) error {
	// Check if already resolved
	if c.Resolution != nil {
		return fmt.Errorf("complaint already resolved by %s at %s",
			c.Resolution.ResolvedBy().String(),
			c.Resolution.Timestamp().Format(time.RFC3339))
	}

	// Create Resolution value object (automatically uses current time)
	resolution, err := NewResolution(resolvedBy)
	if err != nil {
		return fmt.Errorf("failed to create resolution: %w", err)
	}

	// Set the resolution (single value object - no split brain!)
	c.Resolution = &resolution

	return nil
}

// IsResolved checks if the complaint is resolved
// Returns true if Resolution is non-nil (single source of truth - no split brain!)
func (c *Complaint) IsResolved() bool {
	return c.Resolution != nil
}

// GetResolution returns the resolution if resolved, nil otherwise
// Provides safe access to resolution details
func (c *Complaint) GetResolution() *Resolution {
	return c.Resolution
}

// Validate checks if the complaint data is valid using validator
func (c *Complaint) Validate() error {
	// Check value objects first (they have their own validation)
	if c.AgentName.IsEmpty() {
		return fmt.Errorf("agent name is required")
	}

	// Initialize validator once using sync.Once (thread-safe singleton pattern)
	validateOnce.Do(func() {
		validate = validator.New()
	})
	return validate.Struct(c)
}

// GetSummary returns a summary of the complaint
func (c *Complaint) GetSummary() string {
	return fmt.Sprintf("[%s] %s - %s", c.Severity, c.AgentName.String(), c.TaskDescription)
}

// GetDuration returns how long the complaint has been open
func (c *Complaint) GetDuration() time.Duration {
	return time.Since(c.Timestamp)
}
