package delivery

import (
	"time"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
)

// ComplaintDTO represents a type-safe transfer object for complaint data.
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
	ProjectID       string     `json:"project_id,omitempty"`
	Resolved        bool       `json:"resolved"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy      string     `json:"resolved_by,omitempty"`
	FilePath        string     `json:"file_path,omitempty"`
	DocsPath        string     `json:"docs_path,omitempty"`
}

// ToDTO converts a domain Complaint to a type-safe DTO (standalone function).
func ToDTO(c *domain.Complaint) ComplaintDTO {
	return ToDTOWithPaths(c, "", "")
}

// ToDTOWithPaths converts a domain Complaint to a type-safe DTO with optional file paths.
func ToDTOWithPaths(c *domain.Complaint, filePath, docsPath string) ComplaintDTO {
	return ComplaintDTO{
		ID:              c.ID.String(),
		AgentName:       c.AgentID.String(),
		SessionName:     c.SessionID.String(),
		TaskDescription: c.TaskDescription,
		ContextInfo:     c.ContextInfo,
		MissingInfo:     c.MissingInfo,
		ConfusedBy:      c.ConfusedBy,
		FutureWishes:    c.FutureWishes,
		Severity:        string(c.Severity),
		Timestamp:       c.Timestamp,
		ProjectID:       c.ProjectID.String(),
		Resolved:        c.IsResolved(),
		ResolvedAt:      c.ResolvedAt,
		ResolvedBy:      c.ResolvedBy,
		FilePath:        filePath,
		DocsPath:        docsPath,
	}
}

// Request DTOs for MCP tool inputs.

// FileComplaintRequest represents the input for filing a complaint.
type FileComplaintRequest struct {
	AgentName       string `json:"agent_name"       validate:"required,min=1,max=100"`
	SessionName     string `json:"session_name"     validate:"required,min=1,max=100"`
	TaskDescription string `json:"task_description" validate:"required,min=1,max=5000"`
	ContextInfo     string `json:"context_info"     validate:"max=5000"`
	MissingInfo     string `json:"missing_info"     validate:"max=2000"`
	ConfusedBy      string `json:"confused_by"      validate:"max=2000"`
	FutureWishes    string `json:"future_wishes"    validate:"max=2000"`
	Severity        string `json:"severity"         validate:"required,oneof=low medium high critical"`
	ProjectID       string `json:"project_id"       validate:"omitempty,min=1,max=100"`
	WorkingDir      string `json:"working_dir"      validate:"omitempty,max=500"`
}

// ListComplaintsRequest represents the input for listing complaints.
type ListComplaintsRequest struct {
	Limit    int    `json:"limit"              validate:"gte=1,lte=100"`
	Severity string `json:"severity"           validate:"omitempty,oneof=low medium high critical"`
	Resolved *bool  `json:"resolved,omitempty"`
}

// ResolveComplaintRequest represents the input for resolving a complaint.
type ResolveComplaintRequest struct {
	ComplaintID string `json:"complaint_id" validate:"required,uuid4"`
	ResolvedBy  string `json:"resolved_by"  validate:"required,min=1,max=100"`
}

// SearchComplaintsRequest represents the input for searching complaints.
type SearchComplaintsRequest struct {
	Query string `json:"query" validate:"required,min=1,max=500"`
	Limit int    `json:"limit" validate:"gte=1,lte=100"`
}

// Response DTOs for MCP tool outputs.

// FileComplaintResponse represents the output after filing a complaint.
type FileComplaintResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Complaint ComplaintDTO `json:"complaint"`
}

// ListComplaintsResponse represents the output for listing complaints.
type ListComplaintsResponse struct {
	Complaints []ComplaintDTO `json:"complaints"`
	Count      int            `json:"count"`
}

// ResolveComplaintResponse represents the output after resolving a complaint.
type ResolveComplaintResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Complaint ComplaintDTO `json:"complaint"`
}

// SearchComplaintsResponse represents the output for searching complaints.
type SearchComplaintsResponse struct {
	Complaints []ComplaintDTO `json:"complaints"`
	Query      string         `json:"query"`
	Count      int            `json:"count"`
}

// CacheStatsResponse represents the output for cache statistics.
type CacheStatsResponse struct {
	CacheEnabled bool            `json:"cache_enabled"`
	Stats        repo.CacheStats `json:"stats"`
	Message      string          `json:"message"`
}
