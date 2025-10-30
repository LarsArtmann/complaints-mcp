package complaint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

// Complaint represents a structured complaint report from an AI agent
type Complaint struct {
	TaskAskedToPerform string    `json:"task_asked_to_perform"`
	ContextInformation string    `json:"context_information"`
	MissingInformation string    `json:"missing_information"`
	ConfusedBy         string    `json:"confused_by"`
	FutureWishes       string    `json:"future_wishes"`
	SessionName        string    `json:"session_name,omitempty"`
	AgentName          string    `json:"agent_name,omitempty"`
	Timestamp          time.Time `json:"timestamp"`
	ID                 string    `json:"id"`
	ProjectName        string    `json:"project_name"`
}

// Save saves the complaint to both project-local and global storage
func (c *Complaint) Save(workingDir string) (string, error) {
	// Generate unique ID and timestamp if not set
	if c.ID == "" {
		id, err := uuid.NewV4()
		if err != nil {
			return "", fmt.Errorf("failed to generate UUID: %w", err)
		}
		c.ID = id.String()
	}
	if c.Timestamp.IsZero() {
		c.Timestamp = time.Now()
	}
	if c.ProjectName == "" {
		c.ProjectName = detectProjectName(workingDir)
	}

	// Generate filename
	filename := c.generateFilename()

	// Save to project-local directory
	projectPath, err := c.saveProjectLocal(workingDir, filename)
	if err != nil {
		return "", fmt.Errorf("failed to save project-local copy: %w", err)
	}

	// Save to global directory
	_, err = c.saveGlobal(filename)
	if err != nil {
		return "", fmt.Errorf("failed to save global copy: %w", err)
	}

	// Return the primary path (project-local)
	return projectPath, nil
}

// generateFilename creates a unique filename for the complaint
func (c *Complaint) generateFilename() string {
	timestamp := c.Timestamp.Format("2006-01-02_15-04")

	sessionName := strings.ToLower(strings.ReplaceAll(c.SessionName, " ", "-"))
	if sessionName == "" {
		sessionName = "session"
	}

	// Sanitize session name
	sessionName = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return '-'
	}, sessionName)

	return fmt.Sprintf("%s-%s.md", timestamp, sessionName)
}

// saveProjectLocal saves the complaint to the project's docs/complaints directory
func (c *Complaint) saveProjectLocal(workingDir, filename string) (string, error) {
	complaintsDir := filepath.Join(workingDir, "docs", "complaints")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(complaintsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create complaints directory: %w", err)
	}

	filePath := filepath.Join(complaintsDir, filename)

	content := c.generateMarkdown()

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write complaint file: %w", err)
	}

	return filePath, nil
}

// saveGlobal saves the complaint to the global ~/.complaints-mcp directory
func (c *Complaint) saveGlobal(filename string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	globalDir := filepath.Join(homeDir, ".complaints-mcp", c.ProjectName)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create global complaints directory: %w", err)
	}

	filePath := filepath.Join(globalDir, filename)

	content := c.generateMarkdown()

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write global complaint file: %w", err)
	}

	return filePath, nil
}

// generateMarkdown creates the markdown content for the complaint
func (c *Complaint) generateMarkdown() string {
	return fmt.Sprintf(`# AI Agent Complaint Report

**ID:** %s  
**Timestamp:** %s  
**Project:** %s  
**Agent:** %s  
**Session:** %s  

## Task Asked To Perform

%s

## Context Information

%s

## Missing Information

%s

## What Confused the Agent

%s

## Future Wishes

%s

---

*This complaint was filed automatically by an AI agent using the complaints-mcp server. Your feedback helps improve the development experience for everyone.*
`,
		c.ID,
		c.Timestamp.Format("2006-01-02 15:04:05"),
		c.ProjectName,
		c.AgentName,
		c.SessionName,
		c.TaskAskedToPerform,
		c.ContextInformation,
		c.MissingInformation,
		c.ConfusedBy,
		c.FutureWishes,
	)
}

// detectProjectName attempts to detect the project name from git remote or folder name
func detectProjectName(workingDir string) string {
	// Try to get from git remote
	if gitRepoName := getGitRepoName(workingDir); gitRepoName != "" {
		return gitRepoName
	}

	// Fallback to folder name
	folderName := filepath.Base(workingDir)
	if folderName != "" && folderName != "." && folderName != "/" {
		return folderName
	}

	// Ultimate fallback
	return "unknown-project"
}

// getGitRepoName extracts the repository name from git remote
func getGitRepoName(workingDir string) string {
	// This is a simplified implementation
	// In a real implementation, you'd parse git config or use git commands
	gitDir := filepath.Join(workingDir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return ""
	}

	// For now, just return the folder name as a reasonable guess
	return filepath.Base(workingDir)
}
