package bdd_test

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/cmd/server"
)

var _ = Describe("Complaint Listing BDD Tests", func() {
	var (
		serverCmd *server.ServerCommand
		ctx    context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		serverCmd = &server.ServerCommand{}
	})

	Context("List all complaints successfully", func() {
		It("should return list of all complaints", func(ctx SpecContext) {
			// Setup mock repository with some test complaints
			repo := &mockRepository{
				complaints: []*domain.Complaint{
					{
						ID:              domain.ComplaintID{Value: "1"},
						AgentName:       "AI Assistant",
						TaskDescription: "Task 1",
						ContextInfo:     "Context 1",
						MissingInfo:     "Missing 1",
						ConfusedBy:      "Confused 1",
						FutureWishes:    "Wishes 1",
						Severity:        "low",
						ProjectName:     "project-a",
						Timestamp:       time.Now().Add(-1 * 24 * time.Hour),
						Resolved:        false,
					},
					{
						ID:              domain.ComplaintID{Value: "2"},
						AgentName:       "AI Assistant",
						TaskDescription: "Task 2",
						ContextInfo:     "Context 2",
						MissingInfo:     "Missing 2",
						ConfusedBy:      "Confused 2",
						FutureWishes:    "Wishes 2",
						Severity:        "medium",
						ProjectName:     "project-b",
						Timestamp:       time.Now().Add(-2 * 24 * time.Hour),
						Resolved:        false,
					},
					{
						ID:              domain.ComplaintID{Value: "3"},
						AgentName:       "AI Assistant",
						TaskDescription: "Task 3",
						ContextInfo:     "Context 3",
						MissingInfo:     "Missing 3",
						ConfusedBy:      "Confused 3",
						FutureWishes:    "Wishes 3",
						Severity:        "high",
						ProjectName:     "project-a",
						Timestamp:       time.Now().Add(-3 * 24 * time.Hour),
						Resolved:        true, // This one is resolved
					},
				},
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(repo, config)

			// Test listing without filters
			complaints, err := svc.ListComplaintsByProject(ctx, "", 50, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints)).To(Equal(3))
			Expect(complaints[0].TaskDescription).To(Equal("Task 1"))
			Expect(complaints[1].TaskDescription).To(Equal("Task 2"))
			Expect(complaints[2].TaskDescription).To(Equal("Task 3"))

			// Verify resolved status
			Expect(complaints[0].Resolved).To(BeFalse())
			Expect(complaints[1].Resolved).To(BeFalse())
			Expect(complaints[2].Resolved).To(BeTrue())
		})

		It("should format response properly as JSON", func(ctx SpecContext) {
			repo := &mockRepository{}
			svc := service.NewComplaintService(repo, nil)

			// Mock the list response
			serverCmd.SetMockListResponse(`[{
				"id": "1",
				"agent_name": "AI Assistant",
				"task_description": "Task 1",
				"severity": "low",
				"project_name": "project-a",
				"resolved": false,
				"timestamp": "2025-01-01T00:00:00Z",
				"context_info": "Context 1",
				"missing_info": "Missing 1",
				"confused_by": "Confused 1",
				"future_wishes": "Wishes 1"
			}, {
				"id": "2",
				"agent_name": "AI Assistant",
				"task_description": "Task 2",
				"severity": "medium",
				"project_name": "project-b",
				"resolved": false,
				"timestamp": "2025-01-02T00:00:00Z",
				"context_info": "Context 2",
				"missing_info": "Missing 2",
				"confused_by": "Confused 2",
				"future_wishes": "Wishes 2"
			}, {
				"id": "3",
				"agent_name": "AI Assistant",
				"task_description": "Task 3",
				"severity": "high",
				"project_name": "project-a",
				"resolved": true,
				"timestamp": "2025-01-03T00:00:00Z",
				"context_info": "Context 3",
				"missing_info": "Missing 3",
				"confused_by": "Confused 3",
				"future_wishes": "Wishes 3"
			}]`)

			response, err := serverCmd.ListComplaints(ctx, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())

			// Parse and verify JSON response
			var complaints []map[string]interface{}
			err = json.Unmarshal([]byte(response.Content), &complaints)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints)).To(Equal(3))
		})
	})

	Context("List complaints filtered by project", func() {
		It("should return only complaints from specified project", func(ctx SpecContext) {
			repo := &mockRepository{
				complaints: []*domain.Complaint{
					{
						ID:              domain.ComplaintID{Value: "1"},
						AgentName:       "AI Assistant",
						TaskDescription: "Task 1",
						ProjectName:     "project-a",
						Severity:        "low",
						Resolved:        false,
					},
					{
						ID:              domain.ComplaintID{Value: "2"},
						AgentName:       "AI Assistant",
						TaskDescription: "Task 2",
						ProjectName:     "project-b", // Different project
						Severity:        "medium",
						Resolved:        false,
					},
					{
						ID:              domain.ComplaintID{Value: "3"},
						AgentName:       "AI Assistant",
						TaskDescription: "Task 3",
						ProjectName:     "project-a", // Same as first
						Severity:        "high",
						Resolved:        true,
					},
				},
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(repo, config)

			// Test filtering by project-a
			complaints, err := svc.ListComplaintsByProject(ctx, "project-a", 50, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints)).To(Equal(2)) // Only 2 from project-a
			Expect(complaints[0].ProjectName).To(Equal("project-a"))
			Expect(complaints[1].ProjectName).To(Equal("project-a"))

			// Test filtering by project-b
			complaintsB, err := svc.ListComplaintsByProject(ctx, "project-b", 50, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaintsB)).To(Equal(1)) // Only 1 from project-b
			Expect(complaintsB[0].ProjectName).To(Equal("project-b"))
		})
	})

	Context("List complaints filtered by unresolved status", func() {
		It("should return only unresolved complaints", func(ctx SpecContext) {
			repo := &mockRepository{
				complaints: []*domain.Complaint{
					{
						ID:              domain.ComplaintID{Value: "1"},
						AgentName:       "AI Assistant",
						TaskDescription: "Task 1",
						ProjectName:     "project-a",
						Severity:        "low",
						Resolved:        false,
					},
					{
						ID:              domain.ComplaintID{Value: "2"},
						AgentName:       "AI Assistant",
						TaskDescription: "Task 2",
						ProjectName:     "project-b",
						Severity:        "medium",
						Resolved:        false,
					},
					{
						ID:              domain.ComplaintID{Value: "3"},
						AgentName:       "AI Assistant",
						TaskDescription: "Task 3",
						ProjectName:     "project-a",
						Severity:        "high",
						Resolved:        true,
					},
				},
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(repo, config)

			// Test filtering for unresolved
			complaints, err := svc.ListUnresolvedComplaints(ctx, 50, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints)).To(Equal(2)) // Only 2 unresolved
			Expect(complaints[0].Resolved).To(BeFalse())
			Expect(complaints[1].Resolved).To(BeFalse())

			// Verify resolved complaints are excluded
			for _, complaint := range complaints {
				Expect(complaint.Resolved).To(BeFalse())
			}
		})
	})

	Context("List complaints with pagination", func() {
		It("should respect limit and offset parameters", func(ctx SpecContext) {
			repo := &mockRepository{
				complaints: make([]*domain.Complaint, 100),
			}

			// Add 100 complaints to repository
			for i := 0; i < 100; i++ {
				complaints := append(repo.complaints, &domain.Complaint{
					ID:              domain.ComplaintID{Value: fmt.Sprintf("%d", i)},
					AgentName:       "AI Assistant",
					TaskDescription: fmt.Sprintf("Task %d", i),
					ProjectName:     "pagination-test",
					Severity:        "low",
					Resolved:        false,
				})
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(repo, config)

			// Test with limit 10, offset 0
			complaints1, err := svc.ListComplaintsByProject(ctx, "pagination-test", 10, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints1)).To(Equal(10))
			Expect(complaints1[0].TaskDescription).To(Equal("Task 0"))

			// Test with limit 10, offset 10 (should get complaints 10-19)
			complaints2, err := svc.ListComplaintsByProject(ctx, "pagination-test", 10, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints2)).To(Equal(10))
			Expect(complaints2[0].TaskDescription).To(Equal("Task 10"))

			// Test with limit 10, offset 20 (should get complaints 20-29, but only have up to 80)
			complaints3, err := svc.ListComplaintsByProject(ctx, "pagination-test", 10, 20)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints3)).To(Equal(10)) // Even though we only have 80 total, we request 10 and get 10
			Expect(complaints3[0].TaskDescription).To(Equal("Task 20"))
		})
	})
})

// mockRepository implements a simple in-memory repository for testing
type mockRepository struct {
	complaints []*domain.Complaint
}

func (m *mockRepository) Store(ctx context.Context, complaint *domain.Complaint) error {
	m.complaints = append(m.complaints, complaint)
	return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id domain.ComplaintID) (*domain.Complaint, error) {
	for _, complaint := range m.complaints {
		if complaint.ID.Value == id.Value {
			return complaint, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) FindByProject(ctx context.Context, projectName string, limit int, offset int) ([]*domain.Complaint, error) {
	var result []*domain.Complaint
	count := 0
	started := false

	for _, complaint := range m.complaints {
		if complaint.ProjectName == projectName {
			if count >= offset && !started {
				started = true
			}
			if started && count < offset+limit {
				result = append(result, complaint)
				count++
			}
		}
	}
	return result, nil
}

func (m *mockRepository) FindUnresolved(ctx context.Context, limit int, offset int) ([]*domain.Complaint, error) {
	var result []*domain.Complaint
	for _, complaint := range m.complaints {
		if !complaint.Resolved {
			result = append(result, complaint)
		}
	}
	return result, nil
}

func (m *mockRepository) MarkResolved(ctx context.Context, id domain.ComplaintID) error {
	for i, complaint := range m.complaints {
		if complaint.ID.Value == id.Value {
			complaint.Resolved = true
			m.complaints[i] = complaint
			break
		}
	}
	return nil
}