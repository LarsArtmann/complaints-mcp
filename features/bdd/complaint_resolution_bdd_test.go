package bdd_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/cmd/server"
)

var _ = Describe("Complaint Resolution BDD Tests", func() {
	var (
		serverCmd *server.ServerCommand
		ctx    context.Context
		repo   *service.MockRepository
		svc    *service.ComplaintService
	)

	BeforeEach(func() {
		ctx = context.Background()
		repo = &service.MockRepository{}
		svc = service.NewComplaintService(repo, nil)
		serverCmd = &server.ServerCommand{}
	})

	Context("Resolve existing complaints", func() {
		It("should successfully resolve an unresolved complaint", func(ctx SpecContext) {
			// Create an unresolved complaint first
			complaintID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			// Store it
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "AI Assistant",
				SessionName:     "test-session",
				TaskDescription: "Test task for resolution",
				ContextInfo:     "Need to fix authentication flow",
				MissingInfo:     "Unclear error messages",
				ConfusedBy:      "Response format inconsistency",
				FutureWishes:    "Better error handling",
				Severity:        "medium",
				ProjectName:     "resolution-test",
				Timestamp:       ctx,
				Resolved:        false,
			}

			err := repo.Store(ctx, complaint)
			Expect(err).NotTo(HaveOccurred())

			// Now resolve it
			err = svc.ResolveComplaint(ctx, complaintID)
			Expect(err).NotTo(HaveOccurred())

			// Verify it's resolved
			resolved, err := svc.GetComplaint(ctx, complaintID)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved).NotTo(BeNil())
			Expect(resolved.Resolved).To(BeTrue())
		})

		It("should fail to resolve a non-existent complaint", func(ctx SpecContext) {
			nonExistentID := domain.ComplaintID{Value: "non-existent"}

			err := svc.ResolveComplaint(ctx, nonExistentID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should handle concurrent resolution requests", func(ctx SpecContext) {
			// Create multiple complaints
			complaints := make([]*domain.Complaint, 3)
			for i := 0; i < 3; i++ {
				id, _ := domain.NewComplaintID()
				complaints[i] = &domain.Complaint{
					ID:              id,
					AgentName:       "AI Assistant",
					TaskDescription: "Concurrent test task",
					Severity:        "low",
					Resolved:        false,
				}
				repo.Store(ctx, complaints[i])
			}

			// Resolve them concurrently
			done := make(chan bool, 3)
			for i, complaint := range complaints {
				go func(idx int) {
					err := svc.ResolveComplaint(ctx, complaint.ID)
					Expect(err).NotTo(HaveOccurred())
					done[idx] <- true
				}(i)
			}

			// Wait for all to complete
			for i := 0; i < 3; i++ {
				Eventually(func() bool {
					return <-done[i]
				}).WithContext(ctx).Should(BeTrue())
			}
		})

		It("should handle resolution conflicts gracefully", func(ctx SpecContext) {
			complaintID, _ := domain.NewComplaintID()
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "AI Assistant",
				TaskDescription: "Conflict test",
				Severity:        "low",
				Resolved:        false,
			}

			err := repo.Store(ctx, complaint)
			Expect(err).NotTo(HaveOccurred())

			// First resolution succeeds
			err = svc.ResolveComplaint(ctx, complaintID)
			Expect(err).NotTo(HaveOccurred())

			// Second resolution should also succeed (not conflict)
			err = svc.ResolveComplaint(ctx, complaintID)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Error handling for resolution", func() {
		It("should handle repository errors gracefully", func(ctx SpecContext) {
			complaintID, _ := domain.NewComplaintID()
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "AI Assistant",
				TaskDescription: "Error handling test",
				Severity:        "low",
				Resolved:        false,
			}

			// Mock repository error
			repo.SetError(true)

			err := repo.Store(ctx, complaint)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("repository error"))

			// Resolution should also fail gracefully
			err = svc.ResolveComplaint(ctx, complaintID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("repository error"))
		})

		It("should handle service errors gracefully", func(ctx SpecContext) {
			complaintID, _ := domain.NewComplaintID()
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "AI Assistant",
				TaskDescription: "Service error test",
				Severity:        "low",
				Resolved:        false,
			}

			err := repo.Store(ctx, complaint)
			Expect(err).NotTo(HaveOccurred())

			// Mock service error
			svc.SetError(true)

			err = svc.ResolveComplaint(ctx, complaintID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("service error"))
		})
	})

	Context("Performance tests", func() {
		It("should handle rapid resolution requests", func(ctx SpecContext) {
			complaints := make([]*domain.Complaint, 50)
			ids := make([]domain.ComplaintID, 50)

			// Create 50 complaints
			for i := 0; i < 50; i++ {
				id, _ := domain.NewComplaintID()
				ids[i] = id
				complaints[i] = &domain.Complaint{
					ID:              id,
					AgentName:       "AI Assistant",
					TaskDescription: fmt.Sprintf("Performance test complaint %d", i),
					Severity:        "low",
					Resolved:        false,
				}
				repo.Store(ctx, complaints[i])
			}

			// Measure resolution time
			start := time.Now()
			for i, id := range ids {
				err := svc.ResolveComplaint(ctx, id)
				Expect(err).NotTo(HaveOccurred())
			}
			duration := time.Since(start)

			// Should complete within reasonable time
			Expect(duration.Milliseconds()).To(BeNumerically("<", 1000)) // 1 second for 50 resolutions
		})
	})
})

// ServiceMockRepository extends mockRepository with error simulation capabilities
type ServiceMockRepository struct {
	*mockRepository
	shouldError bool
	errorMsg   string
}

func (s *ServiceMockRepository) SetError(shouldError bool, errorMsg string) {
	s.shouldError = shouldError
	s.errorMsg = errorMsg
}

func (s *ServiceMockRepository) Store(ctx context.Context, complaint *domain.Complaint) error {
	if s.shouldError {
		return fmt.Errorf("mock repository error: %s", s.errorMsg)
	}
	return s.mockRepository.Store(ctx, complaint)
}

func (s *ServiceMockRepository) MarkResolved(ctx context.Context, id domain.ComplaintID) error {
	if s.shouldError {
		return fmt.Errorf("mock repository error during resolve: %s", s.errorMsg)
	}
	return s.mockRepository.MarkResolved(ctx, id)
}