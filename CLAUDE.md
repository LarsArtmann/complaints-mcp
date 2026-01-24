# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

complaints-mcp is a Model Context Protocol (MCP) server that allows AI agents to file structured complaint reports when they encounter missing, under-specified, or confusing information during development tasks. It uses the official Go MCP SDK and follows modern Go architectural patterns.

## Build and Test Commands

This project uses **Just** (justfile) as its modern task runner, providing better cross-platform support and more features than traditional Makefiles.

### Building

```bash
# Build the binary
just build

# Or directly with go build
go build -ldflags="-s -w" -o complaints-mcp ./cmd/server/main.go
```

### Testing

```bash
# Run all tests
just test

# Run tests with verbose output
just test-verbose

# Run BDD tests specifically (uses Ginkgo/Gomega)
just test-bdd

# Run tests with coverage report
just test-coverage

# Run unit tests for a specific package
go test ./internal/domain -v
go test ./internal/service -v
go test ./internal/repo -v

# Run benchmarks
just bench
```

### Development Workflow

```bash
# Run full CI pipeline locally
just ci

# Development with hot reload (requires air)
just dev-watch

# Format code
just fmt

# Lint code
just lint

# Run code quality checks
just quality

# Install development tools
just install-tools

# View all available commands
just --list
```

### Why Just over Make?

This project migrated from Makefile to Justfile for better cross-platform compatibility and modern development features:

- **Cross-platform**: Works consistently on Windows, macOS, and Linux
- **Better syntax**: No complex Makefile rules or tab/space issues
- **Self-documenting**: Built-in help with `just --list`
- **More features**: Enhanced error handling, private recipes, dependencies
- **Modern**: Designed for contemporary development workflows

**Installation**:

```bash
# macOS/Linux
brew install just

# Or from source
cargo install just

# Verify installation
just --version
```

**Legacy Make Support**:
The Makefile remains for backward compatibility but Just is recommended. All Make targets have equivalent Just commands.

### Running

```bash
# Run the MCP server (stdio transport)
./complaints-mcp

# Run in development mode with enhanced logging
./complaints-mcp --dev --log-level debug

# Run with cache disabled for debugging
./complaints-mcp --cache-enabled=false

# Run with custom cache size (default: 1000)
./complaints-mcp --cache-max-size=500

# Show version information
./complaints-mcp --version
```

### Linting

```bash
make lint
```

## Architecture

This project follows a clean, layered architecture with clear separation of concerns:

### Layer Structure

1. **cmd/server** - Application entry point
   - main.go: CLI setup with Cobra, dependency injection, graceful shutdown

2. **internal/delivery/mcp** - MCP Protocol Layer
   - Implements MCP server using the official go-sdk
   - Provides 5 tools: file_complaint, list_complaints, resolve_complaint, search_complaints, get_cache_stats
   - Handles JSON schema validation and tool registration
   - Uses stdio transport for communication

3. **internal/service** - Business Logic Layer
   - ComplaintService: Core business logic, orchestrates domain and repository operations
   - Enforces business rules and validation
   - All methods are traced for observability

4. **internal/domain** - Domain Model Layer
   - Complaint entity with validation using go-playground/validator
   - Severity enum (low, medium, high, critical)
   - ComplaintID value object using UUID v4
   - Pure domain logic with no external dependencies (except logging)

5. **internal/repo** - Data Access Layer
   - Repository interface defines storage contract
   - FileRepository: Stores complaints as JSON files with timestamp-based naming
   - Supports pagination, severity filtering, and text search

6. **internal/config** - Configuration Management
   - Uses Viper for configuration loading (YAML, env vars, CLI flags)
   - XDG Base Directory compliant
   - Environment variables prefixed with COMPLAINTS*MCP*

7. **internal/tracing** - Observability
   - MockTracer for development (production-ready tracer interface exists)
   - All service and repository methods create spans

8. **internal/errors** - Custom Error Types
   - Domain-specific error types for better error handling

### Key Dependencies

- **charmbracelet/log**: Modern structured logging (replaced zerolog)
- **modelcontextprotocol/go-sdk**: Official MCP protocol implementation
- **spf13/cobra**: CLI framework
- **spf13/viper**: Configuration management
- **go-playground/validator**: Struct validation
- **ginkgo/gomega**: BDD testing framework

### Testing Strategy

- **BDD Tests** (features/bdd/): Behavior-driven tests using Ginkgo/Gomega
  - complaint_filing_bdd_test.go
  - complaint_listing_bdd_test.go
  - complaint_resolution_bdd_test.go
  - cache_stats_bdd_test.go (NEW!)
  - mcp_integration_bdd_test.go
- **Unit Tests**: Each internal package has corresponding \*\_test.go files
- Test files use table-driven tests where appropriate

## Important Patterns

### Context Propagation

Context is passed through all layers for cancellation and timeout support. Logger is attached to context using charmbracelet/log.WithContext().

### Error Handling

Errors are wrapped with context using fmt.Errorf("description: %w", err) to maintain error chains.

### Logging

Use structured logging with key-value pairs:

```go
logger.Info("message", "key", value, "another_key", another_value)
logger.With("persistent_key", value) // Creates child logger
```

### Dependency Injection

Dependencies are injected through constructors (NewComplaintService, NewServer, etc.). No global state except for the logger default.

### File Storage Format

Complaints are stored as JSON files with naming pattern:

- `YYYY-MM-DD_HH-MM-SS-<session_name>.json`
- `YYYY-MM-DD_HH-MM-SS.json` (if no session name)

### MCP Tool Implementation

Tools are registered with schema definitions and handler functions. Handlers use type-safe input/output structs and return (result, output, error).

#### Available Tools:

1. **file_complaint** - File a new structured complaint
2. **list_complaints** - List all complaints with pagination
3. **resolve_complaint** - Mark a complaint as resolved
4. **search_complaints** - Search complaints by content
5. **get_cache_stats** - Get cache performance statistics (NEW!)

#### get_cache_stats Tool Details:

- **Purpose**: Monitor LRU cache performance metrics
- **Input**: No parameters required
- **Output**: JSON with cache statistics:
  ```json
  {
    "cache_enabled": true,
    "stats": {
      "hits": 42,
      "misses": 8,
      "evictions": 3,
      "current_size": 15,
      "max_size": 1000,
      "hit_rate_percent": 84.0
    },
    "message": "Cache statistics retrieved successfully"
  }
  ```
- **Usage**: Call `get_cache_stats` tool to monitor cache health and performance

## Configuration

Configuration is loaded from (in order of precedence):

1. Command-line flags
2. Environment variables (COMPLAINTS*MCP*\*)
3. Config file (./config.yaml, ~/.complaints-mcp/config.yaml, /etc/complaints-mcp/config.yaml)
4. XDG config directory ($XDG_CONFIG_HOME/complaints-mcp/config.yaml)
5. Default values

### Cache Configuration

The repository factory supports configurable caching:

```yaml
storage:
  cache_enabled: true # Use CachedRepository vs FileRepository
  cache_max_size: 1000 # LRU cache size
  cache_eviction: "lru" # Eviction policy (lru, fifo, none)
```

CLI Flags:

```bash
--cache-enabled=true        # Enable/disable cache (default: true)
--cache-max-size=1000       # Set cache size (default: 1000)
--cache-eviction=lru        # Set eviction policy (default: lru)
```

Environment Variables:

```bash
COMPLAINTS_MCP_CACHE_ENABLED=true
COMPLAINTS_MCP_CACHE_MAX_SIZE=1000
COMPLAINTS_MCP_CACHE_EVICTION=lru
```

**Repository Selection Logic:**

- `cache_enabled=true` → Creates CachedRepository (LRU cache for O(1) lookups)
- `cache_enabled=false` → Creates FileRepository (no cache, direct file I/O)
- Explicit `type="file"` always forces FileRepository regardless of cache settings

**Performance Impact:**

- CachedRepository: ~1000x faster for repeated lookups (O(1) vs O(n))
- FileRepository: Lower memory usage, always reads from disk

## Internal Package Note

The `internal/` directory contains vendored/customized versions of third-party packages:

- afero, cast, jwalterweatherman, pflag, semver - Customized from spf13 libraries
  These are modified to integrate with the project's logging and configuration systems.
