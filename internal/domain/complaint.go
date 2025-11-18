package domain

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
)

var (
	// Global validator instance (thread-safe, created once)
	validate *validator.Validate
	// Initialize validator once using sync.Once (thread-safe singleton pattern)
	validateOnce sync.Once
)

// ResolutionState represents the resolution status of a complaint
// Provides single source of truth for resolution state
type ResolutionState string

const (
	ResolutionStateOpen     ResolutionState = "open"
	ResolutionStateResolved ResolutionState = "resolved"
	ResolutionStateRejected ResolutionState = "rejected"
	ResolutionStateDeferred ResolutionState = "deferred"
)

// AllResolutionStates returns all valid resolution states
func AllResolutionStates() []ResolutionState {
	return []ResolutionState{
		ResolutionStateOpen,
		ResolutionStateResolved,
		ResolutionStateRejected,
		ResolutionStateDeferred,
	}
}

// IsValid checks if the resolution state is valid
func (rs ResolutionState) IsValid() bool {
	switch rs {
	case ResolutionStateOpen, ResolutionStateResolved, ResolutionStateRejected, ResolutionStateDeferred:
		return true
	default:
		return false
	}
}

// IsResolved returns true if the complaint is in a resolved state
func (rs ResolutionState) IsResolved() bool {
	return rs == ResolutionStateResolved
}

// CanTransitionTo checks if state transition is allowed
func (rs ResolutionState) CanTransitionTo(new ResolutionState) bool {
	// Define allowed transitions for state machine
	allowedTransitions := map[ResolutionState][]ResolutionState{
		ResolutionStateOpen:     {ResolutionStateResolved, ResolutionStateRejected, ResolutionStateDeferred},
		ResolutionStateDeferred: {ResolutionStateResolved, ResolutionStateRejected},
		ResolutionStateResolved: {}, // Terminal state
		ResolutionStateRejected: {}, // Terminal state
	}

	allowedStates, exists := allowedTransitions[rs]
	if !exists {
		return false
	}

	return slices.Contains(allowedStates, new)
}

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
// NOTE: The mutex now lives in ThreadSafeComplaint. Using *Complaint is recommended only for
// size/identity semantics, not for thread safety.
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

	// Resolution tracking - SINGLE SOURCE OF TRUTH
	// ResolutionState captures the complete resolution status
	ResolutionState ResolutionState `json:"resolution_state"`
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty"` // nil = not resolved
	ResolvedBy      string          `json:"resolved_by,omitempty"` // empty when not resolved
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
		ResolutionState: ResolutionStateOpen, // Start as open
		// ResolvedAt/ResolvedBy are nil/empty by default
	}

	// Validate the complaint (pure domain logic)
	if err := complaint.Validate(); err != nil {
		return nil, err
	}

	// Complaint created - service layer should handle logging
	return complaint, nil
}

// Resolve marks a complaint as resolved - NOT THREAD SAFE
// Use ThreadSafeComplaint wrapper for concurrent operations
func (c *Complaint) Resolve(resolvedBy string) error {
	// Check if state transition is allowed
	if !c.ResolutionState.CanTransitionTo(ResolutionStateResolved) {
		currentState := string(c.ResolutionState)
		return fmt.Errorf("cannot resolve complaint in state '%s', already resolved or invalid transition", currentState)
	}

	// Validate resolver name
	if resolvedBy == "" {
		return fmt.Errorf("resolver name cannot be empty")
	}

	// Set resolution with timestamp and update state
	now := time.Now()
	c.ResolvedAt = &now
	c.ResolvedBy = resolvedBy
	c.ResolutionState = ResolutionStateResolved // Update state machine

	return nil
}

// IsResolved checks if the complaint is resolved - NOT THREAD SAFE
// Returns true if ResolutionState is resolved (single source of truth)
func (c *Complaint) IsResolved() bool {
	return c.ResolutionState.IsResolved()
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

	// Validate resolution state
	if !c.ResolutionState.IsValid() {
		return fmt.Errorf("invalid resolution state: %s", c.ResolutionState)
	}

	// Validate consistency between ResolutionState and ResolvedAt
	if c.ResolutionState.IsResolved() != (c.ResolvedAt != nil) {
		return fmt.Errorf("resolution state inconsistent with resolved timestamp")
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
