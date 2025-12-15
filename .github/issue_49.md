# Issue #49: Replace All String ID Fields with Strongly-Typed Phantom IDs

## 🎯 **Enhancement: Comprehensive Type Safety Improvements**

### **Current State Analysis**

The codebase currently uses string types for many identifier fields, creating potential for type confusion:

```go
// Current implementation (type confusion possible)
type Complaint struct {
    ID          string        // Generic string
    AgentName    string        // Generic string
    SessionName  string        // Generic string
    ProjectName  string        // Generic string
}

// Problem: Can accidentally mix these
func Process(agentName, sessionName string) error {
    // No compile-time protection against mixing IDs
}
```

### **Target State**

Replace string-typed identifiers with strongly-typed phantom types:

```go
// Target implementation (type-safe)
type Complaint struct {
    ID          ComplaintID   // Compile-time safe
    AgentID     AgentID       // Compile-time safe
    SessionID   SessionID     // Compile-time safe
    ProjectID   ProjectID     // Compile-time safe
}

// Problem: Type system prevents mixing
func Process(agentID AgentID, sessionID SessionID) error {
    // Compile-time error if wrong types passed!
}
```

## 🛠️ **Implementation Plan**

### **Phase 1: Phantom Type Definitions**

```go
// Phantom types with compile-time safety
type (
    ComplaintID string  // Primary complaint identifier
    SessionID   string  // Session identifier
    ProjectID   string  // Project identifier
    AgentID     string  // Agent identifier
)

// Constructors for validation
func NewAgentID(name string) (AgentID, error) {
    if !isValidAgentName(name) {
        return AgentID(""), fmt.Errorf("invalid agent name")
    }
    return AgentID(name), nil
}

func NewSessionID(name string) (SessionID, error) {
    if !isValidSessionName(name) {
        return SessionID(""), fmt.Errorf("invalid session name")
    }
    return SessionID(name), nil
}

func NewProjectID(name string) (ProjectID, error) {
    if !isValidProjectName(name) {
        return ProjectID(""), fmt.Errorf("invalid project name")
    }
    return ProjectID(name), nil
}

// Methods
func (id AgentID) String() string     { return string(id) }
func (id AgentID) IsValid() bool      { return isValidAgentName(string(id)) }
func (id AgentID) IsEmpty() bool      { return string(id) == "" }

func (id SessionID) String() string   { return string(id) }
func (id SessionID) IsValid() bool    { return isValidSessionName(string(id)) }
func (id SessionID) IsEmpty() bool    { return string(id) == "" }

func (id ProjectID) String() string   { return string(id) }
func (id ProjectID) IsValid() bool    { return isValidProjectName(string(id)) }
func (id ProjectID) IsEmpty() bool    { return string(id) == "" }
```

### **Phase 2: Domain Entity Updates**

```go
// Updated Complaint struct
type Complaint struct {
    ID          ComplaintID    `json:"id"`
    AgentID     AgentID        `json:"agent_id"`
    SessionID   SessionID      `json:"session_id"`
    ProjectID   ProjectID      `json:"project_id"`
    TaskDescription string       `json:"task_description"`
    ContextInfo     string       `json:"context_info"`
    MissingInfo     string       `json:"missing_info"`
    ConfusedBy      string       `json:"confused_by"`
    FutureWishes    string       `json:"future_wishes"`
    Severity        Severity      `json:"severity"`
    Timestamp       time.Time     `json:"timestamp"`
    ResolutionState ResolutionState `json:"resolution_state"`
    ResolvedAt      *time.Time     `json:"resolved_at,omitempty"`
    ResolvedBy      string         `json:"resolved_by,omitempty"`
}
```

### **Phase 3: Repository Interface Updates**

```go
// Updated repository interface
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
```

### **Phase 4: Service Method Updates**

```go
// Updated service methods
func (s *ComplaintService) CreateComplaint(
    ctx context.Context,
    agentID domain.AgentID,
    sessionID domain.SessionID,
    taskDescription, contextInfo, missingInfo, confusedBy, futureWishes string,
    severity domain.Severity,
    projectID domain.ProjectID,
) (*domain.Complaint, error)

func (s *ComplaintService) GetComplaintsByAgent(
    ctx context.Context,
    agentID domain.AgentID,
    limit int,
) ([]*domain.Complaint, error)

func (s *ComplaintService) GetComplaintsByProject(
    ctx context.Context,
    projectID domain.ProjectID,
    limit int,
) ([]*domain.Complaint, error)
```

### **Phase 5: MCP Handler Updates**

```go
// Updated handler signatures
func (m *MCPServer) handleFileComplaint(ctx context.Context, req *mcp.CallToolRequest, input FileComplaintInput) (*mcp.CallToolResult, FileComplaintOutput, error) {
    // Parse string inputs to typed IDs
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

    // Call service with typed IDs
    complaint, err := m.service.CreateComplaint(
        ctx, agentID, sessionID, input.TaskDescription,
        input.ContextInfo, input.MissingInfo, input.ConfusedBy,
        input.FutureWishes, domainSeverity, projectID,
    )
    // ...
}
```

### **Phase 6: DTO Updates**

```go
// Updated DTO structure
type ComplaintDTO struct {
    ID              string     `json:"id"`
    AgentName       string     `json:"agent_name"`        // Keep for API compatibility
    SessionName     string     `json:"session_name"`      // Keep for API compatibility
    ProjectName     string     `json:"project_name"`      // Keep for API compatibility
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

// Updated conversion with type safety
func ToDTO(complaint *domain.Complaint) ComplaintDTO {
    return ComplaintDTO{
        ID:              complaint.ID.String(),
        AgentName:       complaint.AgentID.String(),       // Convert typed ID to string
        SessionName:     complaint.SessionID.String(),     // Convert typed ID to string
        ProjectName:     complaint.ProjectID.String(),     // Convert typed ID to string
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
```

## 🎯 **Benefits of This Change**

### **1. Compile-Time Type Safety**

```go
// ❌ Before: Possible to mix IDs
func Process(agentName, sessionName string) error {
    // No protection against: Process(sessionName, agentName)
}

// ✅ After: Compile-time protection
func Process(agentID AgentID, sessionID SessionID) error {
    // Process(sessionID, agentID) // ❌ Compile-time error!
}
```

### **2. Better IDE Support**

- **Smart Completion**: Only appropriate IDs suggested
- **Refactoring**: Rename IDs safely across codebase
- **Navigation**: Jump to ID definition easily
- **Type Hints**: Clear what each ID represents

### **3. Reduced Bugs**

- **Prevents ID Mixups**: Type system catches mistakes
- **Clearer Intent**: Code communicates ID purpose
- **Validation**: IDs can include validation logic
- **Self-Documenting**: Type names describe content

### **4. Future Extensibility**

- **Type-Specific Methods**: Add behavior per ID type
- **Validation**: Centralized validation per ID type
- **Serialization**: Custom JSON handling per ID type
- **Testing**: Type-specific test utilities

## 📋 **Files to Modify**

### **Domain Layer**

- `internal/domain/complaint.go` - Update struct and methods
- `internal/domain/agent_name.go` - Convert to AgentID
- `internal/domain/session_name.go` - Convert to SessionID
- `internal/domain/project_name.go` - Convert to ProjectID

### **Repository Layer**

- `internal/repo/repository.go` - Update interface
- `internal/repo/file_repository.go` - Update implementations
- `internal/repo/cached_repository.go` - Update implementations

### **Service Layer**

- `internal/service/complaint_service.go` - Update method signatures
- `internal/service/complaint_service_test.go` - Update tests

### **Delivery Layer**

- `internal/delivery/mcp/mcp_server.go` - Update handlers
- `internal/delivery/mcp/dto.go` - Update DTOs and conversions

### **Test Layer**

- `features/bdd/*.go` - Update BDD tests
- `internal/domain/*_test.go` - Update unit tests

## 🔄 **Migration Strategy**

### **Step 1: Backward Compatibility**

- Keep existing string fields temporarily
- Add new typed ID fields alongside
- Implement both string and ID methods

### **Step 2: Gradual Migration**

- Update domain entities first
- Move to repository layer
- Update service methods
- Modify delivery handlers

### **Step 3: Cleanup**

- Remove old string fields
- Update all method signatures
- Clean up test fixtures
- Update documentation

## 🧪 **Testing Strategy**

### **Unit Tests**

```go
func TestAgentID_New(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantErr  bool
    }{
        {"valid agent", "AI-Assistant", false},
        {"invalid empty", "", true},
        {"invalid too long", strings.Repeat("a", 101), true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewAgentID(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewAgentID() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && got.String() != tt.input {
                t.Errorf("NewAgentID() = %v, want %v", got.String(), tt.input)
            }
        })
    }
}
```

### **Integration Tests**

```go
func TestComplaint_WithTypeSafeIDs(t *testing.T) {
    agentID, _ := NewAgentID("Test-Agent")
    sessionID, _ := NewSessionID("test-session")
    projectID, _ := NewProjectID("test-project")

    complaint := &Complaint{
        ID:        mustNewComplaintID(),
        AgentID:   agentID,
        SessionID: sessionID,
        ProjectID: projectID,
        // ... other fields
    }

    // Test that type system works correctly
    assert.Equal(t, "Test-Agent", complaint.AgentID.String())
    assert.Equal(t, "test-session", complaint.SessionID.String())
    assert.Equal(t, "test-project", complaint.ProjectID.String())
}
```

### **BDD Tests**

```gherkin
Feature: Type-Safe Complaint Management
  As an AI assistant
  I want to use type-safe identifiers
  So that ID confusion errors are prevented

  Scenario: Create complaint with type-safe IDs
    Given I have valid agent, session, and project IDs
    When I create a complaint with these typed IDs
    Then the complaint should store the IDs with type safety
    And the system should prevent ID mixing
```

## ⚠️ **Breaking Changes**

### **API Changes**

- **Method Signatures**: Service methods now accept typed IDs
- **Constructor Arguments**: Complaint creation requires typed IDs
- **Test Updates**: Existing tests need type conversions

### **Migration Impact**

- **External Consumers**: Need to handle typed ID conversion
- **Configuration**: May need ID parsing changes
- **Integration Points**: All ID boundaries affected

### **Mitigation Strategy**

- **Backward Compatibility**: Provide string conversion methods
- **Gradual Migration**: Support both types during transition
- **Clear Documentation**: Migration guide for consumers
- **Tool Support**: Add CLI helpers for ID conversion

## 🏆 **Success Criteria**

- [ ] All ID fields use phantom types
- [ ] Compile-time type safety enforced
- [ ] No performance regression
- [ ] All tests pass with new types
- [ ] API consumers can migrate successfully
- [ ] Documentation updated for new types
- [ ] Migration guide provided

## 🏷️ **Labels**

- `refactoring` - Large-scale code restructuring
- `type-safety` - Compile-time type safety improvements
- `architecture` - Core architectural enhancement
- `high-priority` - Important for maintainability
- `breaking-change` - Changes API contract

## 📊 **Priority**: High

- **Complexity**: High (affects entire codebase)
- **Value**: High (long-term maintainability)
- **Risk**: Medium (requires careful migration)
- **Dependencies**: Issue #48 (phantom type foundation)

## 🤝 **Dependencies**

- **Issue #48**: Must implement base phantom types first
- **Issue #50**: Validation constructors needed
- **Issue #51**: Schema updates for flat structure

---

**This enhancement represents a major architectural improvement that will significantly improve code quality, reduce bugs, and enhance long-term maintainability through compile-time type safety.**
