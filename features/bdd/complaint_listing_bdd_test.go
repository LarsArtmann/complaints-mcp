package bdd_test

import (
	"context"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
		tracer           tracing.Tracer
		testComplaints   []*domain.Complaint
	)

	expectAllUnresolved := func(complaints []*domain.Complaint) {
		for _, complaint := range complaints {
			Expect(complaint.IsResolved()).To(BeFalse())
		}
	}

	expectEmptyResults := func(ctx context.Context, fn func(context.Context) ([]*domain.Complaint, error)) {
		results, err := fn(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(BeEmpty())
	}

	expectUnresolvedWithCount := func(ctx context.Context, fn func(context.Context, int) ([]*domain.Complaint, error), limit, expectedCount int) {
		complaints, err := fn(ctx, limit)
		Expect(err).NotTo(HaveOccurred())
		Expect(complaints).To(HaveLen(expectedCount))
		expectAllUnresolved(complaints)
	}

	expectEmptySearchResults := func(ctx context.Context, limit int) {
		expectEmptyResults(ctx, func(ctx context.Context) ([]*domain.Complaint, error) {
			return complaintService.SearchComplaints(ctx, "nonexistentterm", limit)
		})
	}

	expectEmptyPaginatedResults := func(ctx context.Context, limit, offset int) {
		expectEmptyResults(ctx, func(ctx context.Context) ([]*domain.Complaint, error) {
			return complaintService.ListComplaints(ctx, limit, offset)
		})
	}

	BeforeEach(func() {
		// Create a temporary directory for each test
		tempDir = GinkgoT().TempDir()
		tracer = tracing.NewMockTracer("test")

		// Initialize repository and service
		repository = repo.NewFileRepository(tempDir, tracer)
		complaintService = service.NewComplaintService(repository, tracer)

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
			"auth-project", "")
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
			"api-project", "")
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
			"database-project", "")
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
			"auth-project", "") // Same project as complaint1
		Expect(err).NotTo(HaveOccurred())

		testComplaints = append(testComplaints, complaint4)
	})

	AfterEach(func() {
		// Clean up temporary directory
		os.RemoveAll(tempDir)
	})

	createTestComplaints := func(ctx context.Context, count int, agentID, sessionID, taskDescription, contextInfo, projectID string) {
		for range count {
			_, err := complaintService.CreateComplaint(ctx,
				agentID,
				sessionID,
				taskDescription,
				contextInfo,
				"",
				"",
				"",
				domain.SeverityLow,
				projectID, "")
			Expect(err).NotTo(HaveOccurred())
		}
	}

	Context("List all complaints", func() {
		It("should return all complaints with pagination", func(ctx SpecContext) {
			// Get first page of complaints
			complaints, err := complaintService.ListComplaints(ctx, 2, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(complaints).To(HaveLen(2))

			// Get second page
			complaints2, err := complaintService.ListComplaints(ctx, 2, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(complaints2).To(HaveLen(2))

			// Verify no overlap
			ids1 := map[string]bool{}
			for _, c := range complaints {
				ids1[c.ID.String()] = true
			}

			for _, c := range complaints2 {
				Expect(ids1[c.ID.String()]).To(BeFalse())
			}
		})

		It("should return empty list when offset exceeds total", func(ctx SpecContext) {
			expectEmptyPaginatedResults(ctx, 10, 100)
		})

		It("should return complaints in creation order", func(ctx SpecContext) {
			complaints, err := complaintService.ListComplaints(ctx, 10, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(complaints).To(HaveLen(4))

			// Should be ordered by creation time (oldest first due to file loading)
			for i := 1; i < len(complaints); i++ {
				Expect(complaints[i].Timestamp).To(
					BeTemporally(">=", complaints[i-1].Timestamp.Add(-time.Millisecond)))
			}
		})
	})

	Context("List complaints by severity", func() {
		It("should filter complaints by severity level", func(ctx SpecContext) {
			// Get high severity complaints
			highComplaints, err := repository.FindBySeverity(
				ctx,
				domain.SeverityHigh,
				10,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(highComplaints).To(HaveLen(1))
			Expect(highComplaints[0].Severity).To(Equal(domain.SeverityHigh))
			Expect(highComplaints[0].TaskDescription).To(Equal("Authentication issue"))

			// Get medium severity complaints
			mediumComplaints, err := repository.FindBySeverity(
				ctx,
				domain.SeverityMedium,
				10,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(mediumComplaints).To(HaveLen(1))
			Expect(mediumComplaints[0].Severity).To(Equal(domain.SeverityMedium))

			// Get low severity complaints
			lowComplaints, err := repository.FindBySeverity(
				ctx,
				domain.SeverityLow,
				10,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(lowComplaints).To(HaveLen(1))
			Expect(lowComplaints[0].Severity).To(Equal(domain.SeverityLow))

			// Get critical severity complaints
			criticalComplaints, err := repository.FindBySeverity(
				ctx,
				domain.SeverityCritical,
				10,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(criticalComplaints).To(HaveLen(1))
			Expect(criticalComplaints[0].Severity).To(Equal(domain.SeverityCritical))
		})

		It("should respect limit parameter", func(ctx SpecContext) {
			// Create more complaints of same severity for testing limit
			createTestComplaints(ctx, 5, "Test Agent", "limit-test", "Low severity test", "", "limit-test")

			// Test with limit
			limitedComplaints, err := repository.FindBySeverity(
				ctx,
				domain.SeverityLow,
				3,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(limitedComplaints).To(HaveLen(3))

			// All should be low severity
			for _, complaint := range limitedComplaints {
				Expect(complaint.Severity).To(Equal(domain.SeverityLow))
			}
		})
	})

	Context("List complaints by project", func() {
		It("should filter complaints by project name", func(ctx SpecContext) {
			// Get complaints for auth-project
			authComplaints, err := repository.FindByProject(ctx, "auth-project", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(authComplaints).To(HaveLen(2))

			for _, complaint := range authComplaints {
				Expect(complaint.ProjectID.String()).To(Equal("auth-project"))
			}

			// Get complaints for api-project
			apiComplaints, err := repository.FindByProject(ctx, "api-project", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(apiComplaints).To(HaveLen(1))
			Expect(apiComplaints[0].ProjectID.String()).To(Equal("api-project"))

			// Get complaints for database-project
			dbComplaints, err := repository.FindByProject(
				ctx,
				"database-project",
				10,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbComplaints).To(HaveLen(1))
			Expect(dbComplaints[0].ProjectID.String()).To(Equal("database-project"))
		})

		It("should return empty for non-existent project", func(ctx SpecContext) {
			expectEmptyResults(ctx, func(ctx context.Context) ([]*domain.Complaint, error) {
				return repository.FindByProject(ctx, "non-existent-project", 10)
			})
		})

		It("should respect limit parameter for project filtering", func(ctx SpecContext) {
			// Create more complaints for auth-project
			createTestComplaints(ctx, 3, "Test Agent", "project-limit-test", "Auth project test", "", "auth-project")

			// Test with limit
			limitedComplaints, err := repository.FindByProject(
				ctx,
				"auth-project",
				2,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(limitedComplaints).To(HaveLen(2))

			for _, complaint := range limitedComplaints {
				Expect(complaint.ProjectID.String()).To(Equal("auth-project"))
			}
		})
	})

	Context("List unresolved complaints", func() {
		It("should return only unresolved complaints", func(ctx SpecContext) {
			expectUnresolvedWithCount(ctx, complaintService.ListUnresolvedComplaints, 10, 4)
		})

		It("should exclude resolved complaints", func(ctx SpecContext) {
			// Resolve one complaint
			_, err := complaintService.ResolveComplaint(ctx, testComplaints[0].ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())

			// List unresolved complaints
			unresolvedComplaints, err := repository.FindUnresolved(ctx, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(unresolvedComplaints).To(HaveLen(3)) // One was resolved

			// Verify resolved complaint is not in list
			for _, complaint := range unresolvedComplaints {
				Expect(complaint.ID.String()).NotTo(Equal(testComplaints[0].ID.String()))
				Expect(complaint.IsResolved()).To(BeFalse())
			}
		})

		It("should respect limit parameter for unresolved filtering", func(ctx SpecContext) {
			expectUnresolvedWithCount(ctx, repository.FindUnresolved, 2, 2)
		})
	})

	Context("Search complaints", func() {
		It("should search complaint content", func(ctx SpecContext) {
			// Search for "authentication"
			authResults, err := repository.Search(ctx, "authentication", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(authResults).To(HaveLen(1))
			Expect(
				strings.ToLower(authResults[0].TaskDescription),
			).To(ContainSubstring("authentication"))

			// Search for "API"
			apiResults, err := repository.Search(ctx, "API", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(apiResults).To(HaveLen(1))
			Expect(apiResults[0].TaskDescription).To(ContainSubstring("API"))

			// Search for "database"
			dbResults, err := repository.Search(ctx, "database", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbResults).To(HaveLen(1))
			Expect(strings.ToLower(dbResults[0].TaskDescription)).To(ContainSubstring("database"))
		})

		It("should be case-insensitive", func(ctx SpecContext) {
			// Search in different cases
			lowerResults, err := complaintService.SearchComplaints(ctx, "jwt", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(lowerResults).To(HaveLen(1))

			upperResults, err := complaintService.SearchComplaints(ctx, "JWT", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(upperResults).To(HaveLen(1))

			mixedResults, err := complaintService.SearchComplaints(ctx, "JwT", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(mixedResults).To(HaveLen(1))
		})

		It("should search across multiple fields", func(ctx SpecContext) {
			// Search in context info
			contextResults, err := complaintService.SearchComplaints(ctx, "validation", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(contextResults).To(HaveLen(1))
			Expect(contextResults[0].ContextInfo).To(ContainSubstring("validation"))

			// Search in missing info
			missingResults, err := complaintService.SearchComplaints(ctx, "codes", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(missingResults).To(HaveLen(1))
			Expect(missingResults[0].MissingInfo).To(ContainSubstring("codes"))

			// Search in confused by
			confusedResults, err := complaintService.SearchComplaints(ctx, "Documentation", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(confusedResults).To(HaveLen(1))
			Expect(confusedResults[0].ConfusedBy).To(ContainSubstring("Documentation"))
		})

		It("should return empty for non-matching search", func(ctx SpecContext) {
			expectEmptySearchResults(ctx, 10)
		})

		It("should respect limit parameter for search", func(ctx SpecContext) {
			// Create complaints with common term
			createTestComplaints(ctx, 5, "Search Test Agent", "search-session", "Common search term test", "common term in context", "search-test")

			// Search with limit
			limitedResults, err := complaintService.SearchComplaints(ctx, "common", 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(limitedResults).To(HaveLen(3))
		})
	})
})
