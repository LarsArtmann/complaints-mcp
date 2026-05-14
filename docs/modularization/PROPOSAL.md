# Modularization Proposal — Complaints MCP

**Created:** 2026-05-14
**State:** Phase 3 — Draft Proposal

---

## 1. Executive Summary

Complaints MCP is a small, well-structured Go monolith with a single `go.mod` and 12 packages under `internal/`. The project is young, the dependency graph is already a clean DAG with no circular dependencies, and no god-packages exist. This is the ideal time to modularize — before coupling accumulates.

**Why modularize now:**

- Enforce compile-time boundaries between domain, infrastructure, and delivery layers
- Enable independent versioning of the core domain (reusable by other projects)
- Isolate test-only dependencies from production `go.mod`
- Prepare for future growth (additional delivery mechanisms, storage backends)

**What changes:**

- Split from 1 `go.mod` into 4 sub-modules coordinated by a `go.work` file
- Extract `domain` + `errors` + `types` into a zero-dependency core module
- Keep delivery and infrastructure as separate modules
- Fix 3 banned dependencies (`viper`, `testify`, `go-playground/validator`) as part of the split

---

## 2. Current State Analysis

### Package Dependency Graph

See [DEPENDENCY_GRAPH.md](./DEPENDENCY_GRAPH.md) for full analysis.

### Layer Classification

| Layer              | Packages                                                   | Internal Deps                          | Role                         |
| ------------------ | ---------------------------------------------------------- | -------------------------------------- | ---------------------------- |
| **Domain**         | `domain`, `errors`, `types`                                | None                                   | Business types and rules     |
| **Infrastructure** | `repo`, `config`, `tracing`, `projectdetect`, `validation` | Domain layer                           | External integrations        |
| **Application**    | `service`                                                  | Domain + Infrastructure                | Business logic orchestration |
| **Delivery**       | `delivery/mcp`                                             | All layers                             | MCP protocol adapter         |
| **Entry**          | `cmd/server`                                               | Config, Repo, Service, Tracing         | CLI entry point              |
| **Tests**          | `features/bdd`                                             | Config, Domain, Repo, Service, Tracing | BDD specifications           |

### Coupling Hotspots

1. **`tracing.Tracer` interface** — Widest-reaching abstraction, imported by `repo`, `service`, `delivery/mcp`. Correctly defined as interface in the tracing package.
2. **`repo` → `config`** — Repository uses `config.Config` to determine storage paths. This is a leaky abstraction; the repository should accept paths as constructor parameters.
3. **`domain` is perfectly isolated** — Zero internal imports, only external deps are `gofrs/uuid` and `go-branded-id`.

### Banned Dependencies Found

| Dependency                | Where                         | Replacement |
| ------------------------- | ----------------------------- | ----------- |
| `spf13/viper`             | `config`                      | `koanf`     |
| `stretchr/testify`        | `domain` tests, `types` tests | `gomega`    |
| `go-playground/validator` | `validation`                  | `govalid`   |

---

## 3. Proposed Module Structure

### Module Map

```
complaints-mcp/                   (go.work — workspace root)
├── core/                         Module 1: domain types + errors + types
│   ├── go.mod                    github.com/larsartmann/complaints-mcp/core
│   ├── domain/
│   │   ├── complaint.go
│   │   ├── id_types.go
│   │   ├── id_helpers.go
│   │   ├── severity.go
│   │   └── *_test.go
│   ├── errors/
│   │   ├── app_error.go
│   │   └── complaint.go
│   └── types/
│       ├── cache.go
│       ├── pagination.go
│       └── docs.go
│
├── infra/                        Module 2: infrastructure
│   ├── go.mod                    github.com/larsartmann/complaints-mcp/infra
│   ├── repo/
│   │   └── repository.go
│   ├── config/
│   │   ├── config.go
│   │   └── *_test.go
│   ├── tracing/
│   │   ├── factory.go
│   │   ├── real_tracer.go
│   │   ├── mock_tracer.go
│   │   └── tracing_test.go
│   ├── projectdetect/
│   │   └── detector.go
│   └── validation/
│       └── validator.go
│
├── server/                       Module 3: MCP delivery + CLI entry point
│   ├── go.mod                    github.com/larsartmann/complaints-mcp/server
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   └── delivery/
│       └── mcp/
│           ├── mcp_server.go
│           └── dto.go
│
├── service/                      Module 4: business logic
│   ├── go.mod                    github.com/larsartmann/complaints-mcp/service
│   └── service.go
│
├── features/                     (not a Go module — BDD specs)
│   └── bdd/
│
├── go.work                       Workspace coordination
├── go.work.sum
├── docs/
├── README.md
└── AGENTS.md
```

### Module Definitions

#### Module 1: `core`

| Field                    | Value                                                                                                                                          |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| **Path**                 | `/core`                                                                                                                                        |
| **Module**               | `github.com/larsartmann/complaints-mcp/core`                                                                                                   |
| **Purpose**              | Domain types, errors, and shared value objects — zero internal dependencies                                                                    |
| **Internal deps (prod)** | None                                                                                                                                           |
| **Internal deps (test)** | None                                                                                                                                           |
| **External deps**        | `gofrs/uuid`, `go-branded-id`                                                                                                                  |
| **Public API**           | `Complaint`, all ID types, `Severity`, `ResolutionState`, `AppError`, `ErrorCode`, `CacheSize`, `PageRequest`, `PageResponse[T]`, `DocsFormat` |

#### Module 2: `infra`

| Field                    | Value                                                                                                                              |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------- |
| **Path**                 | `/infra`                                                                                                                           |
| **Module**               | `github.com/larsartmann/complaints-mcp/infra`                                                                                      |
| **Purpose**              | Infrastructure adapters — storage, config, tracing, git detection, validation                                                      |
| **Internal deps (prod)** | `core`                                                                                                                             |
| **Internal deps (test)** | `core`                                                                                                                             |
| **External deps**        | `charm.land/log/v2`, `go-sdk/mcp`, `spf13/cobra`, `spf13/viper` (→ migrate to `koanf`), `adrg/xdg`, `go-git/v5`, `opentelemetry/*` |
| **Public API**           | `Repository` (interface), `FileRepository`, `Tracer` (interface), `MockTracer`, `Config`, `GitDetector`, `Validator`               |

#### Module 3: `server`

| Field                    | Value                                            |
| ------------------------ | ------------------------------------------------ |
| **Path**                 | `/server`                                        |
| **Module**               | `github.com/larsartmann/complaints-mcp/server`   |
| **Purpose**              | MCP protocol delivery and CLI entry point        |
| **Internal deps (prod)** | `core`, `infra`, `service`                       |
| **Internal deps (test)** | `core`, `infra`                                  |
| **External deps**        | `go-sdk/mcp`, `spf13/cobra`, `charm.land/log/v2` |
| **Public API**           | `MCPServer`, `NewServer`, DTO types              |

#### Module 4: `service`

| Field                    | Value                                                                           |
| ------------------------ | ------------------------------------------------------------------------------- |
| **Path**                 | `/service`                                                                      |
| **Module**               | `github.com/larsartmann/complaints-mcp/service`                                 |
| **Purpose**              | Business logic orchestration — the application layer                            |
| **Internal deps (prod)** | `core`, `infra` (only `repo.Repository` interface + `tracing.Tracer` interface) |
| **Internal deps (test)** | `core`, `infra`                                                                 |
| **External deps**        | `charm.land/log/v2`                                                             |
| **Public API**           | `ComplaintService`, `ProjectDetector` (interface)                               |

### Dependency DAG (Proposed)

```
  core (zero internal deps)
    ▲
    │
  infra (depends on core)
    ▲
    │
  service (depends on core + infra interfaces)
    ▲
    │
  server (depends on core + infra + service)
```

**Cycle check:** `core` → `infra` → `service` → `server`. Strict top-down. No cycles possible.

---

## 4. DAG Verification

### Dependency Direction Rules

1. **`core`** — Zero internal dependencies. Pure domain types.
2. **`infra`** — Depends only on `core`. Infrastructure implements interfaces defined in `core` or `infra`.
3. **`service`** — Depends on `core` (domain types) and `infra` (only interfaces: `Repository`, `Tracer`). Never on concrete implementations.
4. **`server`** — Depends on everything. This is the composition root.

### Cross-Module Import Table

| From \ To     | `core` | `infra`             | `service` | `server` |
| ------------- | ------ | ------------------- | --------- | -------- |
| **`core`**    | —      | ✗                   | ✗         | ✗        |
| **`infra`**   | ✓      | —                   | ✗         | ✗        |
| **`service`** | ✓      | ✓ (interfaces only) | —         | ✗        |
| **`server`**  | ✓      | ✓                   | ✓         | —        |

### Bidirectional Test Dependencies

None required. BDD tests in `features/bdd` import from all modules (they live outside the module tree).

---

## 5. Replace / Workspace Strategy

**Chosen: `go.work` at repo root**

Rationale:

- 4 modules is enough to benefit from workspace mode
- Each module's `go.mod` stays clean — no `replace` directives
- `go.work` is automatically ignored by consumers when we publish
- Simpler than per-module `replace` directives

### go.work structure

```go
go 1.26.2

use (
    ./core
    ./infra
    ./service
    ./server
)
```

### Verification steps

1. `go work sync` — ensure workspace is consistent
2. `go build ./...` from repo root — all modules compile
3. `go test ./...` from repo root — all tests pass
4. Remove `go.work` → verify each module builds independently with versioned imports (future)

---

## 6. Test Dependency Isolation

### Production vs Test Dependencies Per Module

| Module    | Production Deps                                                                           | Test-Only Deps                   |
| --------- | ----------------------------------------------------------------------------------------- | -------------------------------- |
| `core`    | `gofrs/uuid`, `go-branded-id`                                                             | `gomega` (migrated from testify) |
| `infra`   | `core`, `charm.land/log/v2`, `go-git/v5`, `opentelemetry/*`, `spf13/cobra`, `spf13/viper` | `gomega`                         |
| `service` | `core`, `infra`, `charm.land/log/v2`                                                      | `gomega`                         |
| `server`  | `core`, `infra`, `service`, `go-sdk/mcp`, `charm.land/log/v2`                             | `gomega`                         |

### BDD Test Module

`features/bdd` is NOT a Go module. It will import from the workspace modules:

- `github.com/larsartmann/complaints-mcp/core/...`
- `github.com/larsartmann/complaints-mcp/infra/...`
- `github.com/larsartmann/complaints-mcp/service/...`

It requires `ginkgo/v2` + `gomega` in its own `go.mod` at the repo root level, or it stays as part of the workspace.

**Decision:** Keep `features/` as part of the workspace but NOT as its own module. It can import workspace modules. Add its test dependencies to `go.work` or use a root-level `go.mod` for test tooling only.

**Alternative (recommended):** Move BDD tests into the `server` module's test files. They test the full stack and `server` already depends on everything. This eliminates the need for a separate test module.

---

## 7. Interface Extraction Plan

### Current State

The codebase already uses interfaces well:

- `repo.Repository` — interface, implemented by `FileRepository` and `SimpleCachedRepository`
- `tracing.Tracer` — interface, implemented by `MockTracer`, `RealTracer`, `NoOpTracer`
- `service.ProjectDetector` — interface, implemented by `GitDetector`

### Changes Needed

1. **`repo.Repository` interface** → Move to `core` module as a domain port. `FileRepository` stays in `infra`.
   - Reason: `service` depends on `Repository` interface, not the implementation. Defining the interface in `core` allows `service` to depend only on `core` for this, not `infra`.
   - **BUT:** This would mean `core` needs to know about storage abstractions. Alternative: keep interface in `infra` and let `service` import `infra` for the interface only.
   - **Decision:** Keep `Repository` interface in `infra/repo`. The `service` module already legitimately depends on `infra` for both `Repository` and `Tracer` interfaces. No split needed.

2. **`tracing.Tracer` interface** → Keep in `infra/tracing`. Already used correctly as interface by consumers.

3. **No new interface/impl splits needed** — The existing interface placements are sound.

---

## 8. Versioning Strategy

**Chosen: Shared version (monorepo tagging)**

Rationale:

- Single team, single repo, tight coupling between modules
- No external consumers of individual modules
- Simpler CI/CD — one tag `v1.2.3` bumps everything
- Avoids the complexity of per-module semver tagging

Tag format: `vMAJOR.MINOR.PATCH` at repo root

**Migration path:** If a module (e.g., `core`) is extracted for reuse by another project later, it can be tagged independently at that point using `core/vMAJOR.MINOR.PATCH`.

---

## 9. Migration Strategy

Ordered steps, each independently executable. See [EXECUTION_PLAN.md](./EXECUTION_PLAN.md) for full details.

1. **Fix banned dependencies** — Replace `viper` → `koanf`, `testify` → `gomega`, `go-playground/validator` → `govalid`
2. **Decouple `repo` from `config`** — Pass paths as constructor params instead of `Config` struct
3. **Create `core` module** — Move `domain/`, `errors/`, `types/` into `core/`
4. **Create `infra` module** — Move `repo/`, `config/`, `tracing/`, `projectdetect/`, `validation/` into `infra/`
5. **Create `service` module** — Move `service/` into `service/`
6. **Create `server` module** — Move `cmd/`, `delivery/mcp/` into `server/`
7. **Create `go.work`** — Wire all modules together
8. **Fix BDD test imports** — Update `features/bdd` to use new module paths
9. **Verify** — Full build + test suite passes
10. **Update docs** — AGENTS.md, README.md

---

## 10. Risk Assessment

| Risk                                             | Likelihood | Impact | Mitigation                                           |
| ------------------------------------------------ | ---------- | ------ | ---------------------------------------------------- |
| Import path changes break BDD tests              | High       | Medium | Update all imports in one commit, verify immediately |
| Banned dependency replacement causes API changes | Medium     | Medium | Replace one at a time, test after each               |
| `go.work` issues with IDE/LSP                    | Low        | Low    | Use Go 1.26 workspace support, well-tested           |
| Circular dependency appears during split         | Low        | High   | Verify DAG at each step with `go vet ./...`          |
| Test-only deps leak into production go.mod       | Medium     | Medium | Run `go mod tidy` per module after each step         |

---

## 11. Build System Impact

### Current State

- `justfile` exists (deprecated per project policy)
- No `flake.nix` found
- CI: `.github/` directory exists

### Changes Needed

1. **`flake.nix`** — Create per-module build derivation if flake is added later. For now, `go.work` handles everything.
2. **CI/CD** — Update to run `go build ./...` and `go test ./...` from repo root with workspace active.
3. **`justfile`** — Update build/test commands to work with workspace structure. Eventually migrate to `flake.nix`.
4. **`.golangci.yml`** — Update paths if linting is per-module.

---

## Key Decisions Summary

1. **4 modules** — `core`, `infra`, `service`, `server` — balancing granularity with project size
2. **`go.work`** for local development, no `replace` directives
3. **Shared versioning** — single tag for the monorepo
4. **`core` module is zero-dependency** — pure domain types, reusable
5. **`infra` owns all interfaces** — `Repository`, `Tracer`, `Validator` — consumed by `service` via import
6. **BDD tests** stay in `features/` as workspace consumers, not their own module
7. **Banned dependency fixes** are prerequisites, not blockers
