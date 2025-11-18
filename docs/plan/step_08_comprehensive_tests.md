# Step 8: Create Comprehensive Test Suite

## 🎯 Objective
Ensure complete test coverage for phantom types, project detection, and schema validation.

## 🧪 Testing Strategy Overview

### Test Categories
1. **Unit Tests**: Individual component testing
2. **Integration Tests**: Component interaction testing
3. **BDD Tests**: User behavior scenario testing
4. **Performance Tests**: Benchmarking and optimization
5. **Schema Tests**: MCP tool contract validation
6. **End-to-End Tests**: Complete workflow testing

## 🏗️ Implementation Tasks

### A. Phantom Type Unit Tests

#### ComplaintID Tests
```go
// internal/domain/complaint_id_test.go
package domain

import (
    "testing"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestComplaintID_New(t *testing.T) {
    t.Run("successful generation", func(t *testing.T) {
        id, err := NewComplaintID()
        
        require.NoError(t, err)
        assert.False(t, id.IsEmpty())
        assert.True(t, id.IsValid())
        
        // Verify UUID format
        parsed, err := uuid.Parse(id.String())
        require.NoError(t, err)
        assert.Equal(t, id.String(), parsed.String())
    })
    
    t.Run("generates unique IDs", func(t *testing.T) {
        ids := make(map[ComplaintID]bool)
        
        for i := 0; i < 1000; i++ {
            id, err := NewComplaintID()
            require.NoError(t, err)
            
            assert.False(t, ids[id], "Generated duplicate ID: %s", id.String())
            ids[id] = true
        }
    })
    
    t.Run("generates version 4 UUIDs", func(t *testing.T) {
        id, err := NewComplaintID()
        require.NoError(t, err)
        
        parsed, err := id.UUID()
        require.NoError(t, err)
        assert.Equal(t, uuid.Version(4), parsed.Version())
    })
}

func TestComplaintID_Parse(t *testing.T) {
    tests := []struct {
        name         string
        input        string
        wantErr      bool
        errorMsg     string
        expectedID   ComplaintID
    }{
        {
            name:       "valid UUID v4",
            input:      "550e8400-e29b-41d4-a716-446655440000",
            wantErr:    false,
            expectedID: ComplaintID("550e8400-e29b-41d4-a716-446655440000"),
        },
        {
            name:       "valid lowercase UUID v4",
            input:      "9cb3bb9e-b6dc-4e02-9767-e396a42b63a6",
            wantErr:    false,
            expectedID: ComplaintID("9cb3bb9e-b6dc-4e02-9767-e396a42b63a6"),
        },
        {
            name:       "empty string",
            input:      "",
            wantErr:    true,
            errorMsg:   "cannot be empty",
            expectedID: ComplaintID(""),
        },
        {
            name:       "invalid format - too short",
            input:      "550e8400-e29b",
            wantErr:    true,
            errorMsg:   "must be valid UUID v4 format",
            expectedID: ComplaintID(""),
        },
        {
            name:       "invalid format - wrong version",
            input:      "550e8400-e29b-11d4-a716-446655440000", // Version 1
            wantErr:    true,
            errorMsg:   "must be valid UUID v4 format",
            expectedID: ComplaintID(""),
        },
        {
            name:       "invalid format - non-hex characters",
            input:      "550e8400-e29b-41d4-a716-44665544zzzz",
            wantErr:    true,
            errorMsg:   "must be valid UUID v4 format",
            expectedID: ComplaintID(""),
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseComplaintID(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
                assert.True(t, got.IsEmpty())
                assert.False(t, got.IsValid())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedID, got)
                assert.False(t, got.IsEmpty())
                assert.True(t, got.IsValid())
                assert.Equal(t, tt.input, got.String())
            }
        })
    }
}

func TestComplaintID_JSONSerialization(t *testing.T) {
    t.Run("marshal produces flat JSON", func(t *testing.T) {
        id := ComplaintID("550e8400-e29b-41d4-a716-446655440000")
        
        data, err := json.Marshal(id)
        require.NoError(t, err)
        
        // Verify flat structure
        var result map[string]any
        err = json.Unmarshal(data, &result)
        require.NoError(t, err)
        
        // Should be flat string, not nested object
        idValue, exists := result["id"]
        require.True(t, exists, "id field should exist")
        
        idStr, isString := idValue.(string)
        assert.True(t, isString, "id should be string, got %T", idValue)
        assert.Equal(t, id.String(), idStr)
    })
    
    t.Run("unmarshal from flat JSON", func(t *testing.T) {
        jsonData := `{"id":"550e8400-e29b-41d4-a716-446655440000"}`
        
        var id ComplaintID
        err := json.Unmarshal([]byte(jsonData), &id)
        require.NoError(t, err)
        
        assert.Equal(t, ComplaintID("550e8400-e29b-41d4-a716-446655440000"), id)
        assert.True(t, id.IsValid())
    })
    
    t.Run("reject nested JSON", func(t *testing.T) {
        jsonData := `{"id":{"Value":"550e8400-e29b-41d4-a716-446655440000"}}`
        
        var id ComplaintID
        err := json.Unmarshal([]byte(jsonData), &id)
        assert.Error(t, err)
    })
}

func TestComplaintID_Methods(t *testing.T) {
    validID := ComplaintID("550e8400-e29b-41d4-a716-446655440000")
    invalidID := ComplaintID("")
    emptyID := ComplaintID("")
    
    t.Run("String method", func(t *testing.T) {
        assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", validID.String())
        assert.Equal(t, "", emptyID.String())
    })
    
    t.Run("IsEmpty method", func(t *testing.T) {
        assert.False(t, validID.IsEmpty())
        assert.True(t, emptyID.IsEmpty())
    })
    
    t.Run("IsValid method", func(t *testing.T) {
        assert.True(t, validID.IsValid())
        assert.False(t, invalidID.IsValid())
        assert.False(t, emptyID.IsValid())
    })
    
    t.Run("Validate method", func(t *testing.T) {
        assert.NoError(t, validID.Validate())
        assert.Error(t, invalidID.Validate())
    })
    
    t.Run("UUID method", func(t *testing.T) {
        parsed, err := validID.UUID()
        assert.NoError(t, err)
        assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", parsed.String())
        
        _, err = emptyID.UUID()
        assert.Error(t, err)
    })
}
```

#### AgentID, SessionID, ProjectID Tests
```go
// Similar comprehensive test files for each phantom type
// internal/domain/agent_id_test.go
// internal/domain/session_id_test.go
// internal/domain/project_id_test.go

func TestAgentID_New(t *testing.T) {
    tests := []struct {
        name         string
        input        string
        wantErr      bool
        errorMsg     string
        expectedID   AgentID
    }{
        {
            name:       "valid simple name",
            input:      "AI-Assistant",
            wantErr:    false,
            expectedID: AgentID("AI-Assistant"),
        },
        {
            name:       "valid with underscores",
            input:      "Code_Reviewer_Bot",
            wantErr:    false,
            expectedID: AgentID("Code_Reviewer_Bot"),
        },
        {
            name:       "trims whitespace",
            input:      "  AI-Assistant  ",
            wantErr:    false,
            expectedID: AgentID("AI-Assistant"),
        },
        {
            name:       "empty string",
            input:      "",
            wantErr:    true,
            errorMsg:   "cannot be empty",
            expectedID: AgentID(""),
        },
        {
            name:       "too long",
            input:      strings.Repeat("a", 101),
            wantErr:    true,
            errorMsg:   "cannot exceed 100 characters",
            expectedID: AgentID(""),
        },
        {
            name:       "invalid characters - @",
            input:      "AI@Assistant",
            wantErr:    true,
            errorMsg:   "contains invalid characters",
            expectedID: AgentID(""),
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewAgentID(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
                assert.True(t, got.IsEmpty())
                assert.False(t, got.IsValid())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedID, got)
                assert.False(t, got.IsEmpty())
                assert.True(t, got.IsValid())
                assert.Equal(t, strings.TrimSpace(tt.input), got.String())
            }
        })
    }
}
```

### B. Project Detection Tests

#### SystemGitDetector Tests
```go
// internal/detection/system_git_detector_test.go
package detection

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSystemGitDetector_DetectProjectName(t *testing.T) {
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
                
                runGitCommand(dir, "init")
                runGitCommand(dir, "remote", "add", "origin", "https://github.com/user/my-project.git")
                
                return dir, nil
            },
            cleanupFunc: func() error {
                return os.Chdir(t.TempDir())
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
        {
            name: "git remote with various formats",
            setupFunc: func() (string, error) {
                dir := t.TempDir()
                err := os.Chdir(dir)
                require.NoError(t, err)
                
                runGitCommand(dir, "init")
                runGitCommand(dir, "remote", "add", "origin", "git://github.com/user/git-protocol-project.git")
                
                return dir, nil
            },
            cleanupFunc: func() error {
                return os.Chdir(t.TempDir())
            },
            expectedResult: "git-protocol-project",
            expectError: false,
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
            detector := NewSystemGitDetector(workspace, make(map[string]string))
            result, err := detector.DetectProjectName(context.Background())
            
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedResult, result)
            }
        })
    }
}

func TestSystemGitDetector_Caching(t *testing.T) {
    dir := t.TempDir()
    err := os.Chdir(dir)
    require.NoError(t, err)
    
    runGitCommand(dir, "init")
    runGitCommand(dir, "remote", "add", "origin", "https://github.com/user/cache-test.git")
    
    detector := NewSystemGitDetector(dir, make(map[string]string))
    
    // First call should detect
    result1, err1 := detector.DetectProjectName(context.Background())
    require.NoError(t, err1)
    assert.Equal(t, "cache-test", result1)
    
    // Second call should use cache (even if we modify git repo)
    runGitCommand(dir, "remote", "set-url", "origin", "https://github.com/user/different-project.git")
    result2, err2 := detector.DetectProjectName(context.Background())
    require.NoError(t, err2)
    assert.Equal(t, "cache-test", result2) // Should return cached value
    
    // Test cache invalidation (if implemented)
    // This depends on whether you implement cache TTL or manual invalidation
}

func runGitCommand(dir, name string, args ...string) {
    cmd := exec.Command("git", append([]string{name}, args...)...)
    cmd.Dir = dir
    cmd.Run() // Ignore errors for test setup
}
```

### C. Integration Tests

#### Service Layer Tests
```go
// internal/service/complaint_service_integration_test.go
package service

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "github.com/larsartmann/complaints-mcp/internal/domain"
    "github.com/larsartmann/complaints-mcp/internal/detection"
    "github.com/larsartmann/complaints-mcp/internal/repo"
    "github.com/larsartmann/complaints-mcp/internal/tracing"
)

func TestComplaintService_WithPhantomTypes(t *testing.T) {
    // Setup
    tempDir := t.TempDir()
    tracer := tracing.NewNoOpTracer()
    logger := &mockLogger{}
    repository := repo.NewFileRepository(tempDir, tracer)
    mockDetector := &mockProjectDetector{detectResult: "test-project"}
    service := NewComplaintService(repository, tracer, logger, mockDetector)
    
    t.Run("create complaint with typed IDs", func(t *testing.T) {
        complaint, err := service.CreateComplaint(
            context.Background(),
            "AI-Assistant",
            "integration-test",
            "Integration test task",
            "Test context info",
            "Test missing info",
            "Test confused by",
            "Test future wishes",
            "medium",
            "test-project",
        )
        
        require.NoError(t, err)
        require.NotNil(t, complaint)
        
        // Verify phantom types are properly set
        assert.True(t, complaint.ID.IsValid())
        assert.True(t, complaint.AgentID.IsValid())
        assert.True(t, complaint.SessionID.IsValid())
        assert.True(t, complaint.ProjectID.IsValid())
        
        // Verify string conversions work
        assert.Equal(t, "AI-Assistant", complaint.AgentID.String())
        assert.Equal(t, "integration-test", complaint.SessionID.String())
        assert.Equal(t, "test-project", complaint.ProjectID.String())
    })
    
    t.Run("reject invalid typed IDs", func(t *testing.T) {
        _, err := service.CreateComplaint(
            context.Background(),
            "AI@Assistant", // Invalid character
            "test-session",
            "Test task",
            "", "", "", "", "",
            "low",
            "test-project",
        )
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid agent name")
    })
    
    t.Run("get complaint returns typed IDs", func(t *testing.T) {
        // Create complaint
        complaint, err := service.CreateComplaint(
            context.Background(),
            "Test-Agent",
            "get-test-session",
            "Get test task",
            "", "", "", "", "",
            "low",
            "test-project",
        )
        require.NoError(t, err)
        
        // Get complaint
        retrieved, err := service.GetComplaint(context.Background(), complaint.ID)
        require.NoError(t, err)
        require.NotNil(t, retrieved)
        
        // Verify phantom types are preserved
        assert.Equal(t, complaint.ID, retrieved.ID)
        assert.Equal(t, complaint.AgentID, retrieved.AgentID)
        assert.Equal(t, complaint.SessionID, retrieved.SessionID)
        assert.Equal(t, complaint.ProjectID, retrieved.ProjectID)
    })
    
    t.Run("list complaints preserves typed IDs", func(t *testing.T) {
        // Create multiple complaints
        agentIDs := []string{"Agent-1", "Agent-2", "Agent-3"}
        
        for _, agentID := range agentIDs {
            _, err := service.CreateComplaint(
                context.Background(),
                agentID,
                "list-test-session",
                "List test task",
                "", "", "", "", "",
                "low",
                "test-project",
            )
            require.NoError(t, err)
        }
        
        // List complaints
        complaints, err := service.ListComplaints(context.Background(), 100, 0)
        require.NoError(t, err)
        assert.Len(t, complaints, 3)
        
        // Verify all phantom types are valid
        for _, complaint := range complaints {
            assert.True(t, complaint.ID.IsValid())
            assert.True(t, complaint.AgentID.IsValid())
            assert.True(t, complaint.SessionID.IsValid())
            assert.True(t, complaint.ProjectID.IsValid())
        }
    })
    
    t.Run("resolve complaint with typed ID", func(t *testing.T) {
        // Create complaint
        complaint, err := service.CreateComplaint(
            context.Background(),
            "Test-Agent",
            "resolve-test-session",
            "Resolve test task",
            "", "", "", "", "",
            "low",
            "test-project",
        )
        require.NoError(t, err)
        
        // Resolve complaint with typed ID
        resolverID, err := domain.NewAgentID("Test-Resolver")
        require.NoError(t, err)
        
        err = service.ResolveComplaint(
            context.Background(),
            complaint.ID,
            resolverID.String(),
        )
        
        require.NoError(t, err)
        
        // Verify resolution
        resolved, err := service.GetComplaint(context.Background(), complaint.ID)
        require.NoError(t, err)
        
        assert.True(t, resolved.ResolutionState.IsResolved())
        assert.Equal(t, "Test-Resolver", resolved.ResolvedBy)
        assert.NotNil(t, resolved.ResolvedAt)
    })
}

// Mock implementations for testing
type mockProjectDetector struct {
    detectResult string
    detectError   error
}

func (m *mockProjectDetector) DetectProjectName(ctx context.Context) (string, error) {
    return m.detectResult, m.detectError
}

func (m *mockProjectDetector) DetectProjectNameWithWorkspace(ctx context.Context, workspace string) (string, error) {
    return m.detectResult, m.detectError
}

type mockLogger struct {
    warnings []string
    infos    []string
}

func (m *mockLogger) Warn(msg string, args ...interface{}) {
    m.warnings = append(m.warnings, msg)
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
    m.infos = append(m.infos, msg)
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
    // Mock error logging
}
```

### D. BDD Tests

#### Complete Workflow BDD Tests
```gherkin
# features/bdd/phantom_type_workflow_bdd.feature
Feature: Phantom Type Workflow Integration
  As an AI assistant
  I want to use phantom types throughout the complaint system
  So that I get type safety and flat JSON output

  Background:
    Given the MCP server is running
    And phantom types are implemented for all ID fields
    And project detection is configured

  Scenario: Create complaint with type-safe IDs
    Given I have valid typed agent, session, and project IDs
    When I file a complaint with these type-safe IDs
    Then the complaint should be created successfully
    And all ID fields should be valid phantom types
    And the JSON response should use flat ID format

  Scenario: Reject invalid typed IDs
    Given I have invalid agent ID with special characters
    When I attempt to file a complaint with this agent ID
    Then the complaint creation should be rejected
    And the error message should be clear and actionable

  Scenario: Preserve phantom types in storage and retrieval
    Given I create a complaint with type-safe IDs
    When I retrieve the complaint from storage
    Then all phantom types should be preserved
    And type safety should be maintained throughout the lifecycle

  Scenario: Type safety prevents ID mixing
    Given I have agent ID "AI-Assistant" and session ID "dev-session"
    When I attempt to use agent ID as session ID
    Then the type system should prevent this at compile time
    And no runtime ID confusion should occur

  Scenario: JSON serialization produces flat structure
    Given I have a complaint with phantom type IDs
    When I serialize the complaint to JSON
    Then the JSON should use flat string IDs
    And no nested "Value" objects should be present
    And the structure should match the tool schema

  Scenario: Schema validation accepts flat IDs
    Given I call file_complaint with flat agent ID "AI-Assistant"
    And flat project ID "my-project"
    When the tool processes the input
    Then the schema validation should succeed
    And the complaint should be created

  Scenario: Schema validation rejects nested IDs
    Given I attempt to call file_complaint with nested agent ID structure
    When the tool processes the input
    Then the schema validation should reject the input
    And a clear validation error should be returned

  Scenario: Project detection auto-populates project name
    Given I am in a git repository with remote "origin"
    And I file a complaint without specifying project name
    When the tool processes the input
    Then the project name should be auto-detected from git remote
    And the complaint should be associated with the correct project

  Scenario: End-to-end workflow with phantom types
    Given I file a complaint with all typed IDs
    When I list all complaints
    Then the list should show flat ID formats
    And all complaints should have valid phantom types
    When I resolve a complaint
    Then the resolution should preserve phantom types
    And the final state should have consistent flat IDs

  Scenario: Performance with phantom types
    Given I create 1000 complaints with phantom types
    When I measure the creation time
    Then the performance should be acceptable (<100ms per complaint)
    And memory usage should be reasonable
    And file storage should use flat JSON format
```

### E. Schema Validation Tests

#### MCP Schema Tests
```go
// internal/delivery/mcp/schema_validation_test.go
package mcp

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSchemaValidation_FlatIDFormat(t *testing.T) {
    tests := []struct {
        name     string
        input    map[string]any
        tool     string
        wantErr  bool
        errorMsg string
    }{
        {
            name: "file_complaint with flat agent ID",
            input: map[string]any{
                "agent_name": "AI-Assistant",
                "task_description": "Test task",
                "severity": "low",
                "project_name": "my-project",
            },
            tool:    "file_complaint",
            wantErr: false,
        },
        {
            name: "file_complaint with auto-detect project",
            input: map[string]any{
                "agent_name": "AI-Assistant",
                "task_description": "Test task",
                "severity": "low",
                "project_name": "", // Empty for auto-detection
            },
            tool:    "file_complaint",
            wantErr: false,
        },
        {
            name: "resolve_complaint with flat complaint ID",
            input: map[string]any{
                "complaint_id": "550e8400-e29b-41d4-a716-446655440000",
                "resolved_by": "AI-Resolver",
            },
            tool:    "resolve_complaint",
            wantErr: false,
        },
        {
            name: "resolve_complaint with invalid complaint ID format",
            input: map[string]any{
                "complaint_id": "not-a-uuid",
                "resolved_by": "AI-Resolver",
            },
            tool:    "resolve_complaint",
            wantErr: true,
            errorMsg: "must be valid UUID v4 format",
        },
        {
            name: "resolve_complaint with nested complaint ID",
            input: map[string]any{
                "complaint_id": map[string]any{
                    "Value": "550e8400-e29b-41d4-a716-446655440000",
                },
                "resolved_by": "AI-Resolver",
            },
            tool:    "resolve_complaint",
            wantErr: true,
            errorMsg: "complaint_id must be string, got object",
        },
        {
            name: "list_complaints with flat filter IDs",
            input: map[string]any{
                "limit": 10,
                "agent_id": "AI-Assistant",
                "project_id": "my-project",
            },
            tool:    "list_complaints",
            wantErr: false,
        },
        {
            name: "file_complaint with invalid agent name",
            input: map[string]any{
                "agent_name": "AI@Assistant", // Invalid character
                "task_description": "Test task",
                "severity": "low",
            },
            tool:    "file_complaint",
            wantErr: true,
            errorMsg: "contains invalid characters",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup test server
            server := setupTestServer()
            
            // Create tool request
            request := &mcp.CallToolRequest{
                Name: tt.tool,
                Arguments: mcp.ToolArguments(tt.input),
            }
            
            // Process request based on tool
            var err error
            switch tt.tool {
            case "file_complaint":
                _, _, err = server.handleFileComplaint(context.Background(), request)
            case "resolve_complaint":
                _, _, err = server.handleResolveComplaint(context.Background(), request)
            case "list_complaints":
                _, _, err = server.handleListComplaints(context.Background(), request)
            }
            
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errorMsg != "" {
                    assert.Contains(t, err.Error(), tt.errorMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### F. Performance Tests

#### Benchmark Tests
```go
// internal/benchmarks/phantom_type_benchmark_test.go
package benchmarks

import (
    "testing"
    "github.com/google/uuid"
    "github.com/larsartmann/complaints-mcp/internal/domain"
)

// Benchmark phantom type operations
func BenchmarkPhantomType_NewComplaintID(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, _ = domain.NewComplaintID()
    }
}

func BenchmarkPhantomType_ParseComplaintID(b *testing.B) {
    validUUID := "550e8400-e29b-41d4-a716-446655440000"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = domain.ParseComplaintID(validUUID)
    }
}

func BenchmarkPhantomType_String(b *testing.B) {
    id, _ := domain.NewComplaintID()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = id.String()
    }
}

func BenchmarkPhantomType_Validate(b *testing.B) {
    id, _ := domain.NewComplaintID()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = id.Validate()
    }
}

func BenchmarkPhantomType_JSONMarshal(b *testing.B) {
    complaint := &domain.Complaint{
        ID:             mustNewComplaintID(),
        AgentID:        mustNewAgentID("AI-Assistant"),
        SessionID:      mustNewSessionID("test-session"),
        ProjectID:      mustNewProjectID("test-project"),
        TaskDescription: "Benchmark test complaint",
        Severity:       domain.SeverityMedium,
        Timestamp:      time.Now(),
        ResolutionState: domain.ResolutionStateOpen,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = json.Marshal(complaint)
    }
}

// Comparison with struct implementation
func BenchmarkStructID_New(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = &OldComplaintID{
            Value: uuid.New().String(),
        }
    }
}

func BenchmarkStructID_String(b *testing.B) {
    id := &OldComplaintID{
        Value: "550e8400-e29b-41d4-a716-446655440000",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = id.String()
    }
}

func BenchmarkStructID_JSONMarshal(b *testing.B) {
    complaint := &OldComplaint{
        ID: &OldComplaintID{
            Value: "550e8400-e29b-41d4-a716-446655440000",
        },
        AgentName:       "AI-Assistant",
        TaskDescription: "Benchmark test complaint",
        Severity:        "medium",
        Timestamp:       time.Now(),
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = json.Marshal(complaint)
    }
}

// Helper functions
func mustNewComplaintID() domain.ComplaintID {
    id, err := domain.NewComplaintID()
    if err != nil {
        panic(err)
    }
    return id
}

func mustNewAgentID(name string) domain.AgentID {
    id, err := domain.NewAgentID(name)
    if err != nil {
        panic(err)
    }
    return id
}

func mustNewSessionID(name string) domain.SessionID {
    id, err := domain.NewSessionID(name)
    if err != nil {
        panic(err)
    }
    return id
}

func mustNewProjectID(name string) domain.ProjectID {
    id, err := domain.NewProjectID(name)
    if err != nil {
        panic(err)
    }
    return id
}
```

## 🧪 Test Coverage Goals

### Coverage Targets
- **Unit Tests**: 95%+ coverage for phantom type implementations
- **Integration Tests**: 90%+ coverage for service layer with phantom types
- **BDD Tests**: Complete workflow coverage with type safety scenarios
- **Performance Tests**: Benchmark all phantom type operations
- **Schema Tests**: 100% coverage for MCP tool schema validation

### Success Criteria
- [ ] All phantom types have comprehensive unit tests
- [ ] JSON serialization produces correct flat structure
- [ ] Service layer integration tests pass with typed IDs
- [ ] BDD scenarios cover all type safety workflows
- [ ] Performance benchmarks show no regression
- [ ] Schema validation tests cover all edge cases
- [ ] All tests run successfully in CI/CD pipeline
- [ ] Test coverage meets or exceeds targets

## ⏱️ Time Estimate: 10-12 hours
## 🎯 Impact: HIGH (ensures reliability and correctness)
## 💪 Work Required: HIGH (comprehensive test suite)