package complaint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Complaint represents a structured complaint report
type Complaint struct {
	TaskAskedToPerform string
	ContextInformation string
	MissingInformation string
	ConfusedBy         string
	FutureWishes       string
	SessionName        string
	AgentName          string
}

// GenerateFilename creates a filename for the complaint
func (c *Complaint) GenerateFilename() string {
	now := time.Now()
	timestamp := now.Format("2006-01-02_15-04")
	
	sessionName := c.SessionName
	if sessionName == "" {
		sessionName = "default-session"
	}
	
	// Clean session name for filesystem
	sessionName = strings.ReplaceAll(sessionName, " ", "-")
	sessionName = strings.ReplaceAll(sessionName, "/", "-")
	sessionName = strings.ReplaceAll(sessionName, "\\", "-")
	
	return fmt.Sprintf("%s-%s.md", timestamp, sessionName)
}

// GenerateContent creates the markdown content for the complaint
func (c *Complaint) GenerateContent() string {
	now := time.Now()
	
	agentName := c.AgentName
	if agentName == "" {
		agentName = "AI Agent"
	}
	
	return fmt.Sprintf(`# Report about missing/under-specified/confusing information

Date: %s

I was asked to perform:
%s

I was given these context information's:
%s

I was missing these information:
%s

I was confused by:
%s

What I wish for the future is:
%s


Best regards,
%s`,
		now.Format(time.RFC3339),
		c.TaskAskedToPerform,
		c.ContextInformation,
		c.MissingInformation,
		c.ConfusedBy,
		c.FutureWishes,
		agentName)
}

// getProjectName determines the project name from git remote or folder structure
func getProjectName(projectDir string) string {
	// Try to get project name from git remote
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err == nil {
		remoteURL := strings.TrimSpace(string(output))
		// Extract repo name from URL (e.g., https://github.com/user/repo.git -> repo)
		parts := strings.Split(remoteURL, "/")
		if len(parts) > 0 {
			repoName := parts[len(parts)-1]
			repoName = strings.TrimSuffix(repoName, ".git")
			if repoName != "" {
				return repoName
			}
		}
	}
	
	// Fallback to folder name
	folderName := filepath.Base(projectDir)
	if folderName != "" && folderName != "." && folderName != "/" {
		return folderName
	}
	
	// Ultimate fallback
	return "unknown-project"
}

// Save stores the complaint to both project-local and global locations
func (c *Complaint) Save(projectDir string) (string, error) {
	filename := c.GenerateFilename()
	content := c.GenerateContent()
	
	// Save to project-local location
	localPath := filepath.Join(projectDir, "docs", "complaints", filename)
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create local complaints directory: %w", err)
	}
	
	if err := os.WriteFile(localPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write local complaint file: %w", err)
	}
	
	// Save to global location with project name
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return localPath, nil // Return local path even if global fails
	}
	
	projectName := getProjectName(projectDir)
	globalPath := filepath.Join(homeDir, ".complaints-mcp", projectName, filename)
	if err := os.MkdirAll(filepath.Dir(globalPath), 0755); err != nil {
		return localPath, nil // Return local path even if global fails
	}
	
	if err := os.WriteFile(globalPath, []byte(content), 0644); err != nil {
		return localPath, nil // Return local path even if global fails
	}
	
	return localPath, nil
}