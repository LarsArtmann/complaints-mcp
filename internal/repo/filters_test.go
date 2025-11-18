package repo

import (
	"context"
	"fmt"
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

func TestFilterComplaintsContextCancellation(t *testing.T) {
	// Create a context that will be cancelled after a short time
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Create many complaints to ensure we have time to cancel
	complaints := make([]*domain.Complaint, 100)
	for i := 0; i < 100; i++ {
		complaints[i] = createTestComplaint(t, "low")
	}

	// Use a filter that would match all complaints (no early termination via limit)
	filter := SeverityFilter(domain.SeverityLow)
	
	// Add a small delay in the filter to make cancellation more likely
	slowFilter := func(c *domain.Complaint) bool {
		time.Sleep(1 * time.Millisecond) // Small delay to ensure cancellation can happen
		return filter(c)
	}

	// This should return early due to context cancellation
	start := time.Now()
	filtered := filterComplaints(ctx, complaints, slowFilter, 0)
	elapsed := time.Since(start)

	// Should return very quickly due to cancellation (less than 100ms)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected filterComplaints to return early due to cancellation, took %v", elapsed)
	}

	// Should return fewer than all complaints due to cancellation
	if len(filtered) >= 100 {
		t.Errorf("Expected early return due to cancellation, got %d complaints", len(filtered))
	}

	t.Logf("Context cancellation test: filtered %d complaints in %v", len(filtered), elapsed)
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

func BenchmarkSearchFilter(b *testing.B) {
	ctx := context.Background()
	
	// Create a large test dataset
	complaints := make([]*domain.Complaint, 10000)
	for i := 0; i < 10000; i++ {
		severity := domain.SeverityLow
		if i%100 == 0 {
			severity = domain.SeverityHigh
		}
		
		complaint, err := domain.NewComplaint(
			ctx,
			"test-agent",
			"test-session",
			fmt.Sprintf("Test task %d with specific search term", i),
			"test context information",
			"test missing functionality",
			"test confused by behavior",
			"test wishes for improvement",
			severity,
			"test-project",
		)
		if err != nil {
			b.Fatalf("Failed to create test complaint: %v", err)
		}
		complaints[i] = complaint
	}

	// Benchmark search term that will match only a few items (early exit)
	filter := SearchFilter("specific search term")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filterComplaints(ctx, complaints, filter, 10)
	}
}

func BenchmarkSearchFilterNoMatch(b *testing.B) {
	ctx := context.Background()
	
	// Create a large test dataset
	complaints := make([]*domain.Complaint, 10000)
	for i := 0; i < 10000; i++ {
		complaint, err := domain.NewComplaint(
			ctx,
			"test-agent",
			"test-session",
			fmt.Sprintf("Test task %d", i),
			"test context information",
			"test missing functionality",
			"test confused by behavior",
			"test wishes for improvement",
			domain.SeverityLow,
			"test-project",
		)
		if err != nil {
			b.Fatalf("Failed to create test complaint: %v", err)
		}
		complaints[i] = complaint
	}

	// Benchmark search term that will match nothing (worst case)
	filter := SearchFilter("nonexistent term that will never match")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filterComplaints(ctx, complaints, filter, 10)
	}
}
