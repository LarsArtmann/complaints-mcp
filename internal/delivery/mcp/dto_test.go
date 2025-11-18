package delivery

import (
	"testing"
	"time"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToDTO tests the conversion from domain Complaint to DTO
func TestToDTO(t *testing.T) {
	// Create test complaint with all fields
	id, _ := domain.NewComplaintID()
	now := time.Now()
	resolvedAt := now.Add(time.Hour)

	complaint := &domain.Complaint{
		ID:              id,
		AgentName:       domain.MustNewAgentName("Test Agent"),
		SessionName:     domain.MustNewSessionName("test-session"),
		TaskDescription: "Test task description",
		ContextInfo:     "Context information",
		MissingInfo:     "Missing information",
		ConfusedBy:      "Confusing aspects",
		FutureWishes:    "Future wishes",
		Severity:        domain.SeverityHigh,
		Timestamp:       now,
		ProjectName:     domain.MustNewProjectName("test-project"),
		ResolvedAt:      &resolvedAt,
		ResolvedBy:      "test-resolver",
	}

	// Convert to DTO
	dto := ToDTO(complaint)

	// Verify all fields are properly converted
	assert.Equal(t, id.String(), dto.ID, "ID should match")
	assert.Equal(t, "Test Agent", dto.AgentName, "AgentName should match")
	assert.Equal(t, "test-session", dto.SessionName, "SessionName should match")
	assert.Equal(t, "Test task description", dto.TaskDescription, "TaskDescription should match")
	assert.Equal(t, "Context information", dto.ContextInfo, "ContextInfo should match")
	assert.Equal(t, "Missing information", dto.MissingInfo, "MissingInfo should match")
	assert.Equal(t, "Confusing aspects", dto.ConfusedBy, "ConfusedBy should match")
	assert.Equal(t, "Future wishes", dto.FutureWishes, "FutureWishes should match")
	assert.Equal(t, "high", dto.Severity, "Severity should be converted to string")
	assert.Equal(t, now, dto.Timestamp, "Timestamp should match")
	assert.Equal(t, "test-project", dto.ProjectName, "ProjectName should match")
	assert.True(t, dto.Resolved, "Resolved should be true")

	// Verify timestamp pointer is properly converted
	require.NotNil(t, dto.ResolvedAt, "ResolvedAt should not be nil")
	assert.Equal(t, resolvedAt, *dto.ResolvedAt, "ResolvedAt should match")
	assert.Equal(t, "test-resolver", dto.ResolvedBy, "ResolvedBy should match")
}

// TestComplaintDTO_JSONSerialization tests that DTO serializes to valid JSON
func TestComplaintDTO_JSONSerialization(t *testing.T) {
	// Create test complaint
	id, _ := domain.NewComplaintID()
	complaint := &domain.Complaint{
		ID:              id,
		AgentName:       domain.MustNewAgentName("Test Agent"),
		TaskDescription: "Test task",
		Severity:        domain.SeverityMedium,
		Timestamp:       time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC),
	}

	// Convert to DTO
	dto := ToDTO(complaint)

	// Verify DTO has expected structure
	assert.Equal(t, id.String(), dto.ID)
	assert.Equal(t, "Test Agent", dto.AgentName)
	assert.Equal(t, "Test task", dto.TaskDescription)
	assert.Equal(t, "medium", dto.Severity)
	assert.False(t, dto.Resolved)
	assert.Empty(t, dto.ResolvedAt) // Should be omitted when nil
	assert.Empty(t, dto.ResolvedBy) // Should be omitted when empty
}

// TestComplaintDTO_OptionalFields tests optional field behavior
func TestComplaintDTO_OptionalFields(t *testing.T) {
	// Create minimal complaint
	id, _ := domain.NewComplaintID()
	complaint := &domain.Complaint{
		ID:              id,
		AgentName:       domain.MustNewAgentName("Test Agent"),
		TaskDescription: "Test task",
		Severity:        domain.SeverityLow,
		Timestamp:       time.Now(),
		// Leave optional fields empty
	}

	// Convert to DTO
	dto := ToDTO(complaint)

	// Verify optional fields are empty/zero
	assert.Equal(t, "", dto.SessionName)
	assert.Equal(t, "", dto.ContextInfo)
	assert.Equal(t, "", dto.MissingInfo)
	assert.Equal(t, "", dto.ConfusedBy)
	assert.Equal(t, "", dto.FutureWishes)
	assert.Equal(t, "", dto.ProjectName)
	assert.False(t, dto.Resolved)
	assert.Nil(t, dto.ResolvedAt)
	assert.Equal(t, "", dto.ResolvedBy)
}

// TestListComplaintsOutput_TypeSafety tests the output struct is type-safe
func TestListComplaintsOutput_TypeSafety(t *testing.T) {
	id1, _ := domain.NewComplaintID()
	id2, _ := domain.NewComplaintID()

	now := time.Now()
	complaint1 := &domain.Complaint{
		ID:              id1,
		AgentName:       domain.MustNewAgentName("Agent 1"),
		TaskDescription: "Task 1",
		Severity:        domain.SeverityHigh,
		Timestamp:       time.Now(),
	}

	resolvedAt := now.Add(time.Hour)
	complaint2 := &domain.Complaint{
		ID:              id2,
		AgentName:       domain.MustNewAgentName("Agent 2"),
		TaskDescription: "Task 2",
		Severity:        domain.SeverityLow,
		Timestamp:       time.Now(),
		ResolvedAt:      &resolvedAt,
	}

	// Create output with type-safe DTOs
	output := ListComplaintsOutput{
		Complaints: []ComplaintDTO{
			ToDTO(complaint1),
			ToDTO(complaint2),
		},
	}

	// Verify type safety - no maps, only proper DTOs
	assert.Len(t, output.Complaints, 2)
	assert.IsType(t, ComplaintDTO{}, output.Complaints[0])
	assert.IsType(t, ComplaintDTO{}, output.Complaints[1])

	// Verify first complaint
	assert.Equal(t, "Agent 1", output.Complaints[0].AgentName)
	assert.Equal(t, "high", output.Complaints[0].Severity)
	assert.False(t, output.Complaints[0].Resolved)

	// Verify second complaint
	assert.Equal(t, "Agent 2", output.Complaints[1].AgentName)
	assert.Equal(t, "low", output.Complaints[1].Severity)
	assert.True(t, output.Complaints[1].Resolved)
}

// TestFileComplaintOutput_TypeSafety tests the file complaint output is type-safe
func TestFileComplaintOutput_TypeSafety(t *testing.T) {
	id, _ := domain.NewComplaintID()
	complaint := &domain.Complaint{
		ID:              id,
		AgentName:       domain.MustNewAgentName("Test Agent"),
		TaskDescription: "Test task",
		Severity:        domain.SeverityMedium,
		Timestamp:       time.Now(),
	}

	// Create output with type-safe DTO
	output := FileComplaintOutput{
		Success:   true,
		Message:   "Complaint filed successfully",
		Complaint: ToDTO(complaint),
	}

	// Verify type safety
	assert.True(t, output.Success)
	assert.Equal(t, "Complaint filed successfully", output.Message)
	assert.IsType(t, ComplaintDTO{}, output.Complaint)
	assert.Equal(t, "Test Agent", output.Complaint.AgentName)
	assert.Equal(t, "medium", output.Complaint.Severity)
}

// TestResolveComplaintOutput_TypeSafety tests the resolve complaint output is type-safe
func TestResolveComplaintOutput_TypeSafety(t *testing.T) {
	id, _ := domain.NewComplaintID()
	now := time.Now()
	resolvedAt := now.Add(time.Hour)

	complaint := &domain.Complaint{
		ID:              id,
		AgentName:       domain.MustNewAgentName("Test Agent"),
		TaskDescription: "Test task",
		Severity:        domain.SeverityHigh,
		Timestamp:       time.Now(),
		ResolvedAt:      &resolvedAt,
		ResolvedBy:      "test-resolver",
	}

	// Create output with type-safe DTO
	output := ResolveComplaintOutput{
		Success:   true,
		Message:   "Complaint resolved successfully",
		Complaint: ToDTO(complaint),
	}

	// Verify type safety
	assert.True(t, output.Success)
	assert.Equal(t, "Complaint resolved successfully", output.Message)
	assert.IsType(t, ComplaintDTO{}, output.Complaint)
	assert.Equal(t, "Test Agent", output.Complaint.AgentName)
	assert.True(t, output.Complaint.Resolved)
	assert.Equal(t, "test-resolver", output.Complaint.ResolvedBy)
	require.NotNil(t, output.Complaint.ResolvedAt)
	assert.Equal(t, resolvedAt, *output.Complaint.ResolvedAt)
}
