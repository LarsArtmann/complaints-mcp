package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	GetCacheStats() CacheStats           // Optional - only CachedRepository implements
	WarmCache(ctx context.Context) error // Optional - warm cache with context support
}

// CachedRepository implements Repository with in-memory LRU caching for O(1) lookups
type CachedRepository struct {
	baseDir string
	logger  *log.Logger
	tracer  tracing.Tracer

	// LRU cache for automatic memory management
	cache *LRUCache
}

// FileRepository implements Repository using file system storage (legacy)
type FileRepository struct {
	baseDir string
	logger  *log.Logger
	tracer  tracing.Tracer
}

// NewFileRepository creates a new file-based repository (legacy)
func NewFileRepository(baseDir string, tracer tracing.Tracer) Repository {
	return &FileRepository{
		baseDir: baseDir,
		logger:  log.Default(),
		tracer:  tracer,
	}
}

// WarmCache is a no-op for FileRepository (no cache to warm)
func (r *FileRepository) WarmCache(ctx context.Context) error {
	// FileRepository doesn't have a cache, so this is a no-op
	return nil
}

// NewCachedRepository creates a new high-performance cached repository with LRU eviction
func NewCachedRepository(baseDir string, maxCacheSize int, tracer tracing.Tracer) Repository {
	repo := &CachedRepository{
		baseDir: baseDir,
		cache:   NewLRUCache(uint32(maxCacheSize)),
		logger:  log.Default(),
		tracer:  tracer,
	}

	// Don't warm cache automatically - let caller do it with proper context
	return repo
}

// Save saves a complaint to the file system and updates cache
func (r *CachedRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
	ctx, span := r.tracer.Start(ctx, "Save")
	defer span.End()

	logger := r.logger.With("component", "cached-repository", "complaint_id", complaint.ID.String())
	logger.Info("Saving complaint")

	// ✅ FIX: Use UUID-only filename for consistent file updates
	// Old: uuid-timestamp.json (creates duplicates on update)
	// New: uuid.json (single file, updated in-place)
	filename := fmt.Sprintf("%s.json", complaint.ID.String())

	filePath := filepath.Join(r.baseDir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		logger.Error("Failed to create directory", "error", err, "path", filepath.Dir(filePath))
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write complaint as JSON
	data, err := json.MarshalIndent(complaint, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal complaint", "error", err)
		return fmt.Errorf("failed to marshal complaint: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		logger.Error("Failed to write complaint file", "error", err, "path", filePath)
		return fmt.Errorf("failed to write complaint file: %w", err)
	}

	// Update LRU cache (thread-safe, automatic eviction)
	r.cache.Put(complaint.ID.String(), complaint)

	logger.Info("Complaint saved and cached", "path", filePath)
	return nil
}

// Save saves a complaint to the file system (legacy FileRepository)
func (r *FileRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
	ctx, span := r.tracer.Start(ctx, "Save")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "complaint_id", complaint.ID.String())
	logger.Info("Saving complaint")

	// ✅ FIX: Use UUID-only filename for consistent file updates
	// Old: uuid-timestamp.json (creates duplicates on update)
	// New: uuid.json (single file, updated in-place)
	filename := fmt.Sprintf("%s.json", complaint.ID.String())

	filePath := filepath.Join(r.baseDir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		logger.Error("Failed to create directory", "error", err, "path", filepath.Dir(filePath))
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write complaint as JSON
	data, err := json.MarshalIndent(complaint, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal complaint", "error", err)
		return fmt.Errorf("failed to marshal complaint: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		logger.Error("Failed to write complaint file", "error", err, "path", filePath)
		return fmt.Errorf("failed to write complaint file: %w", err)
	}

	logger.Info("Complaint saved successfully", "path", filePath)
	return nil
}

// FindByID retrieves a complaint by ID from LRU cache (O(1) lookup)
func (r *CachedRepository) FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindByID")
	defer span.End()

	logger := r.logger.With("component", "cached-repository", "complaint_id", id.String())
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

// FindByID retrieves a complaint by ID from the file system (legacy)
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

// FindAll retrieves all complaints with pagination from LRU cache
func (r *CachedRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindAll")
	defer span.End()

	logger := r.logger.With("component", "cached-repository", "limit", limit, "offset", offset)
	logger.Debug("Finding all complaints from LRU cache")

	// Get all complaints from LRU cache
	complaints := r.cache.GetAll()

	// Apply pagination
	total := len(complaints)
	if offset >= total {
		return []*domain.Complaint{}, nil
	}

	start := offset
	end := min(offset+limit, total)

	result := complaints[start:end]
	logger.Info("Complaints retrieved from LRU cache", "count", len(result))
	return result, nil
}

// FindAll retrieves all complaints with pagination (legacy)
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
	end := min(offset+limit, total)

	result := complaints[start:end]
	logger.Info("Complaints retrieved", "count", len(result))
	return result, nil
}

// FindBySeverity retrieves complaints by severity level from LRU cache
func (r *CachedRepository) FindBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindBySeverity")
	defer span.End()

	logger := r.logger.With("component", "cached-repository", "severity", string(severity), "limit", limit)
	logger.Debug("Finding complaints by severity from LRU cache")

	allComplaints := r.cache.GetAll()
	filtered := filterComplaints(ctx, allComplaints, SeverityFilter(severity), limit)

	logger.Info("Complaints filtered by severity from LRU cache", "severity", string(severity), "count", len(filtered))
	return filtered, nil
}

// FindBySeverity retrieves complaints by severity level (legacy)
func (r *FileRepository) FindBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindBySeverity")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "severity", string(severity), "limit", limit)
	logger.Debug("Finding complaints by severity")

	complaints, err := r.loadAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	filtered := filterComplaints(ctx, complaints, SeverityFilter(severity), limit)

	logger.Info("Complaints filtered by severity", "severity", string(severity), "count", len(filtered))
	return filtered, nil
}

// Update updates an existing complaint in-place with LRU cache update
func (r *CachedRepository) Update(ctx context.Context, complaint *domain.Complaint) error {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	logger := r.logger.With("component", "cached-repository", "complaint_id", complaint.ID.String())
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

// Update updates an existing complaint in-place (legacy)
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

// Search searches complaints by text content from LRU cache
func (r *CachedRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "Search")
	defer span.End()

	logger := r.logger.With("component", "cached-repository", "query", query, "limit", limit)
	logger.Debug("Searching complaints from LRU cache")

	allComplaints := r.cache.GetAll()
	results := filterComplaints(ctx, allComplaints, SearchFilter(query), limit)

	logger.Info("Complaints searched from LRU cache", "query", query, "count", len(results))
	return results, nil
}

// Search searches complaints by text content (legacy)
func (r *FileRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "Search")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "query", query, "limit", limit)
	logger.Debug("Searching complaints")

	complaints, err := r.loadAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	results := filterComplaints(ctx, complaints, SearchFilter(query), limit)

	logger.Info("Complaints searched", "query", query, "count", len(results))
	return results, nil
}

// WarmCache warms the cache with existing complaint data (public method)
func (r *CachedRepository) WarmCache(ctx context.Context) error {
	logger := r.logger.With("component", "cached-repository")
	logger.Info("Warming LRU cache with existing complaint data")

	// Load all existing complaints into LRU cache
	complaints, err := r.loadAllComplaintsFromDisk()
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

// loadAllComplaintsFromDisk loads all complaints from disk (cache warm-up only)
func (r *CachedRepository) loadAllComplaintsFromDisk() ([]*domain.Complaint, error) {
	logger := r.logger.With("component", "cached-repository")
	logger.Info("Loading all complaints from disk for cache warm-up", "base_dir", r.baseDir)

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
			logger.Warn("Failed to load complaint file during cache warm-up", "error", err, "path", filePath)
			continue
		}

		complaints = append(complaints, complaint)
	}

	logger.Info("Complaints loaded from disk successfully", "base_dir", r.baseDir, "count", len(complaints))
	return complaints, nil
}

// loadComplaintFromFile loads a single complaint from a JSON file
func (r *CachedRepository) loadComplaintFromFile(filePath string) (*domain.Complaint, error) {
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

// loadAllComplaints loads all complaints from the file system (legacy)
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

	// Sort complaints by timestamp (oldest first for consistent ordering)
	sort.Slice(complaints, func(i, j int) bool {
		return complaints[i].Timestamp.Before(complaints[j].Timestamp)
	})

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

// FindByProject retrieves complaints by project name from LRU cache
func (r *CachedRepository) FindByProject(ctx context.Context, projectName string, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindByProject")
	defer span.End()

	logger := r.logger.With("component", "cached-repository", "project_name", projectName, "limit", limit)
	logger.Debug("Finding complaints by project from LRU cache")

	allComplaints := r.cache.GetAll()
	filtered := filterComplaints(ctx, allComplaints, ProjectFilter(projectName), limit)

	logger.Info("Complaints filtered by project from LRU cache", "project_name", projectName, "count", len(filtered))
	return filtered, nil
}

// FindByProject retrieves complaints by project name (legacy)
func (r *FileRepository) FindByProject(ctx context.Context, projectName string, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindByProject")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "project_name", projectName, "limit", limit)
	logger.Debug("Finding complaints by project")

	complaints, err := r.loadAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	filtered := filterComplaints(ctx, complaints, ProjectFilter(projectName), limit)

	logger.Info("Complaints filtered by project", "project_name", projectName, "count", len(filtered))
	return filtered, nil
}

// FindUnresolved retrieves unresolved complaints from LRU cache
func (r *CachedRepository) FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindUnresolved")
	defer span.End()

	logger := r.logger.With("component", "cached-repository", "limit", limit)
	logger.Debug("Finding unresolved complaints from LRU cache")

	allComplaints := r.cache.GetAll()
	filtered := filterComplaints(ctx, allComplaints, UnresolvedFilter(), limit)

	logger.Info("Unresolved complaints filtered from LRU cache", "count", len(filtered))
	return filtered, nil
}

// FindUnresolved retrieves unresolved complaints (legacy)
func (r *FileRepository) FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error) {
	ctx, span := r.tracer.Start(ctx, "FindUnresolved")
	defer span.End()

	logger := r.logger.With("component", "file-repository", "limit", limit)
	logger.Debug("Finding unresolved complaints")

	complaints, err := r.loadAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	filtered := filterComplaints(ctx, complaints, UnresolvedFilter(), limit)

	logger.Info("Unresolved complaints filtered", "count", len(filtered))
	return filtered, nil
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

// GetCacheStats returns current LRU cache performance statistics
func (r *CachedRepository) GetCacheStats() CacheStats {
	return r.cache.GetStats()
}
