# Micro-Task Breakdown - 15min Tasks

**Date**: 2025-11-02 07:02
**Total Tasks**: 50 (max 15min each)
**Total Effort**: ~22 hours
**Target Value**: 80% of total possible value

---

## 📊 COMPLETE MICRO-TASK TABLE

| # | Micro-Task | Time | Impact | Customer | Priority | Parent | Category |
|---|------------|------|--------|----------|----------|--------|----------|
| **1.1** | Verify domain tests pass | 10min | CRITICAL | BLOCKING | P0 | T1 | 1% |
| **1.2** | Document domain test coverage | 5min | LOW | LOW | P3 | T1 | 1% |
| **2.1** | Add context to service test calls | 15min | CRITICAL | BLOCKING | P0 | T2 | 1% |
| **2.2** | Fix mock repository creation | 15min | CRITICAL | BLOCKING | P0 | T2 | 1% |
| **2.3** | Update Resolve() test calls | 10min | CRITICAL | BLOCKING | P0 | T2 | 1% |
| **2.4** | Add ResolvedAt assertions | 5min | HIGH | MEDIUM | P1 | T2 | 1% |
| **3.1** | Add tracer to repo test setup | 10min | CRITICAL | BLOCKING | P0 | T3 | 1% |
| **3.2** | Fix type assertions in repo tests | 15min | CRITICAL | BLOCKING | P0 | T3 | 1% |
| **3.3** | Update filename expectations (UUID) | 15min | CRITICAL | BLOCKING | P0 | T3 | 1% |
| **3.4** | Run and verify all repo tests | 5min | CRITICAL | BLOCKING | P0 | T3 | 1% |
| **4.1** | Add ResolvedBy field to Complaint struct | 5min | HIGH | HIGH | P0 | T4 | 1% |
| **4.2** | Update Resolve() signature | 10min | HIGH | HIGH | P0 | T4 | 1% |
| **4.3** | Update all Resolve() callers | 10min | HIGH | HIGH | P0 | T4 | 1% |
| **4.4** | Add ResolvedBy tests | 5min | MEDIUM | MEDIUM | P1 | T4 | 1% |
| **5.1** | Define Service interface | 15min | HIGH | MEDIUM | P1 | T5 | 4% |
| **5.2** | Verify ComplaintService implements it | 10min | HIGH | MEDIUM | P1 | T5 | 4% |
| **5.3** | Update main.go to use interface | 10min | MEDIUM | LOW | P2 | T5 | 4% |
| **5.4** | Update tests to use interface | 10min | MEDIUM | MEDIUM | P2 | T5 | 4% |
| **6.1** | Analyze current Update() bug | 10min | CRITICAL | HIGH | P1 | T6 | 4% |
| **6.2** | Design fix strategy | 15min | CRITICAL | HIGH | P1 | T6 | 4% |
| **6.3** | Implement Update() fix | 20min | CRITICAL | HIGH | P1 | T6 | 4% |
| **6.4** | Add Update() test (no duplicates) | 15min | HIGH | MEDIUM | P1 | T6 | 4% |
| **7.1** | Define ComplaintDTO struct | 15min | HIGH | MEDIUM | P1 | T7 | 4% |
| **7.2** | Add ToDTO() conversion method | 15min | HIGH | MEDIUM | P1 | T7 | 4% |
| **7.3** | Define all output DTO types | 20min | HIGH | MEDIUM | P1 | T7 | 4% |
| **7.4** | Add DTO validation | 15min | MEDIUM | LOW | P2 | T7 | 4% |
| **7.5** | Add DTO tests | 15min | MEDIUM | MEDIUM | P2 | T7 | 4% |
| **7.6** | Add DTO documentation | 10min | LOW | LOW | P3 | T7 | 4% |
| **8.1** | Update ListComplaintsOutput | 10min | HIGH | MEDIUM | P1 | T8 | 4% |
| **8.2** | Update SearchComplaintsOutput | 10min | HIGH | MEDIUM | P1 | T8 | 4% |
| **8.3** | Remove map construction code | 10min | HIGH | MEDIUM | P1 | T8 | 4% |
| **9.1** | Design cache interface | 15min | HIGH | HIGH | P1 | T9 | 4% |
| **9.2** | Document cache strategies | 10min | MEDIUM | LOW | P2 | T9 | 4% |
| **9.3** | Choose data structures | 5min | MEDIUM | LOW | P2 | T9 | 4% |
| **10.1** | Create CachedRepository struct | 15min | HIGH | HIGH | P1 | T10 | 4% |
| **10.2** | Implement cache for FindByID | 20min | HIGH | HIGH | P1 | T10 | 4% |
| **10.3** | Implement cache invalidation | 15min | HIGH | HIGH | P1 | T10 | 4% |
| **10.4** | Add cache logging | 5min | LOW | LOW | P3 | T10 | 4% |
| **10.5** | Write cache tests | 15min | MEDIUM | MEDIUM | P2 | T10 | 4% |
| **11** | Create AgentName value object | 60min | MEDIUM | LOW | P2 | - | 20% |
| **12** | Create ProjectName value object | 45min | MEDIUM | LOW | P2 | - | 20% |
| **13** | Create SessionName value object | 45min | MEDIUM | LOW | P2 | - | 20% |
| **14** | Update domain to use value objects | 60min | MEDIUM | LOW | P2 | - | 20% |
| **15** | Create dto package structure | 45min | MEDIUM | LOW | P2 | - | 20% |
| **16** | Move DTOs to dto package | 45min | MEDIUM | LOW | P2 | - | 20% |
| **17** | Strengthen Severity enum | 60min | MEDIUM | MEDIUM | P2 | - | 20% |
| **18** | Use custom error types | 90min | MEDIUM | LOW | P2 | - | 20% |
| **19** | Add service layer tests | 90min | MEDIUM | MEDIUM | P2 | - | 20% |
| **20** | Fix 7 BDD test failures | 90min | MEDIUM | MEDIUM | P2 | - | 20% |

---

## 🎯 THE 1% - 15MIN BREAKDOWN (14 tasks = 2.5 hours)

### Task 1: Fix domain test failures (15min)
- **1.1** (10min): Run `go test ./internal/domain -v`, verify 100% pass
- **1.2** (5min): Document test coverage in README

### Task 2: Fix service test failures (45min)
- **2.1** (15min): Add `ctx := context.Background()` to all test functions
- **2.2** (15min): Fix mock repository constructor calls
- **2.3** (10min): Update all `complaint.Resolve()` to `complaint.Resolve(ctx)`
- **2.4** (5min): Add assertions for `ResolvedAt` field

### Task 3: Fix repo test failures (45min)
- **3.1** (10min): Add tracer parameter to all `NewFileRepository()` calls
- **3.2** (15min): Fix type assertions from `*FileRepository` to `Repository`
- **3.3** (15min): Update filename generation test expectations (UUID-timestamp)
- **3.4** (5min): Run tests, verify 100% pass

### Task 4: Add ResolvedBy field (30min)
- **4.1** (5min): Add `ResolvedBy string` field to Complaint struct
- **4.2** (10min): Update `Resolve(ctx)` to `Resolve(ctx, resolvedBy string)`
- **4.3** (10min): Update all callers in service layer, MCP handlers
- **4.4** (5min): Add test verifying ResolvedBy is set

---

## 🎯 THE 4% - 15MIN BREAKDOWN (24 tasks = 5 hours)

### Task 5: Create ComplaintService interface (45min)
- **5.1** (15min): Define `Service` interface with all methods
- **5.2** (10min): Verify `ComplaintService` satisfies interface
- **5.3** (10min): Update main.go to use `Service` type
- **5.4** (10min): Update test mocks to implement `Service`

### Task 6: Fix repository Update() bug (60min)
- **6.1** (10min): Trace through Update() code, identify bug
- **6.2** (15min): Design fix (delete old file OR track filename)
- **6.3** (20min): Implement fix, handle edge cases
- **6.4** (15min): Write test verifying no duplicate files created

### Task 7: Create typed DTO structs (90min)
- **7.1** (15min): Define `ComplaintDTO` struct with all fields
- **7.2** (15min): Add `ToDTO()` method on `domain.Complaint`
- **7.3** (20min): Define output DTOs (List, Search, File, Resolve)
- **7.4** (15min): Add validation tags to DTOs
- **7.5** (15min): Write DTO conversion tests
- **7.6** (10min): Add godoc comments for all DTOs

### Task 8: Replace map[string]interface{} (30min)
- **8.1** (10min): Change `ListComplaintsOutput.Complaints` to `[]ComplaintDTO`
- **8.2** (10min): Change `SearchComplaintsOutput.Complaints` to `[]ComplaintDTO`
- **8.3** (10min): Remove map construction loops, use `ToDTO()`

### Task 9: Design repository cache (30min)
- **9.1** (15min): Define `CachedRepository` interface
- **9.2** (10min): Document cache invalidation strategy
- **9.3** (5min): Choose data structures (map + RWMutex)

### Task 10: Implement in-memory cache (70min)
- **10.1** (15min): Create `CachedRepository` struct with mutex
- **10.2** (20min): Implement `FindByID` with cache lookup
- **10.3** (15min): Implement cache invalidation on Save/Update
- **10.4** (5min): Add debug logging for cache hits/misses
- **10.5** (15min): Write cache correctness tests

---

## 🎯 THE 20% - TASKS 11-20 (12 tasks = 14.5 hours)

**Note**: Tasks 11-20 are larger (45-90min each) and are not broken down further as they represent cohesive units of work that shouldn't be interrupted.

### Task 11: AgentName value object (60min)
- Single cohesive task: Create value object with validation

### Task 12: ProjectName value object (45min)
- Single cohesive task: Create value object with validation

### Task 13: SessionName value object (45min)
- Single cohesive task: Create value object with validation

### Task 14: Update domain (60min)
- Single cohesive task: Integrate value objects into domain

### Task 15: Create dto package (45min)
- Single cohesive task: Package structure setup

### Task 16: Move DTOs (45min)
- Single cohesive task: Refactor to new package

### Task 17: Severity enum (60min)
- Single cohesive task: Implement iota-based enum

### Task 18: Error types (90min)
- Single cohesive task: Replace error handling throughout

### Task 19: Service tests (90min)
- Single cohesive task: Comprehensive test suite

### Task 20: BDD fixes (90min)
- Single cohesive task: Fix all 7 failing tests

---

## 📊 SUMMARY STATISTICS

### By Priority
```
P0 (CRITICAL): 14 micro-tasks = 2.5h  (51% value)
P1 (HIGH):     24 micro-tasks = 5.0h  (13% more → 64% cumulative)
P2 (MEDIUM):   12 macro-tasks = 14.5h (16% more → 80% cumulative)
─────────────────────────────────────
TOTAL:         50 tasks       = 22h   (80% total value)
```

### By Time Range
```
0-10min:   12 tasks
11-15min:  20 tasks
16-45min:   6 tasks
46-90min:  12 tasks
```

### By Category
```
1% (51% value):   14 tasks = 2.5h
4% (64% value):   24 tasks = 5.0h
20% (80% value):  12 tasks = 14.5h
```

---

## 🚀 EXECUTION FLOW

### Sprint 1: THE 1% (Day 1 - 2.5h)
```
[1.1] → [1.2]
         ↓
[2.1] → [2.2] → [2.3] → [2.4]
         ↓
[3.1] → [3.2] → [3.3] → [3.4]
         ↓
[4.1] → [4.2] → [4.3] → [4.4]
```

### Sprint 2: THE 4% (Days 2-4 - 5h)
```
Day 2 (2.5h):
[5.1] → [5.2] → [5.3] → [5.4]
         ↓
[6.1] → [6.2] → [6.3] → [6.4]

Day 3 (2h):
[7.1] → [7.2] → [7.3] → [7.4] → [7.5] → [7.6]
         ↓
[8.1] → [8.2] → [8.3]

Day 4 (1.5h):
[9.1] → [9.2] → [9.3]
         ↓
[10.1] → [10.2] → [10.3] → [10.4] → [10.5]
```

### Sprint 3: THE 20% (Week 2 - 14.5h)
```
[11] → [12] → [13] → [14] (Value objects)
        ↓
[15] → [16]               (DTO package)
        ↓
[17]                      (Severity enum)
        ↓
[18]                      (Error types)
        ↓
[19]                      (Service tests)
        ↓
[20]                      (BDD fixes)
```

---

## ✅ ACCEPTANCE CRITERIA

### After Micro-Tasks 1.1 - 4.4 (THE 1%)
- [ ] All 100% tests passing
- [ ] ResolvedBy field exists and is populated
- [ ] Can deploy to production
- [ ] Complete audit trail (who + when resolved)

### After Micro-Tasks 5.1 - 10.5 (THE 4%)
- [ ] Service interface defined
- [ ] Update() bug fixed (no duplicates)
- [ ] Type-safe DTOs (no maps)
- [ ] Cache implemented (10-100x faster)
- [ ] Production-ready performance

### After Tasks 11-20 (THE 20%)
- [ ] Value objects implemented
- [ ] Clean package structure
- [ ] Strong Severity enum
- [ ] Custom error types
- [ ] Comprehensive tests
- [ ] All BDD tests passing (47/47)

---

## 🎯 SUCCESS METRICS

| Metric | Before | After 1% | After 4% | After 20% |
|--------|--------|----------|----------|-----------|
| **Tests Passing** | 85% | 100% | 100% | 100% |
| **Type Safety** | 6/10 | 6/10 | 7/10 | 9/10 |
| **Query Speed** | O(n) | O(n) | O(1) | O(1) |
| **Tech Debt** | Medium | Low | Low | Very Low |
| **Audit Trail** | Partial | Complete | Complete | Complete |
| **Deployable** | No | Yes | Yes | Yes |

---

**Total Effort**: 22 hours (50 tasks)
**Total Value**: 80% of all possible improvements
**ROI**: The 1% delivers 20:1 return, 4% delivers 8:1 return, 20% delivers 3.6:1 return
