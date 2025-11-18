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
//
// IMPORTANT: Complaint should only be used as a pointer (*Complaint) due to the embedded sync.RWMutex.
// Copying Complaint by value is undefined behavior if the mutex has been used.
type Complaint struct {
	ID              ComplaintID `json:"id" validate:"required"`
	AgentName       AgentName   `json:"agent_name"`   // Value object: required, max 100 chars
	SessionName     SessionName `json:"session_name"` // Value object: optional, max 100 chars
	TaskDescription string      `json:"task_description" validate:"required,min=1,max=1000"`
	ContextInfo     string      `json:"context_info" validate:"max=2000000"`
	MissingInfo     string      `json:"missing_info" validate:"max=2000000"`
	ConfusedBy      string      `json:"confused_by" validate:"max=2000000"`
	FutureWishes    string      `json:"future_wishes" validate:"max=2000000"`
	Severity        Severity    `json:"severity" validate:"required,oneof=low medium high critical"`
	Timestamp       time.Time   `json:"timestamp"`
	ProjectName     ProjectName `json:"project_name"` // Value object: optional, max 100 chars

	// Resolution tracking (single source of truth)
	// ResolvedAt is nil when not resolved, non-nil when resolved
	// This eliminates split-brain state - use IsResolved() to check status
	ResolvedAt *time.Time `json:"resolved_at,omitempty"` // nil = not resolved, non-nil = resolved
	ResolvedBy string     `json:"resolved_by,omitempty"` // Who resolved it (empty when not resolved)

	// Thread safety for concurrent resolution attempts
	mu sync.RWMutex `json:"-"` // Don't serialize mutex
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

// Resolve marks a complaint as resolved - thread-safe with proper error handling
func (c *Complaint) Resolve(resolvedBy string) error {
	// Use write lock to ensure thread-safe resolution
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already resolved (fixes issue #37)
	if c.ResolvedAt != nil {
		return fmt.Errorf("complaint already resolved by %s at %s", c.ResolvedBy, c.ResolvedAt.Format(time.RFC3339))
	}

	// Validate resolver name
	if resolvedBy == "" {
		return fmt.Errorf("resolver name cannot be empty")
	}

	// Set resolution with timestamp (single source of truth)
	now := time.Now()
	c.ResolvedAt = &now       // Set resolution timestamp
	c.ResolvedBy = resolvedBy // Set who resolved it

	return nil
}

// IsResolved checks if the complaint is resolved - thread-safe
// Returns true if ResolvedAt is non-nil (single source of truth)
func (c *Complaint) IsResolved() bool {
	// Use read lock for thread-safe resolution status check
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ResolvedAt != nil
}

// Validate checks if the complaint data is valid using validator
func (c *Complaint) Validate() error {
	// Validate all value objects to enforce their invariants
	if err := c.AgentName.Validate(); err != nil {
		return fmt.Errorf("agent name validation failed: %w", err)
	}

	if err := c.SessionName.Validate(); err != nil {
		return fmt.Errorf("session name validation failed: %w", err)
	}

	if err := c.ProjectName.Validate(); err != nil {
		return fmt.Errorf("project name validation failed: %w", err)
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
