# Crush Integration for complaints-mcp

This file contains Crush-specific commands and configuration for the complaints-mcp project.

## Project Overview

The complaints-mcp is a Model Context Protocol (MCP) server that allows AI coding agents to file structured complaint reports about missing, under-specified, or confusing information they encounter during development tasks.

## Crush Configuration

Add this MCP server to your Crush configuration:

```json
{
  "$schema": "https://charm.land/crush.json",
  "mcp": {
    "complaints": {
      "type": "stdio",
      "command": "/path/to/complaints-mcp",
      "args": [],
      "timeout": 120,
      "disabled": false
    }
  }
}
```

## Building the Server

```bash
# Clone and build
git clone https://github.com/LarsArtmann/complaints-mcp.git
cd complaints-mcp
go build -o complaints-mcp ./cmd/server

# Test the build
./complaints-mcp --help
```

## Testing with Crush

After adding to Crush configuration, test the integration:

1. Start Crush in a project directory
2. Ask the AI to do something with unclear instructions
3. The AI should use the `file_complaint` tool when it encounters issues

## Common Usage Patterns

### When AI Agents Should File Complaints

- **Missing Documentation**: When setup instructions are incomplete
- **Unclear Requirements**: When task specifications are ambiguous
- **Confusing Code**: When code structure or naming is unclear
- **Missing Context**: When additional information is needed
- **Tool Issues**: When development tools are not working as expected

### Example Complaint Triggers

```
"I notice the README doesn't specify which Node.js version is required."
"The environment variables mentioned in .env.example are not documented."
"The function signature lacks information about expected input formats."
"I'm confused about the project structure - there's no architectural overview."
```

## File Locations

Complaints are stored in two locations:

1. **Project-local**: `docs/complaints/<YYYY-MM-DD_HH_MM-SESSION_NAME>.md`
2. **Global**: `~/.complaints-mcp/<PROJECT_NAME>/<YYYY-MM-DD_HH_MM-SESSION_NAME>.md`

The project name is detected from:
1. Git remote repository name
2. Folder name (fallback)
3. "unknown-project" (ultimate fallback)

## Integration Benefits

- **Continuous Feedback Loop**: AI agents can systematically report issues
- **Documentation Improvement**: Identifies areas needing better documentation
- **Developer Experience**: Helps improve onboarding and project clarity
- **Quality Assurance**: Catches confusing or incomplete instructions

## Troubleshooting

### Server Not Starting
```bash
# Check if binary exists and is executable
ls -la /path/to/complaints-mcp

# Test manually
/path/to/complaints-mcp
```

### Tool Not Available in Crush
- Check Crush configuration syntax
- Verify binary path is correct
- Restart Crush after configuration changes
- Check Crush logs for MCP connection errors

### Permissions Issues
```bash
# Make binary executable
chmod +x /path/to/complaints-mcp

# Check global directory permissions
ls -la ~/.complaints-mcp/
```

## Development Commands

### Using Just (Recommended)
```bash
# Build the server
just build

# Build optimized release version
just build-release

# Run tests
just test

# Run BDD tests (when implemented)
just test-bdd

# Lint and format code
just lint
just fmt

# Find code duplicates
just fd          # Standard threshold (15 tokens)
just fd-strict   # Strict threshold (50 tokens)

# Install development tools
just install-tools

# Run full CI pipeline
just ci

# Clean build artifacts
just clean

# Development server
just dev

# Update dependencies
just update

# Show all available commands
just help
```

### Using Go directly
```bash
# Run tests (if any)
go test ./...

# Build for development
go build -o complaints-mcp ./cmd/server

# Build for release
go build -ldflags="-s -w" -o complaints-mcp ./cmd/server

# Clean build artifacts
go clean
```