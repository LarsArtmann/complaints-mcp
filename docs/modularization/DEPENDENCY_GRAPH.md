# Dependency Graph — Complaints MCP

**Generated:** 2026-05-14
**State:** Monolith (single `go.mod`)

## Module Info

- **Module path:** `github.com/larsartmann/complaints-mcp`
- **Go version:** 1.26.2
- **Packages:** 12

## Current Internal Dependency Graph

```
                    ┌─────────────┐
                    │  cmd/server │  (entry point)
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
              ▼            ▼            ▼
        ┌──────────┐ ┌──────────┐ ┌──────────────┐
        │  config  │ │   repo   │ │   service    │
        └────┬─────┘ └────┬─────┘ └──────┬───────┘
             │            │               │
             │       ┌────┼───────┬───────┼──────────────┐
             │       │    │       │       │              │
             │       ▼    │       ▼       ▼              │
             │    domain  │    tracing  projectdetect   │
             │            │                            │
             │            ▼                            │
             │         tracing                         │
             │                                        │
             │       ┌────────────────────────────────┘
             │       │
             │       ▼
             │   ┌───────────────┐
             └──►│ delivery/mcp  │
                 └───────┬───────┘
                         │
                    ┌────┼────┬────────┬────────┐
                    │    │    │        │        │
                    ▼    ▼    ▼        ▼        ▼
                 config repo service tracing  domain

    Leaf packages (zero internal deps):
    ┌─────────────┐  ┌────────────┐  ┌───────────┐  ┌────────────┐  ┌─────────────┐
    │   domain    │  │   errors   │  │   types   │  │ validation │  │   tracing   │
    └─────────────┘  └────────────┘  └───────────┘  └────────────┘  └─────────────┘
    ┌─────────────┐
    │projectdetect│  (only external dep: go-git)
    └─────────────┘
```

## Package-by-Package Dependencies

### Leaf Packages (zero internal dependencies)

| Package | External Dependencies | Exported Surface |
|---|---|---|
| `domain` | `gofrs/uuid`, `go-branded-id` | `Complaint`, `ComplaintID`, `AgentID`, `ProjectID`, `SessionID`, `Severity`, `ResolutionState`, validation helpers |
| `errors` | (stdlib only) | `AppError`, `ErrorCode`, constructor functions |
| `types` | (stdlib only) | `CacheSize`, `CacheEvictionPolicy`, `PageRequest`, `PageResponse[T]`, `CursorRequest`, `CursorResponse[T]`, `DocsFormat`, `DocsConfig` |
| `validation` | (stdlib only) | `Validator`, `ValidationError(s)`, `Validate`, `ValidateStruct` |
| `tracing` | `charm.land/log/v2` (mock), `opentelemetry` (real) | `Tracer` (interface), `Span`, `MockTracer`, `RealTracer`, `NoOpTracer`, factory |
| `projectdetect` | `go-git/v5` | `ProjectInfo`, `GitDetector`, `DetectProject`, `IsGitRepository` |

### Mid-Layer Packages

| Package | Internal Dependencies | External Dependencies | Exported Surface |
|---|---|---|---|
| `config` | `types` | `spf13/viper`, `spf13/cobra`, `adrg/xdg`, `charm.land/log/v2` | `Config`, `ServerConfig`, `StorageConfig`, `Load` |
| `repo` | `config`, `domain`, `tracing` | `charm.land/log/v2` | `Repository` (interface), `FileRepository`, `SimpleCachedRepository`, `CacheStats` |

### High-Level Packages

| Package | Internal Dependencies | External Dependencies | Exported Surface |
|---|---|---|---|
| `service` | `domain`, `repo`, `tracing`, `projectdetect` | `charm.land/log/v2` | `ComplaintService`, `ProjectDetector` (interface) |
| `delivery/mcp` | `config`, `domain`, `repo`, `service`, `tracing` | `go-sdk/mcp`, `charm.land/log/v2` | `MCPServer`, `NewServer`, DTOs |
| `cmd/server` | `config`, `repo`, `service`, `tracing` | `go-sdk/mcp`, `spf13/cobra` | `main` |

### Test-Only Packages

| Package | Test Imports From |
|---|---|
| `features/bdd` | `config`, `domain`, `repo`, `service`, `tracing` |

## Coupling Analysis

### God-Package Detection

No god-packages detected. The largest packages are:
- `repo/` (1 file, ~400 lines) — single concern, clean
- `domain/` (6 files) — all complaint domain, cohesive
- `delivery/mcp/` (2 files + 1 test) — MCP delivery, cohesive
- `tracing/` (3 files + 1 test) — tracing factory + impls, cohesive

### Circular Dependencies

None detected. The dependency graph is a clean DAG.

### Cross-Boundary Concerns

1. **`repo` depends on `config`** — Configuration struct used to construct repository paths. Could be decoupled by passing paths as parameters.
2. **`tracing` defines the `Tracer` interface** — Used by `repo`, `service`, `delivery/mcp`. This is the widest-reaching abstraction in the codebase.
3. **`domain` is truly leaf** — Zero internal imports. Ideal core module candidate.
4. **`errors` is truly leaf** — Zero internal imports. Could be merged into core or stay standalone.
5. **`types` is truly leaf** — Zero internal imports (only used by `config`). Could be merged into core.
6. **`validation` is truly leaf** — Zero internal imports. Used nowhere in production code currently (only imported in tests).

## External Dependency Classification

### Production Dependencies

| Dependency | Used By | Purpose |
|---|---|---|
| `charm.land/log/v2` | config, repo, service, delivery/mcp, tracing(mock) | Structured logging |
| `go-branded-id` | domain | Branded/phantom type IDs |
| `gofrs/uuid` | domain | UUID generation |
| `go-sdk/mcp` | delivery/mcp, cmd/server | MCP protocol |
| `spf13/cobra` | config, cmd/server | CLI framework |
| `spf13/viper` | config | Configuration management |
| `adrg/xdg` | config | XDG directory paths |
| `go-git/v5` | projectdetect | Git repository detection |
| `opentelemetry/*` | tracing(real) | Distributed tracing |

### Test Dependencies

| Dependency | Used By | Purpose |
|---|---|---|
| `onsi/ginkgo/v2` | features/bdd | BDD test framework |
| `onsi/gomega` | features/bdd | BDD assertions |
| `stretchr/testify` | domain tests, types tests | Test assertions |
| `go-playground/validator` | validation | Struct validation (used in validation package) |

## Banned Dependency Flags

Per the how-to-golang skill:

| Dependency | Status | Recommended Replacement |
|---|---|---|
| `spf13/viper` | **BANNED** | `koanf` |
| `stretchr/testify` | **BANNED** | `ginkgo/v2 + gomega` |
| `go-playground/validator` | **BANNED** | `govalid` |

These should be addressed during or before modularization.
