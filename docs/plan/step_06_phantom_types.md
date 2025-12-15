# Step 6: Implement Phantom Types for All ID Fields

## 🎯 Objective

Implement comprehensive phantom type system for type safety and flat JSON structure.

## 🏗️ Implementation Tasks

### A. Create Phantom Type Framework

```go
// internal/domain/types.go
package domain

import (
    "fmt"
    "regexp"
    "strings"
    "time"
)

// Phantom types with compile-time safety and zero runtime overhead
type (
    ComplaintID string  // UUID v4 format
    SessionID   string  // Session identifier
    ProjectID   string  // Project identifier
    AgentID     string  // Agent identifier
)

// Validation patterns
var (
    uuidPattern      = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
    namePattern      = regexp.MustCompile(`^[a-zA-Z0-9\-_\s\.]{1,100}$`)
)

// Common validation functions
func validateUUID(s string) error {
    if s == "" {
        return fmt.Errorf("cannot be empty")
    }
    if !uuidPattern.MatchString(s) {
        return fmt.Errorf("must be valid UUID v4 format")
    }
    return nil
}

func validateName(s string) error {
    trimmed := strings.TrimSpace(s)
    if trimmed == "" {
        return fmt.Errorf("cannot be empty")
    }
    if len(trimmed) > 100 {
        return fmt.Errorf("cannot exceed 100 characters")
    }
    if !namePattern.MatchString(trimmed) {
        return fmt.Errorf("contains invalid characters")
    }
    return nil
}
```

### B. ComplaintID Implementation

```go
// internal/domain/complaint_id.go
type ComplaintID string

func NewComplaintID() (ComplaintID, error) {
    id, err := uuid.NewRandom()
    if err != nil {
        return ComplaintID(""), fmt.Errorf("failed to generate ComplaintID: %w", err)
    }
    return ComplaintID(id.String()), nil
}

func ParseComplaintID(s string) (ComplaintID, error) {
    if err := validateUUID(s); err != nil {
        return ComplaintID(""), fmt.Errorf("invalid ComplaintID: %w", err)
    }
    return ComplaintID(s), nil
}

func MustParseComplaintID(s string) ComplaintID {
    id, err := ParseComplaintID(s)
    if err != nil {
        panic(fmt.Sprintf("invalid ComplaintID: %s", err))
    }
    return id
}

func (id ComplaintID) Validate() error {
    return validateUUID(string(id))
}

func (id ComplaintID) IsValid() bool {
    return id.Validate() == nil
}

func (id ComplaintID) IsEmpty() bool {
    return string(id) == ""
}

func (id ComplaintID) String() string {
    return string(id)
}

func (id ComplaintID) UUID() (uuid.UUID, error) {
    return uuid.Parse(string(id))
}

// JSON serialization (flat structure)
func (id ComplaintID) MarshalJSON() ([]byte, error) {
    return json.Marshal(id.String())
}

func (id *ComplaintID) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    parsed, err := ParseComplaintID(s)
    if err != nil {
        return err
    }
    *id = parsed
    return nil
}
```

### C. AgentID Implementation

```go
// internal/domain/agent_id.go
type AgentID string

func NewAgentID(name string) (AgentID, error) {
    trimmed := strings.TrimSpace(name)
    if err := validateName(trimmed); err != nil {
        return AgentID(""), fmt.Errorf("invalid AgentID: %w", err)
    }
    return AgentID(trimmed), nil
}

func ParseAgentID(s string) (AgentID, error) {
    if err := validateName(s); err != nil {
        return AgentID(""), fmt.Errorf("invalid AgentID: %w", err)
    }
    return AgentID(s), nil
}

func (id AgentID) Validate() error {
    return validateName(string(id))
}

func (id AgentID) IsValid() bool {
    return id.Validate() == nil
}

func (id AgentID) IsEmpty() bool {
    return strings.TrimSpace(string(id)) == ""
}

func (id AgentID) String() string {
    return string(id)
}

// JSON serialization (flat structure)
func (id AgentID) MarshalJSON() ([]byte, error) {
    return json.Marshal(id.String())
}

func (id *AgentID) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    parsed, err := ParseAgentID(s)
    if err != nil {
        return err
    }
    *id = parsed
    return nil
}
```

### D. SessionID Implementation

```go
// internal/domain/session_id.go
type SessionID string

func NewSessionID(name string) (SessionID, error) {
    trimmed := strings.TrimSpace(name)
    if err := validateName(trimmed); err != nil {
        return SessionID(""), fmt.Errorf("invalid SessionID: %w", err)
    }
    return SessionID(trimmed), nil
}

func ParseSessionID(s string) (SessionID, error) {
    if err := validateName(s); err != nil {
        return SessionID(""), fmt.Errorf("invalid SessionID: %w", err)
    }
    return SessionID(s), nil
}

func (id SessionID) Validate() error {
    return validateName(string(id))
}

func (id SessionID) IsValid() bool {
    return id.Validate() == nil
}

func (id SessionID) IsEmpty() bool {
    return strings.TrimSpace(string(id)) == ""
}

func (id SessionID) String() string {
    return string(id)
}

// JSON serialization
func (id SessionID) MarshalJSON() ([]byte, error) {
    return json.Marshal(id.String())
}

func (id *SessionID) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    parsed, err := ParseSessionID(s)
    if err != nil {
        return err
    }
    *id = parsed
    return nil
}
```

### E. ProjectID Implementation

```go
// internal/domain/project_id.go
type ProjectID string

func NewProjectID(name string) (ProjectID, error) {
    trimmed := strings.TrimSpace(name)
    if err := validateName(trimmed); err != nil {
        return ProjectID(""), fmt.Errorf("invalid ProjectID: %w", err)
    }
    return ProjectID(trimmed), nil
}

func ParseProjectID(s string) (ProjectID, error) {
    if err := validateName(s); err != nil {
        return ProjectID(""), fmt.Errorf("invalid ProjectID: %w", err)
    }
    return ProjectID(s), nil
}

func (id ProjectID) Validate() error {
    return validateName(string(id))
}

func (id ProjectID) IsValid() bool {
    return id.Validate() == nil
}

func (id ProjectID) IsEmpty() bool {
    return strings.TrimSpace(string(id)) == ""
}

func (id ProjectID) String() string {
    return string(id)
}

// JSON serialization
func (id ProjectID) MarshalJSON() ([]byte, error) {
    return json.Marshal(id.String())
}

func (id *ProjectID) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    parsed, err := ParseProjectID(s)
    if err != nil {
        return err
    }
    *id = parsed
    return nil
}
```

### F. Update Domain Entities

```go
// internal/domain/complaint.go (updated)
type Complaint struct {
    ID              ComplaintID     `json:"id"`              // ✅ Phantom type - flat JSON
    AgentID         AgentID         `json:"agent_id"`          // ✅ Phantom type
    SessionID       SessionID       `json:"session_id"`        // ✅ Phantom type
    ProjectID       ProjectID       `json:"project_id"`        // ✅ Phantom type
    TaskDescription  string          `json:"task_description"`
    ContextInfo      string          `json:"context_info"`
    MissingInfo      string          `json:"missing_info"`
    ConfusedBy       string          `json:"confused_by"`
    FutureWishes     string          `json:"future_wishes"`
    Severity        Severity        `json:"severity"`
    Timestamp       time.Time       `json:"timestamp"`
    ResolutionState ResolutionState `json:"resolution_state"`
    ResolvedAt      *time.Time     `json:"resolved_at,omitempty"`
    ResolvedBy      string          `json:"resolved_by,omitempty"`
}

func (c *Complaint) Validate() error {
    // Validate all phantom types
    if err := c.ID.Validate(); err != nil {
        return fmt.Errorf("invalid ComplaintID: %w", err)
    }
    if err := c.AgentID.Validate(); err != nil {
        return fmt.Errorf("invalid AgentID: %w", err)
    }
    if err := c.SessionID.Validate(); err != nil {
        return fmt.Errorf("invalid SessionID: %w", err)
    }
    if err := c.ProjectID.Validate(); err != nil {
        return fmt.Errorf("invalid ProjectID: %w", err)
    }

    // Validate other fields...
    return nil
}

func (c *Complaint) IsValid() bool {
    return c.Validate() == nil
}
```

### G. Update Repository Layer

```go
// internal/repo/repository.go (updated)
type Repository interface {
    Save(ctx context.Context, complaint *domain.Complaint) error
    FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error)
    FindBySession(ctx context.Context, sessionID domain.SessionID, limit int) ([]*domain.Complaint, error)
    FindByProject(ctx context.Context, projectID domain.ProjectID, limit int) ([]*domain.Complaint, error)
    FindByAgent(ctx context.Context, agentID domain.AgentID, limit int) ([]*domain.Complaint, error)
    Update(ctx context.Context, complaint *domain.Complaint) error
    Delete(ctx context.Context, id domain.ComplaintID) error
    FindAll(ctx context.Context, limit, offset int) ([]*domain.Complaint, error)
    FindBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error)
    FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error)
    Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error)
    WarmCache(ctx context.Context) error
    GetCacheStats() CacheStats
    GetFilePath(ctx context.Context, id domain.ComplaintID) (string, error)
    GetDocsPath(ctx context.Context, id domain.ComplaintID) (string, error)
}

// internal/repo/file_repository.go (updated)
func (r *FileRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
    // Validate complaint
    if err := complaint.Validate(); err != nil {
        return fmt.Errorf("invalid complaint: %w", err)
    }

    // Serialize with flat JSON structure
    data, err := json.Marshal(complaint)
    if err != nil {
        return fmt.Errorf("failed to marshal complaint: %w", err)
    }

    // Use phantom type for file naming
    fileName := fmt.Sprintf("%s.json", complaint.ID.String())
    return r.writeFile(ctx, fileName, data)
}

func (r *FileRepository) FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
    if err := id.Validate(); err != nil {
        return nil, fmt.Errorf("invalid ComplaintID: %w", err)
    }

    fileName := fmt.Sprintf("%s.json", id.String())
    data, err := r.readFile(ctx, fileName)
    if err != nil {
        return nil, err
    }

    var complaint domain.Complaint
    if err := json.Unmarshal(data, &complaint); err != nil {
        return nil, fmt.Errorf("failed to unmarshal complaint: %w", err)
    }

    return &complaint, nil
}
```

### H. Update Service Layer

```go
// internal/service/complaint_service.go (updated)
func (s *ComplaintService) CreateComplaint(
    ctx context.Context,
    agentName string,
    sessionName string,
    taskDescription, contextInfo, missingInfo, confusedBy, futureWishes string,
    severity domain.Severity,
    projectName string,
) (*domain.Complaint, error) {
    // Create typed IDs with validation
    agentID, err := domain.NewAgentID(agentName)
    if err != nil {
        return nil, fmt.Errorf("invalid agent name: %w", err)
    }

    sessionID, err := domain.NewSessionID(sessionName)
    if err != nil {
        return nil, fmt.Errorf("invalid session name: %w", err)
    }

    projectID, err := domain.NewProjectID(projectName)
    if err != nil {
        return nil, fmt.Errorf("invalid project name: %w", err)
    }

    complaintID, err := domain.NewComplaintID()
    if err != nil {
        return nil, fmt.Errorf("failed to generate ComplaintID: %w", err)
    }

    // Create complaint with phantom types
    complaint := &domain.Complaint{
        ID:              complaintID,
        AgentID:         agentID,
        SessionID:       sessionID,
        ProjectID:       projectID,
        TaskDescription:  taskDescription,
        ContextInfo:      contextInfo,
        MissingInfo:      missingInfo,
        ConfusedBy:       confusedBy,
        FutureWishes:     futureWishes,
        Severity:        severity,
        Timestamp:       time.Now(),
        ResolutionState: domain.ResolutionStateOpen,
    }

    // Validate complete complaint
    if err := complaint.Validate(); err != nil {
        return nil, fmt.Errorf("invalid complaint: %w", err)
    }

    return s.repo.Save(ctx, complaint)
}
```

### I. Update MCP Handlers

```go
// internal/delivery/mcp/mcp_server.go (updated)
func (m *MCPServer) handleFileComplaint(ctx context.Context, req *mcp.CallToolRequest, input FileComplaintInput) (*mcp.CallToolResult, FileComplaintOutput, error) {
    // Convert strings to phantom types with validation
    agentID, err := domain.NewAgentID(input.AgentName)
    if err != nil {
        return nil, FileComplaintOutput{}, fmt.Errorf("invalid agent name: %w", err)
    }

    sessionID, err := domain.NewSessionID(input.SessionName)
    if err != nil {
        return nil, FileComplaintOutput{}, fmt.Errorf("invalid session name: %w", err)
    }

    projectID, err := domain.NewProjectID(input.ProjectName)
    if err != nil {
        return nil, FileComplaintOutput{}, fmt.Errorf("invalid project name: %w", err)
    }

    // Create complaint with phantom types
    complaint, err := m.service.CreateComplaint(
        ctx,
        agentID.String(),    // Service expects strings
        sessionID.String(),
        input.TaskDescription,
        input.ContextInfo,
        input.MissingInfo,
        input.ConfusedBy,
        input.FutureWishes,
        domain.Severity(input.Severity),
        projectID.String(),
    )
    if err != nil {
        return nil, FileComplaintOutput{}, fmt.Errorf("failed to create complaint: %w", err)
    }

    // Get file paths using phantom type ID
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

### J. Update DTOs for API Compatibility

```go
// internal/delivery/mcp/dto.go (updated)
type ComplaintDTO struct {
    ID              string     `json:"id"`              // Flat string from phantom type
    AgentName       string     `json:"agent_name"`       // Backward compatibility
    SessionName     string     `json:"session_name"`     // Backward compatibility
    ProjectName     string     `json:"project_name"`     // Backward compatibility
    TaskDescription string     `json:"task_description"`
    ContextInfo     string     `json:"context_info"`
    MissingInfo     string     `json:"missing_info"`
    ConfusedBy      string     `json:"confused_by"`
    FutureWishes    string     `json:"future_wishes"`
    Severity        string     `json:"severity"`
    Timestamp       time.Time  `json:"timestamp"`
    Resolved        bool       `json:"resolved"`
    ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
    ResolvedBy      string     `json:"resolved_by,omitempty"`
    FilePath        string     `json:"file_path,omitempty"`
    DocsPath        string     `json:"docs_path,omitempty"`
}

func ToDTO(complaint *domain.Complaint) ComplaintDTO {
    return ComplaintDTO{
        ID:              complaint.ID.String(),      // ✅ Convert phantom type to string
        AgentName:       complaint.AgentID.String(),
        SessionName:     complaint.SessionID.String(),
        ProjectName:     complaint.ProjectID.String(),
        TaskDescription: complaint.TaskDescription,
        ContextInfo:     complaint.ContextInfo,
        MissingInfo:     complaint.MissingInfo,
        ConfusedBy:      complaint.ConfusedBy,
        FutureWishes:    complaint.FutureWishes,
        Severity:        string(complaint.Severity),
        Timestamp:       complaint.Timestamp,
        Resolved:        complaint.ResolutionState.IsResolved(),
        ResolvedAt:      complaint.ResolvedAt,
        ResolvedBy:      complaint.ResolvedBy,
    }
}

func ToDTOWithPaths(complaint *domain.Complaint, filePath, docsPath string) ComplaintDTO {
    dto := ToDTO(complaint)
    dto.FilePath = filePath
    dto.DocsPath = docsPath
    return dto
}
```

## 📝 Implementation Details

### Type Safety Benefits

```go
// Compile-time type safety - prevents ID mixing
func ProcessComplaint(id domain.ComplaintID, agentID domain.AgentID) error { ... }

agentID := domain.NewAgentID("AI-Assistant")
sessionID := domain.NewSessionID("dev-session")

// ProcessComplaint(sessionID, agentID) // ❌ Compile-time error!
ProcessComplaint(agentID, sessionID)    // ✅ Type-safe!
```

### Flat JSON Output

```go
// Before (nested)
{
  "id": {"Value": "550e8400-e29b-41d4-a716-446655440000"}  // ❌ Nested
}

// After (flat)
{
  "id": "550e8400-e29b-41d4-a716-446655440000"          // ✅ Flat
}
```

## 🧪 Testing Strategy

### Unit Tests

```go
// Test phantom type creation and validation
func TestComplaintID_New(t *testing.T) {
    id, err := domain.NewComplaintID()
    assert.NoError(t, err)
    assert.True(t, id.IsValid())
    assert.False(t, id.IsEmpty())
}

func TestAgentID_Validation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "AI-Assistant", false},
        {"empty", "", true},
        {"invalid chars", "AI@Assistant", true},
        {"too long", strings.Repeat("a", 101), true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            id, err := domain.NewAgentID(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.True(t, id.IsValid())
            }
        })
    }
}
```

### JSON Serialization Tests

```go
func TestPhantomType_JSONSerialization(t *testing.T) {
    id := domain.MustParseComplaintID("550e8400-e29b-41d4-a716-446655440000")

    data, err := json.Marshal(id)
    require.NoError(t, err)

    var result map[string]any
    err = json.Unmarshal(data, &result)
    require.NoError(t, err)

    // Verify flat structure
    idValue := result["id"]
    idStr, isString := idValue.(string)
    assert.True(t, isString, "ID should be flat string")
    assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", idStr)
    assert.NotContains(t, string(data), `"Value"`)  // No nesting
}
```

### Integration Tests

```go
func TestComplaintService_WithPhantomTypes(t *testing.T) {
    service := setupService()

    complaint, err := service.CreateComplaint(
        context.Background(),
        "AI-Assistant",    // Valid agent name
        "dev-session",     // Valid session name
        "Test task", "", "", "", "",
        "low",
        "my-project",     // Valid project name
    )

    assert.NoError(t, err)
    assert.NotNil(t, complaint)

    // Verify phantom types are valid
    assert.True(t, complaint.ID.IsValid())
    assert.True(t, complaint.AgentID.IsValid())
    assert.True(t, complaint.SessionID.IsValid())
    assert.True(t, complaint.ProjectID.IsValid())
}
```

## ⏱️ Time Estimate: 8-10 hours

## 🎯 Impact: HIGH (complete type safety + JSON fix)

## 💪 Work Required: HIGH (comprehensive refactoring)
