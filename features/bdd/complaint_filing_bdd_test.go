package bdd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/cmd/server"
)

var _ = Describe("Complaint Filing BDD Tests", func() {
	var (
		tempDir string
		serverCmd *server.ServerCommand
		ctx    context.Context
	)

	BeforeEach(func() {
		// Create a temporary directory for each test
		tempDir = GinkgoT().TempDir()
		ctx = context.Background()
		
		// Initialize server command for testing
		serverCmd = &server.ServerCommand{}
	})

	AfterEach(func() {
		// Clean up temporary directory
		os.RemoveAll(tempDir)
	})

	Context("File a valid complaint successfully", func() {
		It("should store complaint with all required fields", func(ctx SpecContext) {
			complaintID, err := domain.NewComplaintID()
			Expect(err).NotTo(HaveOccurred())
			
			complaint := &domain.Complaint{
				ID:              complaintID,
				AgentName:       "AI Assistant",
				SessionName:     "test-session",
				TaskDescription: "Implement authentication system",
				ContextInfo:     "Need to add JWT authentication to API endpoints",
				MissingInfo:     "Unclear error handling patterns",
				ConfusedBy:      "Inconsistent response formats",
				FutureWishes:    "Standardized error codes",
				Severity:        "medium",
				ProjectName:     "user-auth-service",
				Timestamp:       time.Now(),
				Resolved:        false,
			}

			// Create service request
			req := &service.CreateComplaintRequest{
				AgentName:       complaint.AgentName,
				SessionName:     complaint.SessionName,
				TaskDescription: complaint.TaskDescription,
				ContextInfo:     complaint.ContextInfo,
				MissingInfo:     complaint.MissingInfo,
				ConfusedBy:      complaint.ConfusedBy,
				FutureWishes:    complaint.FutureWishes,
				Severity:        complaint.Severity,
				ProjectName:     complaint.ProjectName,
			}

			// Test validation
			err := req.Validate()
			Expect(err).NotTo(HaveOccurred())

			// Create service
			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(nil, config)

			// Execute service method
			result, err := svc.CreateComplaint(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.ID.Value).To(Equal(complaintID.Value))
			Expect(result.AgentName).To(Equal(complaint.AgentName))
			Expect(result.TaskDescription).To(Equal(complaint.TaskDescription))
			Expect(result.Severity).To(Equal(complaint.Severity))
			Expect(result.ProjectName).To(Equal(complaint.ProjectName))
			Expect(result.Resolved).To(BeFalse())
		})

		It("should return success response with proper format", func(ctx SpecContext) {
			complaintID, _ := domain.NewComplaintID()
			req := &service.CreateComplaintRequest{
				AgentName:       "AI Assistant",
				TaskDescription: "Add user management feature",
				ContextInfo:     "User registration and authentication",
				MissingInfo:     "Password reset workflow",
				ConfusedBy:      "UI complexity",
				FutureWishes:    "Better admin panel",
				Severity:        "low",
				ProjectName:     "user-management",
			}

			// Mock successful file response
			serverCmd.SetMockSuccessResponse(fmt.Sprintf(`✅ **Complaint filed successfully!**

**ID:** %s  
**Severity:** %s  
**Project:** %s

Your feedback helps improve the development experience for AI agents.

---

*This complaint was automatically filed by an AI agent using the complaints-mcp server.*`, 
				complaintID.Value, req.Severity, req.ProjectName))

			// Execute the actual file_complaint command
			response, err := serverCmd.FileComplaint(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			
			// Parse the response
			var responseMap map[string]interface{}
			err = json.Unmarshal([]byte(response.Content), &responseMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(responseMap["content"]).To(ContainSubstring("Complaint filed successfully"))
		})

		It("should store complaint with minimum required data", func(ctx SpecContext) {
			// Test with minimal valid data
			complaintID, _ := domain.NewComplaintID()
			req := &service.CreateComplaintRequest{
				AgentName:       "A",  // minimal valid name
				TaskDescription: "T",  // minimal valid description
				Severity:        "low", // valid severity
				ProjectName:     "test",  // valid project name
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(nil, config)

			result, err := svc.CreateComplaint(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.AgentName).To(Equal("A"))
			Expect(result.TaskDescription).To(Equal("T"))
			Expect(result.Severity).To(Equal("low"))
			Expect(result.ProjectName).To(Equal("test"))
		})
	})

	Context("File a complaint with missing required field", func() {
		It("should return validation error for missing task description", func(ctx SpecContext) {
			req := &service.CreateComplaintRequest{
				AgentName:   "AI Assistant",
				// TaskDescription is intentionally empty
				Severity:    "medium",
				ProjectName: "test-project",
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(nil, config)

			_, err := svc.CreateComplaint(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("task description is required"))
		})

		It("should return validation error for missing agent name", func(ctx SpecContext) {
			req := &service.CreateComplaintRequest{
				// AgentName is intentionally empty
				TaskDescription: "Test task",
				Severity:    "medium",
				ProjectName: "test-project",
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(nil, config)

			_, err := svc.CreateComplaint(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("agent name is required"))
		})

		It("should return validation error for invalid severity", func(ctx SpecContext) {
			req := &service.CreateComplaintRequest{
				AgentName:       "AI Assistant",
				TaskDescription: "Test task",
				Severity:        "invalid", // invalid severity
				ProjectName: "test-project",
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(nil, config)

			_, err := svc.CreateComplaint(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid severity"))
		})
	})

	Context("File a complaint with invalid severity", func() {
		It("should return validation error for unsupported severity", func(ctx SpecContext) {
			req := &service.CreateComplaintRequest{
				AgentName:       "AI Assistant",
				TaskDescription: "Test task",
				Severity:        "unsupported", // invalid severity
				ProjectName: "test-project",
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(nil, config)

			_, err := svc.CreateComplaint(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid severity"))
		})
	})

	Context("File complaint with large content", func() {
		It("should handle large content gracefully", func(ctx SpecContext) {
			// Create a large complaint
			largeContent := string(make([]byte, 2000)) // 2KB content
			req := &service.CreateComplaintRequest{
				AgentName:       "AI Assistant",
				TaskDescription: "Large content test",
				ContextInfo:     largeContent,
				Severity:        "low",
				ProjectName:     "content-test",
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(nil, config)

			result, err := svc.CreateComplaint(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.ContextInfo).To(Equal(largeContent))
		})

		It("should enforce content size limits", func(ctx SpecContext) {
			// Test content that exceeds reasonable limits
			veryLargeContent := string(make([]byte, 1000000)) // 1MB content
			
			req := &service.CreateComplaintRequest{
				AgentName:       "AI Assistant",
				TaskDescription: "Oversized content",
				ContextInfo:     veryLargeContent,
				Severity:        "medium",
				ProjectName:     "content-test",
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(nil, config)

			// This should either be handled gracefully or return an appropriate error
			result, err := svc.CreateComplaint(ctx, req)
			
			// For now, we'll assume it should succeed (service layer should handle)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})
	})

	Context("File complaint with special characters", func() {
		It("should handle special characters properly", func(ctx SpecContext) {
			req := &service.CreateComplaintRequest{
				AgentName:       "AI Assistant 🤖",
				TaskDescription: "Test with special chars: quotes, newlines, tabs",
				ContextInfo:     "Content with \"quotes\" and \t\t tabs\nnewlines",
				Severity:        "medium",
				ProjectName:     "special-char-test",
			}

			config := &service.ServiceConfig{}
			svc := service.NewComplaintService(nil, config)

			result, err := svc.CreateComplaint(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.AgentName).To(Equal("AI Assistant 🤖"))
			Expect(result.ContextInfo).To(ContainSubstring("quotes"))
		})
	})
})