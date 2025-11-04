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

// BaseRepository provides shared functionality for all repository implementations
type BaseRepository struct {
	baseDir string
	logger  *log.Logger
	tracer  tracing.Tracer
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(baseDir string, tracer tracing.Tracer) BaseRepository {
	return BaseRepository{
		baseDir: baseDir,
		logger:  log.Default(),
		tracer:  tracer,
	}
}

// BaseDir returns the base directory path
func (b BaseRepository) BaseDir() string {
	return b.baseDir
}

// Logger returns the logger instance
func (b BaseRepository) Logger() *log.Logger {
	return b.logger
}

// Tracer returns the tracer instance
func (b BaseRepository) Tracer() tracing.Tracer {
	return b.tracer
}

// EnsureDir creates the base directory if it doesn't exist
func (b BaseRepository) EnsureDir() error {
	if err := os.MkdirAll(b.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}
	return nil
}

// GenerateFilePath generates a file path for a complaint
func (b BaseRepository) GenerateFilePath(complaint *domain.Complaint) string {
	timestamp := complaint.Timestamp.Format("2006-01-02_15-04-05")
	if complaint.SessionName != "" {
		// Sanitize session name to prevent path traversal attacks
		safeSessionName := filepath.Base(complaint.SessionName) // Remove any directory components
		safeSessionName = strings.ReplaceAll(safeSessionName, " ", "_")
		safeSessionName = strings.ReplaceAll(safeSessionName, "..", "_") // Remove path traversal
		safeSessionName = strings.ReplaceAll(safeSessionName, "/", "_")
		return filepath.Join(b.baseDir, fmt.Sprintf("%s-%s.json", timestamp, safeSessionName))
	}
	return filepath.Join(b.baseDir, fmt.Sprintf("%s.json", timestamp))
}

// LoadAllComplaints loads all complaints from the filesystem
func (b BaseRepository) LoadAllComplaints() ([]*domain.Complaint, error) {
	ctx := context.Background()
	ctx, span := b.tracer.Start(ctx, "LoadAllComplaints")
	defer span.End()

	logger := b.logger.With("component", "file-repository")

	files, err := os.ReadDir(b.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*domain.Complaint{}, nil
		}
		logger.Error("Failed to read base directory", "error", err, "path", b.baseDir)
		return nil, fmt.Errorf("failed to read base directory: %w", err)
	}

	// Filter JSON files
	jsonFiles := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			jsonFiles = append(jsonFiles, file.Name())
		}
	}

	// Sort by filename to maintain consistent ordering
	sort.Strings(jsonFiles)

	complaints := make([]*domain.Complaint, 0, len(jsonFiles))
	for _, fileName := range jsonFiles {
		filePath := filepath.Join(b.baseDir, fileName)
		complaint, err := b.loadComplaintFromFile(filePath)
		if err != nil {
			logger.Warn("Failed to load complaint file", "error", err, "path", filePath)
			continue
		}

		complaints = append(complaints, complaint)
	}

	// Sort complaints by timestamp (oldest first)
	sort.Slice(complaints, func(i, j int) bool {
		return complaints[i].Timestamp.Before(complaints[j].Timestamp)
	})

	logger.Info("Complaints loaded successfully", "count", len(complaints))
	return complaints, nil
}

// loadComplaintFromFile loads a single complaint from a JSON file
func (b BaseRepository) loadComplaintFromFile(filePath string) (*domain.Complaint, error) {
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

// SaveComplaintToFile saves a complaint to a JSON file
func (b BaseRepository) SaveComplaintToFile(ctx context.Context, complaint *domain.Complaint) error {
	_, span := b.tracer.Start(ctx, "SaveComplaintToFile")
	defer span.End()

	logger := b.logger.With("component", "file-repository", "complaint_id", complaint.ID.String())

	if err := complaint.Validate(); err != nil {
		logger.Error("Invalid complaint data", "error", err)
		return fmt.Errorf("invalid complaint data: %w", err)
	}

	// Ensure directory exists
	if err := b.EnsureDir(); err != nil {
		return err
	}

	filePath := b.GenerateFilePath(complaint)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(complaint, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal complaint", "error", err)
		return fmt.Errorf("failed to marshal complaint: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		logger.Error("Failed to write complaint file", "error", err, "path", filePath)
		return fmt.Errorf("failed to write complaint file: %w", err)
	}

	logger.Info("Complaint saved successfully", "path", filePath)
	return nil
}
