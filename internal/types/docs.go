package types

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// DocsFormat represents supported documentation export formats.
type DocsFormat string

const (
	DocsFormatMarkdown DocsFormat = "markdown"
	DocsFormatHTML     DocsFormat = "html"
	DocsFormatText     DocsFormat = "text"
)

// IsValid validates that the docs format is supported.
func (df DocsFormat) IsValid() bool {
	switch df {
	case DocsFormatMarkdown, DocsFormatHTML, DocsFormatText:
		return true
	default:
		return false
	}
}

// String returns string representation.
func (df DocsFormat) String() string {
	return string(df)
}

// FileExtension returns the file extension for this format.
func (df DocsFormat) FileExtension() string {
	switch df {
	case DocsFormatMarkdown:
		return ".md"
	case DocsFormatHTML:
		return ".html"
	case DocsFormatText:
		return ".txt"
	default:
		return ".txt"
	}
}

// GenerateFilename generates a safe filename for documentation export.
func GenerateFilename(timestamp time.Time, sessionName string, format DocsFormat) string {
	// Format timestamp
	timeStr := timestamp.Format("2006-01-02_15-04-05")

	// Handle session name
	if sessionName == "" {
		sessionName = "no-session"
	}

	// Sanitize session name for filename
	sessionName = strings.ReplaceAll(sessionName, " ", "_")
	sessionName = strings.ReplaceAll(sessionName, "/", "_")
	sessionName = strings.ReplaceAll(sessionName, "..", "_")
	sessionName = strings.ReplaceAll(sessionName, ":", "-")
	sessionName = strings.ReplaceAll(sessionName, "\"", "")
	sessionName = strings.ReplaceAll(sessionName, "'", "")
	sessionName = strings.ReplaceAll(sessionName, "\\", "_")
	sessionName = strings.ReplaceAll(sessionName, "<", "")
	sessionName = strings.ReplaceAll(sessionName, ">", "")
	sessionName = strings.ReplaceAll(sessionName, "|", "")
	sessionName = strings.ReplaceAll(sessionName, "?", "")
	sessionName = strings.ReplaceAll(sessionName, "*", "")

	// Remove multiple underscores
	for strings.Contains(sessionName, "__") {
		sessionName = strings.ReplaceAll(sessionName, "__", "_")
	}

	// Trim underscores
	sessionName = strings.Trim(sessionName, "_")

	// Limit length
	if len(sessionName) > 50 {
		sessionName = sessionName[:50]
	}

	return fmt.Sprintf("%s-%s%s", timeStr, sessionName, format.FileExtension())
}

// ValidateDocsDir ensures docs directory is valid and safe.
func ValidateDocsDir(docsDir string) error {
	if docsDir == "" {
		return errors.New("docs directory cannot be empty")
	}

	// Clean path
	docsDir = filepath.Clean(docsDir)

	// Check for path traversal attempts
	if strings.Contains(docsDir, "..") {
		return fmt.Errorf("docs directory cannot contain path traversal elements: %s", docsDir)
	}

	// Check absolute paths (should be relative to project root)
	if filepath.IsAbs(docsDir) {
		return fmt.Errorf("docs directory should be relative to project root, not absolute: %s", docsDir)
	}

	return nil
}

// DocsConfig represents strongly-typed documentation configuration.
type DocsConfig struct {
	Dir     string     `validate:"required,dir"`
	Format  DocsFormat `validate:"required,oneof=markdown html text"`
	Enabled bool       `validate:"boolean"`
}

// Validate validates the docs configuration.
func (dc DocsConfig) Validate() error {
	if err := ValidateDocsDir(dc.Dir); err != nil {
		return fmt.Errorf("invalid docs directory: %w", err)
	}

	if !dc.Format.IsValid() {
		return fmt.Errorf("invalid docs format: %s", dc.Format)
	}

	return nil
}

// String returns string representation of config.
func (dc DocsConfig) String() string {
	return fmt.Sprintf("DocsConfig{Dir: %s, Format: %s, Enabled: %t}",
		dc.Dir, dc.Format, dc.Enabled)
}
