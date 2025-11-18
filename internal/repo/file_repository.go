package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

// FileRepository implements Repository using file system storage
type FileRepository struct {
	BaseRepository
}

// NewFileRepository creates a new file-based repository
func NewFileRepository(baseDir string, tracer tracing.Tracer) Repository {
	return &FileRepository{
		BaseRepository: NewBaseRepository(baseDir, tracer),
	}
}

// WarmCache is a no-op for FileRepository (no cache to warm)
func (r *FileRepository) WarmCache(ctx context.Context) error {
	// FileRepository doesn't have a cache, so this is a no-op
	return nil
}

// Save saves a complaint to the file system
func (r *FileRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
	ctx, span := r.GetTracer().Start(ctx, "Save")
	defer span.End()

	logger := r.GetLogger().With("component", "file-repository", "complaint_id", complaint.ID.String())
	logger.Info("Saving complaint")

	return r.SaveComplaintToFile(complaint)
}

// FindByID retrieves a complaint by ID from the file system
func (r *FileRepository) FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "FindByID")
	defer span.End()

	logger := r.GetLogger().With("component", "file-repository", "complaint_id", id.String())
	logger.Debug("Finding complaint by ID")

	// Search through files
	complaints, err := r.LoadAllComplaintsFromDisk()
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
	ctx, span := r.GetTracer().Start(ctx, "FindAll")
	defer span.End()

	logger := r.GetLogger().With("component", "file-repository", "limit", limit, "offset", offset)
	logger.Debug("Finding all complaints")

	complaints, err := r.LoadAllComplaintsFromDisk()
	if err != nil {
		return nil, err
	}

	result := r.ApplyPagination(complaints, limit, offset)
	logger.Info("Complaints retrieved", "count", len(result))
	return result, nil
}

// FindBySeverity retrieves complaints by severity level
func (r *FileRepository) FindBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "FindBySeverity")
	defer span.End()

	logger := r.GetLogger().With("component", "file-repository", "severity", string(severity), "limit", limit)
	logger.Debug("Finding complaints by severity")

	complaints, err := r.LoadAllComplaintsFromDisk()
	if err != nil {
		return nil, err
	}

	filtered := filterComplaints(ctx, complaints, SeverityFilter(severity), limit)
	logger.Info("Complaints filtered by severity", "severity", string(severity), "count", len(filtered))
	return filtered, nil
}

// FindByProject retrieves complaints by project name
func (r *FileRepository) FindByProject(ctx context.Context, projectName string, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "FindByProject")
	defer span.End()

	logger := r.GetLogger().With("component", "file-repository", "project_name", projectName, "limit", limit)
	logger.Debug("Finding complaints by project")

	complaints, err := r.LoadAllComplaintsFromDisk()
	if err != nil {
		return nil, err
	}

	filtered := filterComplaints(ctx, complaints, ProjectFilter(projectName), limit)
	logger.Info("Complaints filtered by project", "project_name", projectName, "count", len(filtered))
	return filtered, nil
}

// FindUnresolved retrieves unresolved complaints
func (r *FileRepository) FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "FindUnresolved")
	defer span.End()

	logger := r.GetLogger().With("component", "file-repository", "limit", limit)
	logger.Debug("Finding unresolved complaints")

	complaints, err := r.LoadAllComplaintsFromDisk()
	if err != nil {
		return nil, err
	}

	filtered := filterComplaints(ctx, complaints, UnresolvedFilter(), limit)
	logger.Info("Unresolved complaints filtered", "count", len(filtered))
	return filtered, nil
}

// Search searches complaints by text content
func (r *FileRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "Search")
	defer span.End()

	logger := r.GetLogger().With("component", "file-repository", "query", query, "limit", limit)
	logger.Debug("Searching complaints")

	complaints, err := r.LoadAllComplaintsFromDisk()
	if err != nil {
		return nil, err
	}

	results := filterComplaints(ctx, complaints, SearchFilter(query), limit)
	logger.Info("Complaints searched", "query", query, "count", len(results))
	return results, nil
}

// Update updates an existing complaint in-place
func (r *FileRepository) Update(ctx context.Context, complaint *domain.Complaint) error {
	ctx, span := r.GetTracer().Start(ctx, "Update")
	defer span.End()

	logger := r.GetLogger().With("component", "file-repository", "complaint_id", complaint.ID.String())
	logger.Info("Updating complaint")

	// Find existing complaint
	existing, err := r.FindByID(ctx, complaint.ID)
	if err != nil {
		return fmt.Errorf("failed to find existing complaint: %w", err)
	}

	// Update fields with new data
	existing.Timestamp = complaint.Timestamp
	existing.TaskDescription = complaint.TaskDescription
	existing.ContextInfo = complaint.ContextInfo
	existing.MissingInfo = complaint.MissingInfo
	existing.ConfusedBy = complaint.ConfusedBy
	existing.FutureWishes = complaint.FutureWishes
	// Update resolution fields (ResolvedAt is single source of truth)
	existing.ResolvedAt = complaint.ResolvedAt
	existing.ResolvedBy = complaint.ResolvedBy

	// Save updated complaint (updates existing file in-place)
	return r.Save(ctx, existing)
}

// GetCacheStats returns "not available" for FileRepository (no cache)
func (r *FileRepository) GetCacheStats() CacheStats {
	return CacheStats{
		Hits:        0,
		Misses:      0,
		Evictions:   0,
		CurrentSize: 0,
		MaxSize:     0,
		HitRate:     0.0,
	}
}

// GetFilePath returns the actual file path where the complaint is stored
func (r *FileRepository) GetFilePath(ctx context.Context, id domain.ComplaintID) (string, error) {
	ctx, span := r.GetTracer().Start(ctx, "GetFilePath")
	defer span.End()

	logger := r.GetLogger().With("component", "file-repository", "complaint_id", id.String())
	logger.Debug("Getting file path for complaint")

	// Use UUID-only filename for consistent file updates
	filename := fmt.Sprintf("%s.json", id.String())
	filePath := filepath.Join(r.GetBaseDir(), filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Warn("Complaint file not found", "path", filePath)
		return "", fmt.Errorf("complaint file not found: %s", id.String())
	}
	
	logger.Info("File path retrieved", "path", filePath)
	return filePath, nil
}

// GetDocsPath returns the documentation path (if applicable) for the complaint
// Note: This is a basic implementation - in a full implementation, this would need
// access to the DocsRepository configuration and format settings
func (r *FileRepository) GetDocsPath(ctx context.Context, id domain.ComplaintID) (string, error) {
	ctx, span := r.GetTracer().Start(ctx, "GetDocsPath")
	defer span.End()

	logger := r.GetLogger().With("component", "file-repository", "complaint_id", id.String())
	logger.Debug("Getting docs path for complaint")

	// First, find the complaint to get its timestamp and session name
	complaint, err := r.FindByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to find complaint: %w", err)
	}

	// Default docs directory and format (this should be configurable)
	// For now, we'll use the default values from the config
	docsDir := "docs/complaints" 
	format := "markdown" // Default format
	
	// Generate filename using the same logic as DocsRepository
	sessionName := complaint.SessionName.String()
	if sessionName == "" {
		sessionName = "no-session"
	}
	
	// Sanitize session name for filename
	sessionName = strings.ReplaceAll(sessionName, " ", "_")
	sessionName = strings.ReplaceAll(sessionName, "/", "_")
	sessionName = strings.ReplaceAll(sessionName, "..", "_")
	sessionName = strings.ReplaceAll(sessionName, ":", "-")
	sessionName = strings.ReplaceAll(sessionName, "\"", "")
	sessionName = strings.ReplaceAll(sessionName, "'", "")
	sessionName = strings.ReplaceAll(sessionName, "\\", "_")
	sessionName = strings.ReplaceAll(sessionName, "<", "")
	sessionName = strings.ReplaceAll(sessionName, ">", "")
	sessionName = strings.ReplaceAll(sessionName, "|", "")
	sessionName = strings.ReplaceAll(sessionName, "?", "")
	sessionName = strings.ReplaceAll(sessionName, "*", "")
	
	// Remove multiple underscores
	for strings.Contains(sessionName, "__") {
		sessionName = strings.ReplaceAll(sessionName, "__", "_")
	}
	
	// Trim underscores
	sessionName = strings.Trim(sessionName, "_")
	
	// Limit length
	if len(sessionName) > 50 {
		sessionName = sessionName[:50]
	}
	
	// Format timestamp
	timeStr := complaint.Timestamp.Format("2006-01-02_15-04-05")
	
	// Generate filename with appropriate extension
	var extension string
	switch format {
	case "html":
		extension = ".html"
	case "text":
		extension = ".txt"
	default:
		extension = ".md" // markdown
	}
	
	filename := fmt.Sprintf("%s-%s%s", timeStr, sessionName, extension)
	docsPath := filepath.Join(docsDir, filename)
	
	logger.Info("Docs path retrieved", "path", docsPath)
	return docsPath, nil
}