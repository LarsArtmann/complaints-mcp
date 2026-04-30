# Complaints MCP - Agent Instructions

## Project Overview

Complaints MCP is a Model Context Protocol (MCP) server for tracking and managing developer complaints about AI assistant interactions. It provides structured storage, search, and resolution tracking for feedback.

## Technology Stack

- **Language:** Go 1.26.1
- **Architecture:** Clean Architecture (Domain, Service, Repository, Delivery layers)
- **Dependencies:**
  - charm.land/log/v2 - Structured logging
  - github.com/larsartmann/go-branded-id - Branded/phantom type IDs
  - github.com/modelcontextprotocol/go-sdk - MCP SDK
  - github.com/go-playground/validator/v10 - Input validation
  - github.com/onsi/ginkgo/v2 - BDD testing

## Project Structure

```
cmd/server/         # Application entry point
internal/
  domain/           # Domain models (Complaint, IDs, Severity)
  service/          # Business logic layer
  repo/             # Data access layer
  delivery/mcp/     # MCP handlers and DTOs
  types/            # Shared types (Cache, Pagination)
  validation/       # Validation utilities
  errors/           # Application errors
  tracing/          # Observability
features/bdd/       # BDD test specifications
```

## Build & Test

```bash
# Build
just build

# Test
just test

# Run all checks
just lint
```

## Key Design Decisions

1. **Phantom Types:** Use branded IDs via `go-branded-id` (ComplaintID, AgentID, etc.) for type safety
2. **Type Safety:** DTOs for all API boundaries
3. **Validation:** go-playground/validator for struct validation
4. **Pagination:** Generic PageRequest/PageResponse types
5. **Error Handling:** Structured AppError with error codes

## Development Guidelines

- Use functional error handling (Result/Option patterns with samber/mo)
- Prefer composition over inheritance
- Add tests for all new functionality
- Follow existing code patterns
- Keep functions small and focused
- Use structured logging with charm.land/log

## Common Tasks

### Adding a new MCP tool

1. Define input/output types in `internal/delivery/mcp/dto.go`
2. Add tool handler in `internal/delivery/mcp/mcp_server.go`
3. Register in `RegisterTools()`
4. Add BDD tests in `features/bdd/`

### Adding domain types

1. Define in `internal/domain/`
2. Use phantom types for IDs
3. Add validation methods
4. Add comprehensive tests

### Repository operations

1. Add method to `internal/repo/repository.go` interface
2. Implement in `FileRepository`
3. Use service layer for business logic

## Configuration

Server configuration via:

- Environment variables
- Config file (YAML/TOML)
- CLI flags

See `internal/config/config.go` for options.
