package service

import (
	"context"
	"fmt"
	"time"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

// ComplaintService handles complaint business logic
type ComplaintService struct {
	repo   repo.Repository
	tracer tracing.Tracer
}

// NewComplaintService creates a new complaint service
func NewComplaintService(repository repo.Repository, tracer tracing.Tracer) *ComplaintService {
	return &ComplaintService{
		repo:   repository,
		tracer: tracer,
	}
}

// CreateComplaint creates a new complaint
func (s *ComplaintService) CreateComplaint(ctx context.Context, agentName, sessionName, taskDescription, contextInfo, missingInfo, confusedBy, futureWishes string, severity domain.Severity, projectName string) (*domain.Complaint, error) {
	// Generate phantom type ID
	id, err := domain.NewComplaintID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate complaint ID: %w", err)
	}

	// Parse phantom types from strings
	agentID, err := domain.ParseAgentID(agentName)
	if err != nil {
		return nil, fmt.Errorf("invalid agent name: %w", err)
	}

	sessionID, err := domain.ParseSessionID(sessionName)
	if err != nil {
		return nil, fmt.Errorf("invalid session name: %w", err)
	}

	projectID, err := domain.ParseProjectID(projectName)
	if err != nil {
		return nil, fmt.Errorf("invalid project name: %w", err)
	}

	// Create complaint with phantom type ID
	complaint := &domain.Complaint{
		ID:              id,
		AgentID:         agentID,
		SessionID:       sessionID,
		ProjectName:     projectID,
		TaskDescription: taskDescription,
		ContextInfo:     contextInfo,
		MissingInfo:     missingInfo,
		ConfusedBy:      confusedBy,
		FutureWishes:    futureWishes,
		Severity:        severity,
		Timestamp:       time.Now(),
		ResolutionState: domain.ResolutionStateOpen,
	}

	if err := complaint.Validate(); err != nil {
		return nil, fmt.Errorf("invalid complaint: %w", err)
	}

	if err := s.repo.Save(ctx, complaint); err != nil {
		return nil, fmt.Errorf("failed to save complaint: %w", err)
	}

	return complaint, nil
}

// GetComplaint retrieves a complaint by ID
func (s *ComplaintService) GetComplaint(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	return s.repo.FindByID(ctx, id)
}

// ListComplaints retrieves a list of complaints
func (s *ComplaintService) ListComplaints(ctx context.Context, limit, offset int) ([]*domain.Complaint, error) {
	return s.repo.FindAll(ctx, limit, offset)
}

// ResolveComplaint marks a complaint as resolved
func (s *ComplaintService) ResolveComplaint(ctx context.Context, id domain.ComplaintID, resolvedBy string) (*domain.Complaint, error) {
	complaint, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find complaint: %w", err)
	}

	if err := complaint.Resolve(resolvedBy); err != nil {
		return nil, fmt.Errorf("failed to resolve complaint: %w", err)
	}

	if err := s.repo.Update(ctx, complaint); err != nil {
		return nil, fmt.Errorf("failed to update complaint: %w", err)
	}

	return complaint, nil
}

// GetFilePaths returns file and docs paths for a complaint
func (s *ComplaintService) GetFilePaths(ctx context.Context, id domain.ComplaintID) (filePath, docsPath string, err error) {
	filePath, err = s.repo.GetFilePath(ctx, id)
	if err != nil {
		return "", "", fmt.Errorf("failed to get file path: %w", err)
	}

	docsPath, err = s.repo.GetDocsPath(ctx, id)
	if err != nil {
		return "", "", fmt.Errorf("failed to get docs path: %w", err)
	}

	return filePath, docsPath, nil
}

// GetComplaintsBySeverity retrieves complaints by severity level
func (s *ComplaintService) GetComplaintsBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error) {
	return s.repo.FindBySeverity(ctx, severity, limit)
}

// SearchComplaints searches complaints by text query
func (s *ComplaintService) SearchComplaints(ctx context.Context, query string, limit int) ([]*domain.Complaint, error) {
	return s.repo.Search(ctx, query, limit)
}

// GetCacheStats returns cache statistics
func (s *ComplaintService) GetCacheStats() repo.CacheStats {
	return s.repo.GetCacheStats()
}

// ListComplaintsByProject retrieves complaints by project name
func (s *ComplaintService) ListComplaintsByProject(ctx context.Context, projectName string, limit int) ([]*domain.Complaint, error) {
	return s.repo.FindByProject(ctx, projectName, limit)
}

// ListUnresolvedComplaints retrieves unresolved complaints
func (s *ComplaintService) ListUnresolvedComplaints(ctx context.Context, limit int) ([]*domain.Complaint, error) {
	return s.repo.FindUnresolved(ctx, limit)
}
