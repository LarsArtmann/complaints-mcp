package delivery

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServer represents MCP server implementation.
type MCPServer struct {
	config  *config.Config
	service *service.ComplaintService
	logger  *log.Logger
	tracer  tracing.Tracer
	server  *mcp.Server
}

// NewServer creates a new MCP server.
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

// SetConfig sets the configuration for the MCP server.
func (m *MCPServer) SetConfig(cfg *config.Config) {
	m.config = cfg
}

// Start starts the MCP server using stdio transport.
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

// Shutdown gracefully shuts down the MCP server.
func (m *MCPServer) Shutdown(ctx context.Context) error {
	logger := m.logger.With("component", "mcp-server")
	logger.Info("Shutting down MCP server")

	// MCP server with stdio transport handles shutdown automatically
	return nil
}

// registerTools registers all available MCP tools.
func (m *MCPServer) registerTools() error {
	// File complaint tool
	fileComplaintTool := &mcp.Tool{
		Name:        "file_complaint",
		Description: "File a structured complaint about missing or confusing information",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent_name": map[string]any{
					"type":        "string",
					"description": "Name of the AI agent filing the complaint",
					"minLength":   1,
					"maxLength":   100,
				},
				"session_name": map[string]any{
					"type":        "string",
					"description": "Name of the current session",
					"maxLength":   100,
				},
				"task_description": map[string]any{
					"type":        "string",
					"description": "Description of the task being performed",
					"minLength":   1,
					"maxLength":   1000,
				},
				"context_info": map[string]any{
					"type":        "string",
					"description": "Additional context information",
					"maxLength":   500,
				},
				"missing_info": map[string]any{
					"type":        "string",
					"description": "What information was missing or unclear",
					"maxLength":   500,
				},
				"confused_by": map[string]any{
					"type":        "string",
					"description": "What aspects were confusing",
					"maxLength":   500,
				},
				"future_wishes": map[string]any{
					"type":        "string",
					"description": "Suggestions for future improvements",
					"maxLength":   500,
				},
				"severity": map[string]any{
					"type":        "string",
					"description": "Severity level (low, medium, high, critical)",
					"enum":        []string{"low", "medium", "high", "critical"},
				},
				"project_name": map[string]any{
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
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"limit": map[string]any{
					"type":        "integer",
					"description": "Maximum number of complaints to return",
					"minimum":     1,
					"maximum":     100,
				},
				"severity": map[string]any{
					"type":        "string",
					"description": "Filter by severity level",
					"enum":        []string{"low", "medium", "high", "critical"},
				},
				"resolved": map[string]any{
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
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"complaint_id": map[string]any{
					"type":        "string",
					"description": "Unique identifier of the complaint",
					"pattern":     "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
				},
				"resolved_by": map[string]any{
					"type":        "string",
					"description": "Identifier of who resolved the complaint (agent name, user ID, etc.)",
				},
			},
			"required": []string{"complaint_id", "resolved_by"},
		},
	}

	// Search complaints tool
	searchComplaintsTool := &mcp.Tool{
		Name:        "search_complaints",
		Description: "Search complaints by content",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query text",
					"minLength":   1,
					"maxLength":   500,
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Maximum number of results",
					"minimum":     1,
					"maximum":     100,
				},
			},
			"required": []string{"query"},
		},
	}

	// Get cache stats tool
	getCacheStatsTool := &mcp.Tool{
		Name:        "get_cache_stats",
		Description: "Get cache performance statistics",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
			"required":   []string{},
		},
	}

	// Register tools with handlers
	mcp.AddTool(m.server, fileComplaintTool, m.handleFileComplaint)
	mcp.AddTool(m.server, listComplaintsTool, m.handleListComplaints)
	mcp.AddTool(m.server, resolveComplaintTool, m.handleResolveComplaint)
	mcp.AddTool(m.server, searchComplaintsTool, m.handleSearchComplaints)
	mcp.AddTool(m.server, getCacheStatsTool, m.handleGetCacheStats)

	return nil
}

// Input types for tool handlers.
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
	ResolvedBy  string `json:"resolved_by"`
}

type SearchComplaintsInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type GetCacheStatsInput struct{}

// Output types for tool handlers.
type FileComplaintOutput struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Complaint ComplaintDTO `json:"complaint"` // ✅ Type-safe instead of string ID
}

type ListComplaintsOutput struct {
	Complaints []ComplaintDTO `json:"complaints"` // ✅ Type-safe instead of []map[string]any
}

type ResolveComplaintOutput struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Complaint ComplaintDTO `json:"complaint"` // ✅ Type-safe instead of string ID
}

type SearchComplaintsOutput struct {
	Complaints []ComplaintDTO `json:"complaints"` // ✅ Type-safe instead of []map[string]any
	Query      string         `json:"query"`
}

type GetCacheStatsOutput struct {
	CacheEnabled bool            `json:"cache_enabled"`
	Stats        repo.CacheStats `json:"stats"`
	Message      string          `json:"message"`
}

// handleFileComplaint handles the file_complaint tool.
func (m *MCPServer) handleFileComplaint(ctx context.Context, req *mcp.CallToolRequest, input FileComplaintInput) (*mcp.CallToolResult, FileComplaintOutput, error) {
	ctx, span := m.tracer.Start(ctx, "handleFileComplaint")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "file_complaint")
	logger.Info("Handling file complaint request")

	// Parse severity with type safety (eliminates runtime errors)
	domainSeverity, err := domain.ParseSeverity(input.Severity)
	if err != nil {
		return nil, FileComplaintOutput{}, fmt.Errorf("invalid severity: %w", err)
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

	// Get file paths for the complaint
	filePath, docsPath, err := m.service.GetFilePaths(ctx, complaint.ID)
	if err != nil {
		logger.Warn("Failed to get file paths for complaint", "error", err, "complaint_id", complaint.ID.String())
		// Continue without file paths - not a fatal error
	}

	output := FileComplaintOutput{
		Success:   true,
		Message:   "Complaint filed successfully",
		Complaint: ToDTOWithPaths(complaint, filePath, docsPath),
	}

	return nil, output, nil
}

// handleListComplaints handles the list_complaints tool.
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
		var err error
		severityFilter, err = domain.ParseSeverity(input.Severity)
		if err != nil {
			return nil, ListComplaintsOutput{}, fmt.Errorf("invalid severity filter: %w", err)
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
	var results []ComplaintDTO
	for _, complaint := range complaints {
		// Apply resolved filter if set
		if input.Resolved && !complaint.IsResolved() {
			continue
		}
		if !input.Resolved && complaint.IsResolved() {
			continue
		}

		results = append(results, ToDTO(complaint))
	}

	logger.Info("Complaints listed successfully", "count", len(results))

	output := ListComplaintsOutput{
		Complaints: results,
	}

	return nil, output, nil
}

// handleResolveComplaint handles the resolve_complaint tool.
func (m *MCPServer) handleResolveComplaint(ctx context.Context, req *mcp.CallToolRequest, input ResolveComplaintInput) (*mcp.CallToolResult, ResolveComplaintOutput, error) {
	ctx, span := m.tracer.Start(ctx, "handleResolveComplaint")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "resolve_complaint")
	logger.Info("Handling resolve complaint request")

	complaintID := domain.ComplaintID(input.ComplaintID)

	complaint, err := m.service.ResolveComplaint(ctx, complaintID, input.ResolvedBy)
	if err != nil {
		logger.Error("Failed to resolve complaint", "error", err, "complaint_id", input.ComplaintID, "resolved_by", input.ResolvedBy)
		return nil, ResolveComplaintOutput{}, err
	}

	logger.Info("Complaint resolved successfully", "complaint_id", input.ComplaintID, "resolved_by", input.ResolvedBy)

	output := ResolveComplaintOutput{
		Success:   true,
		Message:   "Complaint resolved successfully",
		Complaint: ToDTO(complaint),
	}

	return nil, output, nil
}

// handleSearchComplaints handles the search_complaints tool.
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
	var results []ComplaintDTO
	for _, complaint := range complaints {
		results = append(results, ToDTO(complaint))
	}

	logger.Info("Complaints searched successfully", "query", input.Query, "count", len(results))

	output := SearchComplaintsOutput{
		Complaints: results,
		Query:      input.Query,
	}

	return nil, output, nil
}

// handleGetCacheStats handles the get_cache_stats tool.
func (m *MCPServer) handleGetCacheStats(ctx context.Context, req *mcp.CallToolRequest, input GetCacheStatsInput) (*mcp.CallToolResult, GetCacheStatsOutput, error) {
	ctx, span := m.tracer.Start(ctx, "handleGetCacheStats")
	defer span.End()

	logger := m.logger.With("component", "mcp-server", "tool", "get_cache_stats")
	logger.Info("Handling get cache stats request")

	stats := m.service.GetCacheStats()

	// Determine if cache is enabled (non-zero max size indicates cached repository)
	cacheEnabled := stats.MaxCacheSize > 0
	message := "Cache statistics retrieved successfully"
	if !cacheEnabled {
		message = "Cache disabled (using FileRepository)"
	}

	output := GetCacheStatsOutput{
		CacheEnabled: cacheEnabled,
		Stats:        stats,
		Message:      message,
	}

	logger.Info("Cache stats retrieved successfully",
		"cache_enabled", cacheEnabled,
		"hit_rate", stats.HitRate,
		"current_size", stats.CurrentSize,
		"max_size", stats.MaxCacheSize)

	return nil, output, nil
}
