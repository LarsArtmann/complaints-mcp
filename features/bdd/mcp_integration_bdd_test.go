package bdd_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/config"
	mcpdelivery "github.com/larsartmann/complaints-mcp/internal/delivery/mcp"
	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
	"github.com/spf13/cobra"
)

var _ = Describe("MCP Integration BDD Tests", func() {
	var (
		tempDir          string
		repository       repo.Repository
		complaintService *service.ComplaintService
		mcpServer        *mcpdelivery.MCPServer
		logger           *log.Logger
		tracer           tracing.Tracer
		testConfig       *config.Config
		cmd              *cobra.Command
	)

	BeforeEach(func() {
		// Create a temporary directory for each test
		tempDir = GinkgoT().TempDir()
		logger = log.New(os.Stdout)
		tracer = tracing.NewMockTracer("test")

		// Initialize repository and service
		repository = repo.NewFileRepository(tempDir, tracer)
		complaintService = service.NewComplaintService(repository, tracer)

		// Initialize MCP server
		mcpServer = mcpdelivery.NewServer("test-server", "1.0.0", complaintService, logger, tracer)

		// Create test configuration
		testConfig = &config.Config{
			Server: config.ServerConfig{
				Name: "test-server",
				Host: "localhost",
				Port: uint16(8080),
			},
			Storage: config.StorageConfig{
				BaseDir:    tempDir,
				GlobalDir:  tempDir,
				MaxSize:    uint64(10485760), // 10MB
				Retention:  uint(30),
				AutoBackup: true,
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
				Output: "stdout",
			},
		}

		// Set config for MCP server
		mcpServer.SetConfig(testConfig)

		// Create mock command for testing
		cmd = &cobra.Command{}
		cmd.PersistentFlags().String("config", "", "config file path")
		cmd.PersistentFlags().String("log-level", "info", "log level")
		cmd.PersistentFlags().Bool("dev", false, "development mode")
	})

	AfterEach(func() {
		// Clean up temporary directory
		os.RemoveAll(tempDir)
	})

	Context("MCP tool registration", func() {
		It("should initialize MCP server without errors", func() {
			// Verify that MCP server was created successfully
			Expect(mcpServer).NotTo(BeNil())
			Expect(testConfig.Server.Name).To(Equal("test-server"))
		})

		It("should have valid configuration", func() {
			// Verify configuration is properly set
			Expect(testConfig.Server.Host).To(Equal("localhost"))
			Expect(testConfig.Server.Port).To(Equal(uint16(8080)))
			Expect(testConfig.Storage.BaseDir).To(Equal(tempDir))
			Expect(testConfig.Log.Level).To(Equal("info"))
		})
	})

	Context("End-to-end complaint workflow", func() {
		It("should handle complete complaint lifecycle", func(ctx SpecContext) {
			// Step 1: Create a complaint
			complaint, err := complaintService.CreateComplaint(ctx,
				"Test AI Agent",
				"e2e-session",
				"End-to-end test complaint",
				"Testing complete workflow",
				"No issues found",
				"Clear documentation",
				"Better examples",
				domain.SeverityMedium,
				"e2e-test-project")
			Expect(err).NotTo(HaveOccurred())
			Expect(complaint).NotTo(BeNil())
			Expect(complaint.IsResolved()).To(BeFalse())

			// Step 2: Retrieve the complaint
			retrieved, err := complaintService.GetComplaint(ctx, complaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved).NotTo(BeNil())
			Expect(retrieved.ID.String()).To(Equal(complaint.ID.String()))

			// Step 3: List complaints
			complaints, err := complaintService.ListComplaints(ctx, 10, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(complaints)).To(BeNumerically(">=", 1))

			// Step 4: Search for the complaint
			searchResults, err := complaintService.SearchComplaints(ctx, "End-to-end", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(searchResults)).To(BeNumerically(">=", 1))

			// Step 5: Resolve the complaint
			resolvedComplaint, err := complaintService.ResolveComplaint(ctx, complaint.ID, "test-agent")
			Expect(err).NotTo(HaveOccurred())
			Expect(resolvedComplaint).NotTo(BeNil())

			// Step 6: Verify resolution
			resolved, err := complaintService.GetComplaint(ctx, complaint.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved.IsResolved()).To(BeTrue())

			// Step 7: List unresolved complaints (should be empty)
			unresolved, err := complaintService.ListUnresolvedComplaints(ctx, 10)
			Expect(err).NotTo(HaveOccurred())
			found := false
			for _, c := range unresolved {
				if c.ID.String() == complaint.ID.String() {
					found = true
					break
				}
			}
			Expect(found).To(BeFalse())
		})
	})

	Context("MCP server configuration", func() {
		It("should handle configuration changes", func() {
			// Create new configuration
			newConfig := &config.Config{
				Server: config.ServerConfig{
					Name: "updated-server",
					Host: "127.0.0.1",
					Port: uint16(9090),
				},
				Storage: config.StorageConfig{
					BaseDir:    tempDir,
					GlobalDir:  tempDir,
					MaxSize:    uint64(20971520), // 20MB
					Retention:  uint(60),
					AutoBackup: false,
				},
				Log: config.LogConfig{
					Level:  "debug",
					Format: "json",
					Output: "stderr",
				},
			}

			// Update server configuration
			mcpServer.SetConfig(newConfig)

			// Configuration should be updated (no direct way to verify without internals)
			// This test mainly ensures the SetConfig method doesn't panic
			Expect(true).To(BeTrue())
		})
	})

	Context("Error handling", func() {
		It("should handle repository errors gracefully", func(ctx SpecContext) {
			// Try to get non-existent complaint
			nonExistentID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())

			_, err = complaintService.GetComplaint(ctx, nonExistentID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("complaint not found"))
		})

		It("should handle invalid complaint creation", func(ctx SpecContext) {
			// Try to create complaint with invalid data
			_, err := complaintService.CreateComplaint(ctx,
				"", // empty agent name (invalid)
				"test-session",
				"Test complaint",
				"",
				"",
				"",
				"",
				domain.SeverityLow,
				"test-project")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Performance and scalability", func() {
		It("should handle multiple complaints efficiently", func(ctx SpecContext) {
			// Create multiple complaints
			const numComplaints = 10
			complaintIDs := []domain.ComplaintID{}

			for i := range numComplaints {
				complaint, err := complaintService.CreateComplaint(ctx,
					fmt.Sprintf("Test Agent %d", i),
					fmt.Sprintf("perf-session-%d", i),
					fmt.Sprintf("Performance test complaint %d", i),
					"Performance testing content",
					"",
					"",
					"",
					domain.SeverityLow,
					"perf-test")
				Expect(err).NotTo(HaveOccurred())
				complaintIDs = append(complaintIDs, complaint.ID)
			}

			// Verify all complaints were created
			for _, id := range complaintIDs {
				complaint, err := complaintService.GetComplaint(ctx, id)
				Expect(err).NotTo(HaveOccurred())
				Expect(complaint).NotTo(BeNil())
			}

			// List all complaints
			allComplaints, err := complaintService.ListComplaints(ctx, 100, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(allComplaints)).To(BeNumerically(">=", numComplaints))

			// Search complaints
			searchResults, err := complaintService.SearchComplaints(ctx, "Performance", 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(searchResults)).To(BeNumerically(">=", numComplaints))
		})
	})
})
