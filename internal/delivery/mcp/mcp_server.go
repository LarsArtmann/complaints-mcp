package delivery

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServer represents MCP server implementation
type MCPServer struct {
	config  *config.Config
	service *service.ComplaintService
	logger  *log.Logger
	tracer  tracing.Tracer
	server  *mcp.Server
}

// NewServer creates a new MCP server
func NewServer(name, version string, complaintService *service.ComplaintService, logger *log.Logger, tracer tracing.Tracer) *MCPServer {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: version,
	}, nil)

	return &MCPServer{
		config:  nil, // Will be set during initialization
		service: complaintService,
		logger:  logger,
		tracer:  tracer,
		server:  server,
	}
}

// SetConfig sets the configuration for the MCP server
func (m *MCPServer) SetConfig(cfg *config.Config) {
	m.config = cfg
}

// Start starts the MCP server using stdio transport
func (m *MCPServer) Start(ctx context.Context) error {
	logger := m.logger.With("component", "mcp-server")

	// Register tools
	if err := m.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Start server with stdio transport
	logger.Info("Starting MCP server over stdio")

	if err := m.server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the MCP server
func (m *MCPServer) Shutdown(ctx context.Context) error {
	logger := m.logger.With("component", "mcp-server")
	logger.Info("Shutting down MCP server")

	// MCP server with stdio transport handles shutdown automatically
	return nil
}

// registerTools registers all available MCP tools
func (m *MCPServer) registerTools() error {
	// File complaint tool
	fileComplaintTool := &mcp.Tool{
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
	listComplaintsTool := &mcp.Tool{
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
	resolveComplaintTool := &mcp.Tool{
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
	searchComplaintsTool := &mcp.Tool{
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
	mcp.AddTool(m.server, fileComplaintTool, m.handleFileComplaint)
	mcp.AddTool(m.server, listComplaintsTool, m.handleListComplaints)
	mcp.AddTool(m.server, resolveComplaintTool, m.handleResolveComplaint)
	mcp.AddTool(m.server, searchComplaintsTool, m.handleSearchComplaints)

	return nil
}

// Input types for tool handlers
type FileComplaintInput struct {
	AgentName       string `json:"agent_name"`
	SessionName     string `json:"session_name"`
	TaskDescription string `json:"task_description"`
	ContextInfo     string `json:"context_info"`
	MissingInfo     string `json:"missing_info"`
	ConfusedBy      string `json:"confused_by"`
	FutureWishes    string `json:"future_wishes"`
	Severity        string `json:"severity"`
	ProjectName     string `json:"project_name"`
}

type ListComplaintsInput struct {
	Limit    int    `json:"limit"`
	Severity string `json:"severity"`
	Resolved bool   `json:"resolved"`
}

type ResolveComplaintInput struct {
	ComplaintID string `json:"complaint_id"`
}

type SearchComplaintsInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// Output types for tool handlers
type FileComplaintOutput struct {
	ComplaintID string `json:"complaint_id"`
	Message     string `json:"message"`
	Timestamp   string `json:"timestamp"`
}

type ListComplaintsOutput struct {
	Complaints []map[string]interface{} `json:"complaints"`
	Count      int                      `json:"count"`
}

type ResolveComplaintOutput struct {
	Message     string `json:"message"`
	ComplaintID string `json:"complaint_id"`
}

type SearchComplaintsOutput struct {
	Complaints []map[string]interface{} `json:"complaints"`
	Query      string                   `json:"query"`
	Count      int                      `json:"count"`
}

// handleFileComplaint handles the file_complaint tool
func (m *MCPServer) handleFileComplaint(ctx context.Context, req *mcp.CallToolRequest, input FileComplaintInput) (*mcp.CallToolResult, FileComplaintOutput, error) {
	ctx, span := m.tracer.Start(ctx, "handleFileComplaint")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "file_complaint")
	logger.Info("Handling file complaint request")

	// Convert severity string to domain type
	var domainSeverity domain.Severity
	switch input.Severity {
	case "low":
		domainSeverity = domain.SeverityLow
	case "medium":
		domainSeverity = domain.SeverityMedium
	case "high":
		domainSeverity = domain.SeverityHigh
	case "critical":
		domainSeverity = domain.SeverityCritical
	default:
		return nil, FileComplaintOutput{}, fmt.Errorf("invalid severity: %s", input.Severity)
	}

	complaint, err := m.service.CreateComplaint(
		ctx,
		input.AgentName,
		input.SessionName,
		input.TaskDescription,
		input.ContextInfo,
		input.MissingInfo,
		input.ConfusedBy,
		input.FutureWishes,
		domainSeverity,
		input.ProjectName,
	)
	if err != nil {
		logger.Error("Failed to create complaint", "error", err)
		return nil, FileComplaintOutput{}, err
	}

	logger.Info("Complaint filed successfully", "complaint_id", complaint.ID.String())

	output := FileComplaintOutput{
		ComplaintID: complaint.ID.String(),
		Message:     "Complaint filed successfully",
		Timestamp:   complaint.Timestamp.Format(time.RFC3339),
	}

	return nil, output, nil
}

// handleListComplaints handles the list_complaints tool
func (m *MCPServer) handleListComplaints(ctx context.Context, req *mcp.CallToolRequest, input ListComplaintsInput) (*mcp.CallToolResult, ListComplaintsOutput, error) {
	ctx, span := m.tracer.Start(ctx, "handleListComplaints")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "list_complaints")
	logger.Info("Handling list complaints request")

	// Set defaults
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	var severityFilter domain.Severity
	if input.Severity != "" {
		switch input.Severity {
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

	var complaints []*domain.Complaint
	var err error

	if severityFilter != "" {
		complaints, err = m.service.GetComplaintsBySeverity(ctx, severityFilter, limit)
	} else {
		complaints, err = m.service.ListComplaints(ctx, limit, 0)
	}

	if err != nil {
		logger.Error("Failed to list complaints", "error", err)
		return nil, ListComplaintsOutput{}, err
	}

	// Convert to response format
	var results []map[string]interface{}
	for _, complaint := range complaints {
		// Apply resolved filter if set
		if input.Resolved && !complaint.IsResolved() {
			continue
		}
		if !input.Resolved && complaint.IsResolved() {
			continue
		}

		result := map[string]interface{}{
			"complaint_id":     complaint.ID.String(),
			"agent_name":       complaint.AgentName,
			"session_name":     complaint.SessionName,
			"task_description": complaint.TaskDescription,
			"severity":         string(complaint.Severity),
			"timestamp":        complaint.Timestamp.Format(time.RFC3339),
			"resolved":         complaint.IsResolved(),
			"project_name":     complaint.ProjectName,
		}
		results = append(results, result)
	}

	logger.Info("Complaints listed successfully", "count", len(results))

	output := ListComplaintsOutput{
		Complaints: results,
		Count:      len(results),
	}

	return nil, output, nil
}

// handleResolveComplaint handles the resolve_complaint tool
func (m *MCPServer) handleResolveComplaint(ctx context.Context, req *mcp.CallToolRequest, input ResolveComplaintInput) (*mcp.CallToolResult, ResolveComplaintOutput, error) {
	ctx, span := m.tracer.Start(ctx, "handleResolveComplaint")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "resolve_complaint")
	logger.Info("Handling resolve complaint request")

	complaintID := domain.ComplaintID{Value: input.ComplaintID}

	if err := m.service.ResolveComplaint(ctx, complaintID); err != nil {
		logger.Error("Failed to resolve complaint", "error", err, "complaint_id", input.ComplaintID)
		return nil, ResolveComplaintOutput{}, err
	}

	logger.Info("Complaint resolved successfully", "complaint_id", input.ComplaintID)

	output := ResolveComplaintOutput{
		Message:     "Complaint resolved successfully",
		ComplaintID: input.ComplaintID,
	}

	return nil, output, nil
}

// handleSearchComplaints handles the search_complaints tool
func (m *MCPServer) handleSearchComplaints(ctx context.Context, req *mcp.CallToolRequest, input SearchComplaintsInput) (*mcp.CallToolResult, SearchComplaintsOutput, error) {
	ctx, span := m.tracer.Start(ctx, "handleSearchComplaints")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "search_complaints")
	logger.Info("Handling search complaints request")

	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	complaints, err := m.service.SearchComplaints(ctx, input.Query, limit)
	if err != nil {
		logger.Error("Failed to search complaints", "error", err)
		return nil, SearchComplaintsOutput{}, err
	}

	// Convert to response format
	var results []map[string]interface{}
	for _, complaint := range complaints {
		result := map[string]interface{}{
			"complaint_id":     complaint.ID.String(),
			"agent_name":       complaint.AgentName,
			"session_name":     complaint.SessionName,
			"task_description": complaint.TaskDescription,
			"severity":         string(complaint.Severity),
			"timestamp":        complaint.Timestamp.Format(time.RFC3339),
			"resolved":         complaint.IsResolved(),
			"project_name":     complaint.ProjectName,
		}
		results = append(results, result)
	}

	logger.Info("Complaints searched successfully", "query", input.Query, "count", len(results))

	output := SearchComplaintsOutput{
		Complaints: results,
		Query:      input.Query,
		Count:      len(results),
	}

	return nil, output, nil
}
