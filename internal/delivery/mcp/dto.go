package delivery

import (
	"time"

	"github.com/larsartmann/complaints-mcp/internal/domain"
)

// ComplaintDTO represents a type-safe transfer object for complaint data
type ComplaintDTO struct {
	ID              string     `json:"id"`
	AgentName       string     `json:"agent_name"`
	SessionName     string     `json:"session_name,omitempty"`
	TaskDescription string     `json:"task_description"`
	ContextInfo     string     `json:"context_info,omitempty"`
	MissingInfo     string     `json:"missing_info,omitempty"`
	ConfusedBy      string     `json:"confused_by,omitempty"`
	FutureWishes    string     `json:"future_wishes,omitempty"`
	Severity        string     `json:"severity"`
	Timestamp       time.Time  `json:"timestamp"`
	ProjectName     string     `json:"project_name,omitempty"`
	Resolved        bool       `json:"resolved"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy      string     `json:"resolved_by,omitempty"`
	FilePath        string     `json:"file_path,omitempty"`
	DocsPath        string     `json:"docs_path,omitempty"`
}

// ToDTO converts a domain Complaint to a type-safe DTO (standalone function)
func ToDTO(c *domain.Complaint) ComplaintDTO {
	return ToDTOWithPaths(c, "", "")
}

// ToDTOWithPaths converts a domain Complaint to a type-safe DTO with optional file paths
func ToDTOWithPaths(c *domain.Complaint, filePath, docsPath string) ComplaintDTO {
	return ComplaintDTO{
		ID:              c.ID.String(),
		AgentName:       c.AgentID,
		SessionName:     c.SessionID,
		TaskDescription: c.TaskDescription,
		ContextInfo:     c.ContextInfo,
		MissingInfo:     c.MissingInfo,
		ConfusedBy:      c.ConfusedBy,
		FutureWishes:    c.FutureWishes,
		Severity:        string(c.Severity),
		Timestamp:       c.Timestamp,
		ProjectName:     c.ProjectName,
		Resolved:        c.IsResolved(),
		ResolvedAt:      c.ResolvedAt,
		ResolvedBy:      c.ResolvedBy,
		FilePath:        filePath,
		DocsPath:        docsPath,
	}
}
