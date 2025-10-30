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
- **Global location**: `~/.complaints-mcp/<PROJECT_NAME>/<YYYY-MM-DD_HH_MM-SESSION_NAME>.md`

The global location includes the project name, which is determined by:
1. First checking the git remote repository name
2. Falling back to the folder name if no git remote exists
3. Using "unknown-project" as a last resort

## Installation

1. Clone this repository
2. Build the server:
   ```bash
   go build -o complaints-mcp ./cmd/server
   ```

## Usage

### Running the Server

Run the MCP server:
```bash
./complaints-mcp
```

The server will start and listen for MCP client connections over stdio transport.

## How to Use with Crush

[Crush](https://github.com/charmbracelet/crush) is an AI coding assistant that supports MCP servers out of the box. To use complaints-mcp with Crush:

### 1. Add to Crush Configuration

Add the following to your Crush configuration file (`.crush.json`, `crush.json`, or `$HOME/.config/crush/crush.json`):

```json
{
  "$schema": "https://charm.land/crush.json",
  "mcp": {
    "complaints": {
      "type": "stdio",
      "command": "/path/to/your/complaints-mcp",
      "args": [],
      "timeout": 120,
      "disabled": false
    }
  }
}
```

### 2. Build and Install

First, build the complaints-mcp binary:

```bash
git clone https://github.com/LarsArtmann/complaints-mcp.git
cd complaints-mcp
go build -o complaints-mcp ./cmd/server
```

Then update the `command` path in your Crush configuration to point to the built binary.

### 3. Restart Crush

Restart Crush or reload the configuration. The `file_complaint` tool will now be available to AI agents working through Crush.

### 4. Example Usage

When working with Crush, an AI agent can now file complaints whenever it encounters issues:

```
Agent: I notice there are missing environment variables in the setup instructions.
Agent: Let me file a complaint about this.
[Agent calls file_complaint tool]
```

The complaint will be saved to both:
- Project-local: `docs/complaints/<timestamp>-session.md`
- Global: `~/.complaints-mcp/<project-name>/<timestamp>-session.md`

### 5. Benefits

- **Better AI Experience**: Helps identify areas where documentation or instructions are unclear
- **Continuous Improvement**: Collects structured feedback about project pain points
- **Multi-project Organization**: Global storage is organized by project name
- **Historical Tracking**: Maintains a record of issues encountered over time

This integration ensures that AI agents can provide valuable feedback about project documentation, setup processes, and other areas that need improvement.

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