# Step 5: Implement Project Detection Service

## 🎯 Objective
Create a working project detection service based on research findings.

## 🏗️ Implementation Tasks

### A. Create Detector Interface
```go
// internal/detection/project_detector.go
package detection

import "context"

type ProjectDetector interface {
    DetectProjectName(ctx context.Context) (string, error)
    DetectProjectNameWithWorkspace(ctx context.Context, workspace string) (string, error)
}
```

### B. System Git Implementation (Primary)
```go
// internal/detection/system_git_detector.go
type systemGitDetector struct {
    workspace string
    cache     map[string]string
}

func (d *systemGitDetector) DetectProjectName(ctx context.Context) (string, error) {
    return d.DetectProjectNameWithWorkspace(ctx, d.workspace)
}

func (d *systemGitDetector) DetectProjectNameWithWorkspace(ctx context.Context, workspace string) (string, error) {
    // Check cache
    if cached, exists := d.cache[workspace]; exists {
        return cached, nil
    }
    
    // Try git remote detection
    if name, err := d.detectFromGitRemote(ctx, workspace); err == nil {
        d.cache[workspace] = name
        return name, nil
    }
    
    // Fallback to directory name
    if name, err := d.detectFromDirectoryName(workspace); err == nil {
        d.cache[workspace] = name
        return name, nil
    }
    
    // Default fallback
    name := "unknown-project"
    d.cache[workspace] = name
    return name, nil
}

func (d *systemGitDetector) detectFromGitRemote(ctx context.Context, workspace string) (string, error) {
    cmd := exec.CommandContext(ctx, "git", "config", "--get", "remote.origin.url")
    cmd.Dir = workspace
    
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("not in git repository: %w", err)
    }
    
    gitRemote := strings.TrimSpace(string(output))
    return d.extractProjectNameFromRemote(gitRemote)
}
```

### C. Go-Git Implementation (Alternative)
```go
// internal/detection/go_git_detector.go
import "github.com/go-git/go-git/v5"
import "github.com/go-git/go-git/v5/config"

type goGitDetector struct {
    workspace string
    cache     map[string]string
}

func (d *goGitDetector) DetectProjectNameWithWorkspace(ctx context.Context, workspace string) (string, error) {
    repo, err := git.PlainOpen(workspace)
    if err != nil {
        return "", fmt.Errorf("not a git repository: %w", err)
    }
    
    remoteConfig, err := repo.Config()
    if err != nil {
        return "", fmt.Errorf("failed to get git config: %w", err)
    }
    
    if remote, ok := remoteConfig.Remotes["origin"]; ok {
        if len(remote.URLs) > 0 {
            gitRemote := remote.URLs[0]
            return d.extractProjectNameFromRemote(gitRemote), nil
        }
    }
    
    return "", fmt.Errorf("no origin remote found")
}
```

### D. Factory and Configuration
```go
// internal/detection/factory.go
type DetectorType string

const (
    SystemGit DetectorType = "system-git"
    GoGit     DetectorType = "go-git"
    Auto       DetectorType = "auto"
)

type Factory struct {
    type DetectorType
    cache map[string]string
}

func NewFactory(dt DetectorType) *Factory {
    return &Factory{
        type: dt,
        cache: make(map[string]string),
    }
}

func (f *Factory) CreateDetector(workspace string) ProjectDetector {
    switch f.type {
    case SystemGit:
        return NewSystemGitDetector(workspace, f.cache)
    case GoGit:
        return NewGoGitDetector(workspace, f.cache)
    case Auto:
        // Try system git first, fallback to go-git
        if gitCommandExists() {
            return NewSystemGitDetector(workspace, f.cache)
        }
        return NewGoGitDetector(workspace, f.cache)
    default:
        return NewSystemGitDetector(workspace, f.cache)
    }
}

func gitCommandExists() bool {
    _, err := exec.LookPath("git")
    return err == nil
}
```

### E. Comprehensive Tests
```go
// internal/detection/project_detector_test.go
func TestProjectDetector_SystemGit(t *testing.T) {
    tests := []struct {
        name         string
        setupFunc    func() (string, error)
        cleanupFunc  func() error
        expected     string
        expectError  bool
    }{
        {
            name: "github https remote",
            setupFunc: func() (string, error) {
                dir := t.TempDir()
                err := os.Chdir(dir)
                require.NoError(t, err)
                
                runGitCommand(dir, "init")
                runGitCommand(dir, "remote", "add", "origin", "https://github.com/user/my-project.git")
                
                return dir, nil
            },
            cleanupFunc: func() error {
                return os.Chdir(t.TempDir())
            },
            expected: "my-project",
        },
        {
            name: "directory name fallback",
            setupFunc: func() (string, error) {
                dir := t.TempDir()
                subdir := filepath.Join(dir, "awesome-app")
                err := os.Mkdir(subdir, 0755)
                require.NoError(t, err)
                
                return subdir, nil
            },
            cleanupFunc: func() error {
                return os.Chdir(t.TempDir())
            },
            expected: "awesome-app",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            workspace, err := tt.setupFunc()
            require.NoError(t, err)
            defer tt.cleanupFunc()
            
            detector := NewSystemGitDetector(workspace, make(map[string]string))
            
            result, err := detector.DetectProjectName(context.Background())
            
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### F. Integration with Service Layer
```go
// internal/service/complaint_service.go (updated)
type ComplaintService struct {
    repo           domain.Repository
    tracer         tracing.Tracer
    logger         logging.Logger
    projectDetector detection.ProjectDetector
    config         *config.Config
}

func NewComplaintService(repo domain.Repository, tracer tracing.Tracer, logger logging.Logger, projectDetector detection.ProjectDetector) *ComplaintService {
    return &ComplaintService{
        repo:           repo,
        tracer:         tracer,
        logger:         logger,
        projectDetector: projectDetector,
    }
}

func (s *ComplaintService) CreateComplaint(
    ctx context.Context,
    agentName, sessionName, taskDescription, contextInfo, missingInfo, confusedBy, futureWishes string,
    severity domain.Severity,
    projectName string,
) (*domain.Complaint, error) {
    // Auto-detect project name if not provided
    finalProjectName := projectName
    if strings.TrimSpace(finalProjectName) == "" {
        detected, err := s.projectDetector.DetectProjectName(ctx)
        if err != nil {
            s.logger.Warn("Failed to detect project name, using default", "error", err)
            finalProjectName = "unknown-project"
        } else {
            finalProjectName = detected
            s.logger.Info("Auto-detected project name", "project", finalProjectName)
        }
    }
    
    // Rest of existing complaint creation logic...
}
```

## 📝 Implementation Details

### A. Error Handling Strategy
```go
type DetectionError struct {
    Type    string
    Message string
    Cause   error
}

func (e *DetectionError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Error types
const (
    ErrTypeGitRepository = "git_repository_error"
    ErrTypeRemoteURL     = "remote_url_error"
    ErrTypeDirectoryName = "directory_name_error"
    ErrTypeValidation     = "validation_error"
)
```

### B. Caching Strategy
```go
type CacheEntry struct {
    ProjectName string
    Timestamp  time.Time
}

type MemoryCache struct {
    entries map[string]*CacheEntry
    mutex   sync.RWMutex
    ttl     time.Duration
}

func (c *MemoryCache) Get(workspace string) (string, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    entry, exists := c.entries[workspace]
    if !exists {
        return "", false
    }
    
    // Check TTL
    if time.Since(entry.Timestamp) > c.ttl {
        delete(c.entries, workspace)
        return "", false
    }
    
    return entry.ProjectName, true
}

func (c *MemoryCache) Set(workspace, projectName string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.entries[workspace] = &CacheEntry{
        ProjectName: projectName,
        Timestamp:  time.Now(),
    }
}
```

### C. Configuration Integration
```go
// internal/config/config.go (updated)
type ProjectDetectionConfig struct {
    Enabled    bool     `toml:"enabled" default:"true"`
    Type       string   `toml:"type" default:"auto"` // "auto", "system-git", "go-git"
    CacheTTL   int      `toml:"cache_ttl" default:"300"`        // 5 minutes
    Fallbacks   []string `toml:"fallbacks" default:[]`       // Custom fallbacks
    GitRemote   bool     `toml:"git_remote" default:"true"`     // Enable git remote detection
    Directory   bool     `toml:"directory" default:"true"`     // Enable directory name detection
    DefaultName string   `toml:"default_name" default:"unknown-project"`
}

type Config struct {
    // ... existing fields ...
    ProjectDetection ProjectDetectionConfig `toml:"project_detection"`
}
```

## 🧪 Verification Steps

### 1. Unit Tests
```bash
go test ./internal/detection -v
```

### 2. Integration Tests
```go
func TestProjectDetector_Integration(t *testing.T) {
    // Test with real git repository
    dir := t.TempDir()
    err := os.Chdir(dir)
    require.NoError(t, err)
    
    runGitCommand(dir, "init")
    runGitCommand(dir, "remote", "add", "origin", "https://github.com/user/test-project.git")
    
    detector := NewSystemGitDetector(dir, make(map[string]string))
    result, err := detector.DetectProjectName(context.Background())
    
    assert.NoError(t, err)
    assert.Equal(t, "test-project", result)
}
```

### 3. Service Integration Tests
```go
func TestComplaintService_WithProjectDetection(t *testing.T) {
    detector := NewMockProjectDetector("auto-detected-project", nil)
    service := NewComplaintService(repo, tracer, logger, detector)
    
    complaint, err := service.CreateComplaint(
        context.Background(),
        "Test-Agent",
        "test-session",
        "Test task",
        "", "", "", "", "",
        "low",
        "", // Empty - should auto-detect
    )
    
    assert.NoError(t, err)
    assert.Contains(t, complaint.ProjectName, "auto-detected-project")
}
```

### 4. End-to-End Test
```bash
# Setup test repository
mkdir -p /tmp/test-project
cd /tmp/test-project
git init
git remote add origin https://github.com/user/test-project.git

# Test project detection
echo '{"tool":"file_complaint","arguments":{"agent_name":"Test-Agent","task_description":"Test","severity":"low"}}' | ./complaints-mcp

# Verify response contains auto-detected project name
```

## ⏱️ Time Estimate: 6-8 hours
## 🎯 Impact: High (enables project detection)
## 💪 Work Required: Medium (service implementation + tests)

## 🎯 Success Criteria

- [ ] Project detector interface defined
- [ ] System git implementation working
- [ ] Go-git implementation working (optional)
- [ ] Factory pattern for detector selection
- [ ] Caching mechanism implemented
- [ ] Service integration completed
- [ ] Comprehensive test coverage
- [ ] Configuration options working
- [ ] Error handling robust
- [ ] Performance meets requirements (<100ms detection)