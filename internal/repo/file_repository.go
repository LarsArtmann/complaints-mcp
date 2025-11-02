package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

// Repository defines the interface for complaint storage
type Repository interface {
	Save(ctx context.Context, complaint *domain.Complaint) error
	FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error)
	FindAll(ctx context.Context, limit, offset int) ([]*domain.Complaint, error)
	FindBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error)
	FindByProject(ctx context.Context, projectName string, limit int) ([]*domain.Complaint, error)
	FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error)
	Update(ctx context.Context, complaint *domain.Complaint) error
	Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error)
}

// FileRepository implements Repository using file system storage
type FileRepository struct {
	baseDir string
	logger  *log.Logger
	tracer  tracing.Tracer
}

// NewFileRepository creates a new file-based repository
func NewFileRepository(baseDir string, tracer tracing.Tracer) Repository {
	return &FileRepository{
		baseDir: baseDir,
		logger:  log.Default(),
		tracer:  tracer,
	}
}

// Save saves a complaint to the file system
func (r *FileRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
	ctx, span := r.tracer.Start(ctx, "Save")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "complaint_id", complaint.ID.String())
	logger.Info("Saving complaint")

	// ✅ FIX: Use UUID as primary filename component to prevent collisions
	// Old: timestamp-session.json (collision if same second + same session)
	// New: uuid-timestamp.json (guaranteed unique)
	timestamp := complaint.Timestamp.Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s-%s.json", complaint.ID.String(), timestamp)

	filePath := filepath.Join(r.baseDir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		logger.Error("Failed to create directory", "error", err, "path", filepath.Dir(filePath))
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write complaint as JSON
	data, err := json.MarshalIndent(complaint, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal complaint", "error", err)
		return fmt.Errorf("failed to marshal complaint: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		logger.Error("Failed to write complaint file", "error", err, "path", filePath)
		return fmt.Errorf("failed to write complaint file: %w", err)
	}

	logger.Info("Complaint saved successfully", "path", filePath)
	return nil
}

// FindByID retrieves a complaint by ID from the file system
func (r *FileRepository) FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindByID")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "complaint_id", id.String())
	logger.Debug("Finding complaint by ID")

	// Search through files
	complaints, err := r.loadAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	for _, complaint := range complaints {
		if complaint.ID.String() == id.String() {
			logger.Info("Complaint found by ID")
			return complaint, nil
		}
	}

	logger.Warn("Complaint not found", "complaint_id", id.String())
	return nil, fmt.Errorf("complaint not found: %s", id.String())
}

// FindAll retrieves all complaints with pagination
func (r *FileRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindAll")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "limit", limit, "offset", offset)
	logger.Debug("Finding all complaints")

	complaints, err := r.loadAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	// Apply pagination
	total := len(complaints)
	if offset >= total {
		return []*domain.Complaint{}, nil
	}

	start := offset
	end := offset + limit
	if end > total {
		end = total
	}

	result := complaints[start:end]
	logger.Info("Complaints retrieved", "count", len(result))
	return result, nil
}

// FindBySeverity retrieves complaints by severity level
func (r *FileRepository) FindBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindBySeverity")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "severity", string(severity), "limit", limit)
	logger.Debug("Finding complaints by severity")

	complaints, err := r.loadAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Complaint
	count := 0
	for _, complaint := range complaints {
		if complaint.Severity == severity {
			filtered = append(filtered, complaint)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	logger.Info("Complaints filtered by severity", "severity", string(severity), "count", len(filtered))
	return filtered, nil
}

// Update updates an existing complaint
func (r *FileRepository) Update(ctx context.Context, complaint *domain.Complaint) error {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "complaint_id", complaint.ID.String())
	logger.Info("Updating complaint")

	// Find existing complaint
	existing, err := r.FindByID(ctx, complaint.ID)
	if err != nil {
		return fmt.Errorf("failed to find existing complaint: %w", err)
	}

	// Update fields
	existing.Resolved = complaint.Resolved
	existing.TaskDescription = complaint.TaskDescription
	existing.ContextInfo = complaint.ContextInfo
	existing.MissingInfo = complaint.MissingInfo
	existing.ConfusedBy = complaint.ConfusedBy
	existing.FutureWishes = complaint.FutureWishes

	// Save updated complaint
	return r.Save(ctx, existing)
}

// Search searches complaints by text content
func (r *FileRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "Search")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "query", query, "limit", limit)
	logger.Debug("Searching complaints")

	complaints, err := r.loadAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var results []*domain.Complaint
	count := 0

	for _, complaint := range complaints {
		// Simple text search in relevant fields
		if strings.Contains(strings.ToLower(complaint.TaskDescription), queryLower) ||
			strings.Contains(strings.ToLower(complaint.ContextInfo), queryLower) ||
			strings.Contains(strings.ToLower(complaint.MissingInfo), queryLower) ||
			strings.Contains(strings.ToLower(complaint.ConfusedBy), queryLower) ||
			strings.Contains(strings.ToLower(complaint.FutureWishes), queryLower) ||
			strings.Contains(strings.ToLower(complaint.AgentName), queryLower) ||
			strings.Contains(strings.ToLower(complaint.ProjectName), queryLower) {

			results = append(results, complaint)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	logger.Info("Complaints searched", "query", query, "count", len(results))
	return results, nil
}

// loadAllComplaints loads all complaints from the file system
func (r *FileRepository) loadAllComplaints(ctx context.Context) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "loadAllComplaints")
	defer span.End()

	logger := r.logger.With("component", "file-repository")
	logger.Debug("Loading all complaints")

	entries, err := os.ReadDir(r.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("Base directory does not exist, returning empty list")
			return []*domain.Complaint{}, nil
		}
		logger.Error("Failed to read base directory", "error", err, "path", r.baseDir)
		return nil, fmt.Errorf("failed to read base directory: %w", err)
	}

	var complaints []*domain.Complaint
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(r.baseDir, entry.Name())
		complaint, err := r.loadComplaintFromFile(filePath)
		if err != nil {
			logger.Warn("Failed to load complaint file", "error", err, "path", filePath)
			continue
		}

		complaints = append(complaints, complaint)
	}

	logger.Info("Complaints loaded successfully", "count", len(complaints))
	return complaints, nil
}

// loadComplaintFromFile loads a single complaint from a JSON file
func (r *FileRepository) loadComplaintFromFile(filePath string) (*domain.Complaint, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var complaint domain.Complaint
	if err := json.Unmarshal(data, &complaint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal complaint: %w", err)
	}

	return &complaint, nil
}

// FindByProject retrieves complaints by project name
func (r *FileRepository) FindByProject(ctx context.Context, projectName string, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindByProject")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "project_name", projectName, "limit", limit)
	logger.Debug("Finding complaints by project")

	complaints, err := r.loadAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Complaint
	count := 0
	for _, complaint := range complaints {
		if complaint.ProjectName == projectName {
			filtered = append(filtered, complaint)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	logger.Info("Complaints filtered by project", "project_name", projectName, "count", len(filtered))
	return filtered, nil
}

// FindUnresolved retrieves unresolved complaints
func (r *FileRepository) FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindUnresolved")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "limit", limit)
	logger.Debug("Finding unresolved complaints")

	complaints, err := r.loadAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Complaint
	count := 0
	for _, complaint := range complaints {
		if !complaint.Resolved {
			filtered = append(filtered, complaint)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	logger.Info("Unresolved complaints filtered", "count", len(filtered))
	return filtered, nil
}
