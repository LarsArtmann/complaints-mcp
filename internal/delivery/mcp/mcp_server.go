package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/charmbracelet/log"
	"github.com/ilyakaz/tracey"
	"github.com/modelcontextprotocol/go-sdk/server"
)

// MCPServer represents the MCP server implementation
type MCPServer struct {
	config       *config.Config
	service      *service.ComplaintService
	logger       *log.Logger
	tracer       tracey.Tracer
	server       *server.Server
}

// NewServer creates a new MCP server
func NewServer(name, version string, complaintService *service.ComplaintService, logger *log.Logger, tracer tracing.Tracer) *MCPServer {
	return &MCPServer{
		config:  nil, // Will be set during initialization
		service: complaintService,
		logger:   logger,
		tracer:   tracey.NewTracer("mcp-server"),
	}
}

// Start starts the MCP server
func (m *MCPServer) Start(ctx context.Context) error {
	logger := m.logger.With("component", "mcp-server")
	
	// Initialize server
	m.server = server.New(
		server.WithName(m.config.Server.Name),
		server.WithVersion(version),
		server.WithLogger(logger),
	)
	
	// Register tools
	if err := m.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}
	
	// Register capabilities
	m.server.AddCapabilities(server.Capability{
		Tools: map[string]interface{}{},
	})
	
	// Start server
	logger.Info("Starting MCP server", 
		"address", m.config.Server.Address(),
		"name", m.config.Server.Name)
	
	if err := m.server.Serve(ctx); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}
	
	return nil
}

// Shutdown gracefully shuts down the MCP server
func (m *MCPServer) Shutdown(ctx context.Context) error {
	logger := m.logger.With("component", "mcp-server")
	logger.Info("Shutting down MCP server")
	
	if m.server != nil {
		return m.server.Shutdown(ctx)
	}
	
	return nil
}

// registerTools registers all available MCP tools
func (m *MCPServer) registerTools() error {
	// File complaint tool
	fileComplaintTool := server.Tool{
		Name:        "file_complaint",
		Description: "File a structured complaint about missing or confusing information",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"agent_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the AI agent filing the complaint",
					"minLength":   1,
					"maxLength":   100,
				},
				"session_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the current session",
					"maxLength":   100,
				},
				"task_description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the task being performed",
					"minLength":   1,
					"maxLength":   1000,
				},
				"context_info": map[string]interface{}{
					"type":        "string",
					"description": "Additional context information",
					"maxLength":   500,
				},
				"missing_info": map[string]interface{}{
					"type":        "string",
					"description": "What information was missing or unclear",
					"maxLength":   500,
				},
				"confused_by": map[string]interface{}{
					"type":        "string",
					"description": "What aspects were confusing",
					"maxLength":   500,
				},
				"future_wishes": map[string]interface{}{
					"type":        "string",
					"description": "Suggestions for future improvements",
					"maxLength":   500,
				},
				"severity": map[string]interface{}{
					"type":        "string",
					"description": "Severity level (low, medium, high, critical)",
					"enum":        []string{"low", "medium", "high", "critical"},
				},
				"project_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the project (auto-detected if not provided)",
					"maxLength":   100,
				},
			},
			"required": []string{"agent_name", "task_description", "severity"},
		},
	}
	
	// List complaints tool
	listComplaintsTool := server.Tool{
		Name:        "list_complaints",
		Description: "List all filed complaints with optional filtering",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of complaints to return",
					"minimum":     1,
					"maximum":     100,
				},
				"severity": map[string]interface{}{
					"type":        "string",
					"description": "Filter by severity level",
					"enum":        []string{"low", "medium", "high", "critical"},
				},
				"resolved": map[string]interface{}{
					"type":        "boolean",
					"description": "Filter by resolved status",
				},
			},
		},
	}
	
	// Resolve complaint tool
	resolveComplaintTool := server.Tool{
		Name:        "resolve_complaint",
		Description: "Mark a complaint as resolved",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"complaint_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier of the complaint",
					"pattern":     "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
				},
			},
			"required": []string{"complaint_id"},
		},
	}
	
	// Search complaints tool
	searchComplaintsTool := server.Tool{
		Name:        "search_complaints",
		Description: "Search complaints by content",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query text",
					"minLength":   1,
					"maxLength":   500,
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results",
					"minimum":     1,
					"maximum":     100,
				},
			},
			"required": []string{"query"},
		},
	}
	
	// Register tools with handlers
	m.server.AddTool(fileComplaintTool, m.handleFileComplaint)
	m.server.AddTool(listComplaintsTool, m.handleListComplaints)
	m.server.AddTool(resolveComplaintTool, m.handleResolveComplaint)
	m.server.AddTool(searchComplaintsTool, m.handleSearchComplaints)
	
	return nil
}

// handleFileComplaint handles the file_complaint tool
func (m *MCPServer) handleFileComplaint(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
	ctx, span := m.tracer.Start(ctx, "handleFileComplaint")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "file_complaint")
	logger.Info("Handling file complaint request")

	// Extract arguments
	agentName, _ := arguments["agent_name"].(string)
	sessionName, _ := arguments["session_name"].(string)
	taskDescription, _ := arguments["task_description"].(string)
	contextInfo, _ := arguments["context_info"].(string)
	missingInfo, _ := arguments["missing_info"].(string)
	confusedBy, _ := arguments["confused_by"].(string)
	futureWishes, _ := arguments["future_wishes"].(string)
	severity, _ := arguments["severity"].(string)
	projectName, _ := arguments["project_name"].(string)

	// Convert severity string to domain type
	var domainSeverity domain.Severity
	switch severity {
	case "low":
		domainSeverity = domain.SeverityLow
	case "medium":
		domainSeverity = domain.SeverityMedium
	case "high":
		domainSeverity = domain.SeverityHigh
	case "critical":
		domainSeverity = domain.SeverityCritical
	default:
		return nil, fmt.Errorf("invalid severity: %s", severity)
	}

	complaint, err := m.service.CreateComplaint(
		ctx,
		agentName,
		sessionName,
		taskDescription,
		contextInfo,
		missingInfo,
		confusedBy,
		futureWishes,
		domainSeverity,
		projectName,
	)
	if err != nil {
		logger.Error("Failed to create complaint", "error", err)
		return nil, err
	}

	logger.Info("Complaint filed successfully", "complaint_id", complaint.ID.String())
	
	return map[string]interface{}{
		"complaint_id": complaint.ID.String(),
		"message":      "Complaint filed successfully",
		"timestamp":    complaint.Timestamp.Format(time.RFC3339),
	}, nil
}

// handleListComplaints handles the list_complaints tool
func (m *MCPServer) handleListComplaints(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
	ctx, span := m.tracer.Start(ctx, "handleListComplaints")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "list_complaints")
	logger.Info("Handling list complaints request")

	// Extract arguments
	limit := 50 // default limit
	if l, ok := arguments["limit"].(float64); ok {
		limit = int(l)
	}
	
	var severityFilter domain.Severity
	if severity, ok := arguments["severity"].(string); ok {
		switch severity {
		case "low":
			severityFilter = domain.SeverityLow
		case "medium":
			severityFilter = domain.SeverityMedium
		case "high":
			severityFilter = domain.SeverityHigh
		case "critical":
			severityFilter = domain.SeverityCritical
		}
	}
	
	resolvedFilter := false // default to show all
	if resolved, ok := arguments["resolved"].(bool); ok {
		resolvedFilter = resolved
	}

	var complaints []*domain.Complaint
	var err error
	
	if severityFilter != "" {
		complaints, err = m.service.GetComplaintsBySeverity(ctx, severityFilter, limit)
	} else {
		complaints, err = m.service.ListComplaints(ctx, limit, 0)
	}
	
	if err != nil {
		logger.Error("Failed to list complaints", "error", err)
		return nil, err
	}

	// Convert to response format
	var results []map[string]interface{}
	for _, complaint := range complaints {
		result := map[string]interface{}{
			"complaint_id":    complaint.ID.String(),
			"agent_name":      complaint.AgentName,
			"session_name":    complaint.SessionName,
			"task_description": complaint.TaskDescription,
			"severity":        string(complaint.Severity),
			"timestamp":       complaint.Timestamp.Format(time.RFC3339),
			"resolved":        complaint.IsResolved(),
			"project_name":    complaint.ProjectName,
		}
		
		if resolvedFilter == complaint.IsResolved() {
			results = append(results, result)
		}
	}

	logger.Info("Complaints listed successfully", "count", len(results))
	
	return map[string]interface{}{
		"complaints": results,
		"count":      len(results),
	}, nil
}

// handleResolveComplaint handles the resolve_complaint tool
func (m *MCPServer) handleResolveComplaint(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
	ctx, span := m.tracer.Start(ctx, "handleResolveComplaint")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "resolve_complaint")
	logger.Info("Handling resolve complaint request")

	complaintIDStr, _ := arguments["complaint_id"].(string)
	complaintID := domain.ComplaintID{Value: complaintIDStr}

	if err := m.service.ResolveComplaint(ctx, complaintID); err != nil {
		logger.Error("Failed to resolve complaint", "error", err, "complaint_id", complaintIDStr)
		return nil, err
	}

	logger.Info("Complaint resolved successfully", "complaint_id", complaintIDStr)
	
	return map[string]interface{}{
		"message": "Complaint resolved successfully",
		"complaint_id": complaintIDStr,
	}, nil
}

// handleSearchComplaints handles the search_complaints tool
func (m *MCPServer) handleSearchComplaints(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
	ctx, span := m.tracer.Start(ctx, "handleSearchComplaints")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "search_complaints")
	logger.Info("Handling search complaints request")

	query, _ := arguments["query"].(string)
	limit := 50 // default limit
	if l, ok := arguments["limit"].(float64); ok {
		limit = int(l)
	}

	complaints, err := m.service.SearchComplaints(ctx, query, limit)
	if err != nil {
		logger.Error("Failed to search complaints", "error", err)
		return nil, err
	}

	// Convert to response format
	var results []map[string]interface{}
	for _, complaint := range complaints {
		result := map[string]interface{}{
			"complaint_id":    complaint.ID.String(),
			"agent_name":      complaint.AgentName,
			"session_name":    complaint.SessionName,
			"task_description": complaint.TaskDescription,
			"severity":        string(complaint.Severity),
			"timestamp":       complaint.Timestamp.Format(time.RFC3339),
			"resolved":        complaint.IsResolved(),
			"project_name":    complaint.ProjectName,
		}
		results = append(results, result)
	}

	logger.Info("Complaints searched successfully", "query", query, "count", len(results))
	
	return map[string]interface{}{
		"complaints": results,
		"query":      query,
		"count":      len(results),
	}, nil
}