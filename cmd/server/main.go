package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/larsartmann/complaints-mcp/internal/complaint"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FileComplaintArgs represents the arguments for the file_complaint tool
type FileComplaintArgs struct {
	TaskAskedToPerform string `json:"task_asked_to_perform" jsonschema:"What the agent was asked to do"`
	ContextInformation string `json:"context_information" jsonschema:"What context information was provided"`
	MissingInformation string `json:"missing_information" jsonschema:"What information was missing"`
	ConfusedBy         string `json:"confused_by" jsonschema:"What confused the agent"`
	FutureWishes       string `json:"future_wishes" jsonschema:"What the agent wishes for in the future"`
	SessionName        string `json:"session_name,omitempty" jsonschema:"Session name for the filename (optional)"`
	AgentName          string `json:"agent_name,omitempty" jsonschema:"Name of the agent filing the complaint (optional)"`
}

// FileComplaint handles the file_complaint tool
func FileComplaint(ctx context.Context, req *mcp.CallToolRequest, args FileComplaintArgs) (*mcp.CallToolResult, struct{}, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get current working directory: %v", err)},
			},
			IsError: true,
		}, struct{}{}, err
	}

	// Create complaint
	c := &complaint.Complaint{
		TaskAskedToPerform: args.TaskAskedToPerform,
		ContextInformation: args.ContextInformation,
		MissingInformation: args.MissingInformation,
		ConfusedBy:         args.ConfusedBy,
		FutureWishes:       args.FutureWishes,
		SessionName:        args.SessionName,
		AgentName:          args.AgentName,
	}

	// Save complaint
	filepath, err := c.Save(cwd)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to save complaint: %v", err)},
			},
			IsError: true,
		}, struct{}{}, err
	}

	// Success response
	message := fmt.Sprintf("Complaint filed successfully!\n\nSaved to: %s\n\nThank you for your feedback - this helps improve the development experience for AI agents.", filepath)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}, struct{}{}, nil
}

func main() {
	// Create MCP server with instructions for AI agents
	opts := &mcp.ServerOptions{
		Instructions: "You are connected to a complaints-mcp server. Use the file_complaint tool whenever you encounter missing, under-specified, or confusing information during development tasks. Your feedback is valuable and helps improve the development experience.",
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

	// Run the server over stdio transport
	log.Println("Starting complaints-mcp server...")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}