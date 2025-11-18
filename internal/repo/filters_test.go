package repo

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/complaints-mcp/internal/domain"
)

func TestSeverityFilter(t *testing.T) {
	ctx := context.Background()

	// Create test complaints
	highComplaint := createTestComplaint(t, "high")
	lowComplaint := createTestComplaint(t, "low")

	complaints := []*domain.Complaint{highComplaint, lowComplaint}

	// Filter for high severity
	filtered := filterComplaints(ctx, complaints, SeverityFilter(domain.SeverityHigh), 0)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 high severity complaint, got %d", len(filtered))
	}

	if filtered[0].Severity != domain.SeverityHigh {
		t.Errorf("Expected high severity, got %s", filtered[0].Severity)
	}
}

func TestProjectFilter(t *testing.T) {
	ctx := context.Background()

	complaint1 := createTestComplaint(t, "low")
	complaint1.ProjectName = domain.MustNewProjectName("project-a")

	complaint2 := createTestComplaint(t, "high")
	complaint2.ProjectName = domain.MustNewProjectName("project-b")

	complaints := []*domain.Complaint{complaint1, complaint2}

	// Filter for project-a
	filtered := filterComplaints(ctx, complaints, ProjectFilter("project-a"), 0)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 complaint for project-a, got %d", len(filtered))
	}

	if filtered[0].ProjectName.String() != "project-a" {
		t.Errorf("Expected project-a, got %s", filtered[0].ProjectName.String())
	}
}

func TestUnresolvedFilter(t *testing.T) {
	ctx := context.Background()

	unresolvedComplaint := createTestComplaint(t, "low")

	resolvedComplaint := createTestComplaint(t, "high")
	now := time.Now()
	resolvedComplaint.ResolvedAt = &now
	resolvedComplaint.ResolvedBy = "test-agent"

	complaints := []*domain.Complaint{unresolvedComplaint, resolvedComplaint}

	// Filter for unresolved
	filtered := filterComplaints(ctx, complaints, UnresolvedFilter(), 0)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 unresolved complaint, got %d", len(filtered))
	}

	if filtered[0].IsResolved() {
		t.Error("Expected unresolved complaint, got resolved")
	}
}

func TestSearchFilter(t *testing.T) {
	ctx := context.Background()

	complaint1 := createTestComplaint(t, "low")
	complaint1.TaskDescription = "Fix performance issue"

	complaint2 := createTestComplaint(t, "high")
	complaint2.TaskDescription = "Add new feature"

	complaints := []*domain.Complaint{complaint1, complaint2}

	// Search for "performance"
	filtered := filterComplaints(ctx, complaints, SearchFilter("performance"), 0)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 complaint matching 'performance', got %d", len(filtered))
	}

	if filtered[0].TaskDescription != "Fix performance issue" {
		t.Errorf("Wrong complaint filtered: %s", filtered[0].TaskDescription)
	}
}

func TestSearchFilterCaseInsensitive(t *testing.T) {
	ctx := context.Background()

	complaint := createTestComplaint(t, "low")
	complaint.TaskDescription = "FIX PERFORMANCE ISSUE"

	complaints := []*domain.Complaint{complaint}

	// Search with lowercase
	filtered := filterComplaints(ctx, complaints, SearchFilter("performance"), 0)

	if len(filtered) != 1 {
		t.Error("Search filter should be case-insensitive")
	}
}

func TestFilterComplaintsWithLimit(t *testing.T) {
	ctx := context.Background()

	// Create 5 high severity complaints
	var complaints []*domain.Complaint
	for i := 0; i < 5; i++ {
		complaints = append(complaints, createTestComplaint(t, "high"))
	}

	// Filter with limit of 2
	filtered := filterComplaints(ctx, complaints, SeverityFilter(domain.SeverityHigh), 2)

	if len(filtered) != 2 {
		t.Errorf("Expected limit of 2, got %d", len(filtered))
	}
}

func TestAndFilter(t *testing.T) {
	ctx := context.Background()

	complaint1 := createTestComplaint(t, "high")
	complaint1.ProjectName = domain.MustNewProjectName("project-a")

	complaint2 := createTestComplaint(t, "high")
	complaint2.ProjectName = domain.MustNewProjectName("project-b")

	complaint3 := createTestComplaint(t, "low")
	complaint3.ProjectName = domain.MustNewProjectName("project-a")

	complaints := []*domain.Complaint{complaint1, complaint2, complaint3}

	// Filter for high severity AND project-a
	filter := AndFilter(
		SeverityFilter(domain.SeverityHigh),
		ProjectFilter("project-a"),
	)

	filtered := filterComplaints(ctx, complaints, filter, 0)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 complaint (high + project-a), got %d", len(filtered))
	}

	if filtered[0].Severity != domain.SeverityHigh || filtered[0].ProjectName.String() != "project-a" {
		t.Error("AndFilter failed to filter correctly")
	}
}

func TestOrFilter(t *testing.T) {
	ctx := context.Background()

	highComplaint := createTestComplaint(t, "high")
	highComplaint.ProjectName = domain.MustNewProjectName("project-other")

	mediumComplaint := createTestComplaint(t, "medium")
	mediumComplaint.ProjectName = domain.MustNewProjectName("project-a")

	lowComplaint := createTestComplaint(t, "low")
	lowComplaint.ProjectName = domain.MustNewProjectName("project-other")

	complaints := []*domain.Complaint{highComplaint, mediumComplaint, lowComplaint}

	// Filter for high severity OR project-a
	filter := OrFilter(
		SeverityFilter(domain.SeverityHigh),
		ProjectFilter("project-a"),
	)

	filtered := filterComplaints(ctx, complaints, filter, 0)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 complaints (high OR project-a), got %d", len(filtered))
	}
}

func TestNotFilter(t *testing.T) {
	ctx := context.Background()

	unresolvedComplaint := createTestComplaint(t, "low")

	resolvedComplaint := createTestComplaint(t, "high")
	now := time.Now()
	resolvedComplaint.ResolvedAt = &now
	resolvedComplaint.ResolvedBy = "test-agent"

	complaints := []*domain.Complaint{unresolvedComplaint, resolvedComplaint}

	// Filter for NOT unresolved (i.e., resolved)
	filter := NotFilter(UnresolvedFilter())

	filtered := filterComplaints(ctx, complaints, filter, 0)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 resolved complaint, got %d", len(filtered))
	}

	if !filtered[0].IsResolved() {
		t.Error("Expected resolved complaint")
	}
}

// Helper function to create test complaints
func createTestComplaint(t *testing.T, severityStr string) *domain.Complaint {
	t.Helper()

	severity := domain.Severity(severityStr)
	complaint, err := domain.NewComplaint(
		context.Background(),
		"test-agent",
		"test-session",
		"test task",
		"test context",
		"test missing",
		"test confused",
		"test wishes",
		severity,
		"test-project",
	)

	if err != nil {
		t.Fatalf("Failed to create test complaint: %v", err)
	}

	return complaint
}
