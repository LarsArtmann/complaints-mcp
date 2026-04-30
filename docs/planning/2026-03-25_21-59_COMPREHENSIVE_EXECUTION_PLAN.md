# Comprehensive Multi-Step Execution Plan

**Date:** 2026-03-25 21:59  
**Status:** Planning Phase  
**Objective:** Systematic improvement of complaints-mcp with high-impact, low-effort prioritization

---

## 1. REFLECTION: What I Forgot & Could Improve

### What I Forgot in Previous Work:

1. **Didn't verify test fixtures** - The 5 failing tests need proper setup data
2. **Didn't add structured logging** - Log v2 is imported but not fully utilized with context
3. **Didn't check for existing validation libraries** - We're using custom validation when libraries exist
4. **Didn't consider pagination types** - Missing standardized pagination model
5. **Didn't add request/response DTO types** - Delivery layer lacks proper DTOs

### What Could Be Better:

1. **Type Safety**: Complaint struct uses `ProjectID` field named `ProjectName` - confusing
2. **Error Handling**: Service layer returns generic errors, not using `apperrors` package consistently
3. **Context Propagation**: Tracing is passed but not fully integrated with logging
4. **Repository Interface**: Missing `Count()` and bulk operations
5. **Validation**: Manual validation instead of using `go-playground/validator` or similar

### Architecture Improvements Needed:

1. **CQRS Pattern**: Separate read/write models for better scalability
2. **Event Sourcing**: Complaint lifecycle events not captured
3. **Unit of Work**: Repository operations lack transaction boundaries
4. **Specification Pattern**: Search/filter logic scattered
5. **Result Type**: Using naked returns instead of Result/Option types

---

## 2. COMPREHENSIVE MULTI-STEP EXECUTION PLAN

### Phase 0: Foundation (Immediate - Day 1)

#### Step 0.1: Fix the 5 Failing BDD Tests

**Files:** `features/bdd/*_test.go`  
**Work:** Medium (2-3 hours)  
**Impact:** High (90% → 100% test pass rate)  
**Existing Code:** Test setup exists, just needs proper fixture data

**Action:**

- Add `project_name` and `session_name` to test fixtures
- Fix timing race condition in listing tests
- Fix search indexing in test setup

**Verification:** `go test ./features/bdd -v`

---

#### Step 0.2: Clear Go Module Cache & Fix LSP

**Work:** Low (15 minutes)  
**Impact:** High (fixes IDE errors)

**Action:**

```bash
go clean -modcache
go mod download
```

**Verification:** gopls stops showing false errors

---

#### Step 0.3: Add Pagination Types

**File:** `internal/types/pagination.go` (NEW)  
**Work:** Low (30 minutes)  
**Impact:** Medium (standardizes pagination)  
**Library Option:** Could use `github.com/pilagod/gorm-cursor-paginator` but better to keep simple

**Proposed Types:**

```go
// PageRequest represents pagination parameters
type PageRequest struct {
    Page    int `json:"page" validate:"min=0"`
    PerPage int `json:"per_page" validate:"min=1,max=100"`
}

// PageResponse represents paginated response
type PageResponse[T any] struct {
    Data       []T    `json:"data"`
    Total      int64  `json:"total"`
    Page       int    `json:"page"`
    PerPage    int    `json:"per_page"`
    TotalPages int    `json:"total_pages"`
    HasNext    bool   `json:"has_next"`
    HasPrev    bool   `json:"has_prev"`
}
```

---

### Phase 1: Type System Enhancement (Day 1-2)

#### Step 1.1: Fix Field Naming Inconsistency

**File:** `internal/domain/complaint.go`  
**Work:** Low (1 hour)  
**Impact:** Medium (clarity)  
**Breaking Change:** Yes (JSON field name change)

**Action:**

```go
// BEFORE:
ProjectName ProjectID `json:"project_id"`

// AFTER:
ProjectID ProjectID `json:"project_id"`
```

**Verification:** Update all usages, tests pass

---

#### Step 1.2: Add Request/Response DTOs

**File:** `internal/delivery/mcp/dto.go`  
**Work:** Medium (2 hours)  
**Impact:** High (API contract clarity)  
**Existing Code:** Basic DTOs exist, need expansion

**Proposed DTOs:**

```go
// FileComplaintRequest represents the request to file a complaint
type FileComplaintRequest struct {
    AgentID         string          `json:"agent_id" validate:"required"`
    SessionID       string          `json:"session_id" validate:"required"`
    ProjectID       string          `json:"project_id" validate:"required"`
    TaskDescription string          `json:"task_description" validate:"required,max=1000"`
    ContextInfo     string          `json:"context_info,omitempty"`
    MissingInfo     string          `json:"missing_info,omitempty"`
    ConfusedBy      string          `json:"confused_by,omitempty"`
    FutureWishes    string          `json:"future_wishes,omitempty"`
    Severity        string          `json:"severity" validate:"required,oneof=low medium high critical"`
}

// FileComplaintResponse represents the response from filing a complaint
type FileComplaintResponse struct {
    ID        string    `json:"id"`
    FilePath  string    `json:"file_path"`
    DocsPath  string    `json:"docs_path"`
    CreatedAt time.Time `json:"created_at"`
}
```

---

#### Step 1.3: Integrate Structured Validation Library

**Files:** `internal/domain/*.go`, `internal/delivery/mcp/dto.go`  
**Work:** Medium (3 hours)  
**Impact:** High (removes boilerplate, adds validation rules)  
**Library:** `github.com/go-playground/validator/v10` - industry standard

**Benefits:**

- Tag-based validation
- Built-in validators (email, UUID, etc.)
- Custom validators
- Error messages in multiple languages
- Performance optimized

**Action:**

```go
import "github.com/go-playground/validator/v10"

var validate = validator.New()

func (r *FileComplaintRequest) Validate() error {
    return validate.Struct(r)
}
```

---

### Phase 2: Error Handling Standardization (Day 2)

#### Step 2.1: Create Service Error Wrappers

**File:** `internal/service/errors.go` (NEW)  
**Work:** Low (1 hour)  
**Impact:** High (consistent error handling)  
**Existing Code:** `internal/errors/app_error.go` exists but not used in service

**Action:**
Create service-specific error wrappers that use the apperrors package:

```go
func WrapRepositoryError(op string, err error) error {
    return apperrors.NewRepositoryError(
        fmt.Sprintf("failed to %s complaint", op),
        err,
    )
}
```

**Verification:** All service methods return apperrors

---

#### Step 2.2: Add Error Middleware/Interceptor

**File:** `internal/delivery/mcp/middleware.go` (NEW)  
**Work:** Medium (2 hours)  
**Impact:** Medium (consistent error responses)

**Action:** Create MCP middleware that converts apperrors to proper MCP error responses

---

### Phase 3: Repository Enhancement (Day 2-3)

#### Step 3.1: Add Count Methods

**Files:** `internal/repo/repository.go`  
**Work:** Low (1 hour)  
**Impact:** Medium (needed for pagination)

**Add to Interface:**

```go
Count(ctx context.Context) (int64, error)
CountBySeverity(ctx context.Context, severity domain.Severity) (int64, error)
CountByProject(ctx context.Context, projectID string) (int64, error)
CountUnresolved(ctx context.Context) (int64, error)
```

---

#### Step 3.2: Add Bulk Operations

**Files:** `internal/repo/repository.go`  
**Work:** Medium (2 hours)  
**Impact:** Low-Medium (performance for batch operations)

**Add to Interface:**

```go
SaveBatch(ctx context.Context, complaints []*domain.Complaint) error
DeleteBatch(ctx context.Context, ids []domain.ComplaintID) error
```

---

#### Step 3.3: Implement Specification Pattern for Search

**File:** `internal/repo/specification.go` (NEW)  
**Work:** Medium-High (4 hours)  
**Impact:** High (cleaner search/filter logic)  
**Library Reference:** Similar to `github.com/kentaro/gormspec` but for file-based

**Proposed Design:**

```go
// Specification interface for filtering
type Specification interface {
    IsSatisfiedBy(complaint *domain.Complaint) bool
}

// Composite specifications
type AndSpecification struct {
    specs []Specification
}

type SeveritySpecification struct {
    Severity domain.Severity
}

type ProjectSpecification struct {
    ProjectID string
}
```

---

### Phase 4: Logging & Observability (Day 3)

#### Step 4.1: Add Structured Logging to Service Layer

**Files:** `internal/service/service.go`  
**Work:** Medium (2 hours)  
**Impact:** High (observability)  
**Existing Code:** Logger passed to MCP server but not service

**Action:**

- Add `*log.Logger` to `ComplaintService`
- Add contextual logging with `log.With()`
- Log all operations with timing

---

#### Step 4.2: Integrate Tracing with Logging

**Files:** `internal/tracing/*.go`, `internal/service/service.go`  
**Work:** Medium (2 hours)  
**Impact:** Medium (distributed tracing)

**Action:**

- Extract trace ID from context
- Add trace ID to log entries
- Create spans for all service operations

---

### Phase 5: Testing Infrastructure (Day 3-4)

#### Step 5.1: Add Test Fixtures Package

**File:** `internal/testutil/fixtures.go` (NEW)  
**Work:** Medium (2 hours)  
**Impact:** High (DRY test data)

**Action:**
Create reusable test fixtures:

```go
func NewValidComplaint() *domain.Complaint
func NewValidFileComplaintRequest() *dto.FileComplaintRequest
func NewTestRepository(t *testing.T) repo.Repository
```

---

#### Step 5.2: Add Contract Tests

**Files:** `internal/delivery/mcp/contract_test.go` (NEW)  
**Work:** Medium-High (3 hours)  
**Impact:** High (API stability)  
**Library:** Use existing Ginkgo/Gomega

**Action:**
Test that DTOs serialize/deserialize correctly and match domain models

---

### Phase 6: Advanced Features (Day 4-5)

#### Step 6.1: Implement Project Auto-Detection

**Files:** `internal/service/detection.go` (NEW)  
**Work:** High (6 hours)  
**Impact:** Very High (key feature)  
**Library:** `github.com/go-git/go-git/v5` - already planned

**Algorithm:**

1. Check environment variables (COMPLAINTS_PROJECT_NAME)
2. Check git config (complaints.project.name)
3. Walk up directory tree looking for:
   - `.git/config` → extract remote origin
   - `go.mod` → extract module path
   - `package.json` → extract name
4. Cache result

---

#### Step 6.2: Add Caching Layer Improvements

**Files:** `internal/types/cache.go`  
**Work:** Medium (3 hours)  
**Impact:** Medium (performance)  
**Library:** Could use `github.com/patrickmn/go-cache` but we have custom LRU

**Improvements:**

- Add TTL support
- Add cache warming strategies
- Add cache metrics (prometheus)

---

#### Step 6.3: Add Full-Text Search

**Files:** `internal/search/` (NEW PACKAGE)  
**Work:** High (6 hours)  
**Impact:** High (user experience)  
**Library:** `github.com/blevesearch/bleve` - excellent Go search library

**Action:**

- Create search index on complaint text fields
- Index on Save/Update
- Search across: TaskDescription, ContextInfo, MissingInfo, ConfusedBy

---

## 3. WORK VS IMPACT ANALYSIS

| Priority | Step                       | Work Hours | Impact          | ROI        | Library Used            |
| -------- | -------------------------- | ---------- | --------------- | ---------- | ----------------------- |
| P0       | 0.1 Fix failing tests      | 2-3h       | High (90%→100%) | ⭐⭐⭐⭐⭐ | None                    |
| P0       | 0.2 Clear cache            | 0.25h      | High (IDE fix)  | ⭐⭐⭐⭐⭐ | None                    |
| P0       | 0.3 Pagination types       | 0.5h       | Medium          | ⭐⭐⭐⭐   | None                    |
| P1       | 1.3 Validation lib         | 3h         | High            | ⭐⭐⭐⭐   | go-playground/validator |
| P1       | 2.1 Service errors         | 1h         | High            | ⭐⭐⭐⭐   | None (use existing)     |
| P1       | 1.2 Request/Response DTOs  | 2h         | High            | ⭐⭐⭐⭐   | None                    |
| P2       | 6.1 Project auto-detection | 6h         | Very High       | ⭐⭐⭐⭐⭐ | go-git/go-git           |
| P2       | 3.3 Specification pattern  | 4h         | High            | ⭐⭐⭐     | None                    |
| P2       | 4.1 Structured logging     | 2h         | High            | ⭐⭐⭐⭐   | charmbracelet/log       |
| P3       | 6.3 Full-text search       | 6h         | High            | ⭐⭐⭐     | blevesearch/bleve       |
| P3       | 5.1 Test fixtures          | 2h         | High            | ⭐⭐⭐⭐   | None                    |
| P4       | 1.1 Fix field naming       | 1h         | Medium          | ⭐⭐⭐     | Breaking change         |
| P4       | 3.1 Count methods          | 1h         | Medium          | ⭐⭐⭐     | None                    |
| P4       | 3.2 Bulk operations        | 2h         | Low-Med         | ⭐⭐       | None                    |
| P4       | 5.2 Contract tests         | 3h         | High            | ⭐⭐⭐     | None                    |

---

## 4. EXISTING CODE REUSE ANALYSIS

### Already Have (Reuse):

1. **Phantom Types** - `go-branded-id` ✓
2. **Structured Errors** - `internal/errors/app_error.go` ✓ (but not used consistently)
3. **Repository Pattern** - `internal/repo/repository.go` ✓
4. **Tracing Abstraction** - `internal/tracing/*.go` ✓
5. **Cache Stats** - `internal/types/cache.go` ✓
6. **Config Management** - `internal/config/config.go` ✓

### Should Use More:

1. `apperrors` package in service layer
2. `tracing` in all operations
3. `cache` metrics collection
4. `config` validation

### Missing (Implement or Import):

1. Pagination types
2. Request/Response DTOs
3. Validation library
4. Full-text search
5. Specification pattern

---

## 5. TYPE MODEL ARCHITECTURE IMPROVEMENTS

### Current Issues:

1. **Inconsistent naming** - `ProjectName` field with `ProjectID` type
2. **No pagination types** - Using naked int parameters
3. **No Result type** - Using naked returns with error
4. **No Option type** - Required fields not distinguishable

### Proposed Improvements:

#### 5.1 Result Type

```go
// Result represents an operation result
type Result[T any] struct {
    Value T
    Error error
}

func (r Result[T]) IsSuccess() bool { return r.Error == nil }
func (r Result[T]) IsFailure() bool { return r.Error != nil }
```

#### 5.2 Option Type

```go
// Option represents optional values
type Option[T any] struct {
    value *T
}

func Some[T any](v T) Option[T]  { return Option[T]{&v} }
func None[T any]() Option[T]     { return Option[T]{} }
func (o Option[T]) IsSome() bool { return o.value != nil }
func (o Option[T]) IsNone() bool { return o.value == nil }
```

#### 5.3 Generic Specifications

```go
// Specification interface using generics
type Specification[T any] interface {
    IsSatisfiedBy(entity T) bool
    And(other Specification[T]) Specification[T]
    Or(other Specification[T]) Specification[T]
    Not() Specification[T]
}
```

---

## 6. WELL-ESTABLISHED LIBRARIES TO USE

### Validation

**Library:** `github.com/go-playground/validator/v10`  
**Why:** Industry standard, 17k+ stars, tag-based, extensible  
**Use For:** Request DTO validation, domain model validation

### Git Operations

**Library:** `github.com/go-git/go-git/v5`  
**Why:** Pure Go, no CGO, comprehensive  
**Use For:** Project auto-detection

### Full-Text Search

**Library:** `github.com/blevesearch/bleve`  
**Why:** Leading Go search library, supports faceting, aggregations  
**Use For:** Complaint search functionality

### Caching (Alternative)

**Library:** `github.com/patrickmn/go-cache`  
**Why:** Simple, TTL support, thread-safe  
**Consider:** If our LRU cache needs TTL

### Pagination (Alternative)

**Library:** `github.com/pilagod/gorm-cursor-paginator`  
**Why:** Cursor-based pagination (better for real-time)  
**Consider:** If we add SQL backend

### Result/Option Types

**Library:** `github.com/samber/mo`  
**Why:** Comprehensive monads (Option, Result, Either)  
**Consider:** Adding functional programming patterns

### Testing

**Library:** `github.com/stretchr/testify`  
**Already Using:** ✓  
**Add:** `github.com/vektra/mockery/v2` for mocks

---

## 7. QUESTIONS I CANNOT FIGURE OUT MYSELF

### Q1: Should we use `mo` Result/Option types or create our own?

**Context:** We have `go-branded-id` for phantom types. Should we:

- A) Add Result/Option to that library
- B) Use `samber/mo` directly
- C) Create a local `internal/types/functional.go`

**Tradeoffs:**

- A) Consistent with our philosophy, but more maintenance
- B) Battle-tested, but external dependency
- C) Simple, but duplicates `mo` functionality

### Q2: How should we handle the field naming breaking change?

**Context:** `ProjectName ProjectID` → `ProjectID ProjectID`

**Options:**

- A) Just change it (we're pre-1.0)
- B) Add JSON alias for backward compat
- C) Create migration script

### Q3: Should we implement Specification pattern or use a library?

**Context:** File-based repository, not SQL

**Considerations:**

- Specification pattern adds abstraction
- But we don't have SQL where it shines (query composition)
- Might be overkill for file-based filtering

---

## 8. RECOMMENDED EXECUTION ORDER

### This Week (Immediate Value):

1. **Step 0.1** - Fix failing tests (90%→100%)
2. **Step 0.2** - Clear cache (developer experience)
3. **Step 1.3** - Add validation library (robustness)
4. **Step 2.1** - Service error wrappers (consistency)
5. **Step 6.1** - Project auto-detection (key feature)

### Next Week (Architecture):

6. **Step 0.3** - Pagination types
7. **Step 1.2** - Request/Response DTOs
8. **Step 4.1** - Structured logging
9. **Step 5.1** - Test fixtures

### Following Weeks (Advanced):

10. **Step 6.3** - Full-text search
11. **Step 3.3** - Specification pattern
12. **Step 6.2** - Cache improvements

---

**End of Plan**  
**Ready for Execution**
