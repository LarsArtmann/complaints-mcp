# Issue #50: Implement Compile-Time Safe ID Validation with Phantom Type Constructors

## 🎯 **Enhancement: Centralized Validation with Type Safety**

### **Current State Analysis**

While implementing phantom types fixes JSON nesting, we need robust validation to ensure data integrity:

```go
// Current phantom type (no validation)
type ComplaintID string

func NewComplaintID() ComplaintID {
    return ComplaintID(uuid.New().String())  // ❌ No error handling
}

func ParseComplaintID(s string) ComplaintID {
    return ComplaintID(s)  // ❌ No validation
}
```

### **Target State**

Implement comprehensive validation with proper error handling:

```go
// Target implementation (validation + type safety)
type ComplaintID string

func NewComplaintID() (ComplaintID, error) {
    uuid := uuid4.New()
    return ComplaintID(uuid.String()), nil
}

func ParseComplaintID(s string) (ComplaintID, error) {
    if !isValidUUID(s) {
        return ComplaintID(""), fmt.Errorf("invalid ComplaintID format: %s", s)
    }
    return ComplaintID(s), nil
}

func (id ComplaintID) Validate() error {
    if id.IsEmpty() {
        return fmt.Errorf("ComplaintID cannot be empty")
    }

    if !isValidUUID(id.String()) {
        return fmt.Errorf("ComplaintID must be valid UUID format")
    }

    return nil
}
```

## 🛠️ **Comprehensive Validation Implementation**

### **Phase 1: UUID Validation Utilities**

```go
// internal/validation/uuid.go
package validation

import (
    "github.com/google/uuid"
    "regexp"
)

var (
    // UUID v4 regex pattern
    uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// IsValidUUID checks if string is valid UUID v4 format
func IsValidUUID(s string) bool {
    if s == "" {
        return false
    }
    return uuidPattern.MatchString(s)
}

// ValidateUUID validates and returns error
func ValidateUUID(s string, fieldName string) error {
    if s == "" {
        return fmt.Errorf("%s cannot be empty", fieldName)
    }

    if !IsValidUUID(s) {
        return fmt.Errorf("%s must be valid UUID v4 format, got: %s", fieldName, s)
    }

    return nil
}

// ParseUUID validates and returns parsed UUID
func ParseUUID(s string, fieldName string) (uuid.UUID, error) {
    if err := ValidateUUID(s, fieldName); err != nil {
        return uuid.Nil, err
    }

    parsed, err := uuid.Parse(s)
    if err != nil {
        return uuid.Nil, fmt.Errorf("%s contains invalid UUID: %w", fieldName, err)
    }

    return parsed, nil
}
```

### **Phase 2: ID Type Validation**

#### **ComplaintID Validation**

```go
// internal/domain/complaint_id.go
package domain

import (
    "github.com/larsartmann/complaints-mcp/internal/validation"
    "github.com/google/uuid"
)

type ComplaintID string

// NewComplaintID generates a new complaint ID with validation
func NewComplaintID() (ComplaintID, error) {
    id, err := uuid.NewRandom()
    if err != nil {
        return ComplaintID(""), fmt.Errorf("failed to generate ComplaintID: %w", err)
    }

    return ComplaintID(id.String()), nil
}

// ParseComplaintID parses string to ComplaintID with validation
func ParseComplaintID(s string) (ComplaintID, error) {
    if err := validation.ValidateUUID(s, "ComplaintID"); err != nil {
        return ComplaintID(""), err
    }

    return ComplaintID(s), nil
}

// MustParseComplaintID parses string to ComplaintID, panics on error
func MustParseComplaintID(s string) ComplaintID {
    id, err := ParseComplaintID(s)
    if err != nil {
        panic(fmt.Sprintf("invalid ComplaintID: %s", err))
    }
    return id
}

// Validate checks if ComplaintID is valid
func (id ComplaintID) Validate() error {
    if id.IsEmpty() {
        return fmt.Errorf("ComplaintID cannot be empty")
    }

    return validation.ValidateUUID(id.String(), "ComplaintID")
}

// IsEmpty checks if ComplaintID is empty
func (id ComplaintID) IsEmpty() bool {
    return string(id) == ""
}

// IsValid checks if ComplaintID is valid format
func (id ComplaintID) IsValid() bool {
    return id.Validate() == nil
}

// String returns string representation
func (id ComplaintID) String() string {
    return string(id)
}

// UUID returns parsed UUID value
func (id ComplaintID) UUID() (uuid.UUID, error) {
    return validation.ParseUUID(id.String(), "ComplaintID")
}
```

#### **AgentID Validation**

```go
// internal/domain/agent_id.go
package domain

import (
    "regexp"
    "strings"
)

var (
    // Agent name validation pattern
    agentNamePattern = regexp.MustCompile(`^[a-zA-Z0-9\-_\s]{1,100}$`)
)

type AgentID string

// NewAgentID creates a new agent ID with validation
func NewAgentID(name string) (AgentID, error) {
    name = strings.TrimSpace(name)

    if err := validateAgentName(name); err != nil {
        return AgentID(""), err
    }

    return AgentID(name), nil
}

// ParseAgentID parses string to AgentID with validation
func ParseAgentID(s string) (AgentID, error) {
    if err := validateAgentName(s); err != nil {
        return AgentID(""), err
    }

    return AgentID(s), nil
}

// Validate checks if AgentID is valid
func (id AgentID) Validate() error {
    return validateAgentName(id.String())
}

// IsEmpty checks if AgentID is empty
func (id AgentID) IsEmpty() bool {
    return strings.TrimSpace(id.String()) == ""
}

// IsValid checks if AgentID is valid format
func (id AgentID) IsValid() bool {
    return id.Validate() == nil
}

// String returns string representation
func (id AgentID) String() string {
    return string(id)
}

func validateAgentName(name string) error {
    trimmed := strings.TrimSpace(name)

    if trimmed == "" {
        return fmt.Errorf("agent name cannot be empty")
    }

    if len(trimmed) > 100 {
        return fmt.Errorf("agent name cannot exceed 100 characters, got %d", len(trimmed))
    }

    if !agentNamePattern.MatchString(trimmed) {
        return fmt.Errorf("agent name contains invalid characters: %s", name)
    }

    return nil
}
```

#### **SessionID Validation**

```go
// internal/domain/session_id.go
package domain

import (
    "regexp"
    "strings"
)

var (
    // Session name validation pattern
    sessionNamePattern = regexp.MustCompile(`^[a-zA-Z0-9\-_\s]{1,100}$`)
)

type SessionID string

// NewSessionID creates a new session ID with validation
func NewSessionID(name string) (SessionID, error) {
    name = strings.TrimSpace(name)

    if err := validateSessionName(name); err != nil {
        return SessionID(""), err
    }

    return SessionID(name), nil
}

// ParseSessionID parses string to SessionID with validation
func ParseSessionID(s string) (SessionID, error) {
    if err := validateSessionName(s); err != nil {
        return SessionID(""), err
    }

    return SessionID(s), nil
}

// Validate checks if SessionID is valid
func (id SessionID) Validate() error {
    return validateSessionName(id.String())
}

// IsEmpty checks if SessionID is empty
func (id SessionID) IsEmpty() bool {
    return strings.TrimSpace(id.String()) == ""
}

// IsValid checks if SessionID is valid format
func (id SessionID) IsValid() bool {
    return id.Validate() == nil
}

// String returns string representation
func (id SessionID) String() string {
    return string(id)
}

func validateSessionName(name string) error {
    trimmed := strings.TrimSpace(name)

    if trimmed == "" {
        return fmt.Errorf("session name cannot be empty")
    }

    if len(trimmed) > 100 {
        return fmt.Errorf("session name cannot exceed 100 characters, got %d", len(trimmed))
    }

    if !sessionNamePattern.MatchString(trimmed) {
        return fmt.Errorf("session name contains invalid characters: %s", name)
    }

    return nil
}
```

#### **ProjectID Validation**

```go
// internal/domain/project_id.go
package domain

import (
    "regexp"
    "strings"
)

var (
    // Project name validation pattern
    projectNamePattern = regexp.MustCompile(`^[a-zA-Z0-9\-_\s\.]{1,100}$`)
)

type ProjectID string

// NewProjectID creates a new project ID with validation
func NewProjectID(name string) (ProjectID, error) {
    name = strings.TrimSpace(name)

    if err := validateProjectName(name); err != nil {
        return ProjectID(""), err
    }

    return ProjectID(name), nil
}

// ParseProjectID parses string to ProjectID with validation
func ParseProjectID(s string) (ProjectID, error) {
    if err := validateProjectName(s); err != nil {
        return ProjectID(""), err
    }

    return ProjectID(s), nil
}

// Validate checks if ProjectID is valid
func (id ProjectID) Validate() error {
    return validateProjectName(id.String())
}

// IsEmpty checks if ProjectID is empty
func (id ProjectID) IsEmpty() bool {
    return strings.TrimSpace(id.String()) == ""
}

// IsValid checks if ProjectID is valid format
func (id ProjectID) IsValid() bool {
    return id.Validate() == nil
}

// String returns string representation
func (id ProjectID) String() string {
    return string(id)
}

func validateProjectName(name string) error {
    trimmed := strings.TrimSpace(name)

    if trimmed == "" {
        return fmt.Errorf("project name cannot be empty")
    }

    if len(trimmed) > 100 {
        return fmt.Errorf("project name cannot exceed 100 characters, got %d", len(trimmed))
    }

    if !projectNamePattern.MatchString(trimmed) {
        return fmt.Errorf("project name contains invalid characters: %s", name)
    }

    return nil
}
```

### **Phase 3: Centralized Validation Package**

```go
// internal/validation/validation.go
package validation

import (
    "fmt"
    "strings"
)

// Validator interface for validation logic
type Validator interface {
    Validate() error
    IsValid() bool
    IsEmpty() bool
}

// ValidateAll validates multiple validators
func ValidateAll(validators ...Validator) error {
    for i, v := range validators {
        if err := v.Validate(); err != nil {
            return fmt.Errorf("validation error at index %d: %w", i, err)
        }
    }
    return nil
}

// ValidateRequired checks if required field is not empty
func ValidateRequired(value, fieldName string) error {
    if strings.TrimSpace(value) == "" {
        return fmt.Errorf("%s is required", fieldName)
    }
    return nil
}

// ValidateMaxLength checks string length
func ValidateMaxLength(value string, maxLength int, fieldName string) error {
    if len(value) > maxLength {
        return fmt.Errorf("%s cannot exceed %d characters, got %d", fieldName, maxLength, len(value))
    }
    return nil
}

// ValidateMinLength checks string length
func ValidateMinLength(value string, minLength int, fieldName string) error {
    if len(value) < minLength {
        return fmt.Errorf("%s must be at least %d characters, got %d", fieldName, minLength, len(value))
    }
    return nil
}

// ValidatePattern checks string against regex pattern
func ValidatePattern(value, pattern, fieldName string) error {
    matched, err := regexp.MatchString(pattern, value)
    if err != nil {
        return fmt.Errorf("invalid pattern for %s validation: %w", fieldName, err)
    }

    if !matched {
        return fmt.Errorf("%s contains invalid characters", fieldName)
    }

    return nil
}
```

### **Phase 4: Validation in Domain Entities**

```go
// internal/domain/complaint.go
package domain

import (
    "time"
    "github.com/larsartmann/complaints-mcp/internal/validation"
)

type Complaint struct {
    ID              ComplaintID     `json:"id"`
    AgentID         AgentID         `json:"agent_id"`
    SessionID       SessionID       `json:"session_id"`
    ProjectID       ProjectID       `json:"project_id"`
    TaskDescription  string          `json:"task_description"`
    ContextInfo      string          `json:"context_info"`
    MissingInfo      string          `json:"missing_info"`
    ConfusedBy       string          `json:"confused_by"`
    FutureWishes    string          `json:"future_wishes"`
    Severity        Severity        `json:"severity"`
    Timestamp       time.Time       `json:"timestamp"`
    ResolutionState ResolutionState `json:"resolution_state"`
    ResolvedAt      *time.Time     `json:"resolved_at,omitempty"`
    ResolvedBy      string          `json:"resolved_by,omitempty"`
}

// Validate validates all complaint fields
func (c *Complaint) Validate() error {
    // Validate all ID fields
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

    // Validate string fields
    if err := validation.ValidateRequired(c.TaskDescription, "task description"); err != nil {
        return err
    }

    if err := validation.ValidateMaxLength(c.TaskDescription, 1000, "task description"); err != nil {
        return err
    }

    if err := validation.ValidateMaxLength(c.ContextInfo, 500, "context info"); err != nil {
        return err
    }

    if err := validation.ValidateMaxLength(c.MissingInfo, 500, "missing info"); err != nil {
        return err
    }

    if err := validation.ValidateMaxLength(c.ConfusedBy, 500, "confused by"); err != nil {
        return err
    }

    if err := validation.ValidateMaxLength(c.FutureWishes, 500, "future wishes"); err != nil {
        return err
    }

    // Validate severity
    if !c.Severity.IsValid() {
        return fmt.Errorf("invalid severity: %s", c.Severity)
    }

    // Validate timestamp
    if c.Timestamp.IsZero() {
        return fmt.Errorf("timestamp cannot be zero")
    }

    // Validate resolution state consistency
    if c.ResolutionState.IsResolved() != (c.ResolvedAt != nil) {
        return fmt.Errorf("resolution state and resolved_at timestamp are inconsistent")
    }

    return nil
}

// IsValid checks if complaint is valid
func (c *Complaint) IsValid() bool {
    return c.Validate() == nil
}
```

## 🎯 **Benefits of This Implementation**

### **1. Centralized Validation Logic**

```go
// ❌ Before: Validation scattered
type ComplaintID struct {
    Value string
}

// ✅ After: Centralized with type safety
type ComplaintID string
func (id ComplaintID) Validate() error { ... }
```

### **2. Clear Error Messages**

```go
// ❌ Before: Generic errors
return fmt.Errorf("invalid ID")

// ✅ After: Specific, helpful errors
return fmt.Errorf("ComplaintID must be valid UUID v4 format, got: %s", id)
```

### **3. Consistent Validation Patterns**

```go
// All ID types follow same pattern:
func New[T]ID(input string) (T, error)   // Constructor with validation
func Parse[T]ID(input string) (T, error)  // Parser with validation
func (id T) Validate() error             // Instance validation
func (id T) IsValid() bool               // Quick validity check
func (id T) IsEmpty() bool               // Empty check
```

### **4. Compile-Time Safety + Runtime Validation**

```go
// Type safety prevents mixing
agentID := domain.NewAgentID("AI-Assistant")
sessionID := domain.NewSessionID("dev-session")

// Compile-time error: cannot use AgentID as SessionID
processComplaint(agentID, sessionID) // ✅ OK
processComplaint(sessionID, agentID) // ❌ Compile error!
```

### **5. Better Testing Support**

```go
// Easy to test validation
func TestAgentID_Validation(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantErr  bool
        errorMsg string
    }{
        {"valid", "AI-Assistant", false, ""},
        {"empty", "", true, "agent name cannot be empty"},
        {"too long", strings.Repeat("a", 101), true, "cannot exceed 100 characters"},
        {"invalid chars", "AI@Assistant", true, "contains invalid characters"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := domain.NewAgentID(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 📋 **Files to Create/Modify**

### **New Files**

- `internal/validation/uuid.go` - UUID validation utilities
- `internal/validation/validation.go` - General validation utilities
- `internal/domain/complaint_id.go` - ComplaintID with validation
- `internal/domain/agent_id.go` - AgentID with validation
- `internal/domain/session_id.go` - SessionID with validation
- `internal/domain/project_id.go` - ProjectID with validation

### **Modified Files**

- `internal/domain/complaint.go` - Update to use new ID types
- `internal/delivery/mcp/mcp_server.go` - Update handlers with validation
- `internal/service/complaint_service.go` - Update service validation
- `features/bdd/*.go` - Update BDD tests with validation

## 🧪 **Testing Strategy**

### **Unit Tests for Each ID Type**

```go
func TestComplaintID_New(t *testing.T) {
    tests := []struct {
        name    string
        wantErr bool
    }{
        {"success", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            id, err := NewComplaintID()
            if tt.wantErr {
                assert.Error(t, err)
                assert.True(t, id.IsEmpty())
            } else {
                assert.NoError(t, err)
                assert.False(t, id.IsEmpty())
                assert.True(t, id.IsValid())
            }
        })
    }
}

func TestComplaintID_Parse(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantErr  bool
        errorMsg  string
    }{
        {"valid uuid", "550e8400-e29b-41d4-a716-446655440000", false, ""},
        {"empty", "", true, "cannot be empty"},
        {"invalid format", "not-a-uuid", true, "must be valid UUID v4 format"},
        {"invalid version", "550e8400-e29b-41d4-a716-446655440123", true, "must be valid UUID v4 format"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            id, err := ParseComplaintID(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
                assert.True(t, id.IsEmpty())
            } else {
                assert.NoError(t, err)
                assert.False(t, id.IsEmpty())
                assert.True(t, id.IsValid())
                assert.Equal(t, tt.input, id.String())
            }
        })
    }
}
```

### **Integration Tests**

```go
func TestComplaint_Validation(t *testing.T) {
    tests := []struct {
        name        string
        complaint   *Complaint
        wantErr     bool
        errorMsg    string
    }{
        {
            name: "valid complaint",
            complaint: &Complaint{
                ID:             mustNewComplaintID(),
                AgentID:        mustNewAgentID("AI-Assistant"),
                SessionID:      mustNewSessionID("dev-session"),
                ProjectID:      mustNewProjectID("my-project"),
                TaskDescription: "Test task",
                Severity:       SeverityLow,
                Timestamp:      time.Now(),
            },
            wantErr: false,
        },
        {
            name: "invalid task description",
            complaint: &Complaint{
                ID:             mustNewComplaintID(),
                AgentID:        mustNewAgentID("AI-Assistant"),
                SessionID:      mustNewSessionID("dev-session"),
                ProjectID:      mustNewProjectID("my-project"),
                TaskDescription: "", // Invalid
                Severity:       SeverityLow,
                Timestamp:      time.Now(),
            },
            wantErr: true,
            errorMsg: "task description is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.complaint.Validate()
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## ⚠️ **Breaking Changes**

### **Constructor Changes**

- **New Functions**: Now return `(ID, error)` instead of just `ID`
- **Parse Functions**: Added validation with error handling
- **Service Methods**: Need to handle validation errors

### **Migration Impact**

- **Error Handling**: Call sites need to handle validation errors
- **Test Updates**: Existing tests need to account for validation
- **Documentation**: Update examples to show validation usage

### **Mitigation Strategy**

- **Gradual Migration**: Add validation alongside existing code
- **Backward Compatibility**: Provide both validated and unvalidated constructors
- **Clear Migration Guide**: Step-by-step upgrade instructions

## 🏆 **Success Criteria**

- [ ] All ID types have comprehensive validation
- [ ] Constructor functions properly validate and return errors
- [ ] Validation provides clear, actionable error messages
- [ ] All existing functionality preserved
- [ ] Test suite covers all validation scenarios
- [ ] Performance impact is minimal
- [ ] Migration path is clear and documented

## 🏷️ **Labels**

- `validation` - Data validation implementation
- `type-safety` - Compile-time type safety improvements
- `enhancement` - New feature addition
- `medium-priority` - Important for data integrity
- `breaking-change` - Changes constructor signatures

## 📊 **Priority**: Medium

- **Complexity**: Medium (validation logic + error handling)
- **Value**: High (data integrity + error prevention)
- **Risk**: Low (additive, doesn't break existing)
- **Dependencies**: Issue #48 (phantom type foundation)

## 🤝 **Dependencies**

- **Issue #48**: Must have phantom types implemented first
- **Issue #49**: Should have all ID fields converted
- **Issue #52**: Need tests for validation

---

**This enhancement provides comprehensive validation while maintaining type safety, ensuring data integrity throughout the system with clear error messages and robust testing.**
