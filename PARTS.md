# Component Analysis: Reusable Libraries & SDKs

> Analysis of complaints-mcp components with potential for extraction as standalone libraries.
>
> Generated: 2026-03-03
> Based on: HOW_TO_GOLANG.md v3.1 principles

---

## Executive Summary

The complaints-mcp project contains **8 distinct components** that could be extracted as reusable libraries. This document analyzes each component, compares against well-known alternatives, and identifies value-add opportunities.

| Component | Extraction Priority | Effort | Value |
|-----------|-------------------|--------|-------|
| Phantom Type System | High | Low | High |
| File-Based Repository | Medium | Medium | Medium |
| Structured Errors | Medium | Low | Medium |
| Config Management | Low | Low | Low |
| Tracing Abstraction | Low | Low | Low |
| MCP Server Kit | High | High | Very High |
| Type-Safe Cache Config | Low | Low | Low |
| Domain Primitives | Medium | Low | Medium |

---

## 1. Phantom Type System (ID Types)

**Current Location:** `internal/domain/*_id.go` files

**Description:** Type-safe identifier system using Go's type alias pattern to prevent mixing different ID types at compile time.

```go
// Current implementation pattern
type ComplaintID string
type AgentID string
type SessionID string
type ProjectID string
```

**Features:**
- Compile-time type safety (can't pass AgentID where ComplaintID expected)
- JSON marshaling/unmarshaling support
- Validation with regex patterns
- Constructor functions with error handling
- `Must*` variants for testing
- Consistent API: `Validate()`, `IsValid()`, `IsEmpty()`, `String()`

### Alternatives

| Library | URL | Approach | Pros | Cons |
|---------|-----|----------|------|------|
| **samber/mo** | github.com/samber/mo | Generic Option types | Rich functional API | No ID-specific features |
| **google/uuid** | github.com/google/uuid | UUID-specific | Standard, well-tested | Only UUIDs, not generic |
| **gofrs/uuid** | github.com/gofrs/uuid | UUID with V1-V5 | Multiple UUID versions | Only UUIDs |
| **Custom** | - | Type aliases | Zero dependencies, simple | Manual implementation |

### Value-Add Opportunity

**Library Name:** `github.com/larsartmann/goid`

**Differentiation:**
1. **Code Generation**: Generate phantom types from struct tags
   ```go
   //go:generate goid -type=User -fields=ID,Email
   type User struct {
       ID    UserID    `goid:"uuid"`
       Email EmailID   `goid:"email"`
   }
   ```

2. **Built-in Validators**: Common patterns (UUID, email, slug, semver)

3. **SQL Driver Support**: Automatic `sql.Scanner` / `driver.Valuer` implementation

4. **OpenAPI Integration**: Generate OpenAPI schemas from type definitions

5. **Zero Runtime Cost**: Compile-time safety with no overhead

**Market Gap:** No popular Go library specifically targets the "branded types" / "phantom types" pattern with code generation. TypeScript has `branded`, Rust has `newtype`, Go has nothing standard.

---

## 2. File-Based Repository

**Current Location:** `internal/repo/repository.go`

**Description:** JSON file-based storage with in-memory caching layer, repository pattern implementation.

**Features:**
- Interface-based design (swappable implementations)
- JSON serialization with flat structure
- Simple in-memory cache with LRU eviction
- Search capabilities
- Pagination support
- XDG directory compliance

### Alternatives

| Library | URL | Approach | Pros | Cons |
|---------|-----|----------|------|------|
| **bbolt** | go.etcd.io/bbolt | Embedded KV | ACID, fast | Binary, not human-readable |
| **sqlite** | modernc.org/sqlite | Embedded SQL | Full SQL, mature | SQL complexity |
| **badger** | github.com/dgraph-io/badger | LSM-tree KV | High performance | Complex, overkill for simple use |
| **scribble** | github.com/nanobox-io/golang-scribble | JSON files | Simple, human-readable | Unmaintained, limited features |

### Value-Add Opportunity

**Library Name:** `github.com/larsartmann/filestore`

**Differentiation:**
1. **Git-Native Storage**: Automatic versioning with git commits
   ```go
   store := filestore.New(dir, filestore.WithGitCommits(true))
   ```

2. **Event Sourcing Mode**: Store events as JSONL, project to current state

3. **Multi-Format Support**: JSON, YAML, TOML, MessagePack

4. **Schema Migration**: Automatic version detection and migration

5. **Observability**: Built-in tracing, metrics, query logging

6. **Cache Decorator**: Pluggable cache backends (in-memory, Redis, etc.)

**Market Gap:** No maintained library combines human-readable JSON storage with production features (caching, observability, event sourcing).

---

## 3. Structured Error Handling

**Current Location:** `internal/errors/app_error.go`

**Description:** Rich error types with error codes, HTTP status mapping, and cause chaining.

**Features:**
- Error code enumeration
- HTTP status code mapping
- Error wrapping with `Unwrap()` support
- Constructor functions for common error types
- Details payload for structured error responses

### Alternatives

| Library | URL | Approach | Pros | Cons |
|---------|-----|----------|------|------|
| **cockroachdb/errors** | github.com/cockroachdb/errors | Rich errors | Stack traces, encoding | Heavy dependency |
| **pkg/errors** | github.com/pkg/errors | Error wrapping | Simple, popular | **DEPRECATED** |
| **larsartmann/uniflow** | github.com/larsartmann/uniflow | Railway programming | Composable pipelines | Different paradigm |
| **stdlib errors** | Go 1.13+ | Basic wrapping | Standard, no deps | Minimal features |

### Value-Add Opportunity

**Recommendation:** DON'T extract. Use `cockroachdb/errors` directly.

**Rationale:**
- `cockroachdb/errors` already provides everything we need
- Our current implementation is a subset of its features
- Migration path: Replace `internal/errors` with `cockroachdb/errors` + thin wrapper

**Migration Plan:**
```go
// Instead of custom AppError, use:
import "github.com/cockroachdb/errors"

var ErrComplaintNotFound = errors.New("complaint not found")

func IsComplaintNotFound(err error) bool {
    return errors.Is(err, ErrComplaintNotFound)
}
```

---

## 4. Configuration Management

**Current Location:** `internal/config/config.go`

**Description:** Viper-based configuration with XDG support, environment variables, and validation.

**Features:**
- Multi-source config (file, env, flags)
- XDG Base Directory compliance
- Struct-based configuration with mapstructure
- Post-processing (path expansion)
- Validation with custom rules

### Alternatives

| Library | URL | Approach | Pros | Cons |
|---------|-----|----------|------|------|
| **koanf** | github.com/knadh/koanf | Lightweight | No global state, clean API | Smaller ecosystem |
| **viper** | github.com/spf13/viper | Feature-rich | Widely used, mature | Global state, heavy |
| **envconfig** | github.com/kelseyhightower/envconfig | Env-only | Simple | Limited to env vars |
| **caarlos0/env** | github.com/caarlos0/env/v11 | Struct tags | Clean API | Env-only |

### Value-Add Opportunity

**Recommendation:** DON'T extract. Migrate to `koanf` per HOW_TO_GOLANG.md.

**Rationale:**
- Viper is explicitly banned in HOW_TO_GOLANG.md (global state)
- Koanf provides same features without downsides
- Our wrapper adds no unique value

**Migration Plan:**
```go
// Replace with koanf per HOW_TO_GOLANG.md section 10
import "github.com/knadh/koanf/v2"
```

---

## 5. Tracing Abstraction

**Current Location:** `internal/tracing/*.go`

**Description:** Factory pattern for creating tracers (mock for dev, real for production).

**Features:**
- Interface-based abstraction
- Mock tracer for testing
- Real tracer with Jaeger support
- Context propagation

### Alternatives

| Library | URL | Approach | Pros | Cons |
|---------|-----|----------|------|------|
| **OpenTelemetry** | go.opentelemetry.io/otel | Standard | Industry standard, vendor-neutral | Verbose API |
| **opentracing** | github.com/opentracing/opentracing-go | Deprecated predecessor | - | **DEPRECATED** |
| **jaeger-client** | github.com/jaegertracing/jaeger-client-go | Jaeger-specific | Native Jaeger features | Vendor lock-in |

### Value-Add Opportunity

**Recommendation:** DON'T extract. Use OpenTelemetry SDK directly.

**Rationale:**
- OpenTelemetry is the industry standard
- Our abstraction is too thin to justify a library
- Future: Consider `github.com/larsartmann/otelkit` if we build substantial helpers

---

## 6. MCP Server Framework

**Current Location:** `internal/delivery/mcp/mcp_server.go`

**Description:** Higher-level framework for building MCP servers with tool registration and handlers.

**Features:**
- Tool registration with JSON schema
- Type-safe input/output structs
- Handler function pattern
- Tracing integration
- Structured logging

### Alternatives

| Library | URL | Approach | Pros | Cons |
|---------|-----|----------|------|------|
| **go-sdk** | github.com/modelcontextprotocol/go-sdk | Official SDK | Standard, maintained | Low-level, verbose |
| **mcp-go** | github.com/metoro-io/mcp-golang | Higher-level | Cleaner API | Less mature |
| **custom** | - | Our framework | Tailored to our needs | Maintenance burden |

### Value-Add Opportunity

**Library Name:** `github.com/larsartmann/mcpkit`

**Differentiation:**
1. **Type-Safe Tool Definition**:
   ```go
   type FileComplaintInput struct {
       AgentName string `mcp:"agent_name,required" desc:"Name of the AI agent"`
       Severity  string `mcp:"severity" enum:"low,medium,high,critical"`
   }

   server.Register(FileComplaintInput{}, FileComplaintOutput{}, handler)
   ```

2. **Automatic Schema Generation**: From struct tags using `jsonschema-go`

3. **Middleware Support**: Auth, logging, validation, rate limiting

4. **Transport Abstraction**: stdio, HTTP, WebSocket, SSE

5. **Hot Reload**: Development mode with automatic tool re-registration

6. **Testing Utilities**: Mock server, test helpers, snapshot testing

**Market Gap:** The official MCP SDK is low-level. There's room for a "Gin-like" high-level framework.

**Priority:** **HIGHEST** - MCP ecosystem is nascent, first-mover advantage matters.

---

## 7. Type-Safe Cache Configuration

**Current Location:** `internal/types/cache.go`

**Description:** Type-safe cache size and eviction policy types.

**Features:**
- `CacheSize` type with validation
- `CacheEvictionPolicy` enum
- Bounds checking (Min/Max constants)
- Conversion methods

### Alternatives

| Library | URL | Approach | Pros | Cons |
|---------|-----|----------|------|------|
| **enumer** | github.com/dmarkham/enumer | Code generation | Automatic String(), JSON | Extra build step |
| **go-enum** | github.com/abice/go-enum | Code generation | Rich features | Extra build step |
| **manual** | - | Hand-written | Simple | Repetitive |

### Value-Add Opportunity

**Recommendation:** DON'T extract standalone. Include in `goid` library as "constrained types".

**Extension:**
```go
// In goid library
package goid

type Constrained[T comparable] struct {
    value T
    min   T
    max   T
}

func (c Constrained[T]) Value() T { return c.value }
func (c Constrained[T]) IsValid() bool { return c.value >= c.min && c.value <= c.max }
```

---

## 8. Domain Primitives

**Current Location:** `internal/domain/severity.go`, `internal/domain/complaint.go`

**Description:** Domain-specific types: Severity levels, ResolutionState, validation.

**Features:**
- Severity enum with parsing
- ResolutionState enum
- ValidationError type
- Safe parsing functions

### Alternatives

| Library | URL | Approach | Pros | Cons |
|---------|-----|----------|------|------|
| **enumer** | github.com/dmarkham/enumer | Code generation | Full enum support | Extra build step |
| **go-enum** | github.com/abice/go-enum | Code generation | JSON/SQL support | Extra build step |
| **iota** | Go stdlib | Constant + iota | Standard, simple | Limited features |

### Value-Add Opportunity

**Recommendation:** Include in `goid` library as "validated enums".

**Features:**
```go
// In goid library
type Enum[T ~string] struct {
    allowed []T
    value   T
}

func (e Enum[T]) Parse(s string) (Enum[T], error)
func (e Enum[T]) MustParse(s string) Enum[T]
func (e Enum[T]) Values() []T
func (e Enum[T]) String() string
```

---

## Recommended Extraction Order

### Phase 1: High Priority (Next 4 weeks)

1. **`goid`** - Phantom type system with code generation
   - Start with code generator
   - Add common validators
   - SQL/JSON/OpenAPI support

2. **`mcpkit`** - High-level MCP framework
   - Build on top of official SDK
   - Type-safe tool registration
   - Middleware system

### Phase 2: Medium Priority (Next 8 weeks)

3. **`filestore`** - Production file-based storage
   - Extract from repo/
   - Add git-native mode
   - Event sourcing support

### Phase 3: Consolidation (Ongoing)

4. **Replace internal components:**
   - Replace `internal/errors` → `cockroachdb/errors`
   - Replace `internal/config` → `koanf`
   - Replace `internal/tracing` → OpenTelemetry SDK
   - Consolidate `internal/types` → `goid`

---

## Value Proposition Summary

| Library | Unique Value | Target Users |
|---------|-------------|--------------|
| **goid** | Code-generated phantom types | Teams wanting compile-time safety |
| **mcpkit** | "Gin for MCP" | MCP server developers |
| **filestore** | Human-readable + production features | Local-first apps, CLI tools |

**Strategic Advantage:**
- Early mover in MCP ecosystem
- Fill Go's "branded types" gap
- Opinionated, production-ready defaults

---

## Related Documents

- [HOW_TO_GOLANG.md](/Users/larsartmann/projects/library-policy/HOW_TO_GOLANG.md) - Library selection guidelines
- [CLAUDE.md](/Users/larsartmann/projects/complaints-mcp/CLAUDE.md) - Project architecture
- [PROJECT_SPLIT_EXECUTIVE_REPORT.md](/Users/larsartmann/projects/complaints-mcp/PROJECT_SPLIT_EXECUTIVE_REPORT.md) - Split strategy

---

*Last Updated: 2026-03-03*
