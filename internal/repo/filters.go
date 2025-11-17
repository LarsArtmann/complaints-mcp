package repo

import (
	"context"
	"strings"

	"github.com/larsartmann/complaints-mcp/internal/domain"
)

// FilterStrategy defines a function that filters complaints
// This enables composition and eliminates code duplication across repositories
type FilterStrategy func(*domain.Complaint) bool

// filterComplaints applies a filter strategy to a slice of complaints
// This is the core abstraction that eliminates 60%+ code duplication
func filterComplaints(
	ctx context.Context,
	complaints []*domain.Complaint,
	filter FilterStrategy,
	limit int,
) []*domain.Complaint {
	var filtered []*domain.Complaint
	count := 0

	for _, complaint := range complaints {
		// Apply the filter strategy
		if filter(complaint) {
			filtered = append(filtered, complaint)
			count++

			// Stop if we've reached the limit
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	return filtered
}

// SeverityFilter creates a filter for a specific severity level
func SeverityFilter(severity domain.Severity) FilterStrategy {
	return func(c *domain.Complaint) bool {
		return c.Severity == severity
	}
}

// ProjectFilter creates a filter for a specific project name
func ProjectFilter(projectName string) FilterStrategy {
	return func(c *domain.Complaint) bool {
		return c.ProjectName.String() == projectName
	}
}

// UnresolvedFilter creates a filter for unresolved complaints
func UnresolvedFilter() FilterStrategy {
	return func(c *domain.Complaint) bool {
		return !c.IsResolved()
	}
}

// SearchFilter creates a filter for text search across complaint fields
func SearchFilter(query string) FilterStrategy {
	queryLower := strings.ToLower(query)

	return func(c *domain.Complaint) bool {
		// Search in task description
		if strings.Contains(strings.ToLower(c.TaskDescription), queryLower) {
			return true
		}

		// Search in context info
		if strings.Contains(strings.ToLower(c.ContextInfo), queryLower) {
			return true
		}

		// Search in missing info
		if strings.Contains(strings.ToLower(c.MissingInfo), queryLower) {
			return true
		}

		// Search in confused by
		if strings.Contains(strings.ToLower(c.ConfusedBy), queryLower) {
			return true
		}

		// Search in future wishes
		if strings.Contains(strings.ToLower(c.FutureWishes), queryLower) {
			return true
		}

		// Search in agent name
		if strings.Contains(strings.ToLower(c.AgentName.String()), queryLower) {
			return true
		}

		// Search in session name
		if strings.Contains(strings.ToLower(c.SessionName.String()), queryLower) {
			return true
		}

		// Search in project name
		if strings.Contains(strings.ToLower(c.ProjectName.String()), queryLower) {
			return true
		}

		return false
	}
}

// AndFilter combines multiple filters with AND logic
func AndFilter(filters ...FilterStrategy) FilterStrategy {
	return func(c *domain.Complaint) bool {
		for _, filter := range filters {
			if !filter(c) {
				return false
			}
		}
		return true
	}
}

// OrFilter combines multiple filters with OR logic
func OrFilter(filters ...FilterStrategy) FilterStrategy {
	return func(c *domain.Complaint) bool {
		for _, filter := range filters {
			if filter(c) {
				return true
			}
		}
		return false
	}
}

// NotFilter negates a filter
func NotFilter(filter FilterStrategy) FilterStrategy {
	return func(c *domain.Complaint) bool {
		return !filter(c)
	}
}
