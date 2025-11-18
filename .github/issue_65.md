# Issue #65: Enhance Complaint System with Automatic Project Name Detection

## 🎯 **Enhancement: Intelligent Project Name Detection for Better Complaint Context**

### **Current State Analysis**
The complaints system currently requires manual project name input, but this creates several problems:

**❌ Current Issues:**
- **Manual Input Required**: Users must explicitly provide project name
- **Inconsistent Names**: AI may provide varying project name formats
- **Missing Context**: Project context lost when not provided
- **Duplication Effort**: Manual entry for known git projects
- **Data Quality**: Inconsistent or incorrect project names

### **Target State**
Implement automatic project name detection with intelligent fallback hierarchy:

**✅ Desired Behavior:**
```go
// Automatic detection priority:
1. Git remote repository name (highest priority)
2. Current directory name (fallback)
3. "unknown-project" (last resort)

// Examples:
repo: "github.com/user/my-project" → project: "my-project"
dir: "/path/to/awesome-app" → project: "awesome-app"
fallback: → project: "unknown-project"
```

## 🛠️ **Implementation Plan**

### **Phase 1: Project Name Detection Service**

#### **Git Repository Detection**
```go
// internal/detection/project_detector.go
package detection

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

type ProjectDetector struct {
    workspace string
    cache     map[string]string
}

func NewProjectDetector() *ProjectDetector {
    workspace, _ := os.Getwd()
    return &ProjectDetector{
        workspace: workspace,
        cache:     make(map[string]string),
    }
}

// DetectProjectName detects project name with intelligent fallback
func (pd *ProjectDetector) DetectProjectName() (string, error) {
    // Check cache first
    if cached, exists := pd.cache[pd.workspace]; exists {
        return cached, nil
    }
    
    // Try git remote detection first
    if projectName, err := pd.detectFromGitRemote(); err == nil {
        pd.cache[pd.workspace] = projectName
        return projectName, nil
    }
    
    // Fallback to directory name
    if projectName, err := pd.detectFromDirectoryName(); err == nil {
        pd.cache[pd.workspace] = projectName
        return projectName, nil
    }
    
    // Last resort fallback
    projectName := "unknown-project"
    pd.cache[pd.workspace] = projectName
    return projectName, nil
}

// detectFromGitRemote extracts project name from git remote
func (pd *ProjectDetector) detectFromGitRemote() (string, error) {
    cmd := exec.Command("git", "config", "--get", "remote.origin.url")
    cmd.Dir = pd.workspace
    
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("not in git repository: %w", err)
    }
    
    gitRemote := strings.TrimSpace(string(output))
    return pd.extractProjectNameFromRemote(gitRemote)
}

// extractProjectNameFromRemote parses various git remote formats
func (pd *ProjectDetector) extractProjectNameFromRemote(remote string) (string, error) {
    // Handle different git remote formats:
    // https://github.com/user/project.git
    // git@github.com:user/project.git
    // git://github.com/user/project.git
    // /path/to/project.git (local)
    
    // Remove .git suffix
    if strings.HasSuffix(remote, ".git") {
        remote = remote[:len(remote)-4]
    }
    
    // Extract project name from path
    parts := strings.Split(remote, "/")
    if len(parts) < 2 {
        return "", fmt.Errorf("invalid git remote format: %s", remote)
    }
    
    projectName := parts[len(parts)-1]
    
    // Remove potential user@ prefix
    if strings.Contains(parts[0], "@") {
        // Format: user@host:path
        return projectName, nil
    }
    
    // Validate project name
    if !isValidProjectName(projectName) {
        return "", fmt.Errorf("invalid project name: %s", projectName)
    }
    
    return projectName, nil
}

// detectFromDirectoryName extracts project name from current directory
func (pd *ProjectDetector) detectFromDirectoryName() (string, error) {
    dirName := filepath.Base(pd.workspace)
    
    if !isValidProjectName(dirName) {
        return "", fmt.Errorf("invalid directory name: %s", dirName)
    }
    
    return dirName, nil
}

// isValidProjectName validates project name format
func isValidProjectName(name string) bool {
    if len(name) == 0 || len(name) > 100 {
        return false
    }
    
    // Basic validation - can be enhanced
    for _, char := range name {
        if !isValidProjectChar(char) {
            return false
        }
    }
    
    return true
}

func isValidProjectChar(char rune) bool {
    return (char >= 'a' && char <= 'z') ||
           (char >= 'A' && char <= 'Z') ||
           (char >= '0' && char <= '9') ||
           char == '-' || char == '_' || char == '.'
}
```

### **Phase 2: Integration with Complaint Service**

#### **Enhanced Complaint Creation**
```go
// internal/service/complaint_service.go (enhanced)

func (s *ComplaintService) CreateComplaint(
    ctx context.Context,
    agentName string,
    sessionName string,
    taskDescription, contextInfo, missingInfo, confusedBy, futureWishes string,
    severity domain.Severity,
    projectName string, // Optional - can be empty
) (*domain.Complaint, error) {
    // Detect project name if not provided
    finalProjectName := projectName
    if strings.TrimSpace(finalProjectName) == "" {
        detected, err := s.projectDetector.DetectProjectName()
        if err != nil {
            s.logger.Warn("Failed to detect project name, using default", "error", err)
            finalProjectName = "unknown-project"
        } else {
            finalProjectName = detected
            s.logger.Info("Auto-detected project name", "project", finalProjectName)
        }
    }
    
    // Validate and create typed project ID
    projectID, err := domain.NewProjectID(finalProjectName)
    if err != nil {
        return nil, fmt.Errorf("invalid project name: %w", err)
    }
    
    // Rest of existing complaint creation logic...
    complaint := &domain.Complaint{
        ID:             complaintID,
        AgentID:        agentID,
        SessionID:      sessionID,
        ProjectID:      projectID,
        TaskDescription: taskDescription,
        // ... other fields
    }
    
    return s.repo.Save(ctx, complaint)
}
```

### **Phase 3: MCP Handler Enhancement**

#### **Tool Input Schema Update**
```go
// internal/delivery/mcp/mcp_server.go (updated)

fileComplaintTool := &mcp.Tool{
    Name: "file_complaint",
    Description: "File a structured complaint about missing or confusing information. Project name is automatically detected if not provided.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "agent_name": map[string]any{
                "type": "string",
                "description": "Name of the AI agent filing the complaint",
                "minLength": 1,
                "maxLength": 100,
            },
            "project_name": map[string]any{
                "type": "string",
                "description": "Name of the project (optional - auto-detected from git repository if not provided)",
                "maxLength": 100,
                "examples": []string{
                    "my-project",  // Manual specification
                    "",           // Use auto-detection
                },
            },
            "session_name": map[string]any{
                "type": "string",
                "description": "Name of the current development session",
                "maxLength": 100,
            },
            "task_description": map[string]any{
                "type": "string",
                "description": "Description of the task being performed",
                "minLength": 1,
                "maxLength": 1000,
            },
            "context_info": map[string]any{
                "type": "string",
                "description": "Additional context about the work environment",
                "maxLength": 500,
            },
            "missing_info": map[string]any{
                "type": "string",
                "description": "What information was missing or unclear",
                "maxLength": 500,
            },
            "confused_by": map[string]any{
                "type": "string",
                "description": "What aspects were confusing",
                "maxLength": 500,
            },
            "future_wishes": map[string]any{
                "type": "string",
                "description": "Suggestions for future improvements",
                "maxLength": 500,
            },
            "severity": map[string]any{
                "type": "string",
                "description": "Severity level of the issue",
                "enum": []string{"low", "medium", "high", "critical"},
                "examples": map[string]string{
                    "low": "Minor inconvenience, workarounds available",
                    "medium": "Significant productivity impact",
                    "high": "Blocker, no clear path forward",
                    "critical": "System failure, project stall",
                },
            },
        },
        "required": []string{"agent_name", "task_description", "severity"},
    },
}
```

#### **Handler Implementation**
```go
func (m *MCPServer) handleFileComplaint(ctx context.Context, req *mcp.CallToolRequest, input FileComplaintInput) (*mcp.CallToolResult, FileComplaintOutput, error) {
    // Convert string inputs to domain types
    agentID, err := domain.NewAgentID(input.AgentName)
    if err != nil {
        return nil, FileComplaintOutput{}, fmt.Errorf("invalid agent name: %w", err)
    }
    
    sessionID, err := domain.NewSessionID(input.SessionName)
    if err != nil {
        return nil, FileComplaintOutput{}, fmt.Errorf("invalid session name: %w", err)
    }
    
    // Create complaint - project name detection handled internally
    complaint, err := m.service.CreateComplaint(
        ctx,
        agentID.String(),
        sessionID.String(),
        input.TaskDescription,
        input.ContextInfo,
        input.MissingInfo,
        input.ConfusedBy,
        input.FutureWishes,
        domain.Severity(input.Severity),
        input.ProjectName, // Optional - auto-detected if empty
    )
    if err != nil {
        return nil, FileComplaintOutput{}, fmt.Errorf("failed to create complaint: %w", err)
    }
    
    // Get file paths with enhanced context
    filePath, docsPath, err := m.service.GetFilePaths(ctx, complaint.ID)
    if err != nil {
        m.logger.Warn("Failed to get file paths", "error", err, "complaint_id", complaint.ID.String())
    }
    
    output := FileComplaintOutput{
        Success:   true,
        Message:   "Complaint filed successfully",
        Complaint: delivery.ToDTOWithPaths(complaint, filePath, docsPath),
    }
    
    result := &mcp.CallToolResult{
        Content: []mcp.Content{
            {Type: "text", Text: output.Message},
        },
    }
    
    return result, output, nil
}
```

### **Phase 4: Configuration and Customization**

#### **Configuration Options**
```go
// internal/config/config.go (enhanced)

type Config struct {
    // ... existing fields ...
    
    ProjectDetection ProjectDetectionConfig `toml:"project_detection"`
}

type ProjectDetectionConfig struct {
    Enabled     bool     `toml:"enabled" default:"true"`
    CacheTTL   int      `toml:"cache_ttl" default:"300"`        // 5 minutes
    Fallbacks   []string `toml:"fallbacks" default:[]`       // Custom fallbacks
    GitRemote   bool     `toml:"git_remote" default:"true"`     // Enable git remote detection
    Directory   bool     `toml:"directory" default:"true"`     // Enable directory name detection
    DefaultName string   `toml:"default_name" default:"unknown-project"`
}
```

#### **Environment Variables**
```bash
# Project detection configuration
export COMPLAINTS_MCP_PROJECT_DETECTION_ENABLED=true
export COMPLAINTS_MCP_PROJECT_DETECTION_CACHE_TTL=300
export COMPLAINTS_MCP_PROJECT_DETECTION_GIT_REMOTE=true
export COMPLAINTS_MCP_PROJECT_DETECTION_DIRECTORY=true
export COMPLAINTS_MCP_PROJECT_DETECTION_DEFAULT_NAME="unknown-project"
```

### **Phase 5: Testing Implementation**

#### **Unit Tests**
```go
// internal/detection/project_detector_test.go
package detection

import (
    "os"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestProjectDetector_DetectProjectName(t *testing.T) {
    tests := []struct {
        name         string
        setupFunc    func() (string, error)
        cleanupFunc  func() error
        expectedResult string
        expectError   bool
    }{
        {
            name: "github https remote",
            setupFunc: func() (string, error) {
                dir := t.TempDir()
                err := os.Chdir(dir)
                require.NoError(t, err)
                
                // Initialize git repo with remote
                runGitCommand(dir, "init")
                runGitCommand(dir, "remote", "add", "origin", "https://github.com/user/my-project.git")
                
                return dir, nil
            },
            cleanupFunc: func() error {
                return os.Chdir(t.TempDir()) // Move away from temp dir
            },
            expectedResult: "my-project",
            expectError: false,
        },
        {
            name: "github ssh remote",
            setupFunc: func() (string, error) {
                dir := t.TempDir()
                err := os.Chdir(dir)
                require.NoError(t, err)
                
                runGitCommand(dir, "init")
                runGitCommand(dir, "remote", "add", "origin", "git@github.com:user/awesome-app.git")
                
                return dir, nil
            },
            cleanupFunc: func() error {
                return os.Chdir(t.TempDir())
            },
            expectedResult: "awesome-app",
            expectError: false,
        },
        {
            name: "directory name fallback",
            setupFunc: func() (string, error) {
                dir := t.TempDir()
                subdir := filepath.Join(dir, "my-cool-project")
                err := os.Mkdir(subdir, 0755)
                require.NoError(t, err)
                
                return subdir, nil
            },
            cleanupFunc: func() error {
                return os.Chdir(t.TempDir())
            },
            expectedResult: "my-cool-project",
            expectError: false,
        },
        {
            name: "non-git directory with invalid name",
            setupFunc: func() (string, error) {
                dir := t.TempDir()
                subdir := filepath.Join(dir, "invalid@name")
                err := os.Mkdir(subdir, 0755)
                require.NoError(t, err)
                
                return subdir, nil
            },
            cleanupFunc: func() error {
                return os.Chdir(t.TempDir())
            },
            expectedResult: "unknown-project",
            expectError: false, // Should fall back gracefully
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            originalDir, _ := os.Getwd()
            defer os.Chdir(originalDir) // Ensure we return to original
            
            workspace, err := tt.setupFunc()
            require.NoError(t, err)
            
            // Cleanup
            if tt.cleanupFunc != nil {
                defer tt.cleanupFunc()
            }
            
            // Test detection
            detector := NewProjectDetector()
            result, err := detector.DetectProjectName()
            
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedResult, result)
            }
        })
    }
}

func runGitCommand(dir, name string, args ...string) {
    cmd := exec.Command("git", append([]string{name}, args...)...)
    cmd.Dir = dir
    cmd.Run() // Ignore errors for test setup
}
```

#### **Integration Tests**
```go
// internal/service/complaint_service_test.go (enhanced)

func TestComplaintService_CreateComplaintWithAutoProjectDetection(t *testing.T) {
    // Setup mock detector
    mockDetector := &MockProjectDetector{
        detectResult: "my-test-project",
        detectError:   nil,
    }
    
    service := NewComplaintService(repo, tracer, logger, mockDetector)
    
    // Test complaint creation without project name
    complaint, err := service.CreateComplaint(
        context.Background(),
        "AI-Assistant",
        "test-session",
        "Test task description",
        "", "", "", "", "",
        "medium",
        "", // Empty project name - should auto-detect
    )
    
    require.NoError(t, err)
    require.NotNil(t, complaint)
    
    // Verify project was auto-detected
    assert.Equal(t, "my-test-project", complaint.ProjectID.String())
    
    // Verify detector was called
    assert.True(t, mockDetector.DetectCalled)
}
```

## 🎯 **Benefits of This Enhancement**

### **1. Improved User Experience**
- **Automatic Detection**: No manual project name entry needed
- **Intelligent Fallbacks**: Multiple detection strategies
- **Consistent Names**: Standardized project naming
- **Error Reduction**: Less manual entry errors

### **2. Better Data Quality**
- **Accurate Context**: Project names from actual git repos
- **Standardized Format**: Consistent naming conventions
- **Rich Metadata**: Enhanced search and filtering
- **Reduced Duplication**: Automatic detection prevents variations

### **3. Enhanced AI Context**
- **Project Awareness**: AI knows current project context
- **Better Suggestions**: Context-aware recommendations
- **Smarter Organization**: Automatic project categorization
- **Improved Debugging**: Clear project association

### **4. Configuration Flexibility**
- **Customizable Detection**: Enable/disable detection methods
- **Fallback Options**: Custom project name sources
- **Performance Tuning**: Cache TTL and optimization
- **Integration Ready**: Works with existing workflows

## 📋 **Files to Create/Modify**

### **New Files**
- `internal/detection/project_detector.go` - Project name detection service
- `internal/detection/project_detector_test.go` - Detection service tests
- `internal/detection/mock_project_detector.go` - Test double for detection

### **Modified Files**
- `internal/service/complaint_service.go` - Integration with detection
- `internal/service/complaint_service_test.go` - Enhanced tests
- `internal/config/config.go` - Configuration options
- `internal/delivery/mcp/mcp_server.go` - Tool schema and handlers
- `features/bdd/project_detection_bdd_test.go` - BDD scenarios

### **Configuration Files**
- `examples/project-detection-config.toml` - Example configuration
- `examples/.env.project-detection` - Environment variables example

## 🔄 **Migration Strategy**

### **Phase 1: Backward Compatibility**
- **Optional Field**: Project name remains optional in tool input
- **Graceful Fallback**: Use existing logic if detection fails
- **Manual Override**: Users can still specify project name manually
- **Feature Flag**: Enable/disable detection via configuration

### **Phase 2: Gradual Rollout**
- **Default Enable**: Detection enabled by default for new installations
- **User Control**: Configuration option to disable if problematic
- **Monitoring**: Log detection success/failure rates
- **Feedback**: Collect user experience data

### **Phase 3: Enhanced Features**
- **Advanced Detection**: Support for additional VCS (Mercurial, SVN)
- **Project Mapping**: Custom project name mapping rules
- **Workspace Detection**: Multi-project workspace support
- **Integration**: IDE and editor integration

## 🧪 **Testing Strategy**

### **Detection Scenarios**
- **Git Repositories**: HTTPS, SSH, Git protocols
- **Directory Names**: Valid, invalid, edge cases
- **Non-Git Projects**: Fallback to directory name
- **Invalid Directories**: Graceful fallback behavior
- **Workspace Projects**: Multi-project scenarios

### **Edge Cases**
- **No Git Repository**: Directory name fallback
- **Invalid Directory Name**: Default project name
- **Permission Errors**: Graceful handling
- **Network Issues**: Local fallback options
- **Multiple Remotes**: Intelligent selection

### **Performance Tests**
- **Detection Speed**: Fast project name resolution
- **Cache Performance**: Caching effectiveness measurement
- **Memory Usage**: Low memory footprint
- **Concurrent Safety**: Thread-safe detection

## 🏆 **Success Criteria**

- [ ] Project name detection works for git repositories
- [ ] Directory name fallback works for non-git projects
- [ ] Graceful fallback to default name when all else fails
- [ ] Configuration options allow customization
- [ ] Performance meets requirements (<100ms detection)
- [ ] All edge cases handled gracefully
- [ ] Backward compatibility maintained
- [ ] Comprehensive test coverage achieved
- [ ] Documentation updated with new features

## 🏷️ **Labels**
- `enhancement` - New feature addition
- `automation` - Automatic detection capabilities
- `git-integration` - Git repository integration
- `user-experience` - Improved user workflow
- `configuration` - Configurable behavior
- `medium-priority` - Important UX improvement

## 📊 **Priority**: Medium
- **Complexity**: Medium (git integration + fallback logic)
- **Value**: High (significant UX improvement + data quality)
- **Risk**: Low (graceful fallbacks ensure reliability)
- **Dependencies**: None (standalone enhancement)

## 🤝 **Dependencies**
- **Git Binary**: Requires git command availability
- **File System**: Read access to current directory
- **Configuration**: Optional configuration system
- **Logging**: Existing logging infrastructure

---

**This enhancement provides intelligent project name detection, significantly improving user experience and data quality while maintaining full backward compatibility and offering extensive customization options.**