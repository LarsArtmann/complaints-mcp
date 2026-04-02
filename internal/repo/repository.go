package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

const (
	defaultComplaintsDir = "complaints"
	defaultDocsDir       = "docs/complaints"
	defaultFindAllLimit  = 1000
)

func (r *FileRepository) findByPredicate(
	ctx context.Context,
	matches func(*domain.Complaint) bool,
	limit int,
) ([]*domain.Complaint, error) {
	all, err := r.FindAll(ctx, defaultFindAllLimit, 0)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Complaint

	for _, complaint := range all {
		if matches(complaint) {
			filtered = append(filtered, complaint)
		}

		if len(filtered) >= limit {
			break
		}
	}

	return filtered, nil
}

func (r *FileRepository) findByID(
	ctx context.Context,
	field domain.ComplaintIDField,
	id string,
	limit int,
) ([]*domain.Complaint, error) {
	return r.findByPredicate(ctx, func(c *domain.Complaint) bool {
		return c.GetID(field) == id
	}, limit)
}

// Repository interface for complaint storage.
type Repository interface {
	Save(ctx context.Context, complaint *domain.Complaint) error
	FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error)
	FindAll(ctx context.Context, limit, offset int) ([]*domain.Complaint, error)
	FindBySeverity(
		ctx context.Context,
		severity domain.Severity,
		limit int,
	) ([]*domain.Complaint, error)
	FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error)
	Update(ctx context.Context, complaint *domain.Complaint) error
	Delete(ctx context.Context, id domain.ComplaintID) error
	Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error)
	WarmCache(ctx context.Context) error
	GetCacheStats() CacheStats
	GetFilePath(ctx context.Context, id domain.ComplaintID) (string, error)
	GetDocsPath(ctx context.Context, id domain.ComplaintID) (string, error)
	FindBySession(ctx context.Context, sessionID string, limit int) ([]*domain.Complaint, error)
	FindByProject(ctx context.Context, projectID string, limit int) ([]*domain.Complaint, error)
	FindByAgent(ctx context.Context, agentID string, limit int) ([]*domain.Complaint, error)
}

// FileRepository implements Repository interface using file system.
type FileRepository struct {
	complaintsDir string
	docsDir       string
	tracer        tracing.Tracer
}

// NewFileRepository creates a new file repository.
func NewFileRepository(baseDir string, tracer tracing.Tracer) *FileRepository {
	complaintsDir := filepath.Join(baseDir, defaultComplaintsDir)
	docsDir := filepath.Join(baseDir, defaultDocsDir)

	return &FileRepository{
		complaintsDir: complaintsDir,
		docsDir:       docsDir,
		tracer:        tracer,
	}
}

// Save saves a complaint to file system with FLAT JSON.
func (r *FileRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
	if err := complaint.Validate(); err != nil {
		return fmt.Errorf("invalid complaint: %w", err)
	}

	// Serialize with FLAT JSON structure
	data, err := json.Marshal(complaint)
	if err != nil {
		return fmt.Errorf("failed to marshal complaint: %w", err)
	}

	// Use phantom type ID for file naming
	fileName := complaint.ID.String() + ".json"

	return r.writeFile(fileName, data)
}

// FindByID finds a complaint by ID.
func (r *FileRepository) FindByID(
	ctx context.Context,
	id domain.ComplaintID,
) (*domain.Complaint, error) {
	if id.IsZero() {
		return nil, errors.New("invalid ComplaintID: cannot be empty")
	}

	// Load from file
	fileName := id.String() + ".json"

	data, err := r.readFile(fileName)
	if err != nil {
		return nil, err
	}

	var complaint domain.Complaint
	if err := json.Unmarshal(data, &complaint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal complaint: %w", err)
	}

	return &complaint, nil
}

// FindAll finds all complaints.
func (r *FileRepository) FindAll(
	ctx context.Context,
	limit, offset int,
) ([]*domain.Complaint, error) {
	files, err := r.listComplaintFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list complaint files: %w", err)
	}

	// Sort files by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := files[i].Info()
		infoJ, _ := files[j].Info()

		return infoI.ModTime().After(infoJ.ModTime())
	})

	var complaints []*domain.Complaint

	count := 0

	for i := range files {
		if count < offset {
			count++

			continue
		}

		if len(complaints) >= limit {
			break
		}

		fileName := files[i].Name()
		if !strings.HasSuffix(fileName, ".json") {
			continue
		}

		// Extract ID from filename
		idStr := fileName[:len(fileName)-5]

		id, err := domain.ParseComplaintID(idStr)
		if err != nil {
			continue
		}

		complaint, err := r.FindByID(ctx, id)
		if err != nil {
			continue
		}

		complaints = append(complaints, complaint)
	}

	// Sort by creation time (oldest first) for consistent ordering
	sort.Slice(complaints, func(i, j int) bool {
		return complaints[i].Timestamp.Before(complaints[j].Timestamp)
	})

	return complaints, nil
}

// FindBySeverity finds complaints by severity.
func (r *FileRepository) FindBySeverity(
	ctx context.Context,
	severity domain.Severity,
	limit int,
) ([]*domain.Complaint, error) {
	all, err := r.FindAll(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Complaint

	for _, complaint := range all {
		if complaint.Severity == severity {
			filtered = append(filtered, complaint)
		}

		if len(filtered) >= limit {
			break
		}
	}

	return filtered, nil
}

// FindUnresolved finds unresolved complaints.
func (r *FileRepository) FindUnresolved(
	ctx context.Context,
	limit int,
) ([]*domain.Complaint, error) {
	all, err := r.FindAll(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	var unresolved []*domain.Complaint

	for _, complaint := range all {
		if !complaint.ResolutionState.IsResolved() {
			unresolved = append(unresolved, complaint)
		}

		if len(unresolved) >= limit {
			break
		}
	}

	return unresolved, nil
}

// Update updates a complaint.
func (r *FileRepository) Update(ctx context.Context, complaint *domain.Complaint) error {
	return r.Save(ctx, complaint)
}

// Delete deletes a complaint by ID.
func (r *FileRepository) Delete(ctx context.Context, id domain.ComplaintID) error {
	fileName := id.String() + ".json"

	return os.Remove(filepath.Join(r.complaintsDir, fileName))
}

// Search searches complaints by text.
func (r *FileRepository) Search(
	ctx context.Context,
	query string,
	limit int,
) ([]*domain.Complaint, error) {
	all, err := r.FindAll(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)

	var results []*domain.Complaint

	for _, complaint := range all {
		if strings.Contains(strings.ToLower(complaint.TaskDescription), query) ||
			strings.Contains(strings.ToLower(complaint.ContextInfo), query) ||
			strings.Contains(strings.ToLower(complaint.MissingInfo), query) ||
			strings.Contains(strings.ToLower(complaint.ConfusedBy), query) ||
			strings.Contains(strings.ToLower(complaint.AgentID.String()), query) {
			results = append(results, complaint)
		}

		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// WarmCache loads all complaints into cache.
func (r *FileRepository) WarmCache(ctx context.Context) error {
	return nil // No cache in minimal version
}

// GetCacheStats returns cache statistics.
func (r *FileRepository) GetCacheStats() CacheStats {
	return CacheStats{
		CachedComplaints: int64(0),
		MaxCacheSize:     int64(0),
		MaxSize:          int64(0), // ✅ ADDED for test compatibility
		Hits:             int64(0),
		Misses:           int64(0),
		Evictions:        int64(0),
		CurrentSize:      int64(0),
		HitRate:          0.0,
	}
}

// GetFilePath returns file path for a complaint.
func (r *FileRepository) GetFilePath(ctx context.Context, id domain.ComplaintID) (string, error) {
	fileName := id.String() + ".json"

	return filepath.Join(r.complaintsDir, fileName), nil
}

// GetDocsPath returns documentation path for a complaint.
func (r *FileRepository) GetDocsPath(ctx context.Context, id domain.ComplaintID) (string, error) {
	complaint, err := r.FindByID(ctx, id)
	if err != nil {
		return "", err
	}

	timestamp := complaint.Timestamp.Format("2006-01-02_15-04")

	fileName := fmt.Sprintf(
		"%s-%s-%s.md",
		timestamp,
		complaint.SessionID.String(),
		complaint.TaskDescription[:20],
	)
	if len(fileName) > 100 {
		fileName = fileName[:100]
	}

	return filepath.Join(r.docsDir, fileName), nil
}

// FindBySession finds complaints by session.
func (r *FileRepository) FindBySession(
	ctx context.Context,
	sessionID string,
	limit int,
) ([]*domain.Complaint, error) {
	return r.findByID(ctx, domain.ComplaintFieldSessionID, sessionID, limit)
}

// FindByProject finds complaints by project.
func (r *FileRepository) FindByProject(
	ctx context.Context,
	projectID string,
	limit int,
) ([]*domain.Complaint, error) {
	return r.findByID(ctx, domain.ComplaintFieldProjectID, projectID, limit)
}

// FindByAgent finds complaints by agent.
func (r *FileRepository) FindByAgent(
	ctx context.Context,
	agentID string,
	limit int,
) ([]*domain.Complaint, error) {
	return r.findByID(ctx, domain.ComplaintFieldAgentID, agentID, limit)
}

// CacheStats represents cache statistics.
type CacheStats struct {
	CachedComplaints int64   `json:"cached_complaints"`
	MaxCacheSize     int64   `json:"max_cache_size"`
	MaxSize          int64   `json:"max_size"` // ✅ ADDED for test compatibility
	Hits             int64   `json:"hits"`
	Misses           int64   `json:"misses"`
	Evictions        int64   `json:"evictions"`
	CurrentSize      int64   `json:"current_size"`
	HitRate          float64 `json:"hit_rate_percent"`
}

// listComplaintFiles lists all complaint files.
func (r *FileRepository) listComplaintFiles() ([]os.DirEntry, error) {
	entries, err := os.ReadDir(r.complaintsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []os.DirEntry{}, nil
		}

		return nil, err
	}

	var files []os.DirEntry

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry)
		}
	}

	return files, nil
}

// writeFile writes data to a file.
func (r *FileRepository) writeFile(fileName string, data []byte) error {
	err := os.MkdirAll(r.complaintsDir, 0o755)
	if err != nil {
		return err
	}

	filePath := filepath.Join(r.complaintsDir, fileName)

	return os.WriteFile(filePath, data, 0o644)
}

// NewRepositoryFromConfig creates a repository based on configuration.
func NewRepositoryFromConfig(cfg *config.Config, tracer tracing.Tracer) *FileRepository {
	// For now, always return a FileRepository
	// In the future, this could check cfg.Storage.CacheEnabled to return a cached repository
	return NewFileRepository(cfg.Storage.BaseDir, tracer)
}

// SimpleCachedRepository provides basic caching functionality.
type SimpleCachedRepository struct {
	base    Repository
	cache   map[domain.ComplaintID]*domain.Complaint
	maxSize int
	stats   CacheStats
	mu      sync.RWMutex
}

// NewSimpleCachedRepository creates a simple cached repository wrapper.
func NewSimpleCachedRepository(base Repository, maxSize int) *SimpleCachedRepository {
	return &SimpleCachedRepository{
		base:    base,
		cache:   make(map[domain.ComplaintID]*domain.Complaint),
		maxSize: maxSize,
		stats: CacheStats{
			MaxSize: int64(maxSize),
		},
	}
}

// Save implements Repository interface.
func (r *SimpleCachedRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Save to base repository
	err := r.base.Save(ctx, complaint)
	if err != nil {
		return err
	}

	// Add to cache
	r.cache[complaint.ID] = complaint
	r.updateCacheSize()

	return nil
}

// FindByID implements Repository interface.
func (r *SimpleCachedRepository) FindByID(
	ctx context.Context,
	id domain.ComplaintID,
) (*domain.Complaint, error) {
	r.mu.RLock()

	// Check cache first
	if cached, found := r.cache[id]; found {
		r.mu.RUnlock()
		r.mu.Lock()
		r.stats.Hits++
		r.recalculateHitRate()
		r.mu.Unlock()

		return cached, nil
	}

	r.mu.RUnlock()

	// Get from base repository
	complaint, err := r.base.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Add to cache
	r.mu.Lock()
	r.cache[id] = complaint
	r.stats.Misses++
	r.recalculateHitRate()
	r.updateCacheSize()
	r.mu.Unlock()

	return complaint, nil
}

func (r *SimpleCachedRepository) recalculateHitRate() {
	r.stats.HitRate = float64(r.stats.Hits) / float64(r.stats.Hits+r.stats.Misses) * 100
}

// GetCacheStats implements Repository interface.
func (r *SimpleCachedRepository) GetCacheStats() CacheStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.stats
}

// updateCacheSize updates current cache size and eviction logic.
func (r *SimpleCachedRepository) updateCacheSize() {
	currentSize := int64(len(r.cache))
	r.stats.CurrentSize = currentSize
	r.stats.CachedComplaints = currentSize

	// Simple eviction: if over max size, remove oldest (first) items
	for currentSize > int64(r.maxSize) {
		var oldestID domain.ComplaintID
		for id := range r.cache {
			oldestID = id

			break
		}

		delete(r.cache, oldestID)
		r.stats.Evictions++
		currentSize = int64(len(r.cache))
	}

	r.stats.CurrentSize = currentSize
	r.stats.CachedComplaints = currentSize
}

// Delegate other methods to base repository.
func (r *SimpleCachedRepository) FindAll(
	ctx context.Context,
	limit, offset int,
) ([]*domain.Complaint, error) {
	return r.base.FindAll(ctx, limit, offset)
}

func (r *SimpleCachedRepository) Search(
	ctx context.Context,
	query string,
	limit int,
) ([]*domain.Complaint, error) {
	return r.base.Search(ctx, query, limit)
}

func (r *SimpleCachedRepository) FindByProject(
	ctx context.Context,
	projectID string,
	limit int,
) ([]*domain.Complaint, error) {
	return r.findByID(ctx, projectID, limit, r.base.FindByProject)
}

func (r *SimpleCachedRepository) FindUnresolved(
	ctx context.Context,
	limit int,
) ([]*domain.Complaint, error) {
	return r.base.FindUnresolved(ctx, limit)
}

// Add missing methods from Repository interface.
func (r *SimpleCachedRepository) FindBySeverity(
	ctx context.Context,
	severity domain.Severity,
	limit int,
) ([]*domain.Complaint, error) {
	return r.base.FindBySeverity(ctx, severity, limit)
}

// findByID is a helper that delegates to a finder function with (ctx, id, limit) signature.
func (r *SimpleCachedRepository) findByID(
	ctx context.Context,
	id string,
	limit int,
	finder func(context.Context, string, int) ([]*domain.Complaint, error),
) ([]*domain.Complaint, error) {
	return finder(ctx, id, limit)
}

func (r *SimpleCachedRepository) FindBySession(
	ctx context.Context,
	sessionID string,
	limit int,
) ([]*domain.Complaint, error) {
	return r.findByID(ctx, sessionID, limit, r.base.FindBySession)
}

func (r *SimpleCachedRepository) FindByAgent(
	ctx context.Context,
	agentID string,
	limit int,
) ([]*domain.Complaint, error) {
	return r.findByID(ctx, agentID, limit, r.base.FindByAgent)
}

func (r *SimpleCachedRepository) Delete(ctx context.Context, id domain.ComplaintID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Delete from base repository
	err := r.base.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Remove from cache
	delete(r.cache, id)
	r.updateCacheSize()

	return nil
}

func (r *SimpleCachedRepository) Update(ctx context.Context, complaint *domain.Complaint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Update base repository
	err := r.base.Update(ctx, complaint)
	if err != nil {
		return err
	}

	// Update cache
	if _, exists := r.cache[complaint.ID]; exists {
		r.cache[complaint.ID] = complaint
	}

	return nil
}

func (r *SimpleCachedRepository) WarmCache(ctx context.Context) error {
	return r.base.WarmCache(ctx)
}

func (r *SimpleCachedRepository) GetFilePath(
	ctx context.Context,
	id domain.ComplaintID,
) (string, error) {
	return getCachedPath(ctx, id, r.base.GetFilePath)
}

func (r *SimpleCachedRepository) GetDocsPath(
	ctx context.Context,
	id domain.ComplaintID,
) (string, error) {
	return getCachedPath(ctx, id, r.base.GetDocsPath)
}

type pathFunc func(context.Context, domain.ComplaintID) (string, error)

func getCachedPath(ctx context.Context, id domain.ComplaintID, fn pathFunc) (string, error) {
	return fn(ctx, id)
}

// NewCachedRepository creates a cached repository with minimal cache implementation.
func NewCachedRepository(baseDir string, tracer tracing.Tracer) *SimpleCachedRepository {
	// Create file repository as base
	baseRepo := NewFileRepository(baseDir, tracer)

	// Wrap with simple cache layer
	return NewSimpleCachedRepository(baseRepo, 1000)
}

// readFile reads data from a file.
func (r *FileRepository) readFile(fileName string) ([]byte, error) {
	filePath := filepath.Join(r.complaintsDir, fileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("complaint not found: %s", fileName)
		}

		return nil, err
	}

	return data, nil
}
