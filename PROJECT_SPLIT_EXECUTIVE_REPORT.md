# Project Split Analysis: complaints-mcp

## Executive Summary

**complaints-mcp** is a well-architected Go-based MCP (Model Context Protocol) server that enables AI coding agents to file structured complaint reports. Built with clean architecture principles, it provides complaint management with dual storage (JSON files + caching), type-safe domain models, and MCP protocol integration. **Splitting is NOT recommended** - the project is a cohesive, single-purpose application with appropriate separation already in place.

## Current Architecture

The project follows clean architecture with clear layer separation:

- **cmd/server/** - Application entry point with Cobra CLI
- **internal/domain/** - Core business entities (Complaint, phantom-typed IDs for type safety)
- **internal/service/** - Business logic orchestration layer
- **internal/repo/** - Data access (FileRepository + SimpleCachedRepository with LRU)
- **internal/delivery/mcp/** - MCP protocol handlers and DTOs
- **internal/config/** - Configuration management (Viper/YAML/ENV)
- **internal/tracing/** - OpenTelemetry tracing abstraction
- **internal/types/** - Shared types (cache, docs)
- **internal/errors/** - Domain-specific errors

## Split Recommendation: NOT RECOMMENDED

### Rationale

The project is a **focused, single-purpose application** serving exactly one domain: complaint management for AI agents via MCP protocol. All components are tightly integrated around this core purpose:

| Aspect                 | Assessment                                     |
| ---------------------- | ---------------------------------------------- |
| Domain Cohesion        | High - single bounded context (complaints)     |
| Component Independence | Low - all layers depend on domain types        |
| Functional Diversity   | Low - one delivery mechanism (MCP)             |
| Codebase Size          | Appropriate (~15 internal packages)            |
| Architecture Quality   | Excellent - clean architecture already applied |

### Potential (But Not Recommended) Extractions

| Component                                | Reason Not Extracted                                          |
| ---------------------------------------- | ------------------------------------------------------------- |
| Caching layer (`SimpleCachedRepository`) | Too coupled to domain types; generic cache libraries exist    |
| Phantom type pattern                     | Pattern is trivial; extraction adds indirection without value |
| Tracing abstraction                      | Already a thin wrapper; extraction premature                  |

### Benefits of NOT Splitting

- Maintains architectural coherence
- Single deployment unit simplifies operations
- Single versioning and release cycle
- Easier onboarding for contributors
- No inter-project dependency management

### Risks of Splitting

- Would break clean architecture by creating artificial boundaries
- Increased complexity with multi-module/go.work setup
- Version compatibility headaches between split packages
- Overhead of maintaining multiple repositories
- No clear independent use-cases for extracted components

## Implementation Path

N/A - Splitting not recommended.

### Alternative Recommendations

1. **Keep current structure** - Architecture is well-designed for the project's scope
2. **Consider module extraction only if**:
   - Caching becomes reusable across multiple projects
   - A REST/GraphQL delivery layer is added alongside MCP
   - Plugin architecture is implemented for extensibility

## Conclusion

**Confidence: HIGH (95%)**

The complaints-mcp project is an exemplar of focused, clean-architecture design. It serves a single purpose (MCP server for complaint management) with appropriate internal layering. Splitting would introduce unnecessary complexity without tangible benefits. The existing internal package structure provides adequate separation of concerns for maintainability and testing.

**Recommendation**: Maintain as a single repository. Monitor for opportunities to extract reusable utilities only if concrete external use-cases emerge.
