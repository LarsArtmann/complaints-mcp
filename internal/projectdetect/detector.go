// Package projectdetect provides git-based project auto-detection functionality.
package projectdetect

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// ProjectInfo contains detected project information from git repository.
type ProjectInfo struct {
	Name      string
	RemoteURL string
	Branch    string
	RootPath  string
}

// Detector provides project detection functionality.
type Detector interface {
	Detect(ctx context.Context, workingDir string) (*ProjectInfo, error)
}

// GitDetector detects project information from git repositories.
type GitDetector struct{}

// NewGitDetector creates a new GitDetector.
func NewGitDetector() *GitDetector {
	return &GitDetector{}
}

// Detect finds project information from a git repository at or above workingDir.
func (d *GitDetector) Detect(ctx context.Context, workingDir string) (*ProjectInfo, error) {
	if workingDir == "" {
		return nil, fmt.Errorf("working directory cannot be empty")
	}

	// Open the repository - go-git handles walking up to find .git
	repo, err := git.PlainOpenWithOptions(workingDir, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get repository worktree root
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	rootPath := worktree.Filesystem.Root()

	// Get current branch
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	branch := ""
	if head.Name().IsBranch() {
		branch = head.Name().Short()
	} else {
		// Detached HEAD, use short SHA
		branch = head.Hash().String()[:7]
	}

	// Get remote URL
	remoteURL, err := d.getRemoteURL(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	// Extract project name from remote URL
	name := extractProjectName(remoteURL)

	return &ProjectInfo{
		Name:      name,
		RemoteURL: remoteURL,
		Branch:    branch,
		RootPath:  rootPath,
	}, nil
}

// getRemoteURL retrieves the origin remote URL or falls back to any available remote.
func (d *GitDetector) getRemoteURL(repo *git.Repository) (string, error) {
	// Try origin first
	remote, err := repo.Remote("origin")
	if err == nil && len(remote.Config().URLs) > 0 {
		return remote.Config().URLs[0], nil
	}

	// Fall back to any remote
	remotes, err := repo.Remotes()
	if err != nil {
		return "", fmt.Errorf("failed to list remotes: %w", err)
	}

	for _, r := range remotes {
		if len(r.Config().URLs) > 0 {
			return r.Config().URLs[0], nil
		}
	}

	return "", fmt.Errorf("no remote URLs found")
}

// extractProjectName extracts a readable project name from a remote URL.
func extractProjectName(remoteURL string) string {
	if remoteURL == "" {
		return ""
	}

	// Handle SSH URLs like git@github.com:user/repo.git
	if strings.HasPrefix(remoteURL, "git@") {
		parts := strings.Split(remoteURL, ":")
		if len(parts) == 2 {
			path := parts[1]
			path = strings.TrimSuffix(path, ".git")
			return filepath.Base(path)
		}
	}

	// Handle HTTPS URLs like https://github.com/user/repo.git
	if u, err := url.Parse(remoteURL); err == nil {
		path := strings.TrimSuffix(u.Path, ".git")
		return filepath.Base(path)
	}

	// Fallback: just return the last path component
	path := strings.TrimSuffix(remoteURL, ".git")
	return filepath.Base(path)
}

// DetectProject is a convenience function for direct project detection.
func DetectProject(ctx context.Context, workingDir string) (*ProjectInfo, error) {
	detector := NewGitDetector()
	return detector.Detect(ctx, workingDir)
}

// IsGitRepository checks if the given path is inside a git repository.
func IsGitRepository(path string) bool {
	_, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	return err == nil
}

// Plumbing types for type safety
type (
	// ReferenceName is an alias for plumbing.ReferenceName for type safety.
	ReferenceName = plumbing.ReferenceName
)
