package service

import (
	"context"
	"testing"

	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/larsartmann/complaints-mcp/internal/domain"
)

// mockRepository is a simple in-memory repository for testing
type mockRepository struct {
	complaints []*domain.Complaint
}

func (m *mockRepository) Store(ctx context.Context, complaint *domain.Complaint) error {
	m.complaints = append(m.complaints, complaint)
	return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	for _, c := range m.complaints {
		if c.ID.Value == id.Value {
			return c, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) FindByProject(ctx context.Context, projectName string, limit, offset int) ([]*domain.Complaint, error) {
	var result []*domain.Complaint
	for _, c := range m.complaints {
		if c.ProjectName == projectName {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockRepository) FindUnresolved(ctx context.Context, limit, offset int) ([]*domain.Complaint, error) {
	var result []*domain.Complaint
	for _, c := range m.complaints {
		if !c.Resolved {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockRepository) MarkResolved(ctx context.Context, id domain.ComplaintID) error {
	for _, c := range m.complaints {
		if c.ID.Value == id.Value {
			c.Resolve()
			break
		}
	}
	return nil
}

func TestNewComplaintService(t *testing.T) {
	mockRepo := &mockRepository{}
	cfg := &config.Config{
		Complaints: struct {
			StorageLocation config.StorageLocation `mapstructure:"storage_location" validate:"required,oneof=local global both" json:"storage_location"`
			RetentionDays   int                    `mapstructure:"retention_days" validate:"min=1,max=365" json:"retention_days"`
			ProjectName     string                 `mapstructure:"project_name" json:"project_name"`
			AutoResolve     *bool                  `mapstructure:"auto_resolve" json:"auto_resolve"`
			MaxFileSize     int64                  `mapstructure:"max_file_size" validate:"min=1024,max=1048576" json:"max_file_size"`
		}{
			ProjectName: "test-project",
		},
	}

	service := NewComplaintService(mockRepo, cfg)

	if service == nil {
		t.Error("NewComplaintService() returned nil")
	}

	if service.repo != mockRepo {
		t.Error("NewComplaintService() did not set repository correctly")
	}

	if service.config != cfg {
		t.Error("NewComplaintService() did not set config correctly")
	}
}

func TestCreateComplaintRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateComplaintRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateComplaintRequest{
				AgentName:       "test-agent",
				TaskDescription: "test task",
				Severity:        "high",
			},
			wantErr: false,
		},
		{
			name: "missing agent name",
			req: CreateComplaintRequest{
				TaskDescription: "test task",
				Severity:        "high",
			},
			wantErr: true,
		},
		{
			name: "missing task description",
			req: CreateComplaintRequest{
				AgentName: "test-agent",
				Severity:  "high",
			},
			wantErr: true,
		},
		{
			name: "missing severity",
			req: CreateComplaintRequest{
				AgentName:       "test-agent",
				TaskDescription: "test task",
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			req: CreateComplaintRequest{
				AgentName:       "test-agent",
				TaskDescription: "test task",
				Severity:        "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateComplaintRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComplaintService_CreateComplaint(t *testing.T) {
	mockRepo := &mockRepository{}
	cfg := &config.Config{
		Complaints: struct {
			StorageLocation config.StorageLocation `mapstructure:"storage_location" validate:"required,oneof=local global both" json:"storage_location"`
			RetentionDays   int                    `mapstructure:"retention_days" validate:"min=1,max=365" json:"retention_days"`
			ProjectName     string                 `mapstructure:"project_name" json:"project_name"`
			AutoResolve     *bool                  `mapstructure:"auto_resolve" json:"auto_resolve"`
			MaxFileSize     int64                  `mapstructure:"max_file_size" validate:"min=1024,max=1048576" json:"max_file_size"`
		}{
			ProjectName: "test-project",
		},
	}

	service := NewComplaintService(mockRepo, cfg)

	req := &CreateComplaintRequest{
		AgentName:       "test-agent",
		TaskDescription: "test task",
		Severity:        "high",
		ContextInfo:     "some context",
		ProjectName:     "custom-project",
	}

	complaint, err := service.CreateComplaint(context.Background(), req)

	if err != nil {
		t.Errorf("ComplaintService.CreateComplaint() error = %v", err)
		return
	}

	if complaint == nil {
		t.Error("ComplaintService.CreateComplaint() returned nil")
		return
	}

	if complaint.AgentName != req.AgentName {
		t.Errorf("ComplaintService.CreateComplaint().AgentName = %v, want %v", complaint.AgentName, req.AgentName)
	}

	if complaint.ProjectName != req.ProjectName {
		t.Errorf("ComplaintService.CreateComplaint().ProjectName = %v, want %v", complaint.ProjectName, req.ProjectName)
	}

	if len(mockRepo.complaints) != 1 {
		t.Errorf("Repository should have 1 complaint, got %d", len(mockRepo.complaints))
	}
}

func TestComplaintService_CreateComplaint_UsesDefaultProjectName(t *testing.T) {
	mockRepo := &mockRepository{}
	cfg := &config.Config{
		Complaints: struct {
			StorageLocation config.StorageLocation `mapstructure:"storage_location" validate:"required,oneof=local global both" json:"storage_location"`
			RetentionDays   int                    `mapstructure:"retention_days" validate:"min=1,max=365" json:"retention_days"`
			ProjectName     string                 `mapstructure:"project_name" json:"project_name"`
			AutoResolve     *bool                  `mapstructure:"auto_resolve" json:"auto_resolve"`
			MaxFileSize     int64                  `mapstructure:"max_file_size" validate:"min=1024,max=1048576" json:"max_file_size"`
		}{
			ProjectName: "default-project",
		},
	}

	service := NewComplaintService(mockRepo, cfg)

	req := &CreateComplaintRequest{
		AgentName:       "test-agent",
		TaskDescription: "test task",
		Severity:        "high",
		// No project name provided - should use default
	}

	complaint, err := service.CreateComplaint(context.Background(), req)

	if err != nil {
		t.Errorf("ComplaintService.CreateComplaint() error = %v", err)
		return
	}

	if complaint.ProjectName != cfg.Complaints.ProjectName {
		t.Errorf("ComplaintService.CreateComplaint().ProjectName = %v, want %v", complaint.ProjectName, cfg.Complaints.ProjectName)
	}
}

func TestComplaintService_GetComplaint(t *testing.T) {
	mockRepo := &mockRepository{}
	cfg := &config.Config{
		Complaints: struct {
			StorageLocation config.StorageLocation `mapstructure:"storage_location" validate:"required,oneof=local global both" json:"storage_location"`
			RetentionDays   int                    `mapstructure:"retention_days" validate:"min=1,max=365" json:"retention_days"`
			ProjectName     string                 `mapstructure:"project_name" json:"project_name"`
			AutoResolve     *bool                  `mapstructure:"auto_resolve" json:"auto_resolve"`
			MaxFileSize     int64                  `mapstructure:"max_file_size" validate:"min=1024,max=1048576" json:"max_file_size"`
		}{},
	}

	service := NewComplaintService(mockRepo, cfg)

	// Create a test complaint
	id, _ := domain.NewComplaintID()
	testComplaint := &domain.Complaint{
		ID:              id,
		AgentName:       "test-agent",
		TaskDescription: "test task",
		Severity:        "high",
		ProjectName:     "test-project",
	}

	mockRepo.complaints = append(mockRepo.complaints, testComplaint)

	// Test getting the complaint
	found, err := service.GetComplaint(context.Background(), id)

	if err != nil {
		t.Errorf("ComplaintService.GetComplaint() error = %v", err)
		return
	}

	if found == nil {
		t.Error("ComplaintService.GetComplaint() returned nil")
		return
	}

	if found.ID.Value != id.Value {
		t.Errorf("ComplaintService.GetComplaint().ID = %v, want %v", found.ID.Value, id.Value)
	}

	// Test getting non-existent complaint
	nonExistentID, _ := domain.NewComplaintID()
	notFound, err := service.GetComplaint(context.Background(), nonExistentID)

	if err == nil {
		t.Error("ComplaintService.GetComplaint() should return error for non-existent complaint")
	}

	if notFound != nil {
		t.Error("ComplaintService.GetComplaint() should return nil for non-existent complaint")
	}
}

func TestComplaintService_ResolveComplaint(t *testing.T) {
	mockRepo := &mockRepository{}
	cfg := &config.Config{
		Complaints: struct {
			StorageLocation config.StorageLocation `mapstructure:"storage_location" validate:"required,oneof=local global both" json:"storage_location"`
			RetentionDays   int                    `mapstructure:"retention_days" validate:"min=1,max=365" json:"retention_days"`
			ProjectName     string                 `mapstructure:"project_name" json:"project_name"`
			AutoResolve     *bool                  `mapstructure:"auto_resolve" json:"auto_resolve"`
			MaxFileSize     int64                  `mapstructure:"max_file_size" validate:"min=1024,max=1048576" json:"max_file_size"`
		}{},
	}

	service := NewComplaintService(mockRepo, cfg)

	// Create a test complaint
	id, _ := domain.NewComplaintID()
	testComplaint := &domain.Complaint{
		ID:              id,
		AgentName:       "test-agent",
		TaskDescription: "test task",
		Severity:        "high",
		ProjectName:     "test-project",
		Resolved:        false,
	}

	mockRepo.complaints = append(mockRepo.complaints, testComplaint)

	// Resolve the complaint
	err := service.ResolveComplaint(context.Background(), id)

	if err != nil {
		t.Errorf("ComplaintService.ResolveComplaint() error = %v", err)
	}

	if !testComplaint.Resolved {
		t.Error("ComplaintService.ResolveComplaint() did not resolve the complaint")
	}
}
