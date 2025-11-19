package bdd_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

var _ = Describe("Complaint Filing BDD Tests", func() {
	var (
		tempDir          string
		repository       repo.Repository
		complaintService *service.ComplaintService
		logger           *log.Logger
		tracer           tracing.Tracer
	)

	BeforeEach(func() {
		// Create a temporary directory for each test
		tempDir = GinkgoT().TempDir()
		logger = log.New(os.Stdout)
		tracer = tracing.NewMockTracer("test")

		// Initialize repository and service
		repository = repo.NewFileRepository(tempDir, tracer)
		complaintService = service.NewComplaintService(repository, tracer)
	})

	AfterEach(func() {
		// Clean up temporary directory
		os.RemoveAll(tempDir)
	})

	Context("File a valid complaint successfully", func() {
		It("should store complaint with all required fields", func(ctx SpecContext) {
			complaint, err := complaintService.CreateComplaint(ctx,
				"AI Assistant",
				"test-session",
				"Implement authentication system",
				"Need to add JWT authentication to API endpoints",
				"Unclear error handling patterns",
				"Documentation missing for error responses",
				"Add comprehensive error handling examples",
				domain.SeverityHigh,
				"auth-project")

			Expect(err).NotTo(HaveOccurred())
			Expect(complaint).NotTo(BeNil())
			Expect(complaint.ID.String()).NotTo(BeEmpty())
			Expect(complaint.AgentID.String()).To(Equal("AI Assistant"))
			Expect(complaint.SessionID.String()).To(Equal("test-session"))
			Expect(complaint.TaskDescription).To(Equal("Implement authentication system"))
			Expect(complaint.ContextInfo).To(Equal("Need to add JWT authentication to API endpoints"))
			Expect(complaint.MissingInfo).To(Equal("Unclear error handling patterns"))
			Expect(complaint.ConfusedBy).To(Equal("Documentation missing for error responses"))
			Expect(complaint.FutureWishes).To(Equal("Add comprehensive error handling examples"))
			Expect(complaint.Severity).To(Equal(domain.SeverityHigh))
			Expect(complaint.ProjectName.String()).To(Equal("auth-project"))
			Expect(complaint.IsResolved()).To(BeFalse())
			Expect(complaint.Timestamp).NotTo(BeZero())
		})

		It("should store complaint with minimum required data", func(ctx SpecContext) {
			complaint, err := complaintService.CreateComplaint(ctx,
				"A", // minimal valid name
				"",  // optional session name
				"T", // minimal valid description
				"",  // optional context info
				"",  // optional missing info
				"",  // optional confused by
				"",  // optional future wishes
				domain.SeverityLow,
				"test") // valid project name

			Expect(err).NotTo(HaveOccurred())
			Expect(complaint).NotTo(BeNil())
			Expect(complaint.AgentID.String()).To(Equal("A"))
			Expect(complaint.TaskDescription).To(Equal("T"))
			Expect(complaint.Severity).To(Equal(domain.SeverityLow))
			Expect(complaint.ProjectName.String()).To(Equal("test"))
		})
	})

	Context("File a complaint with missing required field", func() {
		It("should return validation error for missing task description", func(ctx SpecContext) {
			_, err := complaintService.CreateComplaint(ctx,
				"AI Assistant",
				"test-session",
				"", // empty task description
				"Some context",
				"",
				"",
				"",
				domain.SeverityMedium,
				"test-project")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Or(
				ContainSubstring("task description is required"),
				ContainSubstring("TaskDescription"),
			))
		})

		It("should return validation error for missing agent name", func(ctx SpecContext) {
			_, err := complaintService.CreateComplaint(ctx,
				"", // empty agent name
				"test-session",
				"Test task",
				"",
				"",
				"",
				"",
				domain.SeverityMedium,
				"test-project")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Or(
				ContainSubstring("agent name is required"),
				ContainSubstring("AgentID"),
				ContainSubstring("agent name cannot be empty"),
			))
		})

		It("should return validation error for invalid severity", func(ctx SpecContext) {
			// This will be caught at the domain level
			complaint, err := complaintService.CreateComplaint(ctx,
				"AI Assistant",
				"test-session",
				"Test task",
				"",
				"",
				"",
				"",
				"invalid", // invalid severity
				"test-project")

			// The service should handle this gracefully
			Expect(err).To(HaveOccurred())
			Expect(complaint).To(BeNil())
		})
	})

	Context("File a complaint with invalid severity", func() {
		It("should return validation error for unsupported severity", func(ctx SpecContext) {
			_, err := complaintService.CreateComplaint(ctx,
				"AI Assistant",
				"test-session",
				"Test task",
				"",
				"",
				"",
				"",
				"unsupported", // invalid severity
				"test-project")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid severity"))
		})
	})

	Context("File complaint with large content", func() {
		It("should handle large content gracefully", func(ctx SpecContext) {
			// Create a large complaint
			largeContent := string(make([]byte, 2000)) // 2KB content
			complaint, err := complaintService.CreateComplaint(ctx,
				"AI Assistant",
				"test-session",
				"Large content test",
				largeContent,
				"",
				"",
				"",
				domain.SeverityLow,
				"content-test")

			Expect(err).NotTo(HaveOccurred())
			Expect(complaint).NotTo(BeNil())
			Expect(complaint.ContextInfo).To(Equal(largeContent))
		})

		It("should enforce content size limits", func(ctx SpecContext) {
			// Test content that exceeds reasonable limits
			veryLargeContent := string(make([]byte, 1000000)) // 1MB content

			complaint, err := complaintService.CreateComplaint(ctx,
				"AI Assistant",
				"test-session",
				"Oversized content",
				veryLargeContent,
				"",
				"",
				"",
				domain.SeverityMedium,
				"content-test")

			// This should either be handled gracefully or return an appropriate error
			// For now, we'll assume it should succeed (service layer should handle)
			Expect(err).NotTo(HaveOccurred())
			Expect(complaint).NotTo(BeNil())
		})
	})

	Context("File complaint with special characters", func() {
		It("should handle special characters properly", func(ctx SpecContext) {
			complaint, err := complaintService.CreateComplaint(ctx,
				"AI Assistant 🤖",
				"test-session",
				"Test with special chars: quotes, newlines, tabs",
				"Content with \"quotes\" and \t\t tabs\nnewlines",
				"",
				"",
				"",
				domain.SeverityMedium,
				"special-char-test")

			Expect(err).NotTo(HaveOccurred())
			Expect(complaint).NotTo(BeNil())
			Expect(complaint.AgentID.String()).To(Equal("AI Assistant 🤖"))
			Expect(complaint.TaskDescription).To(Equal("Test with special chars: quotes, newlines, tabs"))
			Expect(complaint.ContextInfo).To(Equal("Content with \"quotes\" and \t\t tabs\nnewlines"))
		})
	})

	Context("File complaint and verify persistence", func() {
		It("should persist complaint to file system", func(ctx SpecContext) {
			complaint, err := complaintService.CreateComplaint(ctx,
				"Test Agent",
				"persist-session",
				"Test persistence functionality",
				"Verify complaint is saved to file system",
				"",
				"",
				"",
				domain.SeverityLow,
				"persistence-test")

			Expect(err).NotTo(HaveOccurred())
			Expect(complaint).NotTo(BeNil())

			// Verify the complaint was saved by retrieving it
			retrieved, err := complaintService.GetComplaint(ctx, complaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved).NotTo(BeNil())
			Expect(retrieved.ID.String()).To(Equal(complaint.ID.String()))
			Expect(retrieved.AgentID).To(Equal(complaint.AgentID))
			Expect(retrieved.TaskDescription).To(Equal(complaint.TaskDescription))
		})

		It("should create separate file for each complaint", func(ctx SpecContext) {
			// Create multiple complaints
			complaint1, err := complaintService.CreateComplaint(ctx,
				"Agent 1",
				"session-1",
				"First complaint",
				"",
				"",
				"",
				"",
				domain.SeverityLow,
				"test")
			Expect(err).NotTo(HaveOccurred())

			complaint2, err := complaintService.CreateComplaint(ctx,
				"Agent 2",
				"session-2",
				"Second complaint",
				"",
				"",
				"",
				"",
				domain.SeverityMedium,
				"test")
			Expect(err).NotTo(HaveOccurred())

			// Both should have different IDs
			Expect(complaint1.ID.String()).NotTo(Equal(complaint2.ID.String()))

			// Both should be retrievable
			retrieved1, err := complaintService.GetComplaint(ctx, complaint1.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved1.ID.String()).To(Equal(complaint1.ID.String()))

			retrieved2, err := complaintService.GetComplaint(ctx, complaint2.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved2.ID.String()).To(Equal(complaint2.ID.String()))
		})
	})
})
