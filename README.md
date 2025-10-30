# complaints-mcp

A tiny MCP (Model Context Protocol) server that allows AI Coding Agents to file complaint reports when they encounter under-specified, confusing, or missing information during development tasks.

## Features

- **Complaint Filing**: AI agents can file structured complaint reports about missing/under-specified/confusing information
- **Dual Storage**: Complaints are stored both locally in the project and globally for user-wide tracking
- **Structured Format**: Uses a standardized markdown template for consistent reporting
- **Timestamped Filenames**: Complaints are automatically organized by date and session

## The Complaint Form

When an AI agent files a complaint, it creates a markdown file with the following structure:

```markdown
# Report about missing/under-specified/confusing information

Date: <ISO_DATE+TimeZone>

I was asked to perform:
<FILL_IN_HERE>

I was given these context information's:
<FILL_IN_HERE>

I was missing these information:
<FILL_IN_HERE>

I was confused by:
<FILL_IN_HERE>

What I wish for the future is:
<FILL_IN_HERE>


Best regards,
<YOUR_NAME>
```

## File Organization

Complaints are stored in:
- **Project-local**: `docs/complaints/<YYYY-MM-DD_HH_MM-SESSION_NAME>.md`
- **Global location**: `~/.complaints-mcp/<YYYY-MM-DD_HH_MM-SESSION_NAME>.md`

## Installation

1. Clone this repository
2. Build the server:
   ```bash
   go build -o complaints-mcp ./cmd/server
   ```

## Usage

Run the MCP server:
```bash
./complaints-mcp
```

The server will start and listen for MCP client connections over stdio transport.

## MCP Tool

The server provides one tool:

### `file_complaint`

Files a complaint report about missing/under-specified/confusing information.

**Parameters:**
- `task_asked_to_perform` (string): What the agent was asked to do
- `context_information` (string): What context information was provided
- `missing_information` (string): What information was missing
- `confused_by` (string): What confused the agent
- `future_wishes` (string): What the agent wishes for in the future
- `session_name` (string, optional): Session name for the filename (auto-generated if not provided)
- `agent_name` (string, optional): Name of the agent filing the complaint

**Returns:**
- Success message with file path where the complaint was saved

## Development

This project uses the official [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk) and follows MCP server best practices.

### Dependencies

- Go 1.21+
- github.com/modelcontextprotocol/go-sdk v1.0.0+

## License

MIT License - see LICENSE file for details.