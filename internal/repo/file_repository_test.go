package repo_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

var _ = Describe("FileRepository", func() {
	var (
		tempDir    string
		repository repo.Repository
		tracer     tracing.Tracer
		ctx        context.Context
	)

	BeforeEach(func() {
		tempDir = GinkgoT().TempDir()
		tracer = tracing.NewMockTracer("test")
		repository = repo.NewFileRepository(tempDir, tracer)
		ctx = context.Background()
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Context("Complaint Creation", func() {
		It("should save a new complaint successfully", func() {
			complaintID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       domain.MustNewAgentName("Test Agent"),
				SessionName:     domain.MustNewSessionName("test-session"),
				TaskDescription: "Test task",
				ContextInfo:     "Test context",
				MissingInfo:     "Test missing info",
				ConfusedBy:      "Test confusion",
				FutureWishes:    "Test wishes",
				Severity:        domain.SeverityHigh,
				ProjectName:     domain.MustNewProjectName("test-project"),
			}

			// Test saving complaint
			err = repository.Save(ctx, complaint)
			Expect(err).NotTo(HaveOccurred())

			// Verify file was created
			filename := filepath.Join(tempDir, complaintID.String()+".json")
			_, err = os.Stat(filename)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should store complaint with all fields", func() {
			complaintID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			now := time.Now()
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       domain.MustNewAgentName("Test Agent"),
				SessionName:     domain.MustNewSessionName("test-session"),
				TaskDescription: "Test task description",
				ContextInfo:     "Test context information",
				MissingInfo:     "Missing information",
				ConfusedBy:      "Confused by this",
				FutureWishes:    "Future improvements",
				Severity:        domain.SeverityMedium,
				Timestamp:       now,
				ProjectName:     domain.MustNewProjectName("test-project"),
			}

			err = repository.Save(ctx, complaint)
			Expect(err).NotTo(HaveOccurred())

			// Read back and verify
			filename := filepath.Join(tempDir, complaintID.String()+".json")
			data, err := os.ReadFile(filename)
			Expect(err).NotTo(HaveOccurred())

			var stored domain.Complaint
			err = json.Unmarshal(data, &stored)
			Expect(err).NotTo(HaveOccurred())

			Expect(stored.ID.Value).To(Equal(complaint.ID.Value))
			Expect(stored.AgentName).To(Equal(complaint.AgentName))
			Expect(stored.TaskDescription).To(Equal(complaint.TaskDescription))
			Expect(stored.Severity).To(Equal(complaint.Severity))
			Expect(stored.ProjectName).To(Equal(complaint.ProjectName))
			Expect(stored.IsResolved()).To(BeFalse())
		})
	})

	Context("Complaint Retrieval", func() {
		var savedComplaint *domain.Complaint

		BeforeEach(func() {
			complaintID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			savedComplaint = &domain.Complaint{
				ID:              complaintID,
				AgentName:       domain.MustNewAgentName("Test Agent"),
				SessionName:     domain.MustNewSessionName("test-session"),
				TaskDescription: "Test task",
				ContextInfo:     "Test context",
				MissingInfo:     "Test missing info",
				ConfusedBy:      "Test confusion",
				FutureWishes:    "Test wishes",
				Severity:        domain.SeverityLow,
				ProjectName:     domain.MustNewProjectName("test-project"),
			}

			err = repository.Save(ctx, savedComplaint)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should find complaint by ID", func() {
			found, err := repository.FindByID(ctx, savedComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).NotTo(BeNil())
			Expect(found.ID.Value).To(Equal(savedComplaint.ID.Value))
			Expect(found.AgentName).To(Equal(savedComplaint.AgentName))
			Expect(found.TaskDescription).To(Equal(savedComplaint.TaskDescription))
		})

		It("should return not found for non-existent ID", func() {
			nonExistentID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			found, err := repository.FindByID(ctx, nonExistentID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(found).To(BeNil())
		})
	})

	Context("Complaint Listing", func() {
		BeforeEach(func() {
			// Create multiple complaints
			for range 5 {
				complaintID, err := domain.NewComplaintID()
				Expect(err).NotTo(HaveOccurred())

				complaint := &domain.Complaint{
					ID:              complaintID,
					AgentName:       domain.MustNewAgentName("Test Agent"),
					SessionName:     domain.MustNewSessionName("test-session"),
					TaskDescription: "Test task",
					ContextInfo:     "Test context",
					MissingInfo:     "Test missing info",
					ConfusedBy:      "Test confusion",
					FutureWishes:    "Test wishes",
					Severity:        domain.SeverityMedium,
					ProjectName:     domain.MustNewProjectName("test-project"),
				}

				err = repository.Save(ctx, complaint)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should list all complaints with pagination", func() {
			complaints, err := repository.FindAll(ctx, 10, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints)).To(Equal(5))

			// Test pagination
			complaints2, err := repository.FindAll(ctx, 3, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints2)).To(Equal(3))

			complaints3, err := repository.FindAll(ctx, 3, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints3)).To(Equal(3))
		})
	})

	Context("Complaint Updates", func() {
		var savedComplaint *domain.Complaint

		BeforeEach(func() {
			complaintID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			savedComplaint = &domain.Complaint{
				ID:              complaintID,
				AgentName:       domain.MustNewAgentName("Test Agent"),
				SessionName:     domain.MustNewSessionName("test-session"),
				TaskDescription: "Test task",
				ContextInfo:     "Test context",
				MissingInfo:     "Test missing info",
				ConfusedBy:      "Test confusion",
				FutureWishes:    "Test wishes",
				Severity:        domain.SeverityLow,
				ProjectName:     domain.MustNewProjectName("test-project"),
			}

			err = repository.Save(ctx, savedComplaint)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update an existing complaint", func() {
			// Modify complaint
			resolveTime := time.Now()
			savedComplaint.ResolvedAt = &resolveTime

			err := repository.Update(ctx, savedComplaint)
			Expect(err).NotTo(HaveOccurred())

			// Verify update
			found, err := repository.FindByID(ctx, savedComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(found.IsResolved()).To(BeTrue())
			Expect(found.ResolvedAt).NotTo(BeNil())
			Expect(found.ResolvedAt.Format(time.RFC3339)).To(Equal(resolveTime.Format(time.RFC3339)))
		})

		It("should return error when updating non-existent complaint", func() {
			nonExistentID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			nonExistentComplaint := &domain.Complaint{
				ID:              nonExistentID,
				AgentName:       domain.MustNewAgentName("Test Agent"),
				TaskDescription: "Non-existent",
				Severity:        domain.SeverityLow,
			}

			err = repository.Update(ctx, nonExistentComplaint)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should NOT create duplicate files when updating complaint", func() {
			// Initial save - should create exactly one file
			err := repository.Save(ctx, savedComplaint)
			Expect(err).NotTo(HaveOccurred())

			// List files to verify we have exactly one
			files, err := os.ReadDir(tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(files)).To(Equal(1), "Should have exactly one file after initial save")

			// Update complaint multiple times
			for i := range 3 {
				savedComplaint.TaskDescription = fmt.Sprintf("Updated task %d", i)
				if i%2 == 0 {
					resolveTime := time.Now()
					savedComplaint.ResolvedAt = &resolveTime
				} else {
					savedComplaint.ResolvedAt = nil
				}

				err := repository.Update(ctx, savedComplaint)
				Expect(err).NotTo(HaveOccurred(), "Update %d should succeed", i)
			}

			// Verify still only one file exists
			files, err = os.ReadDir(tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(files)).To(Equal(1), "Should still have exactly one file after multiple updates")

			// Verify the file has the correct name (UUID-only)
			expectedFilename := savedComplaint.ID.String() + ".json"
			Expect(files[0].Name()).To(Equal(expectedFilename), "Filename should be UUID-only")

			// Verify content is correctly updated
			found, err := repository.FindByID(ctx, savedComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(found.TaskDescription).To(Equal("Updated task 2"))
			Expect(found.IsResolved()).To(BeTrue())
		})

		It("should update existing file in-place", func() {
			// Get initial file modification time
			filename := filepath.Join(tempDir, savedComplaint.ID.String()+".json")
			fileInfo, err := os.Stat(filename)
			Expect(err).NotTo(HaveOccurred())
			initialModTime := fileInfo.ModTime()

			// Wait a bit to ensure different timestamp
			time.Sleep(10 * time.Millisecond)

			// Update the complaint
			savedComplaint.TaskDescription = "Updated in-place"
			resolveTime := time.Now()
			savedComplaint.ResolvedAt = &resolveTime

			err = repository.Update(ctx, savedComplaint)
			Expect(err).NotTo(HaveOccurred())

			// Verify file was modified (newer timestamp)
			fileInfo, err = os.Stat(filename)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.ModTime()).To(BeTemporally(">", initialModTime))

			// Verify content was updated
			found, err := repository.FindByID(ctx, savedComplaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(found.TaskDescription).To(Equal("Updated in-place"))
			Expect(found.IsResolved()).To(BeTrue())
			Expect(found.ResolvedAt).NotTo(BeNil())
		})
	})

	Context("Complaint Search", func() {
		BeforeEach(func() {
			// Create complaints with different content
			contents := []struct {
				task     string
				proj     string
				severity domain.Severity
			}{
				{"Database connection issue", "project-alpha", domain.SeverityHigh},
				{"API response parsing", "project-beta", domain.SeverityMedium},
				{"Authentication problems", "project-alpha", domain.SeverityCritical},
				{"Database schema changes", "project-gamma", domain.SeverityLow},
				{"API endpoint changes", "project-beta", domain.SeverityHigh},
			}

			for _, content := range contents {
				complaintID, err := domain.NewComplaintID()
				Expect(err).NotTo(HaveOccurred())

				complaint := &domain.Complaint{
					ID:              complaintID,
					AgentName:       domain.MustNewAgentName("Test Agent"),
					SessionName:     domain.MustNewSessionName("test-session"),
					TaskDescription: content.task,
					ContextInfo:     "Test context",
					MissingInfo:     "Test missing info",
					ConfusedBy:      "Test confusion",
					FutureWishes:    "Test wishes",
					Severity:        content.severity,
					ProjectName:     domain.MustNewProjectName(content.proj),
				}

				err = repository.Save(ctx, complaint)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should search complaints by text", func() {
			results, err := repository.Search(ctx, "database", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(results)).To(Equal(2))

			for _, result := range results {
				Expect(result.TaskDescription).To(ContainSubstring("Database"))
			}
		})

		It("should search with case-insensitive matching", func() {
			results, err := repository.Search(ctx, "API", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(results)).To(Equal(2))

			for _, result := range results {
				Expect(result.TaskDescription).To(ContainSubstring("API"))
			}
		})

		It("should find by severity", func() {
			results, err := repository.FindBySeverity(ctx, domain.SeverityHigh, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(results)).To(Equal(2))

			for _, result := range results {
				Expect(result.Severity).To(Equal(domain.SeverityHigh))
			}
		})

		It("should find by project name", func() {
			results, err := repository.FindByProject(ctx, "project-alpha", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(results)).To(Equal(2))

			for _, result := range results {
				Expect(result.ProjectName).To(Equal("project-alpha"))
			}
		})

		It("should find unresolved complaints", func() {
			results, err := repository.FindUnresolved(ctx, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(results)).To(Equal(5)) // All are unresolved

			for _, result := range results {
				Expect(result.IsResolved()).To(BeFalse())
			}
		})
	})

	Context("Error Handling", func() {
		It("should handle invalid directory gracefully", func() {
			invalidDir := "/invalid/nonexistent/path"
			testRepo := repo.NewFileRepository(invalidDir, tracer)
			complaintID, _ := domain.NewComplaintID()

			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       domain.MustNewAgentName("Test Agent"),
				TaskDescription: "Test task",
				Severity:        domain.SeverityLow,
			}

			err := testRepo.Save(ctx, complaint)
			Expect(err).To(HaveOccurred())
		})

		It("should handle corrupted JSON files", func() {
			complaintID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			// Create a corrupted JSON file
			filename := filepath.Join(tempDir, complaintID.String()+".json")
			err = os.WriteFile(filename, []byte("invalid json content"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Try to read it
			found, err := repository.FindByID(ctx, complaintID)
			Expect(err).To(HaveOccurred())
			Expect(found).To(BeNil())
		})
	})
})
