# Step 3: Fix Critical JSON Nesting Bug

## 🎯 Objective

Fix the immediate JSON nesting bug that's breaking API responses.

## 🚨 Critical Issue

Current JSON output:

```json
{
  "id": {
    "Value": "9cb3bb9e-b6dc-4e02-9767-e396a42b63a6" // ❌ Nested!
  }
}
```

Required JSON output:

```json
{
  "id": "9cb3bb9e-b6dc-4e02-9767-e396a42b63a6" // ✅ Flat!
}
```

## 🏗️ Implementation Tasks

### A. Convert ComplaintID to Phantom Type

- **File**: `internal/domain/complaint_id.go`
- **Change**: `type ComplaintID struct` → `type ComplaintID string`
- **Methods**: New(), Parse(), String(), Validate(), UUID()
- **JSON**: Flat serialization (no nested Value field)

### B. Update Domain Integration

- **File**: `internal/domain/complaint.go`
- **Change**: Use phantom ComplaintID throughout
- **Methods**: Update all Complaint methods
- **Validation**: Add Validate() method to Complaint

### C. Update Repository Layer

- **File**: `internal/repo/file_repository.go`
- **Change**: JSON serialization uses phantom types
- **File Names**: Update file naming strategy
- **Paths**: Update GetFilePath/GetDocsPath methods

### D. Update Service Layer

- **File**: `internal/service/complaint_service.go`
- **Change**: CreateComplaint uses phantom types
- **Methods**: Update all service methods
- **Validation**: Add comprehensive validation

### E. Update MCP Handlers

- **File**: `internal/delivery/mcp/mcp_server.go`
- **Change**: Tool schemas expect flat strings
- **Handlers**: Update input parsing and output
- **DTOs**: Update ToDTO conversion

### F. Update Tests

- **Files**: All test files in affected packages
- **Change**: Update test expectations for flat JSON
- **Fixtures**: Update test data fixtures
- **Coverage**: Maintain or improve test coverage

## 📝 Implementation Details

### Phantom Type Implementation

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
    if !isValidUUID(s) {
        return ComplaintID(""), fmt.Errorf("invalid ComplaintID format: %s", s)
    }
    return ComplaintID(s), nil
}

func (id ComplaintID) String() string { return string(id) }
func (id ComplaintID) MarshalJSON() ([]byte, error) { return json.Marshal(id.String()) }
func (id ComplaintID) Validate() error { return isValidUUID(id.String()) }
```

### Updated Complaint Struct

```go
// internal/domain/complaint.go
type Complaint struct {
    ID              ComplaintID     `json:"id"`              // ✅ Flat string
    AgentID         string          `json:"agent_name"`        // Keep for API compatibility
    SessionID       string          `json:"session_name"`      // Keep for API compatibility
    ProjectName     string          `json:"project_name"`      // Keep for API compatibility
    TaskDescription string          `json:"task_description"`
    ContextInfo     string          `json:"context_info"`
    MissingInfo     string          `json:"missing_info"`
    ConfusedBy      string          `json:"confused_by"`
    FutureWishes    string          `json:"future_wishes"`
    Severity        Severity        `json:"severity"`
    Timestamp       time.Time       `json:"timestamp"`
    ResolutionState ResolutionState `json:"resolution_state"`
    ResolvedAt      *time.Time     `json:"resolved_at,omitempty"`
    ResolvedBy      string          `json:"resolved_by,omitempty"`
}
```

### Updated Repository Methods

```go
func (r *FileRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
    // JSON now serializes flat ID structure
    data, err := json.Marshal(complaint)
    if err != nil {
        return fmt.Errorf("failed to marshal complaint: %w", err)
    }

    // File naming uses flat ID
    fileName := fmt.Sprintf("%s.json", complaint.ID.String())
    return r.writeFile(ctx, fileName, data)
}
```

## 🧪 Verification Steps

### 1. Unit Tests

```bash
go test ./internal/domain -v -run TestComplaintID
```

### 2. JSON Serialization Test

```go
func TestComplaint_MarshalJSON_FlatID(t *testing.T) {
    complaint := &domain.Complaint{
        ID: mustNewComplaintID(),
        // ... other fields
    }

    data, err := json.Marshal(complaint)
    require.NoError(t, err)

    // Verify flat structure
    var result map[string]any
    err = json.Unmarshal(data, &result)
    require.NoError(t, err)

    idValue := result["id"]
    idStr, isString := idValue.(string)
    assert.True(t, isString, "ID should be flat string")
    assert.Contains(t, idStr, "-")  // UUID format
    assert.NotContains(t, string(data), `"Value"`)  // No nesting
}
```

### 3. Integration Test

```bash
# Start server and test tool response
echo '{"tool":"file_complaint","arguments":{"agent_name":"Test-Agent","task_description":"Test task","severity":"low"}}' | ./complaints-mcp
```

Expected response with flat ID:

```json
{
  "success": true,
  "complaint": {
    "id": "9cb3bb9e-b6dc-4e02-9767-e396a42b63a6" // ✅ Flat!
    // ... other fields
  }
}
```

## ⏱️ Time Estimate: 6-8 hours

## 🎯 Impact: CRITICAL (fixes broken API)

## 💪 Work Required: HIGH (core domain changes)

## 🚨 Prerequisite: None (critical bug fix)
