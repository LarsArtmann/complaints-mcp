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

// CachedRepository implements Repository with in-memory LRU caching for O(1) lookups
type CachedRepository struct {
	BaseRepository
	
	// LRU cache for automatic memory management
	cache *LRUCache
}

// NewCachedRepository creates a new high-performance cached repository with LRU eviction
func NewCachedRepository(baseDir string, maxCacheSize int, tracer tracing.Tracer) Repository {
	repo := &CachedRepository{
		BaseRepository: NewBaseRepository(baseDir, tracer),
		cache:          NewLRUCache(uint32(maxCacheSize)),
	}

	// Don't warm cache automatically - let caller do it with proper context
	return repo
}

// Save saves a complaint to the file system and updates cache
func (r *CachedRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
	ctx, span := r.GetTracer().Start(ctx, "Save")
	defer span.End()

	logger := r.GetLogger().With("component", "cached-repository", "complaint_id", complaint.ID.String())
	logger.Info("Saving complaint")

	// Save to file system using base repository
	if err := r.SaveComplaintToFile(complaint); err != nil {
		return err
	}

	// Update LRU cache (thread-safe, automatic eviction)
	r.cache.Put(complaint.ID.String(), complaint)

	logger.Info("Complaint saved and cached")
	return nil
}

// FindByID retrieves a complaint by ID from LRU cache (O(1) lookup)
func (r *CachedRepository) FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "FindByID")
	defer span.End()

	logger := r.GetLogger().With("component", "cached-repository", "complaint_id", id.String())
	logger.Debug("Finding complaint by ID in LRU cache")

	// O(1) LRU cache lookup (also updates access order)
	complaint, exists := r.cache.Get(id.String())
	if exists {
		logger.Info("Complaint found in LRU cache (O(1) lookup)")
		return complaint, nil
	}

	logger.Warn("Complaint not found in cache", "complaint_id", id.String())
	return nil, fmt.Errorf("complaint not found: %s", id.String())
}

// FindAll retrieves all complaints with pagination from LRU cache
func (r *CachedRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "FindAll")
	defer span.End()

	logger := r.GetLogger().With("component", "cached-repository", "limit", limit, "offset", offset)
	logger.Debug("Finding all complaints from LRU cache")

	// Get all complaints from LRU cache
	complaints := r.cache.GetAll()

	// Apply pagination using base repository method
	result := r.ApplyPagination(complaints, limit, offset)
	logger.Info("Complaints retrieved from LRU cache", "count", len(result))
	return result, nil
}

// FindBySeverity retrieves complaints by severity level from LRU cache
func (r *CachedRepository) FindBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "FindBySeverity")
	defer span.End()

	logger := r.GetLogger().With("component", "cached-repository", "severity", string(severity), "limit", limit)
	logger.Debug("Finding complaints by severity from LRU cache")

	allComplaints := r.cache.GetAll()
	filtered := filterComplaints(ctx, allComplaints, SeverityFilter(severity), limit)

	logger.Info("Complaints filtered by severity from LRU cache", "severity", string(severity), "count", len(filtered))
	return filtered, nil
}

// FindByProject retrieves complaints by project name from LRU cache
func (r *CachedRepository) FindByProject(ctx context.Context, projectName string, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "FindByProject")
	defer span.End()

	logger := r.GetLogger().With("component", "cached-repository", "project_name", projectName, "limit", limit)
	logger.Debug("Finding complaints by project from LRU cache")

	allComplaints := r.cache.GetAll()
	filtered := filterComplaints(ctx, allComplaints, ProjectFilter(projectName), limit)

	logger.Info("Complaints filtered by project from LRU cache", "project_name", projectName, "count", len(filtered))
	return filtered, nil
}

// FindUnresolved retrieves unresolved complaints from LRU cache
func (r *CachedRepository) FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "FindUnresolved")
	defer span.End()

	logger := r.GetLogger().With("component", "cached-repository", "limit", limit)
	logger.Debug("Finding unresolved complaints from LRU cache")

	allComplaints := r.cache.GetAll()
	filtered := filterComplaints(ctx, allComplaints, UnresolvedFilter(), limit)

	logger.Info("Unresolved complaints filtered from LRU cache", "count", len(filtered))
	return filtered, nil
}

// Search searches complaints by text content from LRU cache
func (r *CachedRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.GetTracer().Start(ctx, "Search")
	defer span.End()

	logger := r.GetLogger().With("component", "cached-repository", "query", query, "limit", limit)
	logger.Debug("Searching complaints from LRU cache")

	allComplaints := r.cache.GetAll()
	results := filterComplaints(ctx, allComplaints, SearchFilter(query), limit)

	logger.Info("Complaints searched from LRU cache", "query", query, "count", len(results))
	return results, nil
}

// Update updates an existing complaint in-place with LRU cache update
func (r *CachedRepository) Update(ctx context.Context, complaint *domain.Complaint) error {
	ctx, span := r.GetTracer().Start(ctx, "Update")
	defer span.End()

	logger := r.GetLogger().With("component", "cached-repository", "complaint_id", complaint.ID.String())
	logger.Info("Updating complaint in LRU cache")

	// Find existing complaint from LRU cache
	existing, exists := r.cache.Get(complaint.ID.String())
	if !exists {
		return fmt.Errorf("complaint not found: %s", complaint.ID.String())
	}

	// Update fields with new data
	existing.TaskDescription = complaint.TaskDescription
	existing.ContextInfo = complaint.ContextInfo
	existing.MissingInfo = complaint.MissingInfo
	existing.ConfusedBy = complaint.ConfusedBy
	existing.FutureWishes = complaint.FutureWishes
	// Update resolution fields (ResolvedAt is single source of truth)
	existing.ResolvedAt = complaint.ResolvedAt
	existing.ResolvedBy = complaint.ResolvedBy

	// Save updated complaint (updates existing file and LRU cache)
	return r.Save(ctx, existing)
}

// WarmCache warms the cache with existing complaint data (public method)
func (r *CachedRepository) WarmCache(ctx context.Context) error {
	logger := r.GetLogger().With("component", "cached-repository")
	logger.Info("Warming LRU cache with existing complaint data")

	// Load all existing complaints into LRU cache
	complaints, err := r.LoadAllComplaintsFromDisk()
	if err != nil {
		logger.Error("Failed to warm LRU cache", "error", err)
		return fmt.Errorf("failed to warm cache: %w", err)
	}

	for _, complaint := range complaints {
		r.cache.Put(complaint.ID.String(), complaint)
	}

	logger.Info("LRU cache warmed successfully", "complaints_loaded", len(complaints))
	return nil
}

// GetCacheStats returns current LRU cache performance statistics
func (r *CachedRepository) GetCacheStats() CacheStats {
	return r.cache.GetStats()
}

// GetFilePath returns the actual file path where the complaint is stored
func (r *CachedRepository) GetFilePath(ctx context.Context, id domain.ComplaintID) (string, error) {
	ctx, span := r.GetTracer().Start(ctx, "GetFilePath")
	defer span.End()

	logger := r.GetLogger().With("component", "cached-repository", "complaint_id", id.String())
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
func (r *CachedRepository) GetDocsPath(ctx context.Context, id domain.ComplaintID) (string, error) {
	ctx, span := r.GetTracer().Start(ctx, "GetDocsPath")
	defer span.End()

	logger := r.GetLogger().With("component", "cached-repository", "complaint_id", id.String())
	logger.Debug("Getting docs path for complaint")

	// First, find complaint to get its timestamp and session name
	// Try cache first for O(1) lookup
	complaint, exists := r.cache.Get(id.String())
	if !exists {
		return "", fmt.Errorf("complaint not found: %s", id.String())
	}

	// Default docs directory and format (this should be configurable)
	// For now, we'll use default values from the config
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