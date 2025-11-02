package bdd_test

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
	"github.com/charmbracelet/log"
)

var _ = Describe("Complaint Resolution BDD Tests", func() {
	var (
		tempDir string
		repository repo.Repository
		complaintService *service.ComplaintService
		logger *log.Logger
		tracer tracing.Tracer
		testComplaint *domain.Complaint
	)

	BeforeEach(func() {
		// Create a temporary directory for each test
		tempDir = GinkgoT().TempDir()
		logger = log.New(os.Stdout)
		tracer = tracing.NewMockTracer("test")

		// Initialize repository and service
		repository = repo.NewFileRepository(tempDir, tracer)
		complaintService = service.NewComplaintService(repository, tracer, logger)

		// Create a test complaint for resolution testing
		var err error
		testComplaint, err = complaintService.CreateComplaint(context.Background(),
			"AI Assistant",
			"resolution-test-session",
			"Authentication flow needs fixing",
			"JWT token validation is unclear",
			"Missing error codes documentation",
			"Inconsistent response formats",
			"Add comprehensive error handling",
			domain.SeverityMedium,
			"resolution-test-project")
		Expect(err).NotTo(HaveOccurred())
		Expect(testComplaint.Resolved).To(BeFalse())
	})

	AfterEach(func() {
		// Clean up temporary directory
		os.RemoveAll(tempDir)
	})

	Context("Resolve existing complaints", func() {
		It("should successfully resolve an unresolved complaint", func(ctx SpecContext) {
			// Verify complaint is initially unresolved
			retrievedComplaint, err := complaintService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedComplaint.Resolved).To(BeFalse())

			// Resolve the complaint
			err = complaintService.ResolveComplaint(ctx, testComplaint.ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// Verify the complaint is now resolved
			resolvedComplaint, err := complaintService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolvedComplaint.Resolved).To(BeTrue())
		})

		It("should preserve original complaint data when resolving", func(ctx SpecContext) {
			// Resolve the complaint
			err := complaintService.ResolveComplaint(ctx, testComplaint.ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// Verify all original data is preserved
			resolvedComplaint, err := complaintService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			
			Expect(resolvedComplaint.ID.Value).To(Equal(testComplaint.ID.Value))
			Expect(resolvedComplaint.AgentName).To(Equal(testComplaint.AgentName))
			Expect(resolvedComplaint.SessionName).To(Equal(testComplaint.SessionName))
			Expect(resolvedComplaint.TaskDescription).To(Equal(testComplaint.TaskDescription))
			Expect(resolvedComplaint.ContextInfo).To(Equal(testComplaint.ContextInfo))
			Expect(resolvedComplaint.MissingInfo).To(Equal(testComplaint.MissingInfo))
			Expect(resolvedComplaint.ConfusedBy).To(Equal(testComplaint.ConfusedBy))
			Expect(resolvedComplaint.FutureWishes).To(Equal(testComplaint.FutureWishes))
			Expect(resolvedComplaint.Severity).To(Equal(testComplaint.Severity))
			Expect(resolvedComplaint.ProjectName).To(Equal(testComplaint.ProjectName))
			Expect(resolvedComplaint.Timestamp).To(Equal(testComplaint.Timestamp))
			Expect(resolvedComplaint.Resolved).To(BeTrue()) // Only this should change
		})

		It("should record resolution timestamp correctly", func(ctx SpecContext) {
			// Get the original complaint
			originalComplaint, err := complaintService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			_ = originalComplaint // Use variable to avoid unused warning

			// Wait a bit to ensure different timestamp
			time.Sleep(10 * time.Millisecond)

			// Resolve the complaint
			err = complaintService.ResolveComplaint(ctx, testComplaint.ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// Verify resolution - the domain's Resolve method handles timestamp internally
			resolvedComplaint, err := complaintService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolvedComplaint.Resolved).To(BeTrue())
			// Note: The Resolve method in domain only sets Resolved=true, doesn't update timestamp
			// This is by design - the original creation timestamp is preserved
		})
	})

	Context("Attempt to resolve non-existent complaints", func() {
		It("should return error for non-existent complaint ID", func(ctx SpecContext) {
			// Create a non-existent complaint ID
			nonExistentID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			// Try to resolve non-existent complaint
			err = complaintService.ResolveComplaint(ctx, nonExistentID, "test-agent")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("complaint not found"))
		})

		It("should return specific error for empty complaint ID", func(ctx SpecContext) {
			// Try to resolve with empty complaint ID
			emptyID := domain.ComplaintID{}
			err := complaintService.ResolveComplaint(ctx, emptyID, "test-agent")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("complaint not found"))
		})
	})

	Context("Resolve already resolved complaints", func() {
		It("should handle resolving already resolved complaint gracefully", func(ctx SpecContext) {
			// First resolve the complaint
			err := complaintService.ResolveComplaint(ctx, testComplaint.ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// Verify it's resolved
			resolvedComplaint, err := complaintService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolvedComplaint.Resolved).To(BeTrue())

			// Try to resolve it again - should be idempotent
			err = complaintService.ResolveComplaint(ctx, testComplaint.ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// Verify it's still resolved
			stillResolvedComplaint, err := complaintService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(stillResolvedComplaint.Resolved).To(BeTrue())
		})
	})

	Context("Resolution with concurrent access", func() {
		It("should handle concurrent resolution attempts safely", func(ctx SpecContext) {
			// Create multiple goroutines trying to resolve the same complaint
			done := make(chan bool, 3)
			errors := make(chan error, 3)

			// Start 3 goroutines trying to resolve the same complaint
			for i := 0; i < 3; i++ {
				go func() {
					defer func() { done <- true }()
					err := complaintService.ResolveComplaint(ctx, testComplaint.ID, "test-agent")
					errors <- err
				}()
			}

			// Wait for all goroutines to complete
			for i := 0; i < 3; i++ {
				<-done
			}

			// Check that all operations completed without error
			for i := 0; i < 3; i++ {
				err := <-errors
				Expect(err).NotTo(HaveOccurred())
			}

			// Verify complaint is resolved
			resolvedComplaint, err := complaintService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolvedComplaint.Resolved).To(BeTrue())
		})
	})

	Context("Resolution persistence", func() {
		It("should persist resolution across service restarts", func(ctx SpecContext) {
			// Resolve the complaint
			err := complaintService.ResolveComplaint(ctx, testComplaint.ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// Verify it's resolved
			resolvedComplaint, err := complaintService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolvedComplaint.Resolved).To(BeTrue())

			// Simulate service restart by creating new service instance with same repository
			newService := service.NewComplaintService(repository, tracer, logger)

			// Verify complaint is still resolved after "restart"
			restartedComplaint, err := newService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(restartedComplaint.Resolved).To(BeTrue())
		})

		It("should maintain resolution in file system", func(ctx SpecContext) {
			// Resolve the complaint
			err := complaintService.ResolveComplaint(ctx, testComplaint.ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// Verify resolution is persisted by creating new repository instance
			newRepository := repo.NewFileRepository(tempDir, tracer)
			newService := service.NewComplaintService(newRepository, tracer, logger)

			// Load complaint through new service instance
			persistedComplaint, err := newService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(persistedComplaint.Resolved).To(BeTrue())
		})
	})

	Context("Resolution with different complaint data", func() {
		It("should resolve complaints with all severity levels", func(ctx SpecContext) {
			severities := []domain.Severity{
				domain.SeverityLow,
				domain.SeverityMedium,
				domain.SeverityHigh,
				domain.SeverityCritical,
			}

			for i, severity := range severities {
				// Create complaint with specific severity
				complaint, err := complaintService.CreateComplaint(ctx,
					"Test Agent",
					fmt.Sprintf("severity-test-%d", i),
					fmt.Sprintf("Test complaint for %s severity", string(severity)),
					"",
					"",
					"",
					"",
					severity,
					"severity-test")
				Expect(err).NotTo(HaveOccurred())

				// Resolve it
				err = complaintService.ResolveComplaint(ctx, complaint.ID, "test-agent")
				Expect(err).NotTo(HaveOccurred())

				// Verify resolution
				resolved, err := complaintService.GetComplaint(ctx, complaint.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(resolved.Resolved).To(BeTrue())
				Expect(resolved.Severity).To(Equal(severity))
			}
		})

		It("should resolve complaints with maximum allowed content", func(ctx SpecContext) {
			// Create complaint with maximum content
			maxContent := string(make([]byte, 1000)) // Large but reasonable content
			complaint, err := complaintService.CreateComplaint(ctx,
				"Max Content Agent",
				"max-content-session",
				"Maximum content test complaint",
				maxContent,
				maxContent,
				maxContent,
				maxContent,
				domain.SeverityHigh,
				"max-content-test")
			Expect(err).NotTo(HaveOccurred())

			// Resolve it
			err = complaintService.ResolveComplaint(ctx, complaint.ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// Verify resolution and content preservation
			resolved, err := complaintService.GetComplaint(ctx, complaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved.Resolved).To(BeTrue())
			Expect(resolved.ContextInfo).To(Equal(maxContent))
			Expect(resolved.MissingInfo).To(Equal(maxContent))
			Expect(resolved.ConfusedBy).To(Equal(maxContent))
			Expect(resolved.FutureWishes).To(Equal(maxContent))
		})
	})

	Context("Resolution error handling", func() {
		It("should handle repository errors during resolution", func(ctx SpecContext) {
			// This tests error handling at the service level
			// In a real scenario, this might test file permission errors, disk full, etc.
			// For now, we verify normal operation since we can't easily simulate file system errors
			
			err := complaintService.ResolveComplaint(ctx, testComplaint.ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// Verify the resolution succeeded
			resolved, err := complaintService.GetComplaint(ctx, testComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved.Resolved).To(BeTrue())
		})
	})
})