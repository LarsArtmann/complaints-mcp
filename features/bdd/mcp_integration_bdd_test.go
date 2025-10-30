package bdd_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/complaints-mcp/cmd/server"
)

var _ = Describe("MCP Integration BDD Tests", func() {
	var (
		serverCmd *server.ServerCommand
		ctx    context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		serverCmd = &server.ServerCommand{}
	})

	Context("MCP tool registration", func() {
		It("should register all required tools", func() {
			// Verify that the server has all three MCP tools
			tools := serverCmd.GetTools()
			Expect(len(tools)).To(Equal(3))

			// Check for required tools
			toolNames := make(map[string]bool)
			for _, tool := range tools {
				toolNames[tool.Name] = true
			}

			Expect(toolNames["file_complaint"]).To(BeTrue())
			Expect(toolNames["list_complaints"]).To(BeTrue())
			Expect(toolNames["resolve_complaint"]).To(BeTrue())
		})

		It("should provide proper tool descriptions", func() {
			tools := serverCmd.GetTools()

			for _, tool := range tools {
				switch tool.Name {
				case "file_complaint":
					Expect(tool.Description).To(ContainSubstring("File a complaint report"))
				case "list_complaints":
					Expect(tool.Description).To(ContainSubstring("List existing complaints"))
				case "resolve_complaint":
					Expect(tool.Description).To(ContainSubstring("Mark a complaint as resolved"))
				}
			}
		})
	})

	Context("MCP server lifecycle", func() {
		It("should initialize and shutdown gracefully", func() {
			// Test server startup
			serverCmd := &server.ServerCommand{}
			
			// This should not panic
			Expect(func() {
				serverCmd.Start(ctx)
			}).NotTo(Panic())

			// Test shutdown
			Expect(func() {
				serverCmd.Shutdown(ctx)
			}).NotTo(Panic())
		})
	})

	Context("Concurrent request handling", func() {
		It("should handle multiple simultaneous requests", func() {
			serverCmd := &server.ServerCommand{}
			
			// Simulate concurrent requests
			done := make(chan bool, 2)
			
			// Start two concurrent operations
			go func() {
				req := map[string]interface{}{
					"task_description": "Concurrent task 1",
					"severity": "low",
				}
				
				response, err := serverCmd.FileComplaint(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				done[0] <- true
			}()
			
			go func() {
				req := map[string]interface{}{
					"task_description": "Concurrent task 2",
					"severity": "medium",
				}
				
				response, err := serverCmd.FileComplaint(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				done[1] <- true
			}()

			// Wait for both to complete
			Eventually(func() bool {
				select {
				case <-done[0]:
					return true
				case <-done[1]:
					return true
				}
			}).WithContext(ctx).Should(BeTrue())
		})
	})

	Context("Error handling in MCP operations", func() {
		It("should return structured error responses", func() {
			serverCmd := &server.ServerCommand{}
			
			// Test with invalid request
			invalidReq := map[string]interface{}{
				"missing_field": "This field is intentionally missing",
			}

			response, err := serverCmd.FileComplaint(ctx, invalidReq)
			Expect(err).To(HaveOccurred())
			Expect(response).To(BeNil())

			// Parse error response
			var errorResponse map[string]interface{}
			err = json.Unmarshal([]byte(err.Error()), &errorResponse)
			Expect(err).NotTo(HaveOccurred())
			Expect(errorResponse["error"]).To(ContainSubstring("missing_field"))
		})

		It("should handle malformed JSON requests gracefully", func() {
			serverCmd := &server.ServerCommand{}
			
			// Test with malformed JSON
			malformedJSON := `{"task": "test" "malformed}`

			response, err := serverCmd.FileComplaint(ctx, malformedJSON)
			Expect(err).To(HaveOccurred())
			Expect(response).To(BeNil())
		})
	})

	Context("Request validation", func() {
		It("should validate request parameters", func() {
			serverCmd := &server.ServerCommand{}
			
			// Test request validation
			testCases := []struct {
				name        string
				request     map[string]interface{}
				shouldError  bool
			}{
				{
					name: "empty task description",
					request: map[string]interface{}{
						"task_description": "",
						"severity": "low",
					},
					shouldError: true,
				},
				{
					name: "invalid severity",
					request: map[string]interface{}{
						"task_description": "Test",
						"severity": "invalid_severity",
					},
					shouldError: true,
				},
				{
					name: "negative content size test",
					request: map[string]interface{}{
						"task_description": string(make([]byte, 1000000)), // Too large
						"severity": "medium",
					},
					shouldError: false, // Service should handle this gracefully
				},
			}

			for _, testCase := range testCases {
				response, err := serverCmd.FileComplaint(ctx, testCase.request)
				
				if testCase.shouldError {
					Expect(err).To(HaveOccurred(), "Expected validation error for: "+testCase.name)
					Expect(response).To(BeNil(), "Expected nil response for: "+testCase.name)
				} else {
					Expect(err).NotTo(HaveOccurred(), "Did not expect error for: "+testCase.name)
					Expect(response).NotTo(BeNil(), "Expected response for: "+testCase.name)
				}
			}
		})
	})

	Context("Response formatting", func() {
		It("should return properly formatted JSON responses", func() {
			serverCmd := &server.ServerCommand{}
			
			validReq := map[string]interface{}{
				"task_description": "Test response formatting",
				"severity": "medium",
				"agent_name": "Test Agent",
				"project_name": "format-test",
			}

			response, err := serverCmd.FileComplaint(ctx, validReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())

			// Verify JSON structure
			var responseMap map[string]interface{}
			err = json.Unmarshal([]byte(response.Content), &responseMap)
			Expect(err).NotTo(HaveOccurred())
			
			// Check required fields
			Expect(responseMap["content"]).To(ContainSubstring("successfully"))
			Expect(responseMap["task_description"]).To(Equal("Test response formatting"))
			Expect(responseMap["severity"]).To(Equal("medium"))
			Expect(responseMap["agent_name"]).To(Equal("Test Agent"))
			Expect(responseMap["project_name"]).To(Equal("format-test"))
		})
	})

	Context("Memory and resource management", func() {
		It("should handle large volumes of requests", func() {
			serverCmd := &server.ServerCommand{}
			
			// Test memory efficiency with many requests
			for i := 0; i < 100; i++ {
				req := map[string]interface{}{
					"task_description": fmt.Sprintf("Memory test request %d", i),
					"severity": "low",
					"agent_name": "Memory Test Agent",
					"project_name": "memory-test",
				}

				response, err := serverCmd.FileComplaint(ctx, req)
				Expect(err).NotTo(HaveOccurred(), "Request %d should not fail", i)
				Expect(response).NotTo(BeNil(), "Request %d should return response", i)
			}
		})

		It("should clean up resources properly", func() {
			serverCmd := &server.ServerCommand{}
			
			// Test resource cleanup
			for i := 0; i < 10; i++ {
				req := map[string]interface{}{
					"task_description": fmt.Sprintf("Cleanup test %d", i),
					"severity": "low",
					"agent_name": "Cleanup Test Agent",
					"project_name": "cleanup-test",
				}

				response, err := serverCmd.FileComplaint(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			}

			// Server should still be responsive
			finalReq := map[string]interface{}{
				"task_description": "Final request",
				"severity": "low",
				"agent_name": "Final Test Agent",
				"project_name": "cleanup-test",
			}

			finalResponse, err := serverCmd.FileComplaint(ctx, finalReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(finalResponse).NotTo(BeNil())
		})
	})
})