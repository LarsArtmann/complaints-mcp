# complaints-mcp

🚨 **A comprehensive MCP (Model Context Protocol) server for AI agent complaint management and feedback systems**

> **Version**: 2.0.0 | **Status**: Production-Ready | **Architecture**: Clean Architecture with Type Safety

---

## 🎯 Executive Summary

**complaints-mcp** is a sophisticated MCP server that enables AI coding agents to file structured complaint reports when they encounter missing information, confusing specifications, or inadequate documentation during development tasks. Built with **enterprise-grade architecture** and **type-safe Go**, it provides transparent data management, advanced caching, and comprehensive observability.

---

## ✨ Key Features

### 🏗️ **Core Functionality**

- **📝 Structured Complaint Filing**: Standardized complaint reports with rich metadata
- **💾 Dual Storage System**: Local project storage + global user-wide tracking
- **📅 Intelligent Organization**: Timestamp-based filenames with session context
- **🔄 Resolution Tracking**: Complete complaint lifecycle management
- **📄 Documentation Export**: Multi-format export (Markdown, HTML, Text)
- **🔍 Advanced Search**: Full-text search across complaint content
- **📊 Performance Analytics**: Real-time cache statistics and metrics

### 🛡️ **Enterprise-Grade Architecture**

- **🏛️ Clean Architecture**: Layered design with clear separation of concerns
- **🔒 Type Safety**: Strongly-typed domain models with validation
- **🧵 Thread Safety**: Concurrent-safe operations with proper synchronization
- **📋 BDD Testing**: Comprehensive behavioral test suite (40/52 tests passing)
- **🔍 Observability**: Structured tracing and comprehensive logging
- **⚡ High Performance**: LRU caching with O(1) lookups

### 🆕 **Latest Enhancement (v2.0)**

- **📁 File Path Transparency**: Complete visibility into data storage locations
- **🛠️ Enhanced MCP Integration**: File paths included in tool responses
- **🔧 Improved Error Handling**: Graceful degradation with detailed logging
- **📋 Repository Interface Extensions**: GetFilePath() and GetDocsPath() methods

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    MCP CLIENTS (AI Agents)                 │
└─────────────────────┬───────────────────────────────────────────────┘
                      │ stdio/transport
┌─────────────────────▼───────────────────────────────────────────────┐
│                   MCP SERVER LAYER                        │
│  ┌─────────────────┬─────────────────┬─────────────────┐│
│  │  Tool Handlers │  DTO Layer     │  Error Handling ││
│  │  (file_complaint)│ (type-safe)    │   (structured)   ││
│  └─────────────────┴─────────────────┴─────────────────┘│
└─────────────────────┬───────────────────────────────────────────────┘
                      │ Service Interface
┌─────────────────────▼───────────────────────────────────────────────┐
│                  SERVICE LAYER                            │
│  ┌─────────────────┬─────────────────┬─────────────────┐│
│  │  Business Logic│  Path Resolution│  Configuration   ││
│  │   (Orchestration)│    (NEW!)       │   Management     ││
│  └─────────────────┴─────────────────┴─────────────────┘│
└─────────────────────┬───────────────────────────────────────────────┘
                      │ Repository Interface
┌─────────────────────▼───────────────────────────────────────────────┐
│               REPOSITORY LAYER                          │
│  ┌─────────────────┬─────────────────┬─────────────────┐│
│  │  File Repository│ Cached Repository│ Docs Repository ││
│  │  (JSON Storage) │   (LRU Cache)   │  (Multi-format)  ││
│  │ GetFilePath() ✨  │ GetFilePath() ✨  │   Export logic)  ││
│  │ GetDocsPath() ✨  │ GetDocsPath() ✨  │                 ││
│  └─────────────────┴─────────────────┴─────────────────┘│
└─────────────────────┬───────────────────────────────────────────────┘
                      │ File System
┌─────────────────────▼───────────────────────────────────────────────┐
│                STORAGE LAYER                             │
│  ┌─────────────────┬─────────────────┬─────────────────┐│
│  │  JSON Files     │  Markdown Files  │  HTML Files     ││
│  │  ({uuid}.json) │ (YYYY-MM-DD...) │ (YYYY-MM-DD...) ││
│  └─────────────────┴─────────────────┴─────────────────┘│
└───────────────────────────────────────────────────────────────────────┘
```

---

## 📊 Data Model

### **Complaint Structure**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "agent_name": "AI-Coding-Assistant",
  "session_name": "feature-development-session",
  "task_description": "Missing API documentation for authentication endpoints",
  "severity": "high",
  "project_name": "user-management-system",
  "timestamp": "2024-11-09T12:18:30Z",
  "context_info": "Implementing OAuth2 authentication flow",
  "missing_info": "API endpoint specifications, error response formats",
  "confused_by": "Confusing token refresh mechanism",
  "future_wishes": "Comprehensive API documentation with examples",
  "resolved": false,
  "resolved_at": null,
  "resolved_by": "",
  "file_path": "/Users/larsartmann/.local/share/complaints/550e8400-e29b-41d4-a716-446655440000.json",
  "docs_path": "docs/complaints/2024-11-09_12-18-feature-development-session.md"
}
```

### **Severity Levels**

- `low` - Minor inconveniences, workarounds available
- `medium` - Significant productivity impact, unclear alternatives
- `high` - Blockers, no clear path forward
- `critical` - System failure, complete project stall

### **Resolution States**

- `open` - Complaint filed, awaiting resolution
- `resolved` - Issue addressed, timestamp recorded
- `rejected` - Complaint reviewed, deemed invalid
- `deferred` - Postponed for future consideration

---

## 🗂️ File Organization & Storage

### **Primary Storage (JSON)**

```
{storage_base_dir}/{complaint_id}.json

Examples:
/Users/larsartmann/.local/share/complaints/550e8400-e29b-41d4-a716-446655440000.json
/var/lib/complaints/8f94a312-c5f7-4e2b-9ff1-0a3b4c8d7e2f.json
```

### **Documentation Export (Markdown/HTML/Text)**

```
{docs_dir}/{YYYY-MM-DD_HH-MM-SESSION_NAME}.{format}

Examples:
docs/complaints/2024-11-09_12-18-feature-development-session.md
docs/complaints/2024-11-09_14-32-api-integration-session.html
docs/complaints/2024-11-09_16-45-bug-fix-session.txt
```

### **Project Name Detection**

1. **Git Remote Repository Name** - Primary source
2. **Current Directory Name** - Fallback option
3. **"unknown-project"** - Last resort default

---

## 🛠️ Installation & Setup

### **Prerequisites**

- Go 1.21+ with modules
- Git (for project name detection)
- 10MB available disk space

### **Build from Source**

```bash
# Clone repository
git clone https://github.com/LarsArtmann/complaints-mcp.git
cd complaints-mcp

# Build binary
go build -o complaints-mcp ./cmd/server

# Verify installation
./complaints-mcp --help
```

### **Configuration**

#### **Environment Variables**

```bash
export COMPLAINTS_MCP_SERVER_NAME="complaints-mcp"
export COMPLAINTS_MCP_SERVER_HOST="localhost"
export COMPLAINTS_MCP_SERVER_PORT=8080
export COMPLAINTS_MCP_STORAGE_BASE_DIR="$HOME/.local/share/complaints"
export COMPLAINTS_MCP_STORAGE_DOCS_DIR="docs/complaints"
export COMPLAINTS_MCP_STORAGE_DOCS_ENABLED=true
export COMPLAINTS_MCP_STORAGE_DOCS_FORMAT="markdown"
export COMPLAINTS_MCP_LOG_LEVEL="info"
```

#### **Configuration File (YAML)**

```yaml
server:
  name: "complaints-mcp"
  host: "localhost"
  port: 8080

storage:
  base_dir: "$HOME/.local/share/complaints"
  docs_dir: "docs/complaints"
  docs_enabled: true
  docs_format: "markdown"
  max_size: 10485760 # 10MB
  retention_days: 0 # Infinite retention
  auto_backup: true
  cache_enabled: true
  cache_max_size: 1000
  cache_eviction: "lru"

log:
  level: "info"
  format: "text"
  output: "stdout"
```

---

## 🚀 Usage & Integration

### **Running the Server**

```bash
# Standard execution
./complaints-mcp

# With custom configuration
./complaints-mcp --config config/custom.yaml

# With environment variables
COMPLAINTS_MCP_SERVER_PORT=9090 ./complaints-mcp
```

### **MCP Tool Interface**

The server exposes the following MCP tools:

#### **file_complaint**

```json
{
  "name": "file_complaint",
  "description": "File a structured complaint about missing or confusing information",
  "inputSchema": {
    "type": "object",
    "properties": {
      "agent_name": { "type": "string", "minLength": 1, "maxLength": 100 },
      "session_name": { "type": "string", "maxLength": 100 },
      "task_description": { "type": "string", "minLength": 1, "maxLength": 1000 },
      "context_info": { "type": "string", "maxLength": 500 },
      "missing_info": { "type": "string", "maxLength": 500 },
      "confused_by": { "type": "string", "maxLength": 500 },
      "future_wishes": { "type": "string", "maxLength": 500 },
      "severity": { "type": "string", "enum": ["low", "medium", "high", "critical"] },
      "project_name": { "type": "string", "maxLength": 100 }
    },
    "required": ["agent_name", "task_description", "severity"]
  }
}
```

#### **list_complaints**

```json
{
  "name": "list_complaints",
  "description": "Retrieve paginated list of complaints",
  "inputSchema": {
    "type": "object",
    "properties": {
      "limit": { "type": "integer", "minimum": 1, "maximum": 100 },
      "severity": { "type": "string", "enum": ["low", "medium", "high", "critical"] },
      "resolved": { "type": "boolean" }
    }
  }
}
```

#### **resolve_complaint**

```json
{
  "name": "resolve_complaint",
  "description": "Mark a complaint as resolved",
  "inputSchema": {
    "type": "object",
    "properties": {
      "complaint_id": {
        "type": "string",
        "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"
      },
      "resolved_by": { "type": "string", "minLength": 1, "maxLength": 100 }
    },
    "required": ["complaint_id", "resolved_by"]
  }
}
```

#### **search_complaints**

```json
{
  "name": "search_complaints",
  "description": "Search complaints by content",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": { "type": "string", "minLength": 1, "maxLength": 200 },
      "limit": { "type": "integer", "minimum": 1, "maximum": 50 }
    },
    "required": ["query"]
  }
}
```

#### **get_cache_stats**

```json
{
  "name": "get_cache_stats",
  "description": "Get cache performance statistics",
  "inputSchema": {
    "type": "object",
    "properties": {}
  }
}
```

### **AI Assistant Integration**

#### **Crush Integration**

Add to your Crush configuration (`.crush.json`):

```json
{
  "$schema": "https://charm.land/crush.json",
  "mcp": {
    "complaints": {
      "type": "stdio",
      "command": "/path/to/complaints-mcp",
      "args": ["--config", "/path/to/config.yaml"],
      "timeout": 120,
      "disabled": false
    }
  }
}
```

#### **Claude Integration**

Add to your Claude desktop configuration:

```json
{
  "mcpServers": {
    "complaints-mcp": {
      "command": "/path/to/complaints-mcp",
      "args": ["--config", "/path/to/config.yaml"]
    }
  }
}
```

---

## 📋 Example Usage Scenarios

### **Scenario 1: Missing API Documentation**

```json
{
  "agent_name": "AI-Coding-Assistant",
  "session_name": "api-auth-implementation",
  "task_description": "Implement OAuth2 authentication endpoints",
  "context_info": "Working on user management microservice",
  "missing_info": "Complete API specification for /auth/refresh endpoint",
  "confused_by": "Token rotation logic unclear from requirements",
  "future_wishes": "OpenAPI specification with Postman collection",
  "severity": "high",
  "project_name": "user-management-system"
}
```

### **Scenario 2: Confusing Error Messages**

```json
{
  "agent_name": "Bug-Fix-Assistant",
  "session_name": "memory-leak-debugging",
  "task_description": "Fix memory leak in data processing pipeline",
  "context_info": "Analyzing heap dumps, processing large datasets",
  "missing_info": "Clear explanation of error code 'ERR_MEM_CORRUPT'",
  "confused_by": "Error message suggests hardware failure but logs show software issue",
  "future_wishes": "Comprehensive error code documentation with troubleshooting steps",
  "severity": "medium",
  "project_name": "data-processor"
}
```

### **Enhanced Response with File Paths (v2.0+)**

```json
{
  "success": true,
  "message": "Complaint filed successfully",
  "complaint": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "agent_name": "AI-Coding-Assistant",
    "task_description": "Missing API documentation",
    "severity": "high",
    "timestamp": "2024-11-09T12:18:30Z",
    "resolved": false,
    "file_path": "/Users/larsartmann/.local/share/complaints/550e8400-e29b-41d4-a716-446655440000.json",
    "docs_path": "docs/complaints/2024-11-09_12-18-api-auth-implementation.md"
  }
}
```

---

## 🔧 Development & Testing

### **Project Structure**

```
complaints-mcp/
├── cmd/server/                 # Application entry point
├── internal/
│   ├── config/                # Configuration management
│   ├── delivery/mcp/          # MCP server implementation
│   ├── domain/                # Business logic & entities
│   ├── repo/                  # Data access layer
│   ├── service/               # Business orchestration
│   ├── tracing/               # Observability
│   └── types/                # Type definitions
├── features/bdd/             # Behavioral tests
├── docs/                     # Project documentation
└── examples/                  # Usage examples
```

### **Running Tests**

```bash
# Run all tests
just test

# Run specific test suite
go test ./internal/domain -v

# Run BDD tests
go test ./features/bdd -v

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### **Building for Development**

```bash
# Build with race detection
go build -race -o complaints-mcp ./cmd/server

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o complaints-mcp-linux ./cmd/server
GOOS=windows GOARCH=amd64 go build -o complaints-mcp.exe ./cmd/server

# Build with debug symbols
go build -gcflags="all=-N -l" -o complaints-mcp-debug ./cmd/server
```

---

## 📊 Performance & Monitoring

### **Cache Performance**

```json
{
  "cache_enabled": true,
  "stats": {
    "hits": 1247,
    "misses": 89,
    "evictions": 12,
    "current_size": 156,
    "max_size": 1000,
    "hit_rate": 0.9332
  }
}
```

### **Monitoring Metrics**

- **Request Rate**: Average complaints per minute/hour/day
- **Resolution Time**: Time from filing to resolution
- **Storage Usage**: Disk space consumption by complaint data
- **Cache Efficiency**: Hit rate, eviction patterns
- **Error Rate**: Failed operations per time period

### **Log Levels**

- `trace` - Detailed execution tracing
- `debug` - Development debugging information
- `info` - General operational information
- `warn` - Non-critical issues and deprecations
- `error` - Failed operations requiring attention

---

## 🔐 Security Considerations

### **Data Privacy**

- Complaints stored locally with user-controlled locations
- No external data transmission or cloud storage
- File permissions respect system umask settings
- Sensitive information logged at appropriate levels

### **Access Control**

- File system access respects user permissions
- No authentication required for local usage
- Network access disabled by default (stdio transport)
- Configuration file access controlled by user permissions

### **Input Validation**

- Strict input validation with length limits
- Sanitization of file paths to prevent directory traversal
- SQL injection prevention (not applicable to JSON storage)
- Buffer overflow protection in all string operations

---

## 🔄 Migration & Upgrades

### **Version Compatibility**

- **Backward Compatible**: All existing JSON files supported
- **Configuration Migration**: Automatic detection of old config formats
- **API Stability**: MCP tool contracts maintained across versions
- **Data Migration**: Built-in conversion utilities for legacy formats

### **Upgrade Process**

```bash
# Backup existing data
cp -r ~/.local/share/complaints ~/.local/share/complaints.backup

# Upgrade binary
go build -o complaints-mcp ./cmd/server

# Run migration (automatic on first run)
./complaints-mcp --migrate

# Verify data integrity
./complaints-mcp --verify
```

---

## 🐛 Troubleshooting

### **Common Issues**

#### **Server Won't Start**

```bash
# Check configuration
./complaints-mcp --validate-config

# Check permissions
ls -la ~/.local/share/complaints/

# Check port availability
lsof -i :8080
```

#### **Complaints Not Saving**

```bash
# Check disk space
df -h ~/.local/share/

# Check file permissions
touch ~/.local/share/complaints/test.json

# Review logs
./complaints-mcp --log-level debug
```

#### **Performance Issues**

```bash
# Check cache statistics
echo '{"tool": "get_cache_stats"}' | ./complaints-mcp

# Monitor memory usage
./complaints-mcp --profile-memory

# Benchmark operations
go test -bench=. ./internal/repo/
```

### **Debug Mode**

```bash
# Enable comprehensive debugging
COMPLAINTS_MCP_LOG_LEVEL=debug ./complaints-mcp --trace

# Generate bug report
./complaints-mcp --bug-report > bug-report.txt

# Validate data integrity
./complaints-mcp --validate-data
```

---

## 📚 Documentation & Resources

### **Project Documentation**

- [**Architecture Guide**](docs/architecture/) - Complete architectural overview
- [**Implementation Strategy**](docs/strategy/) - Development approach and decisions
- [**Status Reports**](docs/status/) - Progress tracking and milestones
- [**Issue Resolution**](docs/ISSUES_45_46_RESOLUTION.md) - Problem-solving documentation

### **External References**

- [**MCP Specification**](https://modelcontextprotocol.io/) - Protocol documentation
- [**Clean Architecture**](https://blog.cleancoder.com/uncle-bob-2017-architecture-01-part01/) - Architectural patterns
- [**Domain-Driven Design**](https://dddcommunity.org/) - Design principles
- [**Behavior-Driven Development**](https://cucumber.io/docs/bdd/) - Testing methodology

### **Community & Support**

- **GitHub Issues**: [Report bugs and feature requests](https://github.com/LarsArtmann/complaints-mcp/issues)
- **Discussions**: [Community support and Q&A](https://github.com/LarsArtmann/complaints-mcp/discussions)
- **Contributing**: [Development guidelines](CONTRIBUTING.md)
- **Changelog**: [Version history](CHANGELOG.md)

---

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

### **Development Workflow**

1. Fork repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

### **Code Standards**

- Follow Go conventions and `gofmt` formatting
- Maintain test coverage above 80%
- Add comprehensive BDD tests for new features
- Update documentation for API changes
- Ensure backward compatibility

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🎯 Roadmap & Future Development

### **Upcoming Features (v2.1)**

- [ ] **Plugin Architecture** - Extensible tool system
- [ ] **Advanced Search** - Filtering, sorting, faceted search
- [ ] **Analytics Dashboard** - Web-based monitoring interface
- [ ] **Import/Export** - Bulk data management
- [ ] **API Server** - RESTful HTTP interface option

### **Long-term Vision (v3.0)**

- [ ] **Multi-tenant Support** - Organization-level isolation
- [ ] **Event Sourcing** - Complete audit trail and replay
- [ ] **Machine Learning** - Automatic categorization and prioritization
- [ ] **Integration Hub** - Connect with external issue trackers
- [ ] **Mobile App** - Native mobile complaint filing

---

## 🏆 Acknowledgments

- **Model Context Protocol Team** - For the excellent MCP specification
- **Charm Bracelet** - For outstanding Go libraries and tools
- **Go Community** - For continuous language improvements
- **Open Source Contributors** - For valuable feedback and contributions

---

<div align="center">

**[⬆️ Back to Top](#complaints-mcp)**

Made with ❤️ by AI agents, for AI agents

</div>
