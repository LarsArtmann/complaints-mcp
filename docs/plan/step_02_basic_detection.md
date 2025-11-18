# Step 2: Implement Basic Project Detection

## 🎯 Objective
Create a working project detection service using system git commands.

## 🏗️ Implementation Tasks

### A. Create Detection Service
- **File**: `internal/detection/project_detector.go`
- **Interface**: Simple DetectProjectName() method
- **Logic**: Git remote → Directory name → Default fallback
- **Caching**: Simple map-based caching

### B. Basic Git Integration
- **Command**: `git config --get remote.origin.url`
- **Parsing**: Extract project name from remote URL
- **Fallback**: Current directory name extraction
- **Error Handling**: Graceful fallback to default

### C. Unit Tests
- **File**: `internal/detection/project_detector_test.go`
- **Coverage**: Git remote, directory, fallback scenarios
- **Setup**: Temp git repos for testing
- **Validation**: Verify correct project names extracted

## 📝 Implementation Details

### ProjectDetector Interface
```go
type ProjectDetector interface {
    DetectProjectName() (string, error)
}
```

### Basic Implementation
```go
type projectDetector struct {
    workspace string
    cache     map[string]string
}

func (pd *projectDetector) DetectProjectName() (string, error) {
    // Try git remote
    if name, err := pd.fromGitRemote(); err == nil {
        return name, nil
    }
    
    // Fallback to directory
    if name, err := pd.fromDirectoryName(); err == nil {
        return name, nil
    }
    
    // Default
    return "unknown-project", nil
}
```

### Git Remote Formats Supported
- `https://github.com/user/project.git`
- `git@github.com:user/project.git`
- `git://github.com/user/project.git`
- `/path/to/project.git` (local)

## ⏱️ Time Estimate: 4-6 hours
## 🎯 Impact: High (enables project detection)
## 💪 Work Required: Medium (service + tests)