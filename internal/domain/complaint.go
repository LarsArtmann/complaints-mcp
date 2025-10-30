package domain

import (
	"time"
	"github.com/gofrs/uuid"
)

// ComplaintID represents a unique identifier for a complaint
type ComplaintID struct {
	Value string
}

// NewComplaintID creates a new complaint ID
func NewComplaintID() ComplaintID {
	return ComplaintID{Value: uuid.Must(uuid.NewV4()).String()}
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
	AgentName       string        `json:"agent_name" validate:"required"`
	SessionName     string        `json:"session_name"`
	TaskDescription  string        `json:"task_description" validate:"required"`
	ContextInfo     string        `json:"context_info"`
	MissingInfo     string        `json:"missing_info"`
	ConfusedBy      string        `json:"confused_by"`
	FutureWishes    string        `json:"future_wishes"`
	Severity        string        `json:"severity" validate:"required,oneof=low medium high critical"`
	Timestamp       time.Time     `json:"timestamp"`
	ProjectName     string        `json:"project_name"`
	Resolved        bool          `json:"resolved"`
}

// NewComplaint creates a new complaint with the given parameters
func NewComplaint(agentName, sessionName, taskDescription, contextInfo, missingInfo, confusedBy, futureWishes, severity, projectName string) *Complaint {
	now := time.Now()
	return &Complaint{
		ID:              NewComplaintID(),
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
}

// Resolve marks a complaint as resolved
func (c *Complaint) Resolve() {
	c.Resolved = true
}

// IsResolved checks if the complaint is resolved
func (c *Complaint) IsResolved() bool {
	return c.Resolved
}

// Validate checks if the complaint data is valid
func (c *Complaint) Validate() error {
	if c.AgentName == "" {
		return fmt.Errorf("agent name is required")
	}
	
	if c.TaskDescription == "" {
		return fmt.Errorf("task description is required")
	}
	
	if c.Severity == "" {
		return fmt.Errorf("severity is required")
	}
	
	validSeverities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}
	
	if !validSeverities[c.Severity] {
		return fmt.Errorf("invalid severity: %s", c.Severity)
	}
	
	return nil
}