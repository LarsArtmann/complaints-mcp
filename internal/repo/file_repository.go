package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

const defaultComplaintsDir = "complaints"
const defaultDocsDir = "docs/complaints"

// FileRepository implements Repository interface using file system
type FileRepository struct {
	complaintsDir string
	docsDir       string
	cache         map[domain.ComplaintID]*domain.Complaint
	mutex         sync.RWMutex
	tracer        tracing.Tracer
}

// NewFileRepository creates a new file repository
func NewFileRepository(baseDir string, tracer tracing.Tracer) *FileRepository {
	complaintsDir := filepath.Join(baseDir, defaultComplaintsDir)
	docsDir := filepath.Join(baseDir, defaultDocsDir)

	return &FileRepository{
		complaintsDir: complaintsDir,
		docsDir:       docsDir,
		cache:         make(map[domain.ComplaintID]*domain.Complaint),
		tracer:        tracer,
	}
}

// Save saves a complaint to file system
func (r *FileRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
	if err := complaint.Validate(); err != nil {
		return fmt.Errorf("invalid complaint: %w", err)
	}

	// Serialize with flat JSON structure (phantom type)
	data, err := json.Marshal(complaint)
	if err != nil {
		return fmt.Errorf("failed to marshal complaint: %w", err)
	}

	// Use phantom type ID for file naming
	fileName := fmt.Sprintf("%s.json", complaint.ID.String())
	return r.writeFile(ctx, fileName, data)
}

// FindByID finds a complaint by ID
func (r *FileRepository) FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	if err := id.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ComplaintID: %w", err)
	}

	// Check cache first
	r.mutex.RLock()
	if cached, exists := r.cache[id]; exists {
		r.mutex.RUnlock()
		return cached, nil
	}
	r.mutex.RUnlock()

	// Load from file using phantom type ID
	fileName := fmt.Sprintf("%s.json", id.String())
	data, err := r.readFile(ctx, fileName)
	if err != nil {
		return nil, err
	}

	var complaint domain.Complaint
	if err := json.Unmarshal(data, &complaint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal complaint: %w", err)
	}

	// Update cache
	r.mutex.Lock()
	r.cache[id] = &complaint
	r.mutex.Unlock()

	return &complaint, nil
}

// FindAll finds all complaints with pagination
func (r *FileRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Complaint, error) {
	files, err := r.listComplaintFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list complaint files: %w", err)
	}

	// Sort files by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	var complaints []*domain.Complaint
	count := 0

	// Apply pagination
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

		// Extract ID from filename (remove .json)
		idStr := fileName[:len(fileName)-5]
		id, err := domain.ParseComplaintID(idStr)
		if err != nil {
			continue // Skip invalid file names
		}

		complaint, err := r.FindByID(ctx, id)
		if err != nil {
			continue // Skip invalid complaints
		}

		complaints = append(complaints, complaint)
	}

	return complaints, nil
}

// FindBySeverity finds complaints by severity
func (r *FileRepository) FindBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error) {
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

// FindUnresolved finds unresolved complaints
func (r *FileRepository) FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error) {
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

// Update updates a complaint
func (r *FileRepository) Update(ctx context.Context, complaint *domain.Complaint) error {
	if err := complaint.Validate(); err != nil {
		return fmt.Errorf("invalid complaint: %w", err)
	}

	// Save with phantom type ID
	data, err := json.Marshal(complaint)
	if err != nil {
		return fmt.Errorf("failed to marshal complaint: %w", err)
	}

	fileName := fmt.Sprintf("%s.json", complaint.ID.String())
	if err := r.writeFile(ctx, fileName, data); err != nil {
		return err
	}

	// Update cache
	r.mutex.Lock()
	r.cache[complaint.ID] = complaint
	r.mutex.Unlock()

	return nil
}

// Delete deletes a complaint by ID
func (r *FileRepository) Delete(ctx context.Context, id domain.ComplaintID) error {
	if err := id.Validate(); err != nil {
		return fmt.Errorf("invalid ComplaintID: %w", err)
	}

	fileName := fmt.Sprintf("%s.json", id.String())
	err := os.Remove(filepath.Join(r.complaintsDir, fileName))
	if err != nil {
		return fmt.Errorf("failed to delete complaint file: %w", err)
	}

	// Remove from cache
	r.mutex.Lock()
	delete(r.cache, id)
	r.mutex.Unlock()

	return nil
}

// Search searches complaints by text
func (r *FileRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error) {
	all, err := r.FindAll(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []*domain.Complaint

	for _, complaint := range all {
		// Search in all text fields
		if strings.Contains(strings.ToLower(complaint.TaskDescription), query) ||
			strings.Contains(strings.ToLower(complaint.ContextInfo), query) ||
			strings.Contains(strings.ToLower(complaint.MissingInfo), query) ||
			strings.Contains(strings.ToLower(complaint.ConfusedBy), query) ||
			strings.Contains(strings.ToLower(complaint.FutureWishes), query) ||
			strings.Contains(strings.ToLower(complaint.AgentID), query) ||
			strings.Contains(strings.ToLower(complaint.ProjectName), query) {
			results = append(results, complaint)
		}

		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// WarmCache loads all complaints into cache
func (r *FileRepository) WarmCache(ctx context.Context) error {
	files, err := r.listComplaintFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list complaint files: %w", err)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".json") {
			continue
		}

		idStr := fileName[:len(fileName)-5]
		id, err := domain.ParseComplaintID(idStr)
		if err != nil {
			continue
		}

		data, err := r.readFile(ctx, fileName)
		if err != nil {
			continue
		}

		var complaint domain.Complaint
		if err := json.Unmarshal(data, &complaint); err != nil {
			continue
		}

		r.cache[id] = &complaint
	}

	return nil
}

// GetCacheStats returns cache statistics
func (r *FileRepository) GetCacheStats() CacheStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return CacheStats{
		CachedComplaints: len(r.cache),
		MaxCacheSize:     1000, // Default max cache size
	}
}

// GetFilePath returns the file path for a complaint
func (r *FileRepository) GetFilePath(ctx context.Context, id domain.ComplaintID) (string, error) {
	if err := id.Validate(); err != nil {
		return "", fmt.Errorf("invalid ComplaintID: %w", err)
	}

	fileName := fmt.Sprintf("%s.json", id.String())
	return filepath.Join(r.complaintsDir, fileName), nil
}

// GetDocsPath returns the documentation path for a complaint
func (r *FileRepository) GetDocsPath(ctx context.Context, id domain.ComplaintID) (string, error) {
	if err := id.Validate(); err != nil {
		return "", fmt.Errorf("invalid ComplaintID: %w", err)
	}

	// Create timestamp-based documentation path
	complaint, err := r.FindByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to find complaint: %w", err)
	}

	timestamp := complaint.Timestamp.Format("2006-01-02_15-04")
	fileName := fmt.Sprintf("%s-%s-%s.md", timestamp, complaint.SessionID, complaint.TaskDescription[:20])
	
	// Truncate and sanitize filename
	if len(fileName) > 100 {
		fileName = fileName[:100]
	}
	
	fileName = strings.ReplaceAll(fileName, " ", "-")
	fileName = strings.ReplaceAll(fileName, "/", "-")
	
	return filepath.Join(r.docsDir, fileName), nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	CachedComplaints int `json:"cached_complaints"`
	MaxCacheSize     int `json:"max_cache_size"`
}

// listComplaintFiles lists all complaint files
func (r *FileRepository) listComplaintFiles(ctx context.Context) ([]fs.FileInfo, error) {
	entries, err := os.ReadDir(r.complaintsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []fs.FileInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read complaints directory: %w", err)
	}

	var files []fs.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		files = append(files, entry)
	}

	return files, nil
}

// writeFile writes data to a file
func (r *FileRepository) writeFile(ctx context.Context, fileName string, data []byte) error {
	if err := os.MkdirAll(r.complaintsDir, 0755); err != nil {
		return fmt.Errorf("failed to create complaints directory: %w", err)
	}

	filePath := filepath.Join(r.complaintsDir, fileName)
	return os.WriteFile(filePath, data, 0644)
}

// readFile reads data from a file
func (r *FileRepository) readFile(ctx context.Context, fileName string) ([]byte, error) {
	filePath := filepath.Join(r.complaintsDir, fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("complaint not found: %s", fileName)
		}
		return nil, fmt.Errorf("failed to read complaint file: %w", err)
	}
	return data, nil
}

// Implement placeholder methods for compatibility
func (r *FileRepository) FindBySession(ctx context.Context, sessionID string, limit int) ([]*domain.Complaint, error) {
	all, err := r.FindAll(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Complaint
	for _, complaint := range all {
		if complaint.SessionID == sessionID {
			filtered = append(filtered, complaint)
		}
		if len(filtered) >= limit {
			break
		}
	}

	return filtered, nil
}

func (r *FileRepository) FindByProject(ctx context.Context, projectID string, limit int) ([]*domain.Complaint, error) {
	all, err := r.FindAll(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Complaint
	for _, complaint := range all {
		if complaint.ProjectName == projectID {
			filtered = append(filtered, complaint)
		}
		if len(filtered) >= limit {
			break
		}
	}

	return filtered, nil
}

func (r *FileRepository) FindByAgent(ctx context.Context, agentID string, limit int) ([]*domain.Complaint, error) {
	all, err := r.FindAll(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Complaint
	for _, complaint := range all {
		if complaint.AgentID == agentID {
			filtered = append(filtered, complaint)
		}
		if len(filtered) >= limit {
			break
		}
	}

	return filtered, nil
}