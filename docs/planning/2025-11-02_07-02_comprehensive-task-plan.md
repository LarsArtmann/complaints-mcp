# Comprehensive Task Plan - Phase 2

**Date**: 2025-11-02 07:02
**Total Tasks**: 20 (100-30min each)
**Total Effort**: ~22 hours
**Target Value**: 80% of total possible value

---

## 📊 TASK TABLE (sorted by importance/impact/effort/customer-value)

| #      | Task Name                                      | Effort | Impact   | Customer Value | Priority | Category |
| ------ | ---------------------------------------------- | ------ | -------- | -------------- | -------- | -------- |
| **1**  | Fix domain test failures                       | 30min  | CRITICAL | BLOCKING       | P0       | 1%       |
| **2**  | Fix service test failures                      | 45min  | CRITICAL | BLOCKING       | P0       | 1%       |
| **3**  | Fix repo test failures                         | 45min  | CRITICAL | BLOCKING       | P0       | 1%       |
| **4**  | Add ResolvedBy string field to Complaint       | 30min  | HIGH     | HIGH           | P0       | 1%       |
| **5**  | Create ComplaintService interface              | 45min  | HIGH     | MEDIUM         | P1       | 4%       |
| **6**  | Fix repository Update() bug (duplicate files)  | 60min  | CRITICAL | HIGH           | P1       | 4%       |
| **7**  | Create typed DTO structs for MCP responses     | 90min  | HIGH     | MEDIUM         | P1       | 4%       |
| **8**  | Replace map[string]interface{} in MCP handlers | 30min  | HIGH     | MEDIUM         | P1       | 4%       |
| **9**  | Design repository cache interface              | 30min  | HIGH     | HIGH           | P1       | 4%       |
| **10** | Implement in-memory cache with sync.RWMutex    | 60min  | HIGH     | HIGH           | P1       | 4%       |
| **11** | Create AgentName value object                  | 60min  | MEDIUM   | LOW            | P2       | 20%      |
| **12** | Create ProjectName value object                | 45min  | MEDIUM   | LOW            | P2       | 20%      |
| **13** | Create SessionName value object                | 45min  | MEDIUM   | LOW            | P2       | 20%      |
| **14** | Update domain to use value objects             | 60min  | MEDIUM   | LOW            | P2       | 20%      |
| **15** | Create internal/delivery/dto package           | 45min  | MEDIUM   | LOW            | P2       | 20%      |
| **16** | Move DTOs to dto package                       | 45min  | MEDIUM   | LOW            | P2       | 20%      |
| **17** | Strengthen Severity enum with iota             | 60min  | MEDIUM   | MEDIUM         | P2       | 20%      |
| **18** | Use custom error types throughout codebase     | 90min  | MEDIUM   | LOW            | P2       | 20%      |
| **19** | Add service layer tests (full coverage)        | 90min  | MEDIUM   | MEDIUM         | P2       | 20%      |
| **20** | Fix 7 remaining BDD test failures              | 90min  | MEDIUM   | MEDIUM         | P2       | 20%      |

---

## 📈 BREAKDOWN BY CATEGORY

### 🎯 THE 1% (Tasks 1-4) - 2.5 hours → 51% value

**Critical Path**: Must complete before anything else

- Fix all test failures (2h)
- Add ResolvedBy field (30min)

**Why First**: Blocking deployment, unblocks all other work

### 🎯 THE 4% (Tasks 5-10) - 5 hours → 64% cumulative value

**Production Ready**: Type safety + performance

- Interface pattern (45min)
- Update bug fix (1h)
- DTOs (2h)
- Caching (1.5h)

**Why Second**: Makes system production-ready with good performance

### 🎯 THE 20% (Tasks 11-20) - 14.5 hours → 80% cumulative value

**Architecture Polish**: Clean code + maintainability

- Value objects (3.5h)
- DTO extraction (1.5h)
- Severity enum (1h)
- Error types (1.5h)
- Tests (3h)

**Why Third**: Long-term maintainability and code quality

---

## 🔥 DETAILED TASK DESCRIPTIONS

### P0 - CRITICAL (THE 1%)

#### Task 1: Fix domain test failures (30min)

**File**: `internal/domain/complaint_test.go`
**Problem**: Already fixed during Phase 1
**Action**: Verify all tests pass
**Acceptance**: `go test ./internal/domain -v` → 100% pass
**Blocks**: Nothing (already done)

#### Task 2: Fix service test failures (45min)

**File**: `internal/service/complaint_service_test.go`
**Problem**: Missing context parameters, outdated mocks
**Actions**:

- Add `ctx context.Context` to all test calls
- Update mock repository creation
- Fix complaint.Resolve(ctx) calls
- Update assertions for ResolvedAt field
  **Acceptance**: `go test ./internal/service -v` → 100% pass
  **Blocks**: Service layer changes

#### Task 3: Fix repo test failures (45min)

**File**: `internal/repo/file_repository_test.go`
**Problem**: Missing tracer parameter, wrong return type assertions
**Actions**:

- Add tracer to NewFileRepository calls
- Fix type assertions (need .() or change expectations)
- Update filename generation expectations (UUID-based)
- Fix method calls (generateFilename removed)
  **Acceptance**: `go test ./internal/repo -v` → 100% pass
  **Blocks**: Repository changes

#### Task 4: Add ResolvedBy field (30min)

**Files**:

- `internal/domain/complaint.go`
- `internal/domain/complaint_test.go`
- `internal/service/complaint_service.go`
  **Actions**:
- Add `ResolvedBy string` field to Complaint struct
- Update Resolve() to accept resolvedBy parameter
- Update all callers in service layer
- Add test for ResolvedBy field
  **Acceptance**:
- Field exists and is set on resolution
- Tests verify ResolvedBy is populated
  **Customer Value**: Audit trail - know WHO resolved complaint
  **Blocks**: Nothing

---

### P1 - HIGH (THE 4%)

#### Task 5: Create ComplaintService interface (45min)

**File**: `internal/service/interface.go` (new)
**Actions**:

- Define Service interface with all methods
- Extract interface from ComplaintService
- Update main.go to use interface type
- Update tests to use interface
  **Benefits**:
- Easier mocking
- Dependency injection
- Future implementations (e.g., caching service)
  **Acceptance**: Code compiles, tests pass
  **Blocks**: Service layer refactoring

#### Task 6: Fix repository Update() bug (60min)

**File**: `internal/repo/file_repository.go`
**Problem**: Update() calls Save() which creates new file with new timestamp
**Current Behavior**:

```go
func Update(complaint) {
    existing := FindByID()
    existing.Resolved = complaint.Resolved
    Save(existing) // ❌ Creates new file!
}
```

**Solution**:

- Track original filename in metadata OR
- Delete old file before Save OR
- Use UUID-only filenames (no timestamp)
  **Actions**:
- Implement proper update logic
- Add test for update (verify no duplicate files)
- Test concurrent updates
  **Acceptance**: Update modifies existing file, no duplicates
  **Customer Value**: Data integrity
  **Blocks**: Nothing

#### Task 7: Create typed DTO structs (90min)

**File**: `internal/delivery/mcp/dto.go` (new)
**Actions**:

- Create ComplaintDTO struct
- Create proper output types for all tools
- Add ToDTO() methods on domain.Complaint
- Add validation on DTOs
  **Structures Needed**:

```go
type ComplaintDTO struct {
    ID              string
    AgentName       string
    SessionName     string
    TaskDescription string
    Severity        string
    Timestamp       string
    Resolved        bool
    ResolvedAt      *string
    ResolvedBy      string
    ProjectName     string
}
```

**Acceptance**: All DTOs defined, documented
**Blocks**: Task 8

#### Task 8: Replace map[string]interface{} (30min)

**File**: `internal/delivery/mcp/mcp_server.go`
**Actions**:

- Update ListComplaintsOutput to use []ComplaintDTO
- Update SearchComplaintsOutput to use []ComplaintDTO
- Remove map construction in handlers
- Use ToDTO() conversion
  **Acceptance**: No map[string]interface{} in outputs
  **Customer Value**: Type safety, better docs
  **Blocks**: Nothing

#### Task 9: Design repository cache interface (30min)

**File**: `internal/repo/cache.go` (new)
**Actions**:

- Design CachedRepository interface
- Plan cache invalidation strategy
- Document concurrency approach
- Choose data structures (map + RWMutex)
  **Deliverable**: Interface definition + documentation
  **Blocks**: Task 10

#### Task 10: Implement in-memory cache (60min)

**File**: `internal/repo/cached_repository.go` (new)
**Actions**:

- Implement CachedRepository wrapper
- Add sync.RWMutex for concurrency
- Cache FindByID results
- Invalidate on Save/Update
- Add cache hit/miss logging
  **Performance Target**: FindByID from O(n) to O(1)
  **Acceptance**: Benchmarks show 10-100x improvement
  **Customer Value**: Fast queries
  **Blocks**: Nothing

---

### P2 - MEDIUM (THE 20%)

#### Task 11: Create AgentName value object (60min)

**File**: `internal/domain/value/agent_name.go` (new)
**Actions**:

- Create AgentName struct
- Add NewAgentName(string) (AgentName, error) constructor
- Validate: non-empty, 1-100 chars
- Add String(), MarshalJSON, UnmarshalJSON
- Add tests
  **Acceptance**: All validation tests pass
  **Blocks**: Task 14

#### Task 12: Create ProjectName value object (45min)

**File**: `internal/domain/value/project_name.go` (new)
**Actions**: Same as Task 11, for ProjectName
**Blocks**: Task 14

#### Task 13: Create SessionName value object (45min)

**File**: `internal/domain/value/session_name.go` (new)
**Actions**: Same as Task 11, for SessionName
**Blocks**: Task 14

#### Task 14: Update domain to use value objects (60min)

**File**: `internal/domain/complaint.go`
**Actions**:

- Change AgentName from string to value.AgentName
- Change ProjectName from string to value.ProjectName
- Change SessionName from string to value.SessionName
- Update NewComplaint() constructor
- Update all tests
- Update service layer
  **Acceptance**: Tests pass, compile-time safety
  **Customer Value**: Data validation at boundaries
  **Blocks**: Nothing

#### Task 15: Create dto package (45min)

**File**: `internal/delivery/dto/` (new package)
**Actions**:

- Create package structure
- Move DTO definitions from mcp_server.go
- Add package documentation
- Add conversion helpers
  **Blocks**: Task 16

#### Task 16: Move DTOs to dto package (45min)

**Files**: Multiple
**Actions**:

- Import dto package in mcp_server.go
- Update all references
- Clean up mcp_server.go
  **Acceptance**: Clear separation, smaller files
  **Blocks**: Nothing

#### Task 17: Strengthen Severity enum (60min)

**File**: `internal/domain/severity.go` (new)
**Current**:

```go
type Severity string // ❌ Can be ""
```

**New**:

```go
type Severity int
const (
    _ Severity = iota // ✅ Zero value invalid
    SeverityLow
    SeverityMedium
    SeverityHigh
    SeverityCritical
)
```

**Actions**:

- Define iota-based enum
- Add String() method
- Add MarshalJSON/UnmarshalJSON
- Update all usages
- Add tests
  **Acceptance**: Zero value cannot be used
  **Customer Value**: Compile-time safety
  **Blocks**: Nothing

#### Task 18: Use custom error types (90min)

**Files**: All service/repo files
**Actions**:

- Replace fmt.Errorf with errors.New\* functions
- Update error handling to check error types
- Add error wrapping properly
- Add tests for error types
  **Example**:

```go
// Before
return fmt.Errorf("not found: %s", id)

// After
return errors.NewNotFoundError(fmt.Sprintf("complaint %s", id))
```

**Acceptance**: All domain errors use custom types
**Customer Value**: Better error messages
**Blocks**: Nothing

#### Task 19: Add service layer tests (90min)

**File**: `internal/service/complaint_service_test.go`
**Actions**:

- Add test for each service method
- Test error cases
- Test validation
- Test tracing
  **Coverage Target**: 80%+
  **Acceptance**: Comprehensive test suite
  **Blocks**: Nothing

#### Task 20: Fix 7 BDD test failures (90min)

**Files**: `features/bdd/*.go`
**Failures**:

- Large content handling (2)
- Resolution data preservation (1)
- Concurrent resolution (1)
- Maximum content (1)
- Creation order (1)
- Search content (1)
  **Actions**:
- Investigate each failure
- Update test expectations for new file naming
- Update assertions for ResolvedAt field
- Fix any real bugs discovered
  **Acceptance**: 47/47 BDD tests pass
  **Customer Value**: Confidence in features
  **Blocks**: Nothing

---

## 📊 EFFORT DISTRIBUTION

```
P0 (CRITICAL): 2.5h   (11%)  →  51% value
P1 (HIGH):     5.0h   (23%)  →  13% more (64% cumulative)
P2 (MEDIUM):   14.5h  (66%)  →  16% more (80% cumulative)
────────────────────────────
TOTAL:         22h    (100%) →  80% total value
```

---

## 🎯 SUCCESS CRITERIA

### After THE 1% (2.5h)

✅ All tests passing (100%)
✅ Complete audit trail (who + when)
✅ Can deploy to production

### After THE 4% (7.5h cumulative)

✅ Type-safe API layer
✅ 10-100x query performance
✅ No data corruption bugs
✅ Production-ready

### After THE 20% (22h cumulative)

✅ Strong type system
✅ Clean architecture
✅ Comprehensive tests
✅ Long-term maintainable

---

## 🚀 RECOMMENDED EXECUTION ORDER

**Day 1** (2.5h): Tasks 1-4 (THE 1%)
**Day 2** (2.5h): Tasks 5-6
**Day 3** (2.5h): Tasks 7-8
**Day 4** (2.5h): Tasks 9-10
**Week 2** (12h): Tasks 11-17
**Week 3** (3h): Tasks 18-20

**Total Calendar Time**: 3 weeks
**Total Active Time**: 22 hours
