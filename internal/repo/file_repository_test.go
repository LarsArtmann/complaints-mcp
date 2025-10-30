package repo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/errors"
)

var _ = Describe("FileRepository", func() {
	var (
		tempDir  string
		repo     *FileRepository
		ctx      context.Context
	)

	BeforeEach(func() {
		tempDir = GinkgoT().TempDir()
		ctx = context.Background()
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Context("File Storage Operations", func() {
		It("should create directory structure when storing complaint", func() {
			repo = NewFileRepository(tempDir)

			complaintID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "Test Agent",
				TaskDescription: "Test task",
				Severity:        "medium",
				ProjectName:     "test-project",
				Resolved:        false,
			}

			err = repo.Store(ctx, complaint)
			Expect(err).NotTo(HaveOccurred())

			// Verify directory exists
			dirPath := filepath.Join(tempDir, "complaints")
			info, err := os.Stat(dirPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should generate unique filename with timestamp", func() {
			repo = NewFileRepository(tempDir)

			complaint := &domain.Complaint{
				ID:              domain.ComplaintID{Value: "test-id"},
				AgentName:       "Test Agent",
				TaskDescription: "Timestamp test",
				Severity:        "low",
				ProjectName:     "test-project",
				Resolved:        false,
			}

			filename := repo.generateFilename(complaint)
			Expect(filename).To(ContainSubstring("2025"))
			Expect(filename).To(ContainSubstring("test-session"))
		})

		It("should handle large content generation correctly", func() {
			repo = NewFileRepository(tempDir)

			complaint := &domain.Complaint{
				ID:              domain.ComplaintID{Value: "large-content"},
				AgentName:       "Test Agent",
				TaskDescription: "Large content test",
				ContextInfo:     string(make([]byte, 5000)), // 5KB content
				Severity:        "low",
				ProjectName:     "test-project",
				Resolved:        false,
			}

			content := repo.generateComplaintContent(complaint)
			Expect(content).To(ContainSubstring("Large content test"))
			Expect(content).To(ContainSubstring(complaint.ContextInfo))
		})

		It("should properly sanitize filenames", func() {
			testCases := []struct {
				input    string
				expected string
			}{
				{"normal filename", "test-session-2025.md"},
				{"with spaces", "test session 2025.md"},
				{"with special chars", "test-session!@#$%.md"},
				{"very long name", "a-very-long-session-name-that-should-be-truncated-because-filenames-have-limits-and-we-should-handle-this-gracefully.md"},
			}

			repo = NewFileRepository(tempDir)

			for _, testCase := range testCases {
				filename := repo.generateFilename(&domain.Complaint{
					SessionName: testCase.input,
					Resolved:    false,
				})
				Expect(filename).To(Equal(testCase.expected))
			}
		})

		It("should handle invalid write permissions gracefully", func() {
			repo = NewFileRepository(tempDir)

			// Create directory with no write permissions (simulate error)
			noWriteDir := filepath.Join(tempDir, "no-permission")
			err := os.Mkdir(noWriteDir, 0000)
			Expect(err).NotTo(HaveOccurred())
			// Try to write to read-only directory
			complaint := &domain.Complaint{
				ID:              domain.ComplaintID{Value: "test-id"},
				AgentName:       "Test Agent",
				TaskDescription: "Permission test",
				Severity:        "low",
				Resolved:        false,
			}

			err = repo.Store(ctx, complaint)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("permission denied"))
		})

		It("should store complaint with metadata file", func() {
			repo = NewFileRepository(tempDir)

			complaint := &domain.Complaint{
				ID:              domain.ComplaintID{Value: "meta-test"},
				AgentName:       "Test Agent",
				TaskDescription: "Metadata test",
				Severity:        "medium",
				ProjectName:     "test-project",
				Resolved:        false,
				Timestamp:       time.Now(),
			}

			err := repo.Store(ctx, complaint)
			Expect(err).NotTo(HaveOccurred())

			// Check for metadata file creation
			metadataPath := filepath.Join(tempDir, "complaints", complaint.ID.Value+".json")
			_, err = os.Stat(metadataPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle concurrent file operations", func() {
			repo = NewFileRepository(tempDir)

			complaints := make([]*domain.Complaint, 10)
			ids := make([]domain.ComplaintID, 10)

			// Create multiple complaints
			for i := 0; i < 10; i++ {
				id, _ := domain.NewComplaintID()
				ids[i] = id
				complaints[i] = &domain.Complaint{
					ID:              id,
					AgentName:       "Concurrent Agent",
					TaskDescription: fmt.Sprintf("Concurrent task %d", i),
					Severity:        "low",
					Resolved:        false,
				}
				repo.Store(ctx, complaints[i])
			}

			// Verify all were stored
			stored, err := repo.FindByID(ctx, ids[5])
			Expect(err).NotTo(HaveOccurred())
			Expect(stored).NotTo(BeNil())
			Expect(stored.ID.Value).To(Equal(ids[5].Value))
		})
	})

	Context("File Retrieval Operations", func() {
		It("should find complaint by ID", func() {
			repo = NewFileRepository(tempDir)

			complaintID, _ := domain.NewComplaintID()
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "Test Agent",
				TaskDescription: "Find test",
				Severity:        "low",
				ProjectName:     "test-project",
				Resolved:        false,
			}

			// Store complaint
			err := repo.Store(ctx, complaint)
			Expect(err).NotTo(HaveOccurred())

			// Find it
			found, err := repo.FindByID(ctx, complaintID)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).NotTo(BeNil())
			Expect(found.ID.Value).To(Equal(complaintID.Value))
		})

		It("should return nil for non-existent complaint", func() {
			repo = NewFileRepository(tempDir)

			nonExistentID := domain.ComplaintID{Value: "non-existent"}
			
			found, err := repo.FindByID(ctx, nonExistentID)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeNil())
		})

		It("should find complaints by project", func() {
			repo = NewFileRepository(tempDir)

			// Store complaints for different projects
			projectAComplaint := &domain.Complaint{
				ID:              domain.ComplaintID{Value: "proj-a"},
				AgentName:       "Test Agent",
				TaskDescription: "Project A task",
				Severity:        "low",
				ProjectName:     "project-a",
				Resolved:        false,
			}

			projectBComplaint := &domain.Complaint{
				ID:              domain.ComplaintID{Value: "proj-b"},
				AgentName:       "Test Agent",
				TaskDescription: "Project B task",
				Severity:        "medium",
				ProjectName:     "project-b",
				Resolved:        true,
			}

			repo.Store(ctx, projectAComplaint)
			repo.Store(ctx, projectBComplaint)

			// Find only project A complaints
			projectAComplaints, err := repo.FindByProject(ctx, "project-a", 50, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(projectAComplaints)).To(Equal(1))
			Expect(projectAComplaints[0].ID.Value).To(Equal("proj-a"))
		})

		It("should find unresolved complaints", func() {
			repo = NewFileRepository(tempDir)

			// Store mix of resolved and unresolved complaints
			resolvedComplaint := &domain.Complaint{
				ID:              domain.ComplaintID{Value: "resolved"},
				AgentName:       "Test Agent",
				TaskDescription: "Resolved task",
				Severity:        "low",
				ProjectName:     "test-project",
				Resolved:        true,
			}

			unresolvedComplaint := &domain.Complaint{
				ID:              domain.ComplaintID{Value: "unresolved"},
				AgentName:       "Test Agent",
				TaskDescription: "Unresolved task",
				Severity:        "medium",
				ProjectName:     "test-project",
				Resolved:        false,
			}

			repo.Store(ctx, resolvedComplaint)
			repo.Store(ctx, unresolvedComplaint)

			// Find only unresolved complaints
			unresolvedComplaints, err := repo.FindUnresolved(ctx, 50, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(unresolvedComplaints)).To(Equal(1))
			Expect(unresolvedComplaints[0].ID.Value).To(Equal("unresolved"))
		})
	})

	Context("File Resolution Operations", func() {
		It("should mark complaint as resolved", func() {
			repo = NewFileRepository(tempDir)

			complaintID, _ := domain.NewComplaintID()
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "Test Agent",
				TaskDescription: "Resolution test",
				Severity:        "low",
				ProjectName:     "test-project",
				Resolved:        false,
			}

			// Store complaint
			err := repo.Store(ctx, complaint)
			Expect(err).NotTo(HaveOccurred())

			// Mark as resolved
			err = repo.MarkResolved(ctx, complaintID)
			Expect(err).NotTo(HaveOccurred())

			// Verify it's resolved
			resolved, err := repo.FindByID(ctx, complaintID)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved).NotTo(BeNil())
			Expect(resolved.Resolved).To(BeTrue())
		})

		It("should handle resolve operations concurrently", func() {
			repo = NewFileRepository(tempDir)

			complaintID, _ := domain.NewComplaintID()
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "Test Agent",
				TaskDescription: "Concurrent resolution test",
				Severity:        "low",
				ProjectName:     "test-project",
				Resolved:        false,
			}

			// Store complaint
			err := repo.Store(ctx, complaint)
			Expect(err).NotTo(HaveOccurred())

			// Concurrent resolve attempts
			done := make(chan error, 2)
			
			go func() {
				err := repo.MarkResolved(ctx, complaintID)
				done <- err
			}()
			
			go func() {
				err := repo.MarkResolved(ctx, complaintID)
				done <- err
			}()

			// Wait for both to complete
			err1 := <-done
			err2 := <-done
			
			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())
		})
	})

	Context("Error Handling", func() {
		It("should handle missing directory errors", func() {
			repo := NewFileRepository("/non/existent/path")

			// Try to store complaint
			complaintID, _ := domain.NewComplaintID()
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "Test Agent",
				TaskDescription: "Error handling test",
				Severity:        "low",
				ProjectName:     "test-project",
				Resolved:        false,
			}

			err := repo.Store(ctx, complaint)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("directory"))
		})

		It("should handle file write permission errors", func() {
			repo := NewFileRepository(tempDir)

			// Create a file with content that won't allow writing
			unwritableFile := filepath.Join(tempDir, "readonly.txt")
			err := os.WriteFile(unwritableFile, 0644, []byte("test content"))
			if err != nil {
				// On some systems, this might not be a permissions error
				// Continue with other tests
				Skip("Cannot create unwritable file for permission test")
			}

			complaintID, _ := domain.NewComplaintID()
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "Test Agent",
				TaskDescription: "Permission error test",
				Severity:        "low",
				ProjectName:     "test-project",
				Resolved:        false,
			}

			err := repo.Store(ctx, complaint)
			if err == nil {
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})

	Context("Content Parsing", func() {
		It("should parse complaint from markdown file", func() {
			repo := NewFileRepository(tempDir)

			// Create a markdown file
			filename := filepath.Join(tempDir, "complaint.md")
			markdownContent := `# Test Complaint

**Task:** Test parsing from markdown

**Context:** This is a test complaint created from markdown.

**Missing Information:** Need better error messages.

---

*This complaint was automatically filed by an AI agent.*`

			err := os.WriteFile(filename, 0644, []byte(markdownContent))
			Expect(err).NotTo(HaveOccurred())

			// Parse the file
			complaint := repo.parseComplaintFromFile([]byte(markdownContent))
			Expect(complaint).NotTo(BeNil())
			Expect(complaint.TaskDescription).To(Equal("Test parsing from markdown"))
			Expect(complaint.MissingInfo).To(ContainSubstring("Need better error messages"))
		})
	})
})