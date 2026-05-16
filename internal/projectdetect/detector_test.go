package projectdetect

import (
	"os"
	"path/filepath"
	"testing"

	gigit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitDetector_Detect_NotGitRepo(t *testing.T) {
	detector := NewGitDetector()

	// Create a temporary directory outside any git repo
	// Use /var/tmp because /tmp has a .git folder which causes DetectDotGit to find it
	tmpDir, err := os.MkdirTemp("/var/tmp", "not-git-repo-*")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	_, err = detector.Detect(t.Context(), tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open git repository")
}

func TestGitDetector_Detect_Success(t *testing.T) {
	detector := NewGitDetector()

	// Create a temporary git repository
	tmpDir := t.TempDir()
	repo, err := gigit.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Create a remote
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/testuser/testrepo.git"},
	})
	require.NoError(t, err)

	// Create initial commit
	w, err := repo.Worktree()
	require.NoError(t, err)

	// Create a file and commit it
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	_, err = w.Add("test.txt")
	require.NoError(t, err)

	_, err = commitAsTestUser(w, "Initial commit")
	require.NoError(t, err)

	// Test detection from repo root
	info, err := detector.Detect(t.Context(), tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "testrepo", info.Name)
	assert.Equal(t, "https://github.com/testuser/testrepo.git", info.RemoteURL)
	assert.NotEmpty(t, info.Branch)
	assert.Equal(t, tmpDir, info.RootPath)
}

func TestGitDetector_Detect_FromSubdirectory(t *testing.T) {
	detector := NewGitDetector()

	// Create a temporary git repository
	tmpDir := t.TempDir()
	repo, err := gigit.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Create a remote
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@github.com:testuser/myproject.git"},
	})
	require.NoError(t, err)

	// Create initial commit
	w, err := repo.Worktree()
	require.NoError(t, err)

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	_, err = w.Add("test.txt")
	require.NoError(t, err)

	_, err = commitAsTestUser(w, "Initial commit")
	require.NoError(t, err)

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "pkg", "sub")
	err = os.MkdirAll(subDir, 0o755)
	require.NoError(t, err)

	// Test detection from subdirectory
	info, err := detector.Detect(t.Context(), subDir)
	require.NoError(t, err)

	assert.Equal(t, "myproject", info.Name)
	assert.Equal(t, "git@github.com:testuser/myproject.git", info.RemoteURL)
	assert.Equal(t, tmpDir, info.RootPath)
}

func TestExtractProjectName(t *testing.T) {
	tests := []struct {
		name       string
		remoteURL  string
		wantResult string
	}{
		{
			name:       "HTTPS GitHub URL",
			remoteURL:  "https://github.com/user/repo.git",
			wantResult: "repo",
		},
		{
			name:       "SSH GitHub URL",
			remoteURL:  "git@github.com:user/repo.git",
			wantResult: "repo",
		},
		{
			name:       "HTTPS without .git suffix",
			remoteURL:  "https://github.com/user/repo",
			wantResult: "repo",
		},
		{
			name:       "GitLab URL",
			remoteURL:  "https://gitlab.com/group/project.git",
			wantResult: "project",
		},
		{
			name:       "Bitbucket SSH URL",
			remoteURL:  "git@bitbucket.org:team/repo.git",
			wantResult: "repo",
		},
		{
			name:       "empty URL",
			remoteURL:  "",
			wantResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractProjectName(tt.remoteURL)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

func TestIsGitRepository(t *testing.T) {
	t.Run("returns true for git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := v5.PlainInit(tmpDir, false)
		require.NoError(t, err)

		assert.True(t, IsGitRepository(tmpDir))
	})

	t.Run("returns false for non-git directory", func(t *testing.T) {
		// Use /var/tmp because /tmp has a .git folder which causes DetectDotGit to find it
		tmpDir, err := os.MkdirTemp("/var/tmp", "not-git-dir-*")
		require.NoError(t, err)

		defer os.RemoveAll(tmpDir)

		assert.False(t, IsGitRepository(tmpDir))
	})

	t.Run("returns true for subdirectory of git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := v5.PlainInit(tmpDir, false)
		require.NoError(t, err)

		subDir := filepath.Join(tmpDir, "sub")
		err = os.MkdirAll(subDir, 0o755)
		require.NoError(t, err)

		assert.True(t, IsGitRepository(subDir))
	})
}

func TestGitDetector_EmptyWorkingDir(t *testing.T) {
	detector := NewGitDetector()
	_, err := detector.Detect(t.Context(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "working directory cannot be empty")
}

func commitAsTestUser(w *gigit.Worktree, message string) (plumbing.Hash, error) {
	return w.Commit(message, &gigit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
}
