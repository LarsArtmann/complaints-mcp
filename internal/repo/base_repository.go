package repo

import (
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

// BaseRepository provides shared functionality for repository implementations
type BaseRepository struct {
	baseDir string
	logger  *log.Logger
	tracer  tracing.Tracer
}

// NewBaseRepository creates a new base repository with common dependencies
func NewBaseRepository(baseDir string, tracer tracing.Tracer) BaseRepository {
	return BaseRepository{
		baseDir: baseDir,
		logger:  log.Default(),
		tracer:  tracer,
	}
}

// GetBaseDir returns the base directory for storage
func (b *BaseRepository) GetBaseDir() string {
	return b.baseDir
}

// GetLogger returns the logger instance
func (b *BaseRepository) GetLogger() *log.Logger {
	return b.logger
}

// GetTracer returns the tracer instance
func (b *BaseRepository) GetTracer() tracing.Tracer {
	return b.tracer
}

// LoadComplaintFromFile loads a single complaint from a JSON file
// This is shared functionality used by both repository implementations
func (b *BaseRepository) LoadComplaintFromFile(filePath string) (*domain.Complaint, error) {
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

// LoadAllComplaintsFromDisk loads all complaints from the file system
// This provides the common file system scanning logic
func (b *BaseRepository) LoadAllComplaintsFromDisk() ([]*domain.Complaint, error) {
	entries, err := os.ReadDir(b.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			b.logger.Info("Base directory does not exist, returning empty list")
			return []*domain.Complaint{}, nil
		}
		b.logger.Error("Failed to read base directory", "error", err, "path", b.baseDir)
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

		filePath := filepath.Join(b.baseDir, entry.Name())
		complaint, err := b.LoadComplaintFromFile(filePath)
		if err != nil {
			b.logger.Warn("Failed to load complaint file", "error", err, "path", filePath)
			continue
		}

		complaints = append(complaints, complaint)
	}

	// Sort complaints by timestamp (oldest first for consistent ordering)
	sort.Slice(complaints, func(i, j int) bool {
		return complaints[i].Timestamp.Before(complaints[j].Timestamp)
	})

	b.logger.Info("Complaints loaded from disk successfully", "base_dir", b.baseDir, "count", len(complaints))
	return complaints, nil
}

// SaveComplaintToFile saves a complaint to a JSON file
// This provides the common file saving logic
func (b *BaseRepository) SaveComplaintToFile(complaint *domain.Complaint) error {
	// Create logger with context
	logger := b.logger.With("component", "base-repository", "complaint_id", complaint.ID.String())
	logger.Info("Saving complaint")

	// Use UUID-only filename for consistent file updates
	filename := fmt.Sprintf("%s.json", complaint.ID.String())
	filePath := filepath.Join(b.baseDir, filename)

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

// ApplyPagination applies pagination to a slice of complaints
// This provides common pagination logic
func (b *BaseRepository) ApplyPagination(complaints []*domain.Complaint, limit, offset int) []*domain.Complaint {
	total := len(complaints)
	if offset >= total {
		return []*domain.Complaint{}
	}

	start := offset
	end := min(offset+limit, total)

	return complaints[start:end]
}