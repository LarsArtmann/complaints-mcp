package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/errors"
)

// Repository interface for complaint storage
type Repository interface {
	Store(ctx context.Context, complaint *domain.Complaint) error
	FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error)
	FindByProject(ctx context.Context, projectName string, limit int, offset int) ([]*domain.Complaint, error)
	FindUnresolved(ctx context.Context, limit int, offset int) ([]*domain.Complaint, error)
	MarkResolved(ctx context.Context, id domain.ComplaintID) error
}

// FileRepository handles file-based complaint storage
type FileRepository struct {
	basePath string
}

// NewFileRepository creates a new file-based repository
func NewFileRepository(basePath string) Repository {
	return &FileRepository{
		basePath: basePath,
	}
}

// Store saves a complaint to a file
func (r *FileRepository) Store(ctx context.Context, complaint *domain.Complaint) error {
	// Create directory if it doesn't exist
	dir := filepath.Join(r.basePath, "complaints")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.NewStorageError(fmt.Sprintf("failed to create directory: %v", err))
	}
	
	// Generate filename
	filename := r.generateFilename(complaint)
	filePath := filepath.Join(dir, filename)
	
	// Write complaint to file
	content := r.generateComplaintContent(complaint)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return errors.NewStorageError(fmt.Sprintf("failed to write file: %v", err))
	}
	
	return nil
}

// generateFilename creates a unique filename for a complaint
func (r *FileRepository) generateFilename(complaint *domain.Complaint) string {
	timestamp := complaint.Timestamp.Format("2006-01-02_15-04")
	sessionName := complaint.SessionName
	if sessionName == "" {
		sessionName = "session"
	}
	
	// Sanitize session name
	sanitized := sanitizeFilename(sessionName)
	
	return fmt.Sprintf("%s-%s.md", timestamp, sanitized)
}

// generateComplaintContent generates markdown content for a complaint
func (r *FileRepository) generateComplaintContent(complaint *domain.Complaint) string {
	return fmt.Sprintf(`# %s

**Filed:** %s  
**Timestamp:** %s  
**Project:** %s  
**Agent:** %s  
**Session:** %s  
**Severity:** %s  

## Task Description

%s

## Context Information

%s

## Missing Information

%s

## Confusing Points

%s

## Future Wishes

%s

## Status

%s

---

*This complaint was automatically filed by an AI agent using the complaints-mcp server.*

**File:** %s  
**Stored:** %s

---
`,
		complaint.ID.Value,
		complaint.Timestamp.Format("2006-01-02 15:04:05"),
		complaint.Timestamp.Format("2006-01-02 15:04:05"),
		complaint.ProjectName,
		complaint.AgentName,
		complaint.SessionName,
		complaint.Severity,
		complaint.TaskDescription,
		complaint.ContextInfo,
		complaint.MissingInfo,
		complaint.ConfusedBy,
		complaint.FutureWishes,
		fmt.Sprintf("Status: %t", complaint.Resolved),
		complaint.ID.Value,
		time.Now().Format("2006-01-02 15:04:05"),
	)
}

// sanitizeFilename sanitizes a string for use in filenames
func sanitizeFilename(name string) string {
	// Replace spaces and special characters with dashes
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result += string(r)
		} else {
			result += "-"
		}
	}
	return result
}

// FindByID retrieves a complaint by its ID
func (r *FileRepository) FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	complaints, err := r.findByPredicate(ctx, func(c *domain.Complaint) bool {
		return c.ID.Value == id.Value
	})
	if err != nil {
		return nil, err
	}
	if len(complaints) == 0 {
		return nil, nil
	}
	return complaints[0], nil
}

// FindByProject retrieves complaints for a specific project
func (r *FileRepository) FindByProject(ctx context.Context, projectName string, limit int, offset int) ([]*domain.Complaint, error) {
	return r.findByPredicate(ctx, func(c *domain.Complaint) bool {
		return c.ProjectName == projectName
	})
}

// FindUnresolved retrieves unresolved complaints
func (r *FileRepository) FindUnresolved(ctx context.Context, limit int, offset int) ([]*domain.Complaint, error) {
	return r.findByPredicate(ctx, func(c *domain.Complaint) bool {
		return !c.Resolved
	})
}

// findByPredicate finds complaints matching a predicate
func (r *FileRepository) findByPredicate(ctx context.Context, predicate func(*domain.Complaint) bool) ([]*domain.Complaint, error) {
	dir := filepath.Join(r.basePath, "complaints")
	
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.NewStorageError(fmt.Sprintf("failed to read directory: %v", err))
	}
	
	var complaints []*domain.Complaint
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".md" {
			continue
		}
		
		// Read and parse complaint file
		content, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, errors.NewStorageError(fmt.Sprintf("failed to read file: %v", err))
		}
		
		// Create complaint from file content (simplified)
		complaint := r.parseComplaintFromFile(content)
		if complaint != nil && predicate(complaint) {
			complaints = append(complaints, complaint)
		}
	}
	
	return complaints, nil
}

// parseComplaintFromFile parses a complaint from markdown content
func (r *FileRepository) parseComplaintFromFile(content []byte) *domain.Complaint {
	// This is a simplified parser - in a real implementation, we'd use a proper markdown parser
	lines := strings.Split(string(content), "\n")
	
	if len(lines) < 10 {
		return nil
	}
	
	// Extract basic information from filename line
	titleLine := lines[0]
	if !strings.Contains(titleLine, "# ") {
		return nil
	}
	
	// Extract metadata from second line
	metaLine := lines[1]
	if !strings.Contains(metaLine, "**") {
		return nil
	}
	
	return &domain.Complaint{
		// Simplified parsing - in real implementation, extract all fields properly
		TaskDescription: metaLine,
		Timestamp:      time.Now(), // Would extract from filename
		ProjectName:    "unknown", // Would extract from config
		// Add other fields as needed
	}
}

// MarkResolved marks a complaint as resolved
func (r *FileRepository) MarkResolved(ctx context.Context, id domain.ComplaintID) error {
	return r.updateComplaintStatus(ctx, id, func(c *domain.Complaint) bool {
		c.Resolve()
		return true
	})
}

// updateComplaintStatus updates the status of a complaint
func (r *FileRepository) updateComplaintStatus(ctx context.Context, id domain.ComplaintID, updateFunc func(*domain.Complaint) bool) error {
	dir := filepath.Join(r.basePath, "complaints")
	
	files, err := os.ReadDir(dir)
	if err != nil {
		return errors.NewStorageError(fmt.Sprintf("failed to read directory: %v", err))
	}
	
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".md" {
			continue
		}
		
		content, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return errors.NewStorageError(fmt.Sprintf("failed to read file: %v", err))
		}
		
		complaint := r.parseComplaintFromFile(content)
		if complaint != nil && complaint.ID.Value == id.Value {
			if updateFunc(complaint) {
				complaint.Resolve()
			}
			
			// Write updated content back to file
			updatedContent := r.generateComplaintContent(complaint)
			if err := os.WriteFile(filepath.Join(dir, file.Name()), []byte(updatedContent), 0644); err != nil {
				return errors.NewStorageError(fmt.Sprintf("failed to write file: %v", err))
			}
			break
		}
	}
	
	return nil
}