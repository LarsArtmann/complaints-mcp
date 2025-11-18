# Issue #52: Add Comprehensive Tests for Phantom Type Safety

## 🎯 **Enhancement: Complete Test Coverage for Phantom Type Implementation**

### **Current State Analysis**
While implementing phantom types (#48, #49, #50) provides significant architectural improvements, we need comprehensive test coverage to ensure:

- **Type Safety**: Compile-time protections work correctly
- **Validation Logic**: All validation rules are properly tested
- **JSON Serialization**: Flat JSON output works as expected
- **Error Handling**: Error cases are handled gracefully
- **Performance**: No performance regressions introduced

### **Testing Requirements**
We need test coverage for:
- **Phantom Type Construction**: New() and Parse() functions
- **Validation Logic**: All validation rules and edge cases
- **JSON Serialization**: Marshal/Unmarshal with flat structure
- **Type Safety**: Compile-time error prevention
- **Error Handling**: Invalid input handling and error messages
- **Performance**: Benchmarks vs current implementation
- **Integration**: End-to-end workflows with phantom types

## 🛠️ **Comprehensive Test Implementation Plan**

### **Phase 1: Phantom Type Unit Tests**

#### **ComplaintID Tests**
```go
// internal/domain/complaint_id_test.go
package domain

import (
    "testing"
    "time"
    
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewComplaintID(t *testing.T) {
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

func TestParseComplaintID(t *testing.T) {
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
        {
            name:       "invalid format - wrong separator positions",
            input:      "550e8400e29b-41d4-a716-446655440000",
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

func TestMustParseComplaintID(t *testing.T) {
    t.Run("valid UUID", func(t *testing.T) {
        validUUID := "550e8400-e29b-41d4-a716-446655440000"
        
        assert.NotPanics(t, func() {
            id := MustParseComplaintID(validUUID)
            assert.Equal(t, ComplaintID(validUUID), id)
            assert.True(t, id.IsValid())
        })
    })
    
    t.Run("invalid UUID", func(t *testing.T) {
        invalidUUID := "not-a-uuid"
        
        assert.Panics(t, func() {
            MustParseComplaintID(invalidUUID)
        })
    })
}

func TestComplaintID_Validate(t *testing.T) {
    tests := []struct {
        name        string
        id          ComplaintID
        wantErr     bool
        errorMsg    string
    }{
        {
            name:     "valid ID",
            id:       ComplaintID("550e8400-e29b-41d4-a716-446655440000"),
            wantErr:  false,
        },
        {
            name:     "empty ID",
            id:       ComplaintID(""),
            wantErr:  true,
            errorMsg: "cannot be empty",
        },
        {
            name:     "invalid format",
            id:       ComplaintID("not-a-uuid"),
            wantErr:  true,
            errorMsg: "must be valid UUID v4 format",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.id.Validate()
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestComplaintID_IsEmpty(t *testing.T) {
    tests := []struct {
        name     string
        id       ComplaintID
        expected bool
    }{
        {"empty string", ComplaintID(""), true},
        {"empty string with spaces", ComplaintID("   "), true},
        {"valid UUID", ComplaintID("550e8400-e29b-41d4-a716-446655440000"), false},
        {"valid lowercase", ComplaintID("9cb3bb9e-b6dc-4e02-9767-e396a42b63a6"), false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expected, tt.id.IsEmpty())
        })
    }
}

func TestComplaintID_IsValid(t *testing.T) {
    tests := []struct {
        name     string
        id       ComplaintID
        expected bool
    }{
        {"empty string", ComplaintID(""), false},
        {"invalid format", ComplaintID("not-a-uuid"), false},
        {"wrong version", ComplaintID("550e8400-e29b-11d4-a716-446655440000"), false},
        {"valid UUID v4", ComplaintID("550e8400-e29b-41d4-a716-446655440000"), true},
        {"valid lowercase", ComplaintID("9cb3bb9e-b6dc-4e02-9767-e396a42b63a6"), true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expected, tt.id.IsValid())
        })
    }
}

func TestComplaintID_String(t *testing.T) {
    uuidStr := "550e8400-e29b-41d4-a716-446655440000"
    id := ComplaintID(uuidStr)
    
    assert.Equal(t, uuidStr, id.String())
}

func TestComplaintID_UUID(t *testing.T) {
    t.Run("valid ID", func(t *testing.T) {
        uuidStr := "550e8400-e29b-41d4-a716-446655440000"
        id := ComplaintID(uuidStr)
        
        parsed, err := id.UUID()
        require.NoError(t, err)
        assert.Equal(t, uuidStr, parsed.String())
    })
    
    t.Run("empty ID", func(t *testing.T) {
        id := ComplaintID("")
        
        _, err := id.UUID()
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "cannot be empty")
    })
    
    t.Run("invalid ID", func(t *testing.T) {
        id := ComplaintID("not-a-uuid")
        
        _, err := id.UUID()
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "must be valid UUID v4 format")
    })
}
```

#### **AgentID Tests**
```go
// internal/domain/agent_id_test.go
package domain

import (
    "strings"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewAgentID(t *testing.T) {
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
            name:       "valid with spaces",
            input:      "AI Coding Assistant",
            wantErr:    false,
            expectedID: AgentID("AI Coding Assistant"),
        },
        {
            name:       "valid with numbers",
            input:      "Agent-123",
            wantErr:    false,
            expectedID: AgentID("Agent-123"),
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
            name:       "whitespace only",
            input:      "   ",
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
        {
            name:       "invalid characters - #",
            input:      "AI#Assistant",
            wantErr:    true,
            errorMsg:   "contains invalid characters",
            expectedID: AgentID(""),
        },
        {
            name:       "invalid characters - %",
            input:      "AI%Assistant",
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

func TestParseAgentID(t *testing.T) {
    tests := []struct {
        name         string
        input        string
        wantErr      bool
        errorMsg     string
        expectedID   AgentID
    }{
        {
            name:       "valid agent name",
            input:      "AI-Assistant",
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
            name:       "invalid characters",
            input:      "AI@Assistant",
            wantErr:    true,
            errorMsg:   "contains invalid characters",
            expectedID: AgentID(""),
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseAgentID(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedID, got)
            }
        })
    }
}

func TestAgentID_Methods(t *testing.T) {
    validID := AgentID("AI-Assistant")
    invalidID := AgentID("")
    emptyID := AgentID("")
    
    t.Run("String method", func(t *testing.T) {
        assert.Equal(t, "AI-Assistant", validID.String())
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
}
```

#### **SessionID and ProjectID Tests**
```go
// Similar comprehensive test coverage for SessionID and ProjectID
// Follow same pattern as AgentID tests with appropriate validation rules
```

### **Phase 2: JSON Serialization Tests**

#### **Marshal/Unmarshal Tests**
```go
// internal/domain/json_serialization_test.go
package domain

import (
    "encoding/json"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestComplaint_JSONSerialization_FlatIDs(t *testing.T) {
    // Create complaint with phantom types
    complaint := &Complaint{
        ID:             mustNewComplaintID(),
        AgentID:        mustNewAgentID("AI-Assistant"),
        SessionID:      mustNewSessionID("dev-session"),
        ProjectID:      mustNewProjectID("my-project"),
        TaskDescription: "Test task description",
        ContextInfo:     "Test context info",
        MissingInfo:     "Test missing info",
        ConfusedBy:      "Test confused by",
        FutureWishes:    "Test future wishes",
        Severity:        SeverityHigh,
        Timestamp:       time.Date(2024, 11, 9, 12, 18, 30, 0, time.UTC),
        ProjectName:     mustNewProjectName("my-project"),
        ResolutionState: ResolutionStateOpen,
    }
    
    t.Run("Marshal produces flat JSON", func(t *testing.T) {
        data, err := json.Marshal(complaint)
        require.NoError(t, err)
        
        // Verify flat structure
        var result map[string]any
        err = json.Unmarshal(data, &result)
        require.NoError(t, err)
        
        // ID should be flat string, not nested object
        idValue, exists := result["id"]
        require.True(t, exists, "id field should exist")
        
        idStr, isString := idValue.(string)
        assert.True(t, isString, "id should be string, got %T", idValue)
        assert.Equal(t, complaint.ID.String(), idStr)
        
        // Verify no nested "Value" objects
        assert.NotContains(t, string(data), `"Value"`)
    })
    
    t.Run("Unmarshal from flat JSON", func(t *testing.T) {
        // Create flat JSON
        jsonData := map[string]any{
            "id":              complaint.ID.String(),
            "agent_id":         complaint.AgentID.String(),
            "session_id":       complaint.SessionID.String(),
            "project_id":       complaint.ProjectID.String(),
            "task_description": complaint.TaskDescription,
            "severity":        "high",
            "timestamp":       complaint.Timestamp.Format(time.RFC3339),
        }
        
        data, err := json.Marshal(jsonData)
        require.NoError(t, err)
        
        // Unmarshal to complaint
        var unmarshaled Complaint
        err = json.Unmarshal(data, &unmarshaled)
        require.NoError(t, err)
        
        // Verify phantom types
        assert.Equal(t, complaint.ID, unmarshaled.ID)
        assert.Equal(t, complaint.AgentID, unmarshaled.AgentID)
        assert.Equal(t, complaint.SessionID, unmarshaled.SessionID)
        assert.Equal(t, complaint.ProjectID, unmarshaled.ProjectID)
    })
    
    t.Run("Reject nested ID format", func(t *testing.T) {
        // Create JSON with nested ID format (old format)
        nestedJSON := map[string]any{
            "id": map[string]any{
                "Value": complaint.ID.String(),
            },
            "task_description": complaint.TaskDescription,
            "severity":        "high",
        }
        
        data, err := json.Marshal(nestedJSON)
        require.NoError(t, err)
        
        // Attempting to unmarshal should handle gracefully
        // This depends on our unmarshaling strategy
        var unmarshaled Complaint
        err = json.Unmarshal(data, &unmarshaled)
        
        // Either should error or handle gracefully based on implementation
        // For now, we expect some handling strategy
        // The exact behavior will depend on chosen implementation
    })
    
    t.Run("Roundtrip preserves phantom types", func(t *testing.T) {
        // Marshal to JSON
        data, err := json.Marshal(complaint)
        require.NoError(t, err)
        
        // Unmarshal back to complaint
        var unmarshaled Complaint
        err = json.Unmarshal(data, &unmarshaled)
        require.NoError(t, err)
        
        // Verify all phantom types are preserved
        assert.Equal(t, complaint.ID, unmarshaled.ID)
        assert.Equal(t, complaint.AgentID, unmarshaled.AgentID)
        assert.Equal(t, complaint.SessionID, unmarshaled.SessionID)
        assert.Equal(t, complaint.ProjectID, unmarshaled.ProjectID)
        assert.Equal(t, complaint.TaskDescription, unmarshaled.TaskDescription)
        assert.Equal(t, complaint.Severity, unmarshaled.Severity)
    })
}

func TestComplaintDTO_JSONSerialization_FlatIDs(t *testing.T) {
    // Create complaint with phantom types
    complaint := &Complaint{
        ID:             mustNewComplaintID(),
        AgentID:        mustNewAgentID("AI-Assistant"),
        SessionID:      mustNewSessionID("dev-session"),
        ProjectID:      mustNewProjectID("my-project"),
        TaskDescription: "Test task description",
        Severity:        SeverityHigh,
        Timestamp:       time.Date(2024, 11, 9, 12, 18, 30, 0, time.UTC),
        ResolutionState: ResolutionStateOpen,
    }
    
    // Convert to DTO
    dto := ToDTO(complaint)
    
    t.Run("DTO produces flat JSON", func(t *testing.T) {
        data, err := json.Marshal(dto)
        require.NoError(t, err)
        
        // Verify flat structure
        var result map[string]any
        err = json.Unmarshal(data, &result)
        require.NoError(t, err)
        
        // ID should be flat string
        idValue, exists := result["id"]
        require.True(t, exists, "id field should exist")
        
        idStr, isString := idValue.(string)
        assert.True(t, isString, "id should be string, got %T", idValue)
        assert.Equal(t, complaint.ID.String(), idStr)
        
        // Verify no nested objects
        assert.NotContains(t, string(data), `"Value"`)
    })
    
    t.Run("DTO includes all expected fields", func(t *testing.T) {
        data, err := json.Marshal(dto)
        require.NoError(t, err)
        
        // Verify key fields are present
        assert.Contains(t, string(data), `"id"`)
        assert.Contains(t, string(data), `"agent_name"`)
        assert.Contains(t, string(data), `"session_name"`)
        assert.Contains(t, string(data), `"project_name"`)
        assert.Contains(t, string(data), `"task_description"`)
        assert.Contains(t, string(data), `"severity"`)
        assert.Contains(t, string(data), `"timestamp"`)
        
        // Verify file paths are included
        assert.Contains(t, string(data), `"file_path"`)
        assert.Contains(t, string(data), `"docs_path"`)
    })
}
```

### **Phase 3: Integration Tests**

#### **Service Layer Integration Tests**
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
    "github.com/larsartmann/complaints-mcp/internal/repo"
    "github.com/larsartmann/complaints-mcp/internal/tracing"
)

func TestComplaintService_WithPhantomTypes(t *testing.T) {
    t.Run("create complaint with typed IDs", func(t *testing.T) {
        // Setup
        tempDir := t.TempDir()
        tracer := tracing.NewNoOpTracer()
        logger := &mockLogger{}
        repository := repo.NewFileRepository(tempDir, tracer)
        service := NewComplaintService(repository, tracer, logger)
        
        // Create typed IDs
        agentID, err := domain.NewAgentID("AI-Assistant")
        require.NoError(t, err)
        
        sessionID, err := domain.NewSessionID("integration-test")
        require.NoError(t, err)
        
        projectID, err := domain.NewProjectID("test-project")
        require.NoError(t, err)
        
        // Create complaint with typed IDs
        complaint, err := service.CreateComplaint(
            context.Background(),
            agentID.String(),
            sessionID.String(),
            "Integration test task",
            "Test context info",
            "Test missing info",
            "Test confused by",
            "Test future wishes",
            "medium",
            projectID.String(),
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
        // Setup
        tempDir := t.TempDir()
        tracer := tracing.NewNoOpTracer()
        logger := &mockLogger{}
        repository := repo.NewFileRepository(tempDir, tracer)
        service := NewComplaintService(repository, tracer, logger)
        
        // Try to create complaint with invalid agent ID
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
        // Setup
        tempDir := t.TempDir()
        tracer := tracing.NewNoOpTracer()
        logger := &mockLogger{}
        repository := repo.NewFileRepository(tempDir, tracer)
        service := NewComplaintService(repository, tracer, logger)
        
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
        // Setup
        tempDir := t.TempDir()
        tracer := tracing.NewNoOpTracer()
        logger := &mockLogger{}
        repository := repo.NewFileRepository(tempDir, tracer)
        service := NewComplaintService(repository, tracer, logger)
        
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
        // Setup
        tempDir := t.TempDir()
        tracer := tracing.NewNoOpTracer()
        logger := &mockLogger{}
        repository := repo.NewFileRepository(tempDir, tracer)
        service := NewComplaintService(repository, tracer, logger)
        
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
```

### **Phase 4: BDD Tests with Phantom Types**

#### **Updated BDD Scenarios**
```gherkin
// features/bdd/phantom_type_safety_bdd.feature
Feature: Phantom Type Safety for Complaint Management
  As an AI assistant
  I want to use type-safe phantom IDs throughout the system
  So that ID confusion errors are prevented and type safety is enforced

  Background:
    Given the MCP server is running
    And phantom types are implemented for all ID fields

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

  Scenario: Validation provides clear error messages
    Given I attempt to create a ComplaintID with invalid format
    When I call the validation method
    Then I should receive a clear error message
    And the error should indicate the specific validation failure

  Scenario: Phantom types support string conversion
    Given I have a valid phantom type ID
    When I convert it to string
    Then the string should represent the ID value
    And I can convert it back to phantom type when valid
```

#### **BDD Test Implementation**
```go
// features/bdd/phantom_type_safety_bdd_test.go
package bdd

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/cucumber/godog"
    "github.com/stretchr/testify/assert"
    
    "github.com/larsartmann/complaints-mcp/internal/domain"
)

type PhantomTypeBDDContext struct {
    agentID       domain.AgentID
    sessionID     domain.SessionID
    projectID     domain.ProjectID
    complaintID   domain.ComplaintID
    complaint     *domain.Complaint
    lastError     error
    jsonResponse  string
}

func (ctx *PhantomTypeBDDContext) iHaveValidTypedIDs() error {
    var err error
    
    ctx.agentID, err = domain.NewAgentID("AI-Assistant")
    if err != nil {
        return fmt.Errorf("failed to create agent ID: %w", err)
    }
    
    ctx.sessionID, err = domain.NewSessionID("test-session")
    if err != nil {
        return fmt.Errorf("failed to create session ID: %w", err)
    }
    
    ctx.projectID, err = domain.NewProjectID("test-project")
    if err != nil {
        return fmt.Errorf("failed to create project ID: %w", err)
    }
    
    return nil
}

func (ctx *PhantomTypeBDDContext) iFileComplaintWithTypeSafeIDs() error {
    // This would be implemented via the service layer
    // For BDD purposes, we'll create the complaint directly
    var err error
    
    ctx.complaintID, err = domain.NewComplaintID()
    if err != nil {
        return fmt.Errorf("failed to create complaint ID: %w", err)
    }
    
    ctx.complaint = &domain.Complaint{
        ID:             ctx.complaintID,
        AgentID:        ctx.agentID,
        SessionID:      ctx.sessionID,
        ProjectID:      ctx.projectID,
        TaskDescription: "BDD test complaint with phantom types",
        Severity:        domain.SeverityMedium,
        Timestamp:       time.Now(),
        ResolutionState: domain.ResolutionStateOpen,
    }
    
    // Create JSON response for verification
    dto := delivery.ToDTO(ctx.complaint)
    jsonData, err := json.Marshal(dto)
    if err != nil {
        return fmt.Errorf("failed to marshal complaint to JSON: %w", err)
    }
    
    ctx.jsonResponse = string(jsonData)
    return nil
}

func (ctx *PhantomTypeBDDContext) theComplaintShouldBeCreated() error {
    if ctx.lastError != nil {
        return fmt.Errorf("complaint creation failed: %w", ctx.lastError)
    }
    
    if ctx.complaint == nil {
        return fmt.Errorf("complaint is nil after creation")
    }
    
    return nil
}

func (ctx *PhantomTypeBDDContext) allIDFieldsShouldBeValid() error {
    tests := []struct {
        name string
        id   fmt.Stringer
    }{
        {"complaint ID", ctx.complaintID},
        {"agent ID", ctx.agentID},
        {"session ID", ctx.sessionID},
        {"project ID", ctx.projectID},
    }
    
    for _, test := range tests {
        switch v := test.id.(type) {
        case domain.ComplaintID:
            if !v.IsValid() {
                return fmt.Errorf("%s is invalid", test.name)
            }
        case domain.AgentID:
            if !v.IsValid() {
                return fmt.Errorf("%s is invalid", test.name)
            }
        case domain.SessionID:
            if !v.IsValid() {
                return fmt.Errorf("%s is invalid", test.name)
            }
        case domain.ProjectID:
            if !v.IsValid() {
                return fmt.Errorf("%s is invalid", test.name)
            }
        }
    }
    
    return nil
}

func (ctx *PhantomTypeBDDContext) theJSONShouldUseFlatIDFormat() error {
    var response map[string]any
    err := json.Unmarshal([]byte(ctx.jsonResponse), &response)
    if err != nil {
        return fmt.Errorf("failed to parse JSON response: %w", err)
    }
    
    // Check that ID is flat string
    idValue, exists := response["id"]
    if !exists {
        return fmt.Errorf("ID field not found in JSON response")
    }
    
    _, isString := idValue.(string)
    if !isString {
        return fmt.Errorf("ID field is not flat string: %T", idValue)
    }
    
    // Check that no nested "Value" objects exist
    if contains(ctx.jsonResponse, `"Value"`) {
        return fmt.Errorf("JSON response contains nested 'Value' objects")
    }
    
    return nil
}

func (ctx *PhantomTypeBDDContext) iHaveInvalidAgentIDWithSpecialChars() error {
    var err error
    ctx.agentID, err = domain.NewAgentID("AI@Assistant#$")
    ctx.lastError = err
    return nil
}

func (ctx *PhantomTypeBDDContext) theComplaintCreationShouldBeRejected() error {
    if ctx.lastError == nil {
        return fmt.Errorf("expected complaint creation to be rejected, but it succeeded")
    }
    return nil
}

func (ctx *PhantomTypeBDDContext) theErrorMessageShouldBeClear() error {
    if ctx.lastError == nil {
        return fmt.Errorf("expected error but got none")
    }
    
    expectedSubstrings := []string{
        "invalid",
        "agent name",
        "character",
    }
    
    errorStr := ctx.lastError.Error()
    for _, expected := range expectedSubstrings {
        if !contains(errorStr, expected) {
            return fmt.Errorf("error message '%s' should contain '%s'", errorStr, expected)
        }
    }
    
    return nil
}

// Helper functions
func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || 
           (len(s) > len(substr) && 
            (strings.HasPrefix(s, substr) || 
             strings.HasSuffix(s, substr) || 
             indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return i
        }
    }
    return -1
}
```

### **Phase 5: Performance Benchmarks**

#### **Phantom Types vs Struct Performance**
```go
// internal/benchmarks/phantom_type_benchmark_test.go
package benchmarks

import (
    "encoding/json"
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

func BenchmarkPhantomType_MarshalJSON(b *testing.B) {
    complaint := &domain.Complaint{
        ID:             mustNewComplaintID(),
        AgentID:        mustNewAgentID("AI-Assistant"),
        SessionID:      mustNewSessionID("test-session"),
        ProjectID:      mustNewProjectID("test-project"),
        TaskDescription: "Benchmark test complaint",
        Severity:        domain.SeverityMedium,
        Timestamp:       time.Now(),
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = json.Marshal(complaint)
    }
}

// Compare with struct implementation (for regression testing)
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

func BenchmarkStructID_MarshalJSON(b *testing.B) {
    complaint := &OldComplaint{
        ID: &OldComplaintID{
            Value: "550e8400-e29b-41d4-a716-446655440000",
        },
        AgentName:       "AI-Assistant",
        TaskDescription: "Benchmark test complaint",
        Severity:        domain.SeverityMedium,
        Timestamp:       time.Now(),
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = json.Marshal(complaint)
    }
}

// Old struct types for comparison
type OldComplaintID struct {
    Value string `json:"Value"`
}

func (id *OldComplaintID) String() string {
    return id.Value
}

type OldComplaint struct {
    ID              *OldComplaintID `json:"id"`
    AgentName       string          `json:"agent_name"`
    TaskDescription string          `json:"task_description"`
    Severity        domain.Severity `json:"severity"`
    Timestamp       time.Time       `json:"timestamp"`
}
```

### **Phase 6: Property-Based Tests**

#### **Property-Based Testing with Gopter**
```go
// internal/domain/property_based_test.go
package domain

import (
    "testing"
    
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
)

func TestComplaintID_Properties(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    // Property: Valid UUIDs should parse successfully
    properties.Property("valid UUIDs parse successfully", prop.ForAll(
        gen.UUID4(),
        func(uuidVal uuid.UUID) bool {
            uuidStr := uuidVal.String()
            
            id, err := ParseComplaintID(uuidStr)
            
            return err == nil && 
                   !id.IsEmpty() && 
                   id.IsValid() && 
                   id.String() == uuidStr
        },
    ))
    
    // Property: Invalid UUIDs should fail parsing
    properties.Property("invalid UUIDs fail parsing", prop.ForAll(
        gen.AlphaString().WithMinLen(1).WithMaxLen(50),
        func(invalidStr string) bool {
            // Only test strings that are not valid UUIDs
            _, err := uuid.Parse(invalidStr)
            if err == nil {
                return true // Skip valid UUIDs
            }
            
            id, parseErr := ParseComplaintID(invalidStr)
            
            return parseErr != nil && 
                   id.IsEmpty() && 
                   !id.IsValid()
        },
    ))
    
    // Property: NewComplaintID always generates valid IDs
    properties.Property("NewComplaintID always generates valid IDs", prop.ForAll(
        gen.Unit(),
        func(_unit struct{}) bool {
            id, err := NewComplaintID()
            
            return err == nil && 
                   !id.IsEmpty() && 
                   id.IsValid()
        },
    ))
    
    // Property: String() is idempotent
    properties.Property("String() is idempotent", prop.ForAll(
        gen.UUID4(),
        func(uuidVal uuid.UUID) bool {
            uuidStr := uuidVal.String()
            
            id, err := ParseComplaintID(uuidStr)
            if err != nil {
                return true // Skip invalid UUIDs
            }
            
            // Multiple calls to String() should return same result
            first := id.String()
            second := id.String()
            
            return first == second
        },
    ))
    
    // Run properties
    testing.RunT(t, gopter.NewDefaultTestingParameters().WithMinSuccessfulTests(1000), properties)
}

func TestAgentID_Properties(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    // Property: Valid agent names parse successfully
    properties.Property("valid agent names parse successfully", prop.ForAll(
        gen.AlphaNumString().WithMinLen(1).WithMaxLen(100),
        func(name string) bool {
            id, err := NewAgentID(name)
            
            return err == nil && 
                   !id.IsEmpty() && 
                   id.IsValid() && 
                   id.String() == name
        },
    ))
    
    // Property: String() normalizes whitespace consistently
    properties.Property("String() normalizes whitespace", prop.ForAll(
        gen.AlphaNumString().WithMinLen(1).WithMaxLen(50),
        func(name string) bool {
            // Add random whitespace
            withSpaces := "  " + name + "  "
            
            id, err := NewAgentID(withSpaces)
            if err != nil {
                return true // Skip invalid names
            }
            
            // String() should return trimmed version
            return id.String() == name
        },
    ))
    
    testing.RunT(t, gopter.NewDefaultTestingParameters().WithMinSuccessfulTests(1000), properties)
}
```

## 🎯 **Test Coverage Goals**

### **Coverage Targets**
- **Unit Tests**: 95%+ coverage for phantom type implementations
- **Integration Tests**: 90%+ coverage for service layer with phantom types
- **BDD Tests**: Complete workflow coverage with type safety scenarios
- **Performance Tests**: Benchmark all phantom type operations
- **Property Tests**: 1000+ iterations for property-based validation

### **Test Categories**
1. **Construction Tests**: New() and Parse() functions
2. **Validation Tests**: All validation rules and edge cases
3. **Conversion Tests**: String() and type conversions
4. **Serialization Tests**: JSON marshal/unmarshal
5. **Integration Tests**: Service layer workflows
6. **Performance Tests**: Benchmarks vs current implementation
7. **Property Tests**: Mathematical properties of operations
8. **Error Handling Tests**: Invalid input scenarios
9. **Edge Case Tests**: Boundary conditions and rare scenarios
10. **Regression Tests**: Ensure no functionality loss

## 📋 **Files to Create**

### **Test Files**
- `internal/domain/complaint_id_test.go` - ComplaintID unit tests
- `internal/domain/agent_id_test.go` - AgentID unit tests
- `internal/domain/session_id_test.go` - SessionID unit tests
- `internal/domain/project_id_test.go` - ProjectID unit tests
- `internal/domain/json_serialization_test.go` - JSON serialization tests
- `internal/service/complaint_service_integration_test.go` - Service integration tests
- `features/bdd/phantom_type_safety_bdd_test.go` - BDD tests
- `internal/benchmarks/phantom_type_benchmark_test.go` - Performance benchmarks
- `internal/domain/property_based_test.go` - Property-based tests

### **Test Utilities**
- `internal/testing/test_helpers.go` - Test helper functions
- `internal/testing/fixtures.go` - Test data fixtures
- `internal/testing/assertions.go` - Custom assertions

## 🏆 **Success Criteria**

- [ ] All phantom type operations have comprehensive unit tests
- [ ] JSON serialization produces correct flat structure
- [ ] Service layer integration tests pass with typed IDs
- [ ] BDD scenarios cover all type safety workflows
- [ ] Performance benchmarks show no regression
- [ ] Property-based tests validate mathematical properties
- [ ] Test coverage meets or exceeds targets
- [ ] All tests run successfully in CI/CD pipeline
- [ ] Test documentation is complete and clear

## 🏷️ **Labels**
- `testing` - Comprehensive test coverage
- `quality-assurance` - Code quality validation
- `type-safety` - Phantom type testing
- `performance` - Benchmark testing
- `integration` - End-to-end testing
- `medium-priority` - Important for reliability

## 📊 **Priority**: Medium
- **Complexity**: High (comprehensive test suite)
- **Value**: High (ensures reliability and correctness)
- **Risk**: Low (testing is additive)
- **Dependencies**: Issues #48, #49, #50

## 🤝 **Dependencies**
- **Issue #48**: Must have phantom types implemented
- **Issue #49**: Should have all ID fields converted
- **Issue #50**: Need validation constructors for testing
- **Issue #51**: Need schema alignment for JSON tests

---

**This comprehensive test suite ensures phantom type implementation is thoroughly validated, performant, and reliable across all usage scenarios.**