package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

// FileOperations handles file I/O operations for complaints
type FileOperations struct {
	baseDir string
	logger  *log.Logger
	tracer  tracing.Tracer
}

// NewFileOperations creates a new file operations handler
func NewFileOperations(baseDir string, tracer tracing.Tracer) *FileOperations {
	return &FileOperations{
		baseDir: baseDir,
		logger:  log.Default(),
		tracer:  tracer,
	}
}

// SaveComplaintToFile saves a complaint to its dedicated file
func (f *FileOperations) SaveComplaintToFile(ctx context.Context, complaint *domain.Complaint) error {
	ctx, span := f.tracer.Start(ctx, "SaveComplaintToFile")
	defer span.End()

	logger := f.logger.With("operation", "save_to_file", "complaint_id", complaint.ID.String())

	filePath := f.GenerateFilePathForID(complaint.ID.String())
	dir := filepath.Dir(filePath)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("Failed to create directory", "dir", dir, "error", err)
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Convert complaint to JSON with proper formatting
	data, err := json.MarshalIndent(complaint, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal complaint to JSON", "error", err)
		return fmt.Errorf("failed to marshal complaint: %w", err)
	}

	// Write to file atomically (write to temp file, then rename)
	tempFilePath := filePath + ".tmp"
	if err := os.WriteFile(tempFilePath, data, 0644); err != nil {
		logger.Error("Failed to write to temporary file", "file", tempFilePath, "error", err)
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFilePath, filePath); err != nil {
		logger.Error("Failed to rename temporary file", "temp", tempFilePath, "target", filePath, "error", err)
		// Clean up temp file on error
		os.Remove(tempFilePath)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	logger.Info("Complaint saved to file successfully")
	return nil
}

// LoadComplaintFromFile loads a complaint from its file
func (f *FileOperations) LoadComplaintFromFile(ctx context.Context, filePath string) (*domain.Complaint, error) {
	ctx, span := f.tracer.Start(ctx, "LoadComplaintFromFile")
	defer span.End()

	logger := f.logger.With("operation", "load_from_file", "file", filePath)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Warn("Complaint file not found", "file", filePath)
		return nil, fmt.Errorf("complaint file not found: %s", filePath)
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("Failed to read complaint file", "file", filePath, "error", err)
		return nil, fmt.Errorf("failed to read complaint file: %w", err)
	}

	// Unmarshal JSON
	var complaint domain.Complaint
	if err := json.Unmarshal(data, &complaint); err != nil {
		logger.Error("Failed to unmarshal complaint from JSON", "file", filePath, "error", err)
		return nil, fmt.Errorf("failed to unmarshal complaint: %w", err)
	}

	// Validate loaded complaint
	if err := complaint.Validate(); err != nil {
		logger.Error("Loaded complaint failed validation", "file", filePath, "error", err)
		return nil, fmt.Errorf("invalid complaint data: %w", err)
	}

	logger.Info("Complaint loaded from file successfully")
	return &complaint, nil
}

// LoadAllComplaints loads all complaints from the base directory
func (f *FileOperations) LoadAllComplaints(ctx context.Context) ([]*domain.Complaint, error) {
	ctx, span := f.tracer.Start(ctx, "LoadAllComplaints")
	defer span.End()

	logger := f.logger.With("operation", "load_all_complaints")

	// Check if base directory exists
	if _, err := os.Stat(f.baseDir); os.IsNotExist(err) {
		logger.Info("Base directory does not exist, creating it", "dir", f.baseDir)
		if err := os.MkdirAll(f.baseDir, 0755); err != nil {
			logger.Error("Failed to create base directory", "dir", f.baseDir, "error", err)
			return nil, fmt.Errorf("failed to create base directory: %w", err)
		}
		return []*domain.Complaint{}, nil
	}

	// Walk through directory and find all .json files
	var complaints []*domain.Complaint
	err := filepath.WalkDir(f.baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-JSON files
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		// Load complaint from file
		complaint, err := f.LoadComplaintFromFile(ctx, path)
		if err != nil {
			logger.Warn("Failed to load complaint file, skipping", "file", path, "error", err)
			return nil // Continue with other files
		}

		complaints = append(complaints, complaint)
		return nil
	})

	if err != nil {
		logger.Error("Failed to walk directory", "dir", f.baseDir, "error", err)
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	logger.Info("All complaints loaded successfully", "count", len(complaints))
	return complaints, nil
}

// GenerateFilePathForID generates the file path for a complaint ID
func (f *FileOperations) GenerateFilePathForID(complaintID string) string {
	// Use two-level directory structure: /base_dir/ab/ab-cd-ef-gh.json
	// This prevents having too many files in one directory
	id := complaintID
	if len(id) >= 2 {
		prefix := id[:2]
		return filepath.Join(f.baseDir, prefix, complaintID+".json")
	}
	return filepath.Join(f.baseDir, complaintID+".json")
}

// GenerateFilePath generates the file path for a complaint
func (f *FileOperations) GenerateFilePath(complaint *domain.Complaint) string {
	return f.GenerateFilePathForID(complaint.ID.String())
}
