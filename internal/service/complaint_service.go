package service

import (
	"context"
	"fmt"
	"time"

	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/errors"
	"github.com/larsartmann/complaints-mcp/internal/repo"
)

// ComplaintService handles business logic for complaints
type ComplaintService struct {
	repo   repo.Repository
	config *config.Config
}

// NewComplaintService creates a new complaint service
func NewComplaintService(repository repo.Repository, cfg *config.Config) *ComplaintService {
	return &ComplaintService{
		repo:   repository,
		config: cfg,
	}
}

// CreateComplaint creates a new complaint
func (s *ComplaintService) CreateComplaint(ctx context.Context, req *CreateComplaintRequest) (*domain.Complaint, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error(), "")
	}

	// Generate complaint ID
	id, err := domain.NewComplaintID()
	if err != nil {
		return nil, errors.NewStorageError("failed to generate complaint ID", err)
	}

	// Use project name from request or fall back to config
	projectName := req.ProjectName
	if projectName == "" {
		projectName = s.config.Complaints.ProjectName
	}

	// Create complaint
	complaint := &domain.Complaint{
		ID:              id,
		AgentName:       req.AgentName,
		SessionName:     req.SessionName,
		TaskDescription: req.TaskDescription,
		ContextInfo:     req.ContextInfo,
		MissingInfo:     req.MissingInfo,
		ConfusedBy:      req.ConfusedBy,
		FutureWishes:    req.FutureWishes,
		Severity:        req.Severity,
		ProjectName:     projectName,
		Timestamp:       time.Now(),
		Resolved:        false,
	}

	// Validate complaint
	if err := complaint.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error(), "")
	}

	// Store complaint
	if err := s.repo.Store(ctx, complaint); err != nil {
		return nil, fmt.Errorf("failed to store complaint: %w", err)
	}

	return complaint, nil
}

// GetComplaint retrieves a complaint by ID
func (s *ComplaintService) GetComplaint(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	complaint, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find complaint: %w", err)
	}
	if complaint == nil {
		return nil, errors.NewNotFoundError(fmt.Sprintf("complaint with ID %s not found", id.Value))
	}
	return complaint, nil
}

// ListComplaintsByProject retrieves complaints for a project
func (s *ComplaintService) ListComplaintsByProject(ctx context.Context, projectName string, limit, offset int) ([]*domain.Complaint, error) {
	if limit <= 0 {
		limit = 50 // default limit
	}
	if limit > 100 {
		limit = 100 // max limit
	}

	complaints, err := s.repo.FindByProject(ctx, projectName, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list complaints for project %s: %w", projectName, err)
	}
	return complaints, nil
}

// ListUnresolvedComplaints retrieves unresolved complaints
func (s *ComplaintService) ListUnresolvedComplaints(ctx context.Context, limit, offset int) ([]*domain.Complaint, error) {
	if limit <= 0 {
		limit = 50 // default limit
	}
	if limit > 100 {
		limit = 100 // max limit
	}

	complaints, err := s.repo.FindUnresolved(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list unresolved complaints: %w", err)
	}
	return complaints, nil
}

// ResolveComplaint marks a complaint as resolved
func (s *ComplaintService) ResolveComplaint(ctx context.Context, id domain.ComplaintID) error {
	if err := s.repo.MarkResolved(ctx, id); err != nil {
		return fmt.Errorf("failed to resolve complaint %s: %w", id.Value, err)
	}
	return nil
}

// CreateComplaintRequest represents a request to create a complaint
type CreateComplaintRequest struct {
	AgentName       string `json:"agent_name" validate:"required"`
	SessionName     string `json:"session_name"`
	TaskDescription string `json:"task_description" validate:"required"`
	ContextInfo     string `json:"context_info"`
	MissingInfo     string `json:"missing_info"`
	ConfusedBy      string `json:"confused_by"`
	FutureWishes    string `json:"future_wishes"`
	Severity        string `json:"severity" validate:"required,oneof=low medium high critical"`
	ProjectName     string `json:"project_name"`
}

// Validate validates the request
func (r *CreateComplaintRequest) Validate() error {
	if r.AgentName == "" {
		return fmt.Errorf("agent name is required")
	}
	if r.TaskDescription == "" {
		return fmt.Errorf("task description is required")
	}
	if r.Severity == "" {
		return fmt.Errorf("severity is required")
	}
	validSeverities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}
	if !validSeverities[r.Severity] {
		return fmt.Errorf("invalid severity: %s", r.Severity)
	}
	return nil
}