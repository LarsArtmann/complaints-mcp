# Integration Plan: go-composable-business-types/id Library

## Executive Summary

This document outlines the integration of `github.com/larsartmann/go-composable-business-types/id` into the complaints-mcp project to replace custom phantom type ID implementations with a standardized, feature-rich library approach.

## Current State Analysis

### Existing ID Types (4 implementations, ~400 lines)

| Type        | File          | Pattern                   | Validation                          | Lines |
| ----------- | ------------- | ------------------------- | ----------------------------------- | ----- |
| ComplaintID | complaint.go  | `type ComplaintID string` | UUID v4 regex                       | ~90   |
| AgentID     | agent_id.go   | `type AgentID string`     | Length + Unicode                    | ~110  |
| SessionID   | session_id.go | `type SessionID string`   | Alphanumeric + dash/underscore      | ~110  |
| ProjectID   | project_id.go | `type ProjectID string`   | Alphanumeric + dots/dash/underscore | ~110  |

### Common Functionality (Duplicated)

All 4 types implement identical patterns:

- Constructor: `NewXID() (XID, error)`
- Parser: `ParseXID(s string) (XID, error)`
- Must parser: `MustParseXID(s string) XID`
- Validation: `Validate() error`
- Check: `IsValid() bool`, `IsEmpty() bool`
- Stringer: `String() string`
- JSON: `MarshalJSON() ([]byte, error)`, `UnmarshalJSON([]byte) error`

### Issues with Current Approach

1. **Code Duplication**: Same patterns repeated 4 times
2. **Missing Features**: No SQL scanning, binary encoding, comparison
3. **Maintenance Burden**: Changes require updates in 4 places
4. **Inconsistent APIs**: Slight variations in method signatures

## Target State

### Using go-composable-business-types/id

```go
package domain

import "github.com/larsartmann/go-composable-business-types/id"

// Brand types (empty structs for type distinctness)
type ComplaintBrand struct{}
type AgentBrand struct{}
type SessionBrand struct{}
type ProjectBrand struct{}

// Type aliases for convenience
type ComplaintID = id.ID[ComplaintBrand, string]
type AgentID = id.ID[AgentBrand, string]
type SessionID = id.ID[SessionBrand, string]
type ProjectID = id.ID[ProjectBrand, string]

// Constructors with validation
type IDConstructor func(string) (id.ID[B, string], error)
```

### Benefits

| Benefit                 | Description                                            |
| ----------------------- | ------------------------------------------------------ |
| **Type Safety**         | Compile-time prevention of mixing ID types             |
| **Less Code**           | ~400 lines → ~80 lines (80% reduction)                 |
| **More Features**       | SQL, Binary, Gob, Text, Compare, Or, IsZero            |
| **Standardized**        | Consistent with go-composable-business-types ecosystem |
| **Zero Value Handling** | Proper null serialization in JSON                      |
| **Tested**              | Library has comprehensive test coverage                |

## Integration Strategy

### Phase 1: Dependency Addition

1. Add module dependency
2. Run `go mod tidy`
3. Verify no conflicts

### Phase 2: Type Refactoring

For each ID type, replace the implementation:

#### ComplaintID (UUID v4)

```go
// Before: custom type with regex validation
// After: branded ID with validation wrapper

type ComplaintBrand struct{}
type ComplaintID = id.ID[ComplaintBrand, string]

func NewComplaintID() (ComplaintID, error) {
    uuid, err := uuid.NewV4()
    if err != nil {
        return ComplaintID{}, fmt.Errorf("failed to generate ComplaintID: %w", err)
    }
    return id.NewID[ComplaintBrand](uuid.String()), nil
}

func ParseComplaintID(s string) (ComplaintID, error) {
    if err := validateComplaintID(s); err != nil {
        return ComplaintID{}, fmt.Errorf("invalid ComplaintID: %w", err)
    }
    return id.NewID[ComplaintBrand](s), nil
}
```

#### AgentID, SessionID, ProjectID (Named IDs)

```go
type AgentBrand struct{}
type AgentID = id.ID[AgentBrand, string]

func NewAgentID(name string) (AgentID, error) {
    trimmed := strings.TrimSpace(name)
    if err := validateAgentID(trimmed); err != nil {
        return AgentID{}, fmt.Errorf("invalid AgentID: %w", err)
    }
    return id.NewID[AgentBrand](trimmed), nil
}
```

### Phase 3: API Compatibility Layer

Maintain backward compatibility by keeping:

- Constructor functions: `NewXID()`, `ParseXID()`, `MustParseXID()`
- Validation functions: `ValidateXID()`
- Custom validation logic (UUID regex, character patterns)

Removed (replaced by library):

- `MarshalJSON()` / `UnmarshalJSON()` → library handles this
- `String()` → library `String()` or `Get()`
- `IsEmpty()` → library `IsZero()`
- `IsValid()` → use `Validate()` error check

### Phase 4: Consumer Updates

Update call sites:

```go
// Before
if id.IsEmpty() { ... }

// After
if id.IsZero() { ... }

// Before
id.String()

// After (same method exists)
id.String()

// Before
data, _ := id.MarshalJSON()

// After (automatic via library)
data, _ := json.Marshal(id)
```

## Implementation Steps

### Step 1: Add Dependency

```bash
cd /Users/larsartmann/projects/complaints-mcp
go get github.com/larsartmann/go-composable-business-types/id
go mod tidy
```

### Step 2: Refactor domain/complaint.go

- Define `ComplaintBrand` struct
- Create type alias `ComplaintID = id.ID[ComplaintBrand, string]`
- Update `NewComplaintID()` to return `id.ID[ComplaintBrand, string]`
- Update `ParseComplaintID()` with validation
- Remove manual JSON marshaling
- Keep validation regex

### Step 3: Refactor domain/agent_id.go

- Define `AgentBrand` struct
- Create type alias `AgentID = id.ID[AgentBrand, string]`
- Update constructors and parsers
- Remove duplicated methods

### Step 4: Refactor domain/session_id.go

- Define `SessionBrand` struct
- Create type alias `SessionID = id.ID[SessionBrand, string]`
- Update constructors and parsers

### Step 5: Refactor domain/project_id.go

- Define `ProjectBrand` struct
- Create type alias `ProjectID = id.ID[ProjectBrand, string]`
- Update constructors and parsers

### Step 6: Update Consumers

Search for and update:

- `.IsEmpty()` → `.IsZero()`
- Check JSON serialization still works
- Verify repository layer compatibility

### Step 7: Run Tests

```bash
just test
just test-bdd
```

## Risk Assessment

| Risk                       | Likelihood | Impact | Mitigation                   |
| -------------------------- | ---------- | ------ | ---------------------------- |
| JSON serialization changes | Low        | High   | Test backward compatibility  |
| SQL scanning breaks        | Low        | Medium | Add SQL tests if needed      |
| Type inference issues      | Medium     | Low    | Use explicit type parameters |
| Build failures             | Low        | High   | CI pipeline will catch       |

## Verification Checklist

- [ ] All tests pass
- [ ] JSON serialization produces same output
- [ ] Complaint creation works end-to-end
- [ ] Repository operations function correctly
- [ ] MCP tools return correct data

## Long-term Benefits

1. **Ecosystem Alignment**: Part of go-composable-business-types ecosystem
2. **Future Enhancements**: Library updates provide new features automatically
3. **Code Reduction**: 80% less ID-related code
4. **Type Safety**: Strong compile-time guarantees
5. **Developer Experience**: Less boilerplate, more focus on domain logic

## References

- Library: `/Users/larsartmann/projects/go-composable-business-types/id/`
- Current IDs: `/Users/larsartmann/projects/complaints-mcp/internal/domain/`
- Related: AGENTS.md "Strong types over runtime checks"
