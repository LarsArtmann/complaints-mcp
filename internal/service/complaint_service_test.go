package service_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/errors"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

// mockRepository implements service.Repository for testing
type mockRepository struct {
	complaints []*domain.Complaint
}

func (m *mockRepository) Save(ctx context.Context, complaint *domain.Complaint) error {
	m.complaints = append(m.complaints, complaint)
	return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	for _, c := range m.complaints {
		if c.ID.Value == id.Value {
			return c, nil
		}
	}
	return nil, errors.NewNotFoundError("complaint not found")
}

func (m *mockRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Complaint, error) {
	if offset >= len(m.complaints) {
		return []*domain.Complaint{}, nil
	}
	
	end := offset + limit
	if end > len(m.complaints) {
		end = len(m.complaints)
	}
	
	result := make([]*domain.Complaint, 0, end-offset)
	for i := offset; i < end; i++ {
		result = append(result, m.complaints[i])
	}
	
	return result, nil
}

func (m *mockRepository) FindBySeverity(ctx context.Context, severity domain.Severity, limit int) ([]*domain.Complaint, error) {
	var result []*domain.Complaint
	for _, c := range m.complaints {
		if c.Severity == severity {
			result = append(result, c)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockRepository) FindByProject(ctx context.Context, projectName string, limit int) ([]*domain.Complaint, error) {
	var result []*domain.Complaint
	for _, c := range m.complaints {
		if c.ProjectName == projectName {
			result = append(result, c)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockRepository) FindUnresolved(ctx context.Context, limit int) ([]*domain.Complaint, error) {
	var result []*domain.Complaint
	for _, c := range m.complaints {
		if !c.Resolved {
			result = append(result, c)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Complaint, error) {
	var result []*domain.Complaint
	queryLower := strings.ToLower(query)
	
	for _, c := range m.complaints {
		// Simple case-insensitive search in task description and context
		if strings.Contains(strings.ToLower(c.TaskDescription), queryLower) || 
		   strings.Contains(strings.ToLower(c.ContextInfo), queryLower) {
			result = append(result, c)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockRepository) Update(ctx context.Context, complaint *domain.Complaint) error {
	for i, c := range m.complaints {
		if c.ID.Value == complaint.ID.Value {
			m.complaints[i] = complaint
			return nil
		}
	}
	return errors.NewNotFoundError("complaint not found")
}

func (m *mockRepository) Delete(ctx context.Context, id domain.ComplaintID) error {
	for i, c := range m.complaints {
		if c.ID.Value == id.Value {
			m.complaints = append(m.complaints[:i], m.complaints[i+1:]...)
			return nil
		}
	}
	return errors.NewNotFoundError("complaint not found")
}

func TestNewComplaintService(t *testing.T) {
	// ✅ Test with correct constructor signature
	mockRepo := &mockRepository{}
	mockTracer := tracing.NewMockTracer("test")
	mockLogger := log.New(io.Discard)

	// ✅ Use correct NewComplaintService signature
	svc := service.NewComplaintService(mockRepo, mockTracer, mockLogger)

	if svc == nil {
		t.Error("NewComplaintService() should not return nil")
	}
}

func TestComplaintService_CreateComplaint(t *testing.T) {
	// Setup
	mockRepo := &mockRepository{}
	mockTracer := tracing.NewMockTracer("test")
	mockLogger := log.New(io.Discard)
	svc := service.NewComplaintService(mockRepo, mockTracer, mockLogger)
	ctx := context.Background()

	// Test
	complaint, err := svc.CreateComplaint(ctx,
		"test-agent",
		"test-session",
		"test task",
		"test context",
		"missing info",
		"confused by",
		"future wishes",
		domain.SeverityHigh,
		"test-project")

	// Verify
	if err != nil {
		t.Errorf("CreateComplaint error = %v, want nil", err)
		return
	}

	if complaint == nil {
		t.Error("CreateComplaint should return complaint")
		return
	}

	if complaint.AgentName != "test-agent" {
		t.Errorf("AgentName = %v, want %v", complaint.AgentName, "test-agent")
	}

	if complaint.ProjectName != "test-project" {
		t.Errorf("ProjectName = %v, want %v", complaint.ProjectName, "test-project")
	}

	if complaint.Severity != domain.SeverityHigh {
		t.Errorf("Severity = %v, want %v", complaint.Severity, domain.SeverityHigh)
	}

	if complaint.Resolved != false {
		t.Errorf("Resolved = %v, want %v", complaint.Resolved, false)
	}

	if complaint.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

func TestComplaintService_CreateComplaint_ValidationError(t *testing.T) {
	// Setup
	mockRepo := &mockRepository{}
	mockTracer := tracing.NewMockTracer("test")
	mockLogger := log.New(io.Discard)
	svc := service.NewComplaintService(mockRepo, mockTracer, mockLogger)
	ctx := context.Background()

	// Test with empty agent name (should fail validation)
	complaint, err := svc.CreateComplaint(ctx,
		"", // invalid: empty agent name
		"test-session",
		"test task",
		"test context",
		"missing info",
		"confused by",
		"future wishes",
		domain.SeverityHigh,
		"test-project")

	// Verify
	if err == nil {
		t.Error("CreateComplaint should return error for invalid data")
		return
	}

	if complaint != nil {
		t.Error("CreateComplaint should return nil complaint on error")
		return
	}

	if !strings.Contains(err.Error(), "AgentName") && !strings.Contains(err.Error(), "required") {
		t.Errorf("Error should mention AgentName or required, got: %v", err)
	}
}

func TestComplaintService_GetComplaint(t *testing.T) {
	// Setup
	mockRepo := &mockRepository{}
	mockTracer := tracing.NewMockTracer("test")
	mockLogger := log.New(io.Discard)
	svc := service.NewComplaintService(mockRepo, mockTracer, mockLogger)
	ctx := context.Background()

	// Create a complaint first
	created, err := svc.CreateComplaint(ctx,
		"test-agent",
		"test-session",
		"test task",
		"test context",
		"missing info",
		"confused by",
		"future wishes",
		domain.SeverityHigh,
		"test-project")
	if err != nil {
		t.Fatalf("Failed to create complaint: %v", err)
	}

	// Test retrieving it
	retrieved, err := svc.GetComplaint(ctx, created.ID)

	// Verify
	if err != nil {
		t.Errorf("GetComplaint error = %v, want nil", err)
		return
	}

	if retrieved == nil {
		t.Error("GetComplaint should return complaint")
		return
	}

	if retrieved.ID.Value != created.ID.Value {
		t.Errorf("ID = %v, want %v", retrieved.ID.Value, created.ID.Value)
	}

	if retrieved.AgentName != created.AgentName {
		t.Errorf("AgentName = %v, want %v", retrieved.AgentName, created.AgentName)
	}
}

func TestComplaintService_GetComplaint_NotFound(t *testing.T) {
	// Setup
	mockRepo := &mockRepository{}
	mockTracer := tracing.NewMockTracer("test")
	mockLogger := log.New(io.Discard)
	svc := service.NewComplaintService(mockRepo, mockTracer, mockLogger)
	ctx := context.Background()

	// Create a non-existent ID
	nonExistentID, err := domain.NewComplaintID()
	if err != nil {
		t.Fatalf("Failed to create ID: %v", err)
	}

	// Test retrieving non-existent complaint
	retrieved, err := svc.GetComplaint(ctx, nonExistentID)

	// Verify
	if err == nil {
		t.Error("GetComplaint should return error for non-existent complaint")
		return
	}

	if retrieved != nil {
		t.Error("GetComplaint should return nil for non-existent complaint")
		return
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention not found, got: %v", err)
	}
}

func TestComplaintService_ResolveComplaint(t *testing.T) {
	// Setup
	mockRepo := &mockRepository{}
	mockTracer := tracing.NewMockTracer("test")
	mockLogger := log.New(io.Discard)
	svc := service.NewComplaintService(mockRepo, mockTracer, mockLogger)
	ctx := context.Background()

	// Create a complaint first
	created, err := svc.CreateComplaint(ctx,
		"test-agent",
		"test-session",
		"test task",
		"test context",
		"missing info",
		"confused by",
		"future wishes",
		domain.SeverityHigh,
		"test-project")
	if err != nil {
		t.Fatalf("Failed to create complaint: %v", err)
	}

	if created.Resolved {
		t.Fatal("Complaint should start as unresolved")
	}

	// Wait a bit to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Resolve the complaint
	err = svc.ResolveComplaint(ctx, created.ID)
	if err != nil {
		t.Errorf("ResolveComplaint error = %v, want nil", err)
		return
	}

	// Verify resolution
	resolved, err := svc.GetComplaint(ctx, created.ID)
	if err != nil {
		t.Errorf("GetComplaint error = %v, want nil", err)
		return
	}

	if !resolved.Resolved {
		t.Error("Complaint should be marked as resolved")
	}

	if resolved.ResolvedAt == nil {
		t.Error("Complaint should have ResolvedAt timestamp set")
	}
}

func TestComplaintService_ResolveComplaint_NotFound(t *testing.T) {
	// Setup
	mockRepo := &mockRepository{}
	mockTracer := tracing.NewMockTracer("test")
	mockLogger := log.New(io.Discard)
	svc := service.NewComplaintService(mockRepo, mockTracer, mockLogger)
	ctx := context.Background()

	// Create a non-existent ID
	nonExistentID, err := domain.NewComplaintID()
	if err != nil {
		t.Fatalf("Failed to create ID: %v", err)
	}

	// Try to resolve non-existent complaint
	err = svc.ResolveComplaint(ctx, nonExistentID)

	// Verify
	if err == nil {
		t.Error("ResolveComplaint should return error for non-existent complaint")
		return
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention not found, got: %v", err)
	}
}

func TestComplaintService_ListComplaints(t *testing.T) {
	// Setup
	mockRepo := &mockRepository{}
	mockTracer := tracing.NewMockTracer("test")
	mockLogger := log.New(io.Discard)
	svc := service.NewComplaintService(mockRepo, mockTracer, mockLogger)
	ctx := context.Background()

	// Create multiple complaints
	for i := 0; i < 5; i++ {
		_, err := svc.CreateComplaint(ctx,
			"test-agent",
			"test-session",
			fmt.Sprintf("test task %d", i),
			"test context",
			"missing info",
			"confused by",
			"future wishes",
			domain.SeverityLow,
			"test-project")
		if err != nil {
			t.Fatalf("Failed to create complaint %d: %v", i, err)
		}
	}

	// Test listing complaints
	complaints, err := svc.ListComplaints(ctx, 10, 0)
	if err != nil {
		t.Errorf("ListComplaints error = %v, want nil", err)
		return
	}

	if len(complaints) != 5 {
		t.Errorf("ListComplaints returned %d complaints, want 5", len(complaints))
	}

	// Verify all complaints are returned as pointers
	for i, complaint := range complaints {
		if complaint == nil {
			t.Errorf("Complaint %d should not be nil", i)
		}
		
		if !strings.Contains(complaint.TaskDescription, "test task") {
			t.Errorf("Complaint %d should contain 'test task', got: %s", i, complaint.TaskDescription)
		}
	}
}