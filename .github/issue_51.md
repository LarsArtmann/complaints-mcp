# Issue #51: Update All JSON Schemas for Flat ID Field Structure

## 🎯 **Enhancement: API Schema Updates for Phantom Type Integration**

### **Current State Analysis**
The JSON schemas in MCP tool handlers still expect nested ID structures, but phantom types produce flat structures:

**❌ Current Schema Expectation (Nested):**
```json
{
  "complaint_id": {
    "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"
  }
}
```

**✅ New Phantom Type Output (Flat):**
```json
{
  "complaint_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### **Problem Identified**
This schema mismatch will cause:
- **Tool Input Validation Failures**: MCP clients sending flat strings rejected
- **API Documentation Confusion**: Docs show wrong structure
- **Integration Issues**: External tools broken
- **Test Failures**: Test expectations inconsistent

## 🛠️ **Comprehensive Schema Update Plan**

### **Phase 1: MCP Tool Schema Updates**

#### **resolve_complaint Tool Schema**
```go
// ❌ Current (nested)
resolveComplaintTool := &mcp.Tool{
    Name: "resolve_complaint",
    Description: "Mark a complaint as resolved",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "complaint_id": map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "Value": map[string]any{
                        "type": "string",
                        "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
                        "description": "Unique identifier of the complaint",
                    },
                },
                "required": []string{"Value"},
                "description": "Complaint identifier object",
            },
        },
        "required": []string{"complaint_id"},
    },
}

// ✅ Updated (flat)
resolveComplaintTool := &mcp.Tool{
    Name: "resolve_complaint",
    Description: "Mark a complaint as resolved",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "complaint_id": map[string]any{
                "type": "string",
                "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
                "description": "Unique identifier of the complaint (UUID v4 format)",
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

#### **file_complaint Tool Schema (Updated for typed IDs)**
```go
// ✅ Updated (flat with typed ID examples)
fileComplaintTool := &mcp.Tool{
    Name: "file_complaint",
    Description: "File a structured complaint about missing or confusing information",
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
            "project_name": map[string]any{
                "type": "string",
                "description": "Name of the project being worked on",
                "maxLength": 100,
                "pattern": "^[a-zA-Z0-9\\-\\_\\s\\.]{1,100}$",
                "examples": []string{
                    "user-management-system",
                    "data-processor-v2",
                    "api-gateway-service",
                },
            },
        },
        "required": []string{"agent_name", "task_description", "severity"},
    },
}
```

#### **list_complaints Tool Schema**
```go
// ✅ Updated (flat with improved filtering)
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

#### **search_complaints Tool Schema**
```go
// ✅ Updated (enhanced with search options)
searchComplaintsTool := &mcp.Tool{
    Name: "search_complaints",
    Description: "Search complaints by content with advanced options",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "query": map[string]any{
                "type": "string",
                "description": "Search query to find matching complaints",
                "minLength": 1,
                "maxLength": 200,
                "examples": []string{
                    "authentication",
                    "memory leak",
                    "API documentation",
                    "database schema",
                },
            },
            "limit": map[string]any{
                "type": "integer",
                "description": "Maximum number of results to return",
                "minimum": 1,
                "maximum": 50,
                "default": 20,
                "examples": []int{10, 20, 50},
            },
            "severity": map[string]any{
                "type": "string",
                "description": "Optional severity filter for search results",
                "enum": []string{"low", "medium", "high", "critical"},
                "examples": []string{"high", "critical"},
            },
            "resolved": map[string]any{
                "type": "boolean",
                "description": "Optional filter by resolution status",
                "examples": []bool{true, false},
            },
        },
        "required": []string{"query"},
        "default": map[string]any{
            "limit": 20,
        },
    },
}
```

#### **get_cache_stats Tool Schema (Enhanced)**
```go
// ✅ Updated (enhanced with detailed stats)
getCacheStatsTool := &mcp.Tool{
    Name: "get_cache_stats",
    Description: "Get comprehensive cache performance statistics and system health",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "reset": map[string]any{
                "type": "boolean",
                "description": "Reset cache statistics after retrieval",
                "default": false,
                "examples": []bool{true, false},
            },
        },
        "default": map[string]any{
            "reset": false,
        },
    },
}
```

### **Phase 2: Output Schema Documentation**

#### **Expected Tool Responses (Updated)**
```json
// file_complaint response
{
  "success": true,
  "message": "Complaint filed successfully",
  "complaint": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "agent_name": "AI-Coding-Assistant",
    "session_name": "feature-development",
    "task_description": "Implementing OAuth2 authentication",
    "context_info": "Working on user management microservice",
    "missing_info": "API specification for refresh endpoint",
    "confused_by": "Token rotation logic unclear",
    "future_wishes": "Comprehensive API documentation",
    "severity": "high",
    "project_name": "user-management-system",
    "timestamp": "2024-11-09T12:18:30Z",
    "resolved": false,
    "resolved_at": null,
    "resolved_by": "",
    "file_path": "/Users/larsartmann/.local/share/complaints/550e8400-e29b-41d4-a716-446655440000.json",
    "docs_path": "docs/complaints/2024-11-09_12-18-feature-development.md"
  }
}

// resolve_complaint response
{
  "success": true,
  "message": "Complaint resolved successfully",
  "complaint": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "resolved": true,
    "resolved_at": "2024-11-09T14:30:00Z",
    "resolved_by": "AI-Code-Reviewer"
  }
}

// list_complaints response
{
  "complaints": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "agent_name": "AI-Coding-Assistant",
      "task_description": "Implementing OAuth2 authentication",
      "severity": "high",
      "timestamp": "2024-11-09T12:18:30Z",
      "resolved": false
    },
    {
      "id": "9cb3bb9e-b6dc-4e02-9767-e396a42b63a6",
      "agent_name": "Code-Reviewer-Bot",
      "task_description": "Reviewing security implementations",
      "severity": "medium",
      "timestamp": "2024-11-09T11:45:00Z",
      "resolved": true
    }
  ]
}

// search_complaints response
{
  "complaints": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "task_description": "Implementing OAuth2 authentication",
      "missing_info": "API specification for refresh endpoint",
      "severity": "high",
      "relevance_score": 0.95
    }
  ],
  "query": "authentication",
  "total_results": 1
}

// get_cache_stats response
{
  "cache_enabled": true,
  "stats": {
    "hits": 1247,
    "misses": 89,
    "evictions": 12,
    "current_size": 156,
    "max_size": 1000,
    "hit_rate": 0.9332,
    "memory_usage_bytes": 2048000,
    "uptime_seconds": 3600
  },
  "message": "Cache statistics retrieved successfully"
}
```

### **Phase 3: Documentation Updates**

#### **README.md Schema Examples**
```markdown
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
        "description": "Unique identifier in UUID v4 format (flat string)",
        "examples": ["550e8400-e29b-41d4-a716-446655440000"]
      }
    },
    "required": ["agent_name", "task_description", "severity"]
  }
}
```

#### **API Migration Guide**
```markdown
## API Migration Guide: v1.0 → v2.0

### ID Field Structure Changes

#### Before (Nested)
```json
{
  "complaint_id": {
    "Value": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### After (Flat)
```json
{
  "complaint_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Migration Steps

1. **Update Tool Calls**: Change from nested to flat ID format
2. **Update Validation Logic**: Validate strings instead of objects
3. **Update Error Handling**: Handle string format validation errors
4. **Update Tests**: Modify test expectations for flat format
5. **Update Documentation**: Update examples and API docs

### Example Migration

#### Tool Call Update
```json
// ❌ Before (nested)
{
  "tool": "resolve_complaint",
  "arguments": {
    "complaint_id": {
      "Value": "550e8400-e29b-41d4-a716-446655440000"
    },
    "resolved_by": "AI-Assistant"
  }
}

// ✅ After (flat)
{
  "tool": "resolve_complaint", 
  "arguments": {
    "complaint_id": "550e8400-e29b-41d4-a716-446655440000",
    "resolved_by": "AI-Assistant"
  }
}
```

#### Validation Update
```json
// ❌ Before (nested validation)
{
  "complaint_id": {
    "Value": {
      "type": "string",
      "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"
    }
  }
}

// ✅ After (flat validation)
{
  "complaint_id": {
    "type": "string",
    "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
    "description": "UUID v4 format complaint identifier"
  }
}
```
```

### **Phase 4: Test Schema Updates**

#### **BDD Test Updates**
```gherkin
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
    Given I have an existing complaint with ID "550e8400-e29b-41d4-a716-446655440000"
    When I call the resolve_complaint tool with flat ID format
    Then the tool should accept the flat ID format
    And should resolve the complaint successfully

  Scenario: Reject nested ID format
    Given I attempt to use nested ID format
    When I call any tool with nested ID structure
    Then the tool should reject the nested format
    And should return appropriate validation error
```

#### **Unit Test Updates**
```go
func TestToolSchema_FlatIDValidation(t *testing.T) {
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
            errorMsg: "complaint_id must match UUID v4 pattern",
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

## 🎯 **Benefits of Schema Updates**

### **1. Consistency with Phantom Types**
```go
// ✅ Aligned: Schema matches actual output
type ComplaintDTO struct {
    ID string `json:"id"`  // Flat string, matches phantom type
}

// Tool schema expects flat string
{
  "id": {
    "type": "string",  // ✅ Matches DTO
    "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"
  }
}
```

### **2. Better Developer Experience**
```json
{
  "complaint_id": {
    "type": "string",
    "description": "UUID v4 format complaint identifier",
    "examples": ["550e8400-e29b-41d4-a716-446655440000"],
    "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"
  }
}
```

### **3. Improved Validation**
- **Pattern Matching**: Direct string validation
- **Clear Examples**: Real-world format examples
- **Better Error Messages**: Specific validation failures
- **Type Safety**: Schema enforces correct types

### **4. Enhanced Documentation**
- **Migration Guide**: Clear upgrade path
- **Examples**: Before/after comparisons
- **Validation Rules**: Detailed pattern explanations
- **Tool References**: Complete API documentation

## 📋 **Files to Modify**

### **Core Schema Files**
- `internal/delivery/mcp/mcp_server.go` - Update all tool schemas
- `internal/delivery/mcp/dto.go` - Verify DTO structure alignment
- `internal/delivery/mcp/input_validation.go` - Update validation logic

### **Test Files**
- `features/bdd/mcp_integration_bdd_test.go` - Update BDD scenarios
- `internal/delivery/mcp/mcp_server_test.go` - Update schema tests
- `features/bdd/schema_validation_bdd_test.go` - Add new schema tests

### **Documentation Files**
- `README.md` - Update tool interface documentation
- `docs/API_SCHEMA.md` - Create comprehensive schema reference
- `docs/MIGRATION_GUIDE.md` - Create v1→v2 migration guide
- `examples/tool_usage/` - Add updated example files

## 🔄 **Migration Strategy**

### **Phase 1: Schema Updates**
- Update all tool schemas to flat ID format
- Add comprehensive examples and patterns
- Ensure validation logic matches schema

### **Phase 2: Documentation Updates**
- Update README.md with new schema examples
- Create migration guide for existing users
- Add API schema reference documentation

### **Phase 3: Test Updates**
- Update BDD tests for flat ID format
- Add schema validation tests
- Update integration test expectations

### **Phase 4: Validation**
- Test all tool schemas with flat ID inputs
- Verify error messages are clear and helpful
- Ensure backward compatibility guidelines are followed

## 🧪 **Testing Strategy**

### **Schema Validation Tests**
```go
func TestToolSchema_FlatIDSupport(t *testing.T) {
    // Test that all tool schemas accept flat ID format
    tools := []Tool{fileComplaintTool, resolveComplaintTool, listComplaintsTool}
    
    for _, tool := range tools {
        t.Run(tool.Name, func(t *testing.T) {
            // Test flat ID inputs
            flatInputs := generateFlatIDInputs(tool.Name)
            for _, input := range flatInputs {
                err := validateToolInput(tool, input)
                assert.NoError(t, err, "Tool %s should accept flat ID input: %+v", tool.Name, input)
            }
            
            // Test nested ID inputs (should fail)
            nestedInputs := generateNestedIDInputs(tool.Name)
            for _, input := range nestedInputs {
                err := validateToolInput(tool, input)
                assert.Error(t, err, "Tool %s should reject nested ID input: %+v", tool.Name, input)
            }
        })
    }
}
```

### **Integration Tests**
```go
func TestMCPServer_FlatIDIntegration(t *testing.T) {
    server := setupTestServer()
    
    // Test complete workflow with flat IDs
    t.Run("Complete Workflow", func(t *testing.T) {
        // 1. File complaint with flat agent ID
        complaintID := fileComplaintWithFlatID(server)
        
        // 2. Resolve complaint with flat complaint ID
        resolveComplaintWithFlatID(server, complaintID)
        
        // 3. List complaints with flat filter
        listComplaintsWithFlatIDs(server)
        
        // 4. Search complaints with flat query
        searchComplaintsWithFlatQuery(server)
    })
}
```

### **BDD Tests**
```gherkin
Feature: Flat ID Schema Support
  As an AI assistant
  I want to use flat ID formats throughout the MCP interface
  So that tool interactions are consistent with phantom type implementation

  Background:
    Given the MCP server is running with updated schemas
    And phantom types are implemented for all ID fields

  Scenario: File complaint with flat agent and project IDs
    When I file a complaint with flat agent ID "AI-Coding-Assistant"
    And flat project ID "user-management-system"
    Then the complaint should be created successfully
    And the response should use flat ID format

  Scenario: List complaints with flat ID filters
    When I list complaints with agent ID filter "AI-Coding-Assistant"
    And project ID filter "user-management-system"
    Then only matching complaints should be returned
    And all IDs in response should be flat format

  Scenario: Resolve complaint with flat complaint ID
    When I resolve complaint with flat ID "550e8400-e29b-41d4-a716-446655440000"
    And resolver "AI-Code-Reviewer"
    Then the complaint should be resolved successfully
    And the response should use flat ID format
```

## ⚠️ **Breaking Changes**

### **API Contract Changes**
- **Input Format**: IDs change from nested objects to flat strings
- **Validation Rules**: Direct string validation instead of object validation
- **Error Messages**: Updated to reflect flat ID format expectations
- **Examples**: All documentation examples updated to flat format

### **Migration Impact**
- **Tool Calls**: All clients need to use flat ID format
- **Integration Points**: External tools must update ID handling
- **Test Suites**: Existing tests need schema updates
- **Documentation**: All API docs need format updates

### **Mitigation Strategy**
- **Migration Guide**: Step-by-step upgrade instructions
- **Backward Compatibility Window**: Support both formats during transition
- **Clear Error Messages**: Helpful validation errors for format issues
- **Tooling Support**: CLI tools for format conversion

## 🏆 **Success Criteria**

- [ ] All tool schemas updated to flat ID format
- [ ] Schema validation works correctly with flat IDs
- [ ] Nested ID formats are properly rejected
- [ ] Error messages are clear and actionable
- [ ] Documentation updated with new examples
- [ ] Migration guide provided for existing users
- [ ] Test suite covers all schema scenarios
- [ ] Integration tests validate complete workflows
- [ ] Performance impact is minimal

## 🏷️ **Labels**
- `documentation` - API documentation updates
- `breaking-change` - Changes API contract
- `api` - MCP tool interface updates
- `schema` - JSON schema modifications
- `high-priority` - Critical for phantom type integration
- `validation` - Schema validation improvements

## 📊 **Priority**: High
- **Complexity**: Medium (schema updates + documentation)
- **Value**: High (enables phantom type integration)
- **Risk**: Medium (breaking changes require migration)
- **Dependencies**: Issue #48 (phantom types), Issue #49 (ID field conversion)

## 🤝 **Dependencies**
- **Issue #48**: Must have phantom types implemented first
- **Issue #49**: Should have ID fields converted to phantom types
- **Issue #50**: Need validation constructors for schema examples

---

**This issue ensures complete alignment between phantom type implementation and MCP tool schemas, providing consistent, well-documented API interfaces that support the new type-safe ID system.**