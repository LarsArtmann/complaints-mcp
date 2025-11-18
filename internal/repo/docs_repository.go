package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
	"github.com/larsartmann/complaints-mcp/internal/types"
)

// DocsRepository handles exporting complaints to documentation formats
type DocsRepository struct {
	docsDir string
	format  types.DocsFormat
	enabled bool
	logger  *log.Logger
	tracer  tracing.Tracer
}

// NewDocsRepository creates a new documentation repository
func NewDocsRepository(docsDir, format string, enabled bool, logger *log.Logger, tracer tracing.Tracer) *DocsRepository {
	// Convert string to strong type
	docsFormat := types.DocsFormat(format)
	if !docsFormat.IsValid() {
		logger.Warn("Invalid docs format, falling back to markdown", "format", format)
		docsFormat = types.DocsFormatMarkdown
	}

	return &DocsRepository{
		docsDir: docsDir,
		format:  docsFormat,
		enabled: enabled,
		logger:  logger,
		tracer:  tracer,
	}
}

// ExportToDocs exports a complaint to documentation format
func (d *DocsRepository) ExportToDocs(complaint *domain.Complaint) error {
	if !d.enabled {
		d.logger.Debug("Documentation export disabled")
		return nil
	}

	ctx, span := d.tracer.Start(context.Background(), "docs_export")
	_ = ctx // Use context but don't need it for current implementation
	defer span.End()

	switch d.format {
	case types.DocsFormatMarkdown:
		return d.exportToMarkdown(complaint)
	case types.DocsFormatHTML:
		return d.exportToHTML(complaint)
	case types.DocsFormatText:
		return d.exportToText(complaint)
	default:
		return fmt.Errorf("unsupported documentation format: %s", d.format)
	}
}

// GenerateDocsFilename generates human-readable documentation filename using strong types
func (d *DocsRepository) GenerateDocsFilename(complaint *domain.Complaint) string {
	return types.GenerateFilename(complaint.Timestamp, complaint.SessionName.String(), d.format)
}

// exportToMarkdown exports complaint to markdown format
func (d *DocsRepository) exportToMarkdown(complaint *domain.Complaint) error {
	filename := d.GenerateDocsFilename(complaint)
	filepath := filepath.Join(d.docsDir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(d.docsDir, 0o755); err != nil {
		d.logger.Error("Failed to create docs directory", "error", err, "path", d.docsDir)
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Markdown template
	tmpl := `# {{.AgentName}} Complaint

**Created:** {{.Timestamp.Format "2006-01-02 15:04:05"}}  
**Session:** {{.SessionName}}  
**Severity:** {{.Severity}}  
**Project:** {{.ProjectName}}  
**Status:** {{if .Resolved}}✅ Resolved{{else}}🔄 Open{{end}}  

{{if .Resolved}}**Resolved By:** {{.ResolvedBy}}  
**Resolved At:** {{.ResolvedAt.Format "2006-01-02 15:04:05"}}{{end}}

---

## Task Description

{{.TaskDescription}}

---

## Context Information

{{.ContextInfo}}

---

## Missing Information

{{.MissingInfo}}

---

## What Confused Me

{{.ConfusedBy}}

---

## Future Wishes

{{.FutureWishes}}

---

## Metadata

**Complaint ID:** {{.ID}}  
**Timestamp:** {{.Timestamp.Format "2006-01-02T15:04:05Z07:00"}}  

---

*This complaint was filed via complaints-mcp system and is stored for infinite retention as documentation.*`

	// Execute template
	t, err := template.New("complaint").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse markdown template: %w", err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create markdown file: %w", err)
	}
	defer file.Close()

	if err := t.Execute(file, complaint); err != nil {
		return fmt.Errorf("failed to execute markdown template: %w", err)
	}

	d.logger.Info("Complaint exported to documentation",
		"format", "markdown",
		"file", filepath,
		"complaint_id", complaint.ID.String())

	return nil
}

// exportToHTML exports complaint to HTML format
func (d *DocsRepository) exportToHTML(complaint *domain.Complaint) error {
	filename := d.GenerateDocsFilename(complaint)
	filepath := filepath.Join(d.docsDir, strings.Replace(filename, types.DocsFormatMarkdown.FileExtension(), types.DocsFormatHTML.FileExtension(), 1))

	// Ensure directory exists
	if err := os.MkdirAll(d.docsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// HTML template
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.AgentName}} Complaint</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .header { border-bottom: 2px solid #333; padding-bottom: 10px; margin-bottom: 20px; }
        .section { margin-bottom: 20px; }
        .field { font-weight: bold; }
        .status { color: {{if .Resolved}}green{{else}}orange{{end}}; }
        .metadata { background-color: #f5f5f5; padding: 10px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.AgentName}} Complaint</h1>
        <p class="status">Status: {{if .Resolved}}✅ Resolved{{else}}🔄 Open{{end}}</p>
    </div>
    
    <div class="section">
        <p><span class="field">Created:</span> {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
        <p><span class="field">Session:</span> {{.SessionName}}</p>
        <p><span class="field">Severity:</span> {{.Severity}}</p>
        <p><span class="field">Project:</span> {{.ProjectName}}</p>
        {{if .Resolved}}
        <p><span class="field">Resolved By:</span> {{.ResolvedBy}}</p>
        <p><span class="field">Resolved At:</span> {{.ResolvedAt.Format "2006-01-02 15:04:05"}}</p>
        {{end}}
    </div>
    
    <div class="section">
        <h2>Task Description</h2>
        <p>{{.TaskDescription}}</p>
    </div>
    
    <div class="section">
        <h2>Context Information</h2>
        <p>{{.ContextInfo}}</p>
    </div>
    
    <div class="section">
        <h2>Missing Information</h2>
        <p>{{.MissingInfo}}</p>
    </div>
    
    <div class="section">
        <h2>What Confused Me</h2>
        <p>{{.ConfusedBy}}</p>
    </div>
    
    <div class="section">
        <h2>Future Wishes</h2>
        <p>{{.FutureWishes}}</p>
    </div>
    
    <div class="section metadata">
        <h2>Metadata</h2>
        <p><span class="field">Complaint ID:</span> {{.ID}}</p>
        <p><span class="field">Timestamp:</span> {{.Timestamp.Format "2006-01-02T15:04:05Z07:00"}}</p>
    </div>
    
    <footer>
        <p><em>This complaint was filed via the complaints-mcp system and is stored for infinite retention as documentation.</em></p>
    </footer>
</body>
</html>`

	// Execute template
	t, err := template.New("complaint").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()

	if err := t.Execute(file, complaint); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	d.logger.Info("Complaint exported to documentation",
		"format", "html",
		"file", filepath,
		"complaint_id", complaint.ID.String())

	return nil
}

// exportToText exports complaint to plain text format
func (d *DocsRepository) exportToText(complaint *domain.Complaint) error {
	filename := d.GenerateDocsFilename(complaint)
	filepath := filepath.Join(d.docsDir, strings.Replace(filename, types.DocsFormatMarkdown.FileExtension(), types.DocsFormatText.FileExtension(), 1))

	// Ensure directory exists
	if err := os.MkdirAll(d.docsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Plain text template
	tmpl := `{{.AgentName}} Complaint
=====================

Created: {{.Timestamp.Format "2006-01-02 15:04:05"}}
Session: {{.SessionName}}
Severity: {{.Severity}}
Project: {{.ProjectName}}
Status: {{if .Resolved}}✅ Resolved{{else}}🔄 Open{{end}}
{{if .Resolved}}Resolved By: {{.ResolvedBy}}
Resolved At: {{.ResolvedAt.Format "2006-01-02 15:04:05"}}{{end}}

Task Description
----------------
{{.TaskDescription}}

Context Information
-------------------
{{.ContextInfo}}

Missing Information
------------------
{{.MissingInfo}}

What Confused Me
----------------
{{.ConfusedBy}}

Future Wishes
--------------
{{.FutureWishes}}

Metadata
---------
Complaint ID: {{.ID}}
Timestamp: {{.Timestamp.Format "2006-01-02T15:04:05Z07:00"}}

This complaint was filed via the complaints-mcp system and is stored for infinite retention as documentation.
`

	// Execute template
	t, err := template.New("complaint").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse text template: %w", err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create text file: %w", err)
	}
	defer file.Close()

	if err := t.Execute(file, complaint); err != nil {
		return fmt.Errorf("failed to execute text template: %w", err)
	}

	d.logger.Info("Complaint exported to documentation",
		"format", "text",
		"file", filepath,
		"complaint_id", complaint.ID.String())

	return nil
}

// IsEnabled returns whether documentation export is enabled
func (d *DocsRepository) IsEnabled() bool {
	return d.enabled
}

// GetFormat returns the current documentation format
func (d *DocsRepository) GetFormat() types.DocsFormat {
	return d.format
}

// GetDocsDir returns the documentation directory
func (d *DocsRepository) GetDocsDir() string {
	return d.docsDir
}
