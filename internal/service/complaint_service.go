package service

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

// ComplaintService provides business logic for managing complaints
type ComplaintService struct {
	repo     repo.Repository
	docsRepo *repo.DocsRepository
	logger   *log.Logger
	tracer   tracing.Tracer
}

// NewComplaintService creates a new complaint service
func NewComplaintService(repo repo.Repository, tracer tracing.Tracer, logger *log.Logger) *ComplaintService {
	return &ComplaintService{
		repo:   repo,
		logger: logger,
		tracer: tracer,
		// DocsRepository will be set by UpdateConfig method after config is loaded
	}
}

// UpdateConfig updates service configuration including docs repository
func (s *ComplaintService) UpdateConfig(docsDir, docsFormat string, docsEnabled bool) {
	s.docsRepo = repo.NewDocsRepository(docsDir, docsFormat, docsEnabled, s.logger, s.tracer)
}

// CreateComplaint creates a new complaint
func (s *ComplaintService) CreateComplaint(ctx context.Context, agentName, sessionName, taskDescription, contextInfo, missingInfo, confusedBy, futureWishes string, severity domain.Severity, projectName string) (*domain.Complaint, error) {
	ctx, span := s.tracer.Start(ctx, "CreateComplaint")
	defer span.End()

	logger := s.logger.With("operation", "create_complaint")
	logger.Info("Creating new complaint",
		"agent_name", agentName,
		"severity", string(severity),
		"session_name", sessionName)

	complaint, err := domain.NewComplaint(ctx, agentName, sessionName, taskDescription, contextInfo, missingInfo, confusedBy, futureWishes, severity, projectName)
	if err != nil {
		logger.Error("Failed to create complaint", "error", err)
		return nil, err
	}

	if err := s.repo.Save(ctx, complaint); err != nil {
		logger.Error("Failed to save complaint", "error", err, "complaint_id", complaint.ID.String())
		return nil, fmt.Errorf("failed to save complaint: %w", err)
	}

	// Export to documentation format (markdown by default)
	if s.docsRepo != nil {
		if err := s.docsRepo.ExportToDocs(complaint); err != nil {
			logger.Warn("Failed to export complaint to documentation", "error", err, "complaint_id", complaint.ID.String())
			// Don't fail the entire operation if docs export fails
		}
	}

	logger.Info("Complaint created successfully", "complaint_id", complaint.ID.String())
	return complaint, nil
}

// GetComplaint retrieves a complaint by ID
func (s *ComplaintService) GetComplaint(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	ctx, span := s.tracer.Start(ctx, "GetComplaint")
	defer span.End()

	logger := s.logger.With("operation", "get_complaint", "complaint_id", id.String())
	logger.Debug("Retrieving complaint")

	complaint, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.Error("Failed to retrieve complaint", "error", err)
		return nil, fmt.Errorf("failed to retrieve complaint: %w", err)
	}

	if complaint == nil {
		logger.Warn("Complaint not found", "complaint_id", id.String())
		return nil, fmt.Errorf("complaint not found: %s", id.String())
	}

	logger.Info("Complaint retrieved successfully")
	return complaint, nil
}

// ListComplaints retrieves all complaints
func (s *ComplaintService) ListComplaints(ctx context.Context, limit, offset int) ([]*domain.Complaint, error) {
	ctx, span := s.tracer.Start(ctx, "ListComplaints")
	defer span.End()

	logger := s.logger.With("operation", "list_complaints", "limit", limit, "offset", offset)
	logger.Debug("Retrieving complaints list")

	complaints, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to list complaints", "error", err)
		return nil, fmt.Errorf("failed to list complaints: %w", err)
	}

	logger.Info("Complaints listed successfully", "count", len(complaints))
	return complaints, nil
}

// ResolveComplaint marks a complaint as resolved
func (s *ComplaintService) ResolveComplaint(ctx context.Context, id domain.ComplaintID, resolvedBy string) error {
	ctx, span := s.tracer.Start(ctx, "ResolveComplaint")
	defer span.End()

	logger := s.logger.With("operation", "resolve_complaint", "complaint_id", id.String())
	logger.Info("Resolving complaint")

	complaint, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.Error("Failed to retrieve complaint for resolution", "error", err)
		return fmt.Errorf("failed to retrieve complaint: %w", err)
	}

	if complaint == nil {
		logger.Warn("Complaint not found for resolution", "complaint_id", id.String())
		return fmt.Errorf("complaint not found: %s", id.String())
	}

	// Check if already resolved (for idempotency)
	if complaint.IsResolved() {
		logger.Info("Complaint already resolved, returning success for idempotency", "resolved_by", complaint.ResolvedBy, "resolved_at", complaint.ResolvedAt)
		return nil
	}

	if err := complaint.Resolve(resolvedBy); err != nil {
		logger.Error("failed to resolve complaint", "error", err, "complaint_id", id.String())
		return fmt.Errorf("failed to resolve complaint: %w", err)
	}

	if err := s.repo.Update(ctx, complaint); err != nil {
		logger.Error("Failed to update complaint", "error", err)
		return fmt.Errorf("failed to update complaint: %w", err)
	}

	logger.Info("Complaint resolved successfully", "complaint_id", id.String())
	return nil
}

// GetComplaintsBySeverity retrieves complaints by severity level
func (s *ComplaintService) GetComplaintsBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error) {
	ctx, span := s.tracer.Start(ctx, "GetComplaintsBySeverity")
	defer span.End()

	logger := s.logger.With("operation", "get_complaints_by_severity", "severity", string(severity), "limit", limit)
	logger.Debug("Retrieving complaints by severity")

	complaints, err := s.repo.FindBySeverity(ctx, severity, limit)
	if err != nil {
		logger.Error("Failed to retrieve complaints by severity", "error", err)
		return nil, fmt.Errorf("failed to retrieve complaints by severity: %w", err)
	}

	logger.Info("Complaints retrieved by severity successfully", "severity", string(severity), "count", len(complaints))
	return complaints, nil
}

// SearchComplaints searches complaints by text content
func (s *ComplaintService) SearchComplaints(ctx context.Context, query string, limit int) ([]*domain.Complaint, error) {
	ctx, span := s.tracer.Start(ctx, "SearchComplaints")
	defer span.End()

	logger := s.logger.With("operation", "search_complaints", "query", query, "limit", limit)
	logger.Debug("Searching complaints")

	complaints, err := s.repo.Search(ctx, query, limit)
	if err != nil {
		logger.Error("Failed to search complaints", "error", err)
		return nil, fmt.Errorf("failed to search complaints: %w", err)
	}

	logger.Info("Complaints searched successfully", "query", query, "count", len(complaints))
	return complaints, nil
}

// ListComplaintsByProject retrieves complaints by project name
func (s *ComplaintService) ListComplaintsByProject(ctx context.Context, projectName string, limit int) ([]*domain.Complaint, error) {
	ctx, span := s.tracer.Start(ctx, "ListComplaintsByProject")
	defer span.End()

	logger := s.logger.With("operation", "list_complaints_by_project", "project_name", projectName, "limit", limit)
	logger.Debug("Retrieving complaints by project")

	complaints, err := s.repo.FindByProject(ctx, projectName, limit)
	if err != nil {
		logger.Error("Failed to list complaints by project", "error", err)
		return nil, fmt.Errorf("failed to list complaints by project: %w", err)
	}

	logger.Info("Complaints listed by project successfully", "project_name", projectName, "count", len(complaints))
	return complaints, nil
}

// ListUnresolvedComplaints retrieves unresolved complaints
func (s *ComplaintService) ListUnresolvedComplaints(ctx context.Context, limit int) ([]*domain.Complaint, error) {
	ctx, span := s.tracer.Start(ctx, "ListUnresolvedComplaints")
	defer span.End()

	logger := s.logger.With("operation", "list_unresolved_complaints", "limit", limit)
	logger.Debug("Retrieving unresolved complaints")

	complaints, err := s.repo.FindUnresolved(ctx, limit)
	if err != nil {
		logger.Error("Failed to list unresolved complaints", "error", err)
		return nil, fmt.Errorf("failed to list unresolved complaints: %w", err)
	}

	logger.Info("Unresolved complaints listed successfully", "count", len(complaints))
	return complaints, nil
}

// GetCacheStats returns cache performance statistics
func (s *ComplaintService) GetCacheStats() repo.CacheStats {
	return s.repo.GetCacheStats()
}
