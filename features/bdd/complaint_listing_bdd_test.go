package bdd_test

import (
	"context"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

var _ = Describe("Complaint Listing BDD Tests", func() {
	var (
		tempDir          string
		repository       repo.Repository
		complaintService *service.ComplaintService
		logger           *log.Logger
		tracer           tracing.Tracer
		testComplaints   []*domain.Complaint
	)

	BeforeEach(func() {
		// Create a temporary directory for each test
		tempDir = GinkgoT().TempDir()
		logger = log.New(os.Stdout)
		tracer = tracing.NewMockTracer("test")

		// Initialize repository and service
		repository = repo.NewFileRepository(tempDir, tracer)
		complaintService = service.NewComplaintService(repository, tracer, logger)

		// Create test complaints
		testComplaints = []*domain.Complaint{}

		// Create complaints with different data for testing
		complaint1, err := complaintService.CreateComplaint(context.Background(),
			"AI Assistant 1",
			"session-1",
			"Authentication issue",
			"JWT token validation unclear",
			"Missing error codes",
			"Documentation unclear",
			"Add examples",
			domain.SeverityHigh,
			"auth-project")
		Expect(err).NotTo(HaveOccurred())
		testComplaints = append(testComplaints, complaint1)

		complaint2, err := complaintService.CreateComplaint(context.Background(),
			"AI Assistant 2",
			"session-2",
			"API design confusion",
			"REST vs GraphQL unclear",
			"",
			"Inconsistent patterns",
			"Standardize approach",
			domain.SeverityMedium,
			"api-project")
		Expect(err).NotTo(HaveOccurred())
		testComplaints = append(testComplaints, complaint2)

		complaint3, err := complaintService.CreateComplaint(context.Background(),
			"AI Assistant 3",
			"session-3",
			"Database schema missing",
			"No table definitions",
			"Migration scripts absent",
			"Data relationships unclear",
			"Add ERD diagrams",
			domain.SeverityLow,
			"database-project")
		Expect(err).NotTo(HaveOccurred())
		testComplaints = append(testComplaints, complaint3)

		complaint4, err := complaintService.CreateComplaint(context.Background(),
			"AI Assistant 4",
			"session-4",
			"Testing framework confusion",
			"Unit vs integration unclear",
			"Mock patterns missing",
			"Test coverage unknown",
			"Add testing guidelines",
			domain.SeverityCritical,
			"auth-project") // Same project as complaint1
		Expect(err).NotTo(HaveOccurred())
		testComplaints = append(testComplaints, complaint4)
	})

	AfterEach(func() {
		// Clean up temporary directory
		os.RemoveAll(tempDir)
	})

	Context("List all complaints", func() {
		It("should return all complaints with pagination", func(ctx SpecContext) {
			// Get first page of complaints
			complaints, err := complaintService.ListComplaints(ctx, 2, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints)).To(Equal(2))

			// Get second page
			complaints2, err := complaintService.ListComplaints(ctx, 2, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints2)).To(Equal(2))

			// Verify no overlap
			ids1 := map[string]bool{}
			for _, c := range complaints {
				ids1[c.ID.Value] = true
			}
			for _, c := range complaints2 {
				Expect(ids1[c.ID.Value]).To(BeFalse())
			}
		})

		It("should return empty list when offset exceeds total", func(ctx SpecContext) {
			complaints, err := complaintService.ListComplaints(ctx, 10, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints)).To(Equal(0))
		})

		It("should return complaints in creation order", func(ctx SpecContext) {
			complaints, err := complaintService.ListComplaints(ctx, 10, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints)).To(Equal(4))

			// Should be ordered by creation time (oldest first due to file loading)
			for i := 1; i < len(complaints); i++ {
				Expect(complaints[i].Timestamp).To(
					BeTemporally(">=", complaints[i-1].Timestamp))
			}
		})
	})

	Context("List complaints by severity", func() {
		It("should filter complaints by severity level", func(ctx SpecContext) {
			// Get high severity complaints
			highComplaints, err := complaintService.GetComplaintsBySeverity(ctx, domain.SeverityHigh, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(highComplaints)).To(Equal(1))
			Expect(highComplaints[0].Severity).To(Equal(domain.SeverityHigh))
			Expect(highComplaints[0].TaskDescription).To(Equal("Authentication issue"))

			// Get medium severity complaints
			mediumComplaints, err := complaintService.GetComplaintsBySeverity(ctx, domain.SeverityMedium, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(mediumComplaints)).To(Equal(1))
			Expect(mediumComplaints[0].Severity).To(Equal(domain.SeverityMedium))

			// Get low severity complaints
			lowComplaints, err := complaintService.GetComplaintsBySeverity(ctx, domain.SeverityLow, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(lowComplaints)).To(Equal(1))
			Expect(lowComplaints[0].Severity).To(Equal(domain.SeverityLow))

			// Get critical severity complaints
			criticalComplaints, err := complaintService.GetComplaintsBySeverity(ctx, domain.SeverityCritical, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(criticalComplaints)).To(Equal(1))
			Expect(criticalComplaints[0].Severity).To(Equal(domain.SeverityCritical))
		})

		It("should respect limit parameter", func(ctx SpecContext) {
			// Create more complaints of same severity for testing limit
			for range 5 {
				_, err := complaintService.CreateComplaint(ctx,
					"Test Agent",
					"limit-test",
					"Low severity test",
					"",
					"",
					"",
					"",
					domain.SeverityLow,
					"limit-test")
				Expect(err).NotTo(HaveOccurred())
			}

			// Test with limit
			limitedComplaints, err := complaintService.GetComplaintsBySeverity(ctx, domain.SeverityLow, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(limitedComplaints)).To(Equal(3))

			// All should be low severity
			for _, complaint := range limitedComplaints {
				Expect(complaint.Severity).To(Equal(domain.SeverityLow))
			}
		})
	})

	Context("List complaints by project", func() {
		It("should filter complaints by project name", func(ctx SpecContext) {
			// Get complaints for auth-project
			authComplaints, err := complaintService.ListComplaintsByProject(ctx, "auth-project", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(authComplaints)).To(Equal(2))

			for _, complaint := range authComplaints {
				Expect(complaint.ProjectName.String()).To(Equal("auth-project"))
			}

			// Get complaints for api-project
			apiComplaints, err := complaintService.ListComplaintsByProject(ctx, "api-project", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(apiComplaints)).To(Equal(1))
			Expect(apiComplaints[0].ProjectName.String()).To(Equal("api-project"))

			// Get complaints for database-project
			dbComplaints, err := complaintService.ListComplaintsByProject(ctx, "database-project", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(dbComplaints)).To(Equal(1))
			Expect(dbComplaints[0].ProjectName.String()).To(Equal("database-project"))
		})

		It("should return empty for non-existent project", func(ctx SpecContext) {
			complaints, err := complaintService.ListComplaintsByProject(ctx, "non-existent-project", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints)).To(Equal(0))
		})

		It("should respect limit parameter for project filtering", func(ctx SpecContext) {
			// Create more complaints for auth-project
			for range 3 {
				_, err := complaintService.CreateComplaint(ctx,
					"Test Agent",
					"project-limit-test",
					"Auth project test",
					"",
					"",
					"",
					"",
					domain.SeverityLow,
					"auth-project")
				Expect(err).NotTo(HaveOccurred())
			}

			// Test with limit
			limitedComplaints, err := complaintService.ListComplaintsByProject(ctx, "auth-project", 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(limitedComplaints)).To(Equal(2))

			for _, complaint := range limitedComplaints {
				Expect(complaint.ProjectName.String()).To(Equal("auth-project"))
			}
		})
	})

	Context("List unresolved complaints", func() {
		It("should return only unresolved complaints", func(ctx SpecContext) {
			unresolvedComplaints, err := complaintService.ListUnresolvedComplaints(ctx, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(unresolvedComplaints)).To(Equal(4)) // All test complaints are unresolved

			for _, complaint := range unresolvedComplaints {
				Expect(complaint.IsResolved()).To(BeFalse())
			}
		})

		It("should exclude resolved complaints", func(ctx SpecContext) {
			// Resolve one complaint
			err := complaintService.ResolveComplaint(ctx, testComplaints[0].ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// List unresolved complaints
			unresolvedComplaints, err := complaintService.ListUnresolvedComplaints(ctx, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(unresolvedComplaints)).To(Equal(3)) // One was resolved

			// Verify resolved complaint is not in list
			for _, complaint := range unresolvedComplaints {
				Expect(complaint.ID.Value).NotTo(Equal(testComplaints[0].ID.Value))
				Expect(complaint.IsResolved()).To(BeFalse())
			}
		})

		It("should respect limit parameter for unresolved filtering", func(ctx SpecContext) {
			limitedComplaints, err := complaintService.ListUnresolvedComplaints(ctx, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(limitedComplaints)).To(Equal(2))

			for _, complaint := range limitedComplaints {
				Expect(complaint.IsResolved()).To(BeFalse())
			}
		})
	})

	Context("Search complaints", func() {
		It("should search complaint content", func(ctx SpecContext) {
			// Search for "authentication"
			authResults, err := complaintService.SearchComplaints(ctx, "authentication", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(authResults)).To(Equal(1))
			Expect(strings.ToLower(authResults[0].TaskDescription)).To(ContainSubstring("authentication"))

			// Search for "API"
			apiResults, err := complaintService.SearchComplaints(ctx, "API", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(apiResults)).To(Equal(1))
			Expect(apiResults[0].TaskDescription).To(ContainSubstring("API"))

			// Search for "database"
			dbResults, err := complaintService.SearchComplaints(ctx, "database", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(dbResults)).To(Equal(1))
			Expect(strings.ToLower(dbResults[0].TaskDescription)).To(ContainSubstring("database"))
		})

		It("should be case-insensitive", func(ctx SpecContext) {
			// Search in different cases
			lowerResults, err := complaintService.SearchComplaints(ctx, "jwt", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(lowerResults)).To(Equal(1))

			upperResults, err := complaintService.SearchComplaints(ctx, "JWT", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(upperResults)).To(Equal(1))

			mixedResults, err := complaintService.SearchComplaints(ctx, "JwT", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(mixedResults)).To(Equal(1))
		})

		It("should search across multiple fields", func(ctx SpecContext) {
			// Search in context info
			contextResults, err := complaintService.SearchComplaints(ctx, "validation", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(contextResults)).To(Equal(1))
			Expect(contextResults[0].ContextInfo).To(ContainSubstring("validation"))

			// Search in missing info
			missingResults, err := complaintService.SearchComplaints(ctx, "codes", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(missingResults)).To(Equal(1))
			Expect(missingResults[0].MissingInfo).To(ContainSubstring("codes"))

			// Search in confused by
			confusedResults, err := complaintService.SearchComplaints(ctx, "Documentation", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(confusedResults)).To(Equal(1))
			Expect(confusedResults[0].ConfusedBy).To(ContainSubstring("Documentation"))
		})

		It("should return empty for non-matching search", func(ctx SpecContext) {
			results, err := complaintService.SearchComplaints(ctx, "nonexistentterm", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(results)).To(Equal(0))
		})

		It("should respect limit parameter for search", func(ctx SpecContext) {
			// Create complaints with common term
			for range 5 {
				_, err := complaintService.CreateComplaint(ctx,
					"Search Test Agent",
					"search-session",
					"Common search term test",
					"common term in context",
					"",
					"",
					"",
					domain.SeverityLow,
					"search-test")
				Expect(err).NotTo(HaveOccurred())
			}

			// Search with limit
			limitedResults, err := complaintService.SearchComplaints(ctx, "common", 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(limitedResults)).To(Equal(3))
		})
	})
})
