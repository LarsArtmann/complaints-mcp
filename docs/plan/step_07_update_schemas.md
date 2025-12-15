# Step 7: Update MCP Tool Schemas for Flat ID Structure

## 🎯 Objective

Update all MCP tool schemas to expect flat ID strings instead of nested objects.

## 🚨 Critical Issue

Current schemas expect nested structure:

```json
{
  "complaint_id": {
    "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"
  }
}
```

But phantom types produce flat structure:

```json
{
  "complaint_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## 🏗️ Implementation Tasks

### A. Update Tool Input Schemas

#### resolve_complaint Tool

```go
// internal/delivery/mcp/mcp_server.go (updated)

resolveComplaintTool := &mcp.Tool{
    Name: "resolve_complaint",
    Description: "Mark a complaint as resolved",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "complaint_id": map[string]any{
                "type": "string",
                "description": "Unique identifier of the complaint in UUID v4 format (flat string)",
                "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
                "examples": []string{
                    "550e8400-e29b-41d4-a716-446655440000",
                    "9cb3bb9e-b6dc-4e02-9767-e396a42b63a6",
                },
            },
            "resolved_by": map[string]any{
                "type": "string",
                "description": "Identifier of who resolved the complaint",
                "minLength": 1,
                "maxLength": 100,
                "examples": []string{
                    "AI-Assistant",
                    "DevOps-Team",
                    "Human-Reviewer",
                },
            },
        },
        "required": []string{"complaint_id", "resolved_by"},
    },
}
```

#### file_complaint Tool

```go
fileComplaintTool := &mcp.Tool{
    Name: "file_complaint",
    Description: "File a structured complaint about missing or confusing information. Project name is automatically detected if not provided.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "agent_name": map[string]any{
                "type": "string",
                "description": "Name of the AI agent filing the complaint",
                "minLength": 1,
                "maxLength": 100,
                "pattern": "^[a-zA-Z0-9\\-\\_\\s]{1,100}$",
                "examples": []string{
                    "AI-Coding-Assistant",
                    "Code-Reviewer-Bot",
                    "DevOps-Automation",
                },
            },
            "project_name": map[string]any{
                "type": "string",
                "description": "Name of the project (optional - auto-detected from git repository if not provided)",
                "maxLength": 100,
                "pattern": "^[a-zA-Z0-9\\-\\_\\s\\.]{1,100}$",
                "examples": []string{
                    "my-project",  // Manual specification
                    "",           // Use auto-detection
                },
            },
            "session_name": map[string]any{
                "type": "string",
                "description": "Name of the current session",
                "maxLength": 100,
                "pattern": "^[a-zA-Z0-9\\-\\_\\s]{1,100}$",
                "examples": []string{
                    "feature-development",
                    "bug-fix-session",
                    "api-integration",
                },
            },
            "task_description": map[string]any{
                "type": "string",
                "description": "Description of the task being performed",
                "minLength": 1,
                "maxLength": 1000,
                "examples": []string{
                    "Implementing OAuth2 authentication",
                    "Fixing memory leak in data processor",
                    "Adding new API endpoints",
                },
            },
            "context_info": map[string]any{
                "type": "string",
                "description": "Additional context information",
                "maxLength": 500,
                "examples": []string{
                    "Working on user management microservice",
                    "Debugging production data pipeline",
                    "Migrating from legacy authentication",
                },
            },
            "missing_info": map[string]any{
                "type": "string",
                "description": "What information was missing or unclear",
                "maxLength": 500,
                "examples": []string{
                    "API specification for refresh endpoint",
                    "Database schema documentation",
                    "Error code definitions",
                },
            },
            "confused_by": map[string]any{
                "type": "string",
                "description": "What aspects were confusing",
                "maxLength": 500,
                "examples": []string{
                    "Token rotation logic unclear",
                    "Conflicting requirements",
                    "Outdated documentation examples",
                },
            },
            "future_wishes": map[string]any{
                "type": "string",
                "description": "Suggestions for future improvements",
                "maxLength": 500,
                "examples": []string{
                    "Comprehensive API documentation",
                    "Postman collection examples",
                    "Integration test suite",
                },
            },
            "severity": map[string]any{
                "type": "string",
                "description": "Severity level of the complaint",
                "enum": []string{"low", "medium", "high", "critical"},
                "examples": map[string]string{
                    "low": "Minor inconvenience, workarounds available",
                    "medium": "Significant productivity impact",
                    "high": "Blocker, no clear path forward",
                    "critical": "System failure, project stall",
                },
            },
        },
        "required": []string{"agent_name", "task_description", "severity"},
    },
}
```

#### list_complaints Tool

```go
listComplaintsTool := &mcp.Tool{
    Name: "list_complaints",
    Description: "Retrieve paginated list of complaints with optional filtering",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "limit": map[string]any{
                "type": "integer",
                "description": "Maximum number of complaints to return",
                "minimum": 1,
                "maximum": 100,
                "default": 50,
                "examples": []int{10, 25, 50, 100},
            },
            "offset": map[string]any{
                "type": "integer",
                "description": "Number of complaints to skip for pagination",
                "minimum": 0,
                "default": 0,
                "examples": []int{0, 50, 100},
            },
            "severity": map[string]any{
                "type": "string",
                "description": "Filter by severity level",
                "enum": []string{"low", "medium", "high", "critical"},
                "examples": []string{"high", "medium", "low"},
            },
            "resolved": map[string]any{
                "type": "boolean",
                "description": "Filter by resolution status (true=resolved, false=unresolved)",
                "examples": []bool{true, false},
            },
            "agent_id": map[string]any{
                "type": "string",
                "description": "Filter by agent identifier (flat format)",
                "pattern": "^[a-zA-Z0-9\\-\\_\\s]{1,100}$",
                "examples": []string{
                    "AI-Coding-Assistant",
                    "Code-Reviewer-Bot",
                },
            },
            "project_id": map[string]any{
                "type": "string",
                "description": "Filter by project identifier (flat format)",
                "pattern": "^[a-zA-Z0-9\\-\\_\\s\\.]{1,100}$",
                "examples": []string{
                    "user-management-system",
                    "data-processor-v2",
                },
            },
        },
        "default": map[string]any{
            "limit": 50,
            "offset": 0,
        },
    },
}
```

### B. Update Tool Output Examples

#### Expected Output Examples

```go
// file_complaint output example
{
  "success": true,
  "message": "Complaint filed successfully",
  "complaint": {
    "id": "550e8400-e29b-41d4-a716-446655440000",  // ✅ Flat string
    "agent_name": "AI-Coding-Assistant",
    "session_name": "feature-development",
    "project_name": "my-project",
    "task_description": "Implementing OAuth2 authentication",
    "context_info": "Working on user management microservice",
    "missing_info": "API specification for refresh endpoint",
    "confused_by": "Token rotation logic unclear",
    "future_wishes": "Comprehensive API documentation",
    "severity": "high",
    "timestamp": "2024-11-09T12:18:30Z",
    "resolved": false,
    "resolved_at": null,
    "resolved_by": "",
    "file_path": "/Users/larsartmann/.local/share/complaints/550e8400-e29b-41d4-a716-446655440000.json",
    "docs_path": "docs/complaints/2024-11-09_12-18-feature-development.md"
  }
}
```

### C. Update Input Validation Functions

```go
// internal/delivery/mcp/input_validation.go (updated)

func validateResolveComplaintInput(input map[string]any) (string, string, error) {
    // Validate complaint_id as flat string
    complaintIDInterface, exists := input["complaint_id"]
    if !exists {
        return "", "", fmt.Errorf("complaint_id is required")
    }

    complaintIDStr, isString := complaintIDInterface.(string)
    if !isString {
        return "", "", fmt.Errorf("complaint_id must be string, got %T", complaintIDInterface)
    }

    // Validate UUID format (flat string)
    if !isValidUUID(complaintIDStr) {
        return "", "", fmt.Errorf("complaint_id must be valid UUID v4 format, got: %s", complaintIDStr)
    }

    // Validate resolved_by
    resolvedByInterface, exists := input["resolved_by"]
    if !exists {
        return "", "", fmt.Errorf("resolved_by is required")
    }

    resolvedByStr, isString := resolvedByInterface.(string)
    if !isString {
        return "", "", fmt.Errorf("resolved_by must be string, got %T", resolvedByInterface)
    }

    if strings.TrimSpace(resolvedByStr) == "" {
        return "", "", fmt.Errorf("resolved_by cannot be empty")
    }

    if len(resolvedByStr) > 100 {
        return "", "", fmt.Errorf("resolved_by cannot exceed 100 characters")
    }

    return complaintIDStr, resolvedByStr, nil
}

func validateFileComplaintInput(input map[string]any) (FileComplaintInput, error) {
    var result FileComplaintInput

    // Validate agent_name
    if err := validateRequiredString(input, "agent_name", &result.AgentName, 1, 100, agentNamePattern); err != nil {
        return result, err
    }

    // Validate project_name (optional)
    if projectInterface, exists := input["project_name"]; exists && projectInterface != nil {
        if projectStr, isString := projectInterface.(string); isString && projectStr != "" {
            if err := validatePattern(projectStr, projectNamePattern, "project_name"); err != nil {
                return result, err
            }
            result.ProjectName = projectStr
        }
    }

    // Validate session_name (optional)
    if sessionInterface, exists := input["session_name"]; exists && sessionInterface != nil {
        if sessionStr, isString := sessionInterface.(string); isString {
            if err := validatePattern(sessionStr, sessionNamePattern, "session_name"); err != nil {
                return result, err
            }
            result.SessionName = sessionStr
        }
    }

    // Validate task_description (required)
    if err := validateRequiredString(input, "task_description", &result.TaskDescription, 1, 1000, nil); err != nil {
        return result, err
    }

    // Validate optional fields with length limits
    validateOptionalString(input, "context_info", &result.ContextInfo, 500)
    validateOptionalString(input, "missing_info", &result.MissingInfo, 500)
    validateOptionalString(input, "confused_by", &result.ConfusedBy, 500)
    validateOptionalString(input, "future_wishes", &result.FutureWishes, 500)

    // Validate severity (required)
    if err := validateSeverity(input); err != nil {
        return result, err
    }

    return result, nil
}

// Helper functions
func validateRequiredString(input map[string]any, field string, result *string, minLen, maxLen int, pattern *regexp.Regexp) error {
    valueInterface, exists := input[field]
    if !exists {
        return fmt.Errorf("%s is required", field)
    }

    valueStr, isString := valueInterface.(string)
    if !isString {
        return fmt.Errorf("%s must be string, got %T", field, valueInterface)
    }

    trimmed := strings.TrimSpace(valueStr)
    if len(trimmed) < minLen {
        return fmt.Errorf("%s cannot be empty", field)
    }

    if len(trimmed) > maxLen {
        return fmt.Errorf("%s cannot exceed %d characters", field, maxLen)
    }

    if pattern != nil && !pattern.MatchString(trimmed) {
        return fmt.Errorf("%s contains invalid characters", field)
    }

    *result = trimmed
    return nil
}

func validateOptionalString(input map[string]any, field string, result *string, maxLen int) {
    if valueInterface, exists := input[field]; exists && valueInterface != nil {
        if valueStr, isString := valueInterface.(string); isString {
            trimmed := strings.TrimSpace(valueStr)
            if len(trimmed) <= maxLen {
                *result = trimmed
            }
        }
    }
}

func validatePattern(value, field string, pattern *regexp.Regexp) error {
    if !pattern.MatchString(value) {
        return fmt.Errorf("%s contains invalid characters", field)
    }
    return nil
}
```

### D. Update BDD Tests

```gherkin
# features/bdd/schema_validation_bdd.feature
Feature: MCP Tool Schema Validation
  As an AI assistant
  I want to use flat ID formats in tool calls
  So that schema validation works correctly with phantom types

  Scenario: File complaint with flat agent ID
    Given I have a valid complaint with flat agent ID
    When I call the file_complaint tool with flat ID format
    Then the tool should accept the flat ID format
    And should create the complaint successfully

  Scenario: Resolve complaint with flat complaint ID
    Given I have an existing complaint with flat complaint ID
    When I call the resolve_complaint tool with flat ID format
    Then the tool should accept the flat ID format
    And should resolve the complaint successfully

  Scenario: Reject nested ID format
    Given I attempt to use nested ID format
    When I call any tool with nested ID structure
    Then the tool should reject the nested format
    And should return appropriate validation error

  Scenario: Auto-detect project name
    Given I don't provide a project name
    When I file a complaint with auto-detection enabled
    Then the system should auto-detect the project name
    And should associate the complaint with the correct project
```

### E. Update Documentation

#### README.md Schema Examples

````markdown
## MCP Tool Interface

### file_complaint
```json
{
  "name": "file_complaint",
  "description": "File a structured complaint about missing or confusing information",
  "inputSchema": {
    "type": "object",
    "properties": {
      "agent_name": {
        "type": "string",
        "description": "Name of the AI agent filing the complaint",
        "minLength": 1,
        "maxLength": 100,
        "examples": ["AI-Coding-Assistant", "Code-Reviewer-Bot"]
      },
      "complaint_id": {
        "type": "string",
        "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
        "description": "UUID v4 format complaint identifier (flat string)",
        "examples": ["550e8400-e29b-41d4-a716-446655440000"]
      }
    },
    "required": ["agent_name", "task_description", "severity"]
  }
}
````

## 🧪 Verification Steps

### 1. Schema Validation Tests

```go
func TestSchemaValidation_FlatIDFormat(t *testing.T) {
    tests := []struct {
        name     string
        input    map[string]any
        wantErr  bool
        errorMsg string
    }{
        {
            name: "valid flat complaint ID",
            input: map[string]any{
                "complaint_id": "550e8400-e29b-41d4-a716-446655440000",
                "resolved_by": "AI-Assistant",
            },
            wantErr: false,
        },
        {
            name: "invalid nested complaint ID",
            input: map[string]any{
                "complaint_id": map[string]any{
                    "Value": "550e8400-e29b-41d4-a716-446655440000",
                },
                "resolved_by": "AI-Assistant",
            },
            wantErr: true,
            errorMsg: "complaint_id must be string, got object",
        },
        {
            name: "invalid complaint ID format",
            input: map[string]any{
                "complaint_id": "not-a-uuid",
                "resolved_by": "AI-Assistant",
            },
            wantErr: true,
            errorMsg: "must be valid UUID v4 format",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := validateResolveComplaintInput(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2. Integration Tests

```go
func TestMCPServer_FlatIDIntegration(t *testing.T) {
    server := setupTestServer()

    t.Run("File complaint with flat IDs", func(t *testing.T) {
        input := map[string]any{
            "agent_name": "AI-Assistant",
            "task_description": "Test task",
            "severity": "low",
            "project_name": "", // Auto-detect
        }

        // Call file_complaint tool
        result, output, err := server.handleFileComplaint(context.Background(), &mcp.CallToolRequest{
            Arguments: mcp.ToolArguments(input),
        })

        require.NoError(t, err)
        assert.NotNil(t, result)
        assert.True(t, output.Success)

        // Verify flat ID in output
        assert.Contains(t, output.Complaint.ID, "-")  // UUID format
    })

    t.Run("Resolve complaint with flat ID", func(t *testing.T) {
        // First create a complaint
        input := map[string]any{
            "agent_name": "AI-Assistant",
            "task_description": "Test task",
            "severity": "low",
        }

        _, createOutput, err := server.handleFileComplaint(context.Background(), &mcp.CallToolRequest{
            Arguments: mcp.ToolArguments(input),
        })
        require.NoError(t, err)

        // Then resolve with flat ID
        resolveInput := map[string]any{
            "complaint_id": createOutput.Complaint.ID,  // Flat string
            "resolved_by": "Test-Resolver",
        }

        result, output, err := server.handleResolveComplaint(context.Background(), &mcp.CallToolRequest{
            Arguments: mcp.ToolArguments(resolveInput),
        })

        require.NoError(t, err)
        assert.NotNil(t, result)
        assert.True(t, output.Success)
    })
}
```

### 3. End-to-End Tests

```bash
# Test flat ID format with real MCP server
echo '{"tool":"file_complaint","arguments":{"agent_name":"Test-Agent","task_description":"Test task","severity":"low"}}' | ./complaints-mcp

# Expected response should contain flat ID:
# "id": "550e8400-e29b-41d4-a716-446655440000"

# Test resolve with flat ID
echo '{"tool":"resolve_complaint","arguments":{"complaint_id":"550e8400-e29b-41d4-a716-446655440000","resolved_by":"Test-Resolver"}}' | ./complaints-mcp
```

## ⏱️ Time Estimate: 4-6 hours

## 🎯 Impact: HIGH (fixes API contract alignment)

## 💪 Work Required: MEDIUM (schema updates + validation)

## 🚨 Prerequisite: Step 6 (phantom types implementation)
