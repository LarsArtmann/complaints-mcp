# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

complaints-mcp is a Model Context Protocol (MCP) server that allows AI agents to file structured complaint reports when they encounter missing, under-specified, or confusing information during development tasks. It uses the official Go MCP SDK and follows modern Go architectural patterns.

## Build and Test Commands

### Building
```bash
# Build the binary
make build

# Or directly with go build
go build -ldflags="-s -w" -o complaints-mcp ./cmd/server/main.go
```

### Testing
```bash
# Run all tests
make test

# Run BDD tests specifically (uses Ginkgo/Gomega)
go test ./features/bdd/... -v

# Run unit tests for a specific package
go test ./internal/domain -v
go test ./internal/service -v
go test ./internal/repo -v
```

### Running
```bash
# Run the MCP server (stdio transport)
./complaints-mcp

# Run in development mode with enhanced logging
./complaints-mcp --dev --log-level debug

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
   - Provides 4 tools: file_complaint, list_complaints, resolve_complaint, search_complaints
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
   - Environment variables prefixed with COMPLAINTS_MCP_

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
  - mcp_integration_bdd_test.go
- **Unit Tests**: Each internal package has corresponding *_test.go files
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

## Configuration

Configuration is loaded from (in order of precedence):
1. Command-line flags
2. Environment variables (COMPLAINTS_MCP_*)
3. Config file (./config.yaml, ~/.complaints-mcp/config.yaml, /etc/complaints-mcp/config.yaml)
4. XDG config directory ($XDG_CONFIG_HOME/complaints-mcp/config.yaml)
5. Default values

## Internal Package Note

The `internal/` directory contains vendored/customized versions of third-party packages:
- afero, cast, jwalterweatherman, pflag, semver - Customized from spf13 libraries
These are modified to integrate with the project's logging and configuration systems.
