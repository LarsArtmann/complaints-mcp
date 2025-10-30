package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FileComplaintArgs represents the arguments for the file_complaint tool
type FileComplaintArgs struct {
	TaskDescription  string `json:"task_description" jsonschema:"What the agent was asked to do"`
	ContextInfo      string `json:"context_info" jsonschema:"What context information was provided"`
	MissingInfo      string `json:"missing_info" jsonschema:"What information was missing"`
	ConfusedBy       string `json:"confused_by" jsonschema:"What confused the agent"`
	FutureWishes     string `json:"future_wishes" jsonschema:"What the agent wishes for in the future"`
	Severity         string `json:"severity" jsonschema:"Severity level: low, medium, high, or critical"`
	SessionName      string `json:"session_name,omitempty" jsonschema:"Session name for the filename (optional)"`
	AgentName        string `json:"agent_name,omitempty" jsonschema:"Name of the agent filing the complaint (optional)"`
	ProjectName      string `json:"project_name,omitempty" jsonschema:"Project name (optional, will auto-detect)"`
}

// ListComplaintsArgs represents the arguments for listing complaints
type ListComplaintsArgs struct {
	ProjectName string `json:"project_name,omitempty" jsonschema:"Filter by project name (optional)"`
	Unresolved  *bool  `json:"unresolved,omitempty" jsonschema:"Filter for unresolved complaints only (optional)"`
	Limit       int    `json:"limit,omitempty" jsonschema:"Maximum number of complaints to return (default: 50)"`
	Offset      int    `json:"offset,omitempty" jsonschema:"Number of complaints to skip (default: 0)"`
}

// ResolveComplaintArgs represents the arguments for resolving a complaint
type ResolveComplaintArgs struct {
	ComplaintID string `json:"complaint_id" jsonschema:"ID of the complaint to resolve"`
}

// Global service and configuration
var (
	complaintService *service.ComplaintService
	appConfig        *config.Config
)

// initService initializes the complaint service and configuration
func initService() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	appConfig = cfg

	// Get base path for storage
	basePath := cfg.Complaints.ProjectName
	if basePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		basePath = cwd
	}

	// Initialize repository
	repository := repo.NewFileRepository(basePath)

	// Initialize service
	complaintService = service.NewComplaintService(repository, cfg)

	return nil
}

// FileComplaint handles the file_complaint tool
func FileComplaint(ctx context.Context, req *mcp.CallToolRequest, args FileComplaintArgs) (*mcp.CallToolResult, struct{}, error) {
	// Initialize service if not already done
	if complaintService == nil {
		if err := initService(); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to initialize service: %v", err)},
				},
				IsError: true,
			}, struct{}{}, err
		}
	}

	// Set default severity if not provided
	if args.Severity == "" {
		args.Severity = "medium"
	}

	// Create complaint request
	createReq := &service.CreateComplaintRequest{
		AgentName:       args.AgentName,
		SessionName:     args.SessionName,
		TaskDescription: args.TaskDescription,
		ContextInfo:     args.ContextInfo,
		MissingInfo:     args.MissingInfo,
		ConfusedBy:      args.ConfusedBy,
		FutureWishes:    args.FutureWishes,
		Severity:        args.Severity,
		ProjectName:     args.ProjectName,
	}

	// Create complaint
	complaint, err := complaintService.CreateComplaint(ctx, createReq)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create complaint: %v", err)},
			},
			IsError: true,
		}, struct{}{}, err
	}

	// Success response
	message := fmt.Sprintf("✅ **Complaint filed successfully!**\n\n**ID:** %s\n**Severity:** %s\n**Project:** %s\n\nYour feedback helps improve the development experience for AI agents.", 
		complaint.ID.Value, complaint.Severity, complaint.ProjectName)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}, struct{}{}, nil
}

// ListComplaints handles the list_complaints tool
func ListComplaints(ctx context.Context, req *mcp.CallToolRequest, args ListComplaintsArgs) (*mcp.CallToolResult, struct{}, error) {
	// Initialize service if not already done
	if complaintService == nil {
		if err := initService(); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to initialize service: %v", err)},
				},
				IsError: true,
			}, struct{}{}, err
		}
	}

	var complaints []*domain.Complaint
	var err error

	if args.Unresolved != nil && *args.Unresolved {
		complaints, err = complaintService.ListUnresolvedComplaints(ctx, args.Limit, args.Offset)
	} else if args.ProjectName != "" {
		complaints, err = complaintService.ListComplaintsByProject(ctx, args.ProjectName, args.Limit, args.Offset)
	} else {
		// Get all complaints by checking all projects
		complaints, err = complaintService.ListComplaintsByProject(ctx, "", args.Limit, args.Offset)
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list complaints: %v", err)},
			},
			IsError: true,
		}, struct{}{}, err
	}

	if len(complaints) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No complaints found matching the specified criteria."},
			},
		}, struct{}{}, nil
	}

	// Format complaints
	message := "## 📋 Complaints\n\n"
	for _, c := range complaints {
		status := "🔴 Open"
		if c.Resolved {
			status = "✅ Resolved"
		}
		
		message += fmt.Sprintf(`### %s
**ID:** %s  
**Status:** %s  
**Severity:** %s  
**Project:** %s  
**Agent:** %s  
**Session:** %s  
**Date:** %s  

**Task:** %s

---

`, c.TaskDescription, c.ID.Value, status, c.Severity, c.ProjectName, c.AgentName, c.SessionName, c.Timestamp.Format("2006-01-02 15:04:05"), c.TaskDescription)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}, struct{}{}, nil
}

// ResolveComplaint handles the resolve_complaint tool
func ResolveComplaint(ctx context.Context, req *mcp.CallToolRequest, args ResolveComplaintArgs) (*mcp.CallToolResult, struct{}, error) {
	// Initialize service if not already done
	if complaintService == nil {
		if err := initService(); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to initialize service: %v", err)},
				},
				IsError: true,
			}, struct{}{}, err
		}
	}

	complaintID := domain.ComplaintID{Value: args.ComplaintID}
	
	err := complaintService.ResolveComplaint(ctx, complaintID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to resolve complaint: %v", err)},
			},
			IsError: true,
		}, struct{}{}, err
	}

	message := fmt.Sprintf("✅ **Complaint %s marked as resolved!**", args.ComplaintID)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}, struct{}{}, nil
}

func main() {
	// Initialize service on startup
	if err := initService(); err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// Create MCP server with instructions for AI agents
	opts := &mcp.ServerOptions{
		Instructions: `You are connected to a complaints-mcp server. This system helps improve AI agent development experience by tracking and managing complaints about missing, unclear, or confusing information.

Available tools:
- file_complaint: File a new complaint about development issues
- list_complaints: View existing complaints (with filtering options)
- resolve_complaint: Mark complaints as resolved

Use these tools whenever you encounter information gaps, unclear requirements, or confusing documentation. Your feedback helps improve the development ecosystem for everyone.`,
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "complaints-mcp",
		Version: "1.0.0",
	}, opts)

	// Add the file_complaint tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "file_complaint",
		Description: "File a complaint report about missing/under-specified/confusing information. This helps improve the development experience for AI agents by identifying areas where information is unclear or insufficient.",
	}, FileComplaint)

	// Add the list_complaints tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_complaints",
		Description: "List existing complaints with optional filtering by project or resolution status.",
	}, ListComplaints)

	// Add the resolve_complaint tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "resolve_complaint",
		Description: "Mark a complaint as resolved when the issue has been addressed.",
	}, ResolveComplaint)

	// Run the server over stdio transport
	log.Printf("Starting complaints-mcp server on %s...", appConfig.GetServerAddress())
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}