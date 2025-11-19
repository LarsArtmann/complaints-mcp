package bdd_test

import (
	"encoding/json"
	"os"

	"github.com/charmbracelet/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

var _ = Describe("Cache Statistics BDD Tests", func() {
	var (
		tempDir          string
		repository       repo.Repository
		complaintService *service.ComplaintService
		tracer           tracing.Tracer
	)

	BeforeEach(func() {
		// Create a temporary directory for each test
		tempDir = GinkgoT().TempDir()
		tracer = tracing.NewMockTracer("test")
	})

	AfterEach(func() {
		// Clean up temporary directory
		os.RemoveAll(tempDir)
	})

	Describe("get_cache_stats functionality", func() {
		Context("when using CachedRepository", func() {
			BeforeEach(func() {
				// Initialize cached repository and service
				repository = repo.NewCachedRepository(tempDir, tracer)
				complaintService = service.NewComplaintService(repository, tracer)
			})

			It("should return cache performance statistics", func(ctx SpecContext) {
				// Arrange - File some complaints to populate cache
				complaint1, err := complaintService.CreateComplaint(ctx,
					"Agent 1", "session-1", "Test cache functionality",
					"Context info", "Missing info", "Confused by", "Future wishes",
					domain.SeverityLow, "")
				Expect(err).NotTo(HaveOccurred())

				complaint2, err := complaintService.CreateComplaint(ctx,
					"Agent 2", "session-2", "Test cache hit tracking",
					"More context", "More missing", "More confusion", "More wishes",
					domain.SeverityMedium, "")
				Expect(err).NotTo(HaveOccurred())

				// Act - Get cache stats after creation
				stats := complaintService.GetCacheStats()

				// Assert - Verify cache has entries but no access yet
				Expect(stats.MaxSize).To(Equal(int64(1000)), "Cache max size should be 1000")
				Expect(stats.CurrentSize).To(Equal(int64(2)), "Cache should have 2 entries from creation")
				Expect(stats.Hits).To(Equal(int64(0)), "No hits yet - just cached from creation")
				Expect(stats.Misses).To(Equal(int64(0)), "No misses yet - just cached from creation")
				Expect(stats.HitRate).To(Equal(float64(0.0)), "Hit rate should be 0% (no accesses)")

				// Now create cache hits by retrieving complaints
				_, err = complaintService.GetComplaint(ctx, complaint1.ID)
				Expect(err).NotTo(HaveOccurred())
				_, err = complaintService.GetComplaint(ctx, complaint2.ID)
				Expect(err).NotTo(HaveOccurred())

				// Create additional hits
				_, err = complaintService.GetComplaint(ctx, complaint1.ID)
				Expect(err).NotTo(HaveOccurred())
				_, err = complaintService.GetComplaint(ctx, complaint2.ID)
				Expect(err).NotTo(HaveOccurred())

				// Get final stats
				finalStats := complaintService.GetCacheStats()

				// Assert - Verify cache hit statistics
				Expect(finalStats.CurrentSize).To(Equal(int64(2)), "Cache should still have 2 entries")
				Expect(finalStats.Hits).To(Equal(int64(4)), "Should have 4 hits from retrieving twice each")
				Expect(finalStats.Misses).To(Equal(int64(0)), "No misses since items were cached from creation")
				Expect(finalStats.HitRate).To(Equal(float64(100.0)), "Hit rate should be 100% (all accesses were hits")
			})

			It("should show cache enabled flag", func() {
				// Act
				stats := complaintService.GetCacheStats()

				// Assert
				Expect(stats.MaxSize).To(BeNumerically(">", 0), "CachedRepository should have max size > 0")
			})
		})

		Context("when using FileRepository", func() {
			BeforeEach(func() {
				// Create file repository directly (no cache)
				repository = repo.NewFileRepository(tempDir, tracer)
				complaintService = service.NewComplaintService(repository, tracer)
			})

			It("should return cache disabled statistics", func() {
				// Act
				stats := complaintService.GetCacheStats()

				// Assert
				Expect(stats.MaxSize).To(Equal(int64(0)), "FileRepository should have max size 0")
				Expect(stats.CurrentSize).To(Equal(int64(0)), "FileRepository should have current size 0")
				Expect(stats.Hits).To(Equal(int64(0)), "FileRepository should have 0 hits")
				Expect(stats.Misses).To(Equal(int64(0)), "FileRepository should have 0 misses")
				Expect(stats.Evictions).To(Equal(int64(0)), "FileRepository should have 0 evictions")
				Expect(stats.HitRate).To(Equal(float64(0.0)), "FileRepository should have 0% hit rate")
			})
		})

		Context("cache performance characteristics", func() {
			BeforeEach(func() {
				// Use cached repository for performance tests
				repository = repo.NewCachedRepository(tempDir, tracer)
				complaintService = service.NewComplaintService(repository, tracer)
			})

			It("should track statistics accurately", func(ctx SpecContext) {
				// Arrange - Create complaint first (cached during creation)
				complaint, err := complaintService.CreateComplaint(ctx,
					"Test Agent", "session", "Test tracking",
					"", "", "", "", domain.SeverityLow, "")
				Expect(err).NotTo(HaveOccurred())

				// Act - Access complaint multiple times to generate hits
				_, err = complaintService.GetComplaint(ctx, complaint.ID)
				Expect(err).NotTo(HaveOccurred())
				_, err = complaintService.GetComplaint(ctx, complaint.ID)
				Expect(err).NotTo(HaveOccurred())
				_, err = complaintService.GetComplaint(ctx, complaint.ID)
				Expect(err).NotTo(HaveOccurred())

				// Get stats
				stats := complaintService.GetCacheStats()

				// Assert - Verify tracking accuracy (all hits since cached from creation)
				Expect(stats.Hits+stats.Misses).To(Equal(int64(3)), "Should have 3 total lookups")
				Expect(stats.Hits).To(Equal(int64(3)), "Should have 3 hits (all accesses)")
				Expect(stats.Misses).To(Equal(int64(0)), "Should have 0 misses (cached from creation)")
				Expect(stats.HitRate).To(Equal(float64(100.0)), "Hit rate should be 100%")
			})
		})
	})

	Describe("Cache statistics JSON serialization", func() {
		Context("when stats are returned", func() {
			BeforeEach(func() {
				repository = repo.NewCachedRepository(tempDir, tracer)
				complaintService = service.NewComplaintService(repository, tracer)
			})

			It("should serialize correctly to JSON", func() {
				// Act
				stats := complaintService.GetCacheStats()

				// Verify JSON serialization works correctly
				jsonData, err := json.Marshal(stats)
				Expect(err).NotTo(HaveOccurred())

				var unmarshaledStats repo.CacheStats
				err = json.Unmarshal(jsonData, &unmarshaledStats)
				Expect(err).NotTo(HaveOccurred())

				// Verify all fields are present and reasonable
				Expect(unmarshaledStats.Hits).To(BeNumerically(">=", 0))
				Expect(unmarshaledStats.Misses).To(BeNumerically(">=", 0))
				Expect(unmarshaledStats.Evictions).To(BeNumerically(">=", 0))
				Expect(unmarshaledStats.CurrentSize).To(BeNumerically(">=", 0))
				Expect(unmarshaledStats.MaxSize).To(BeNumerically(">=", 0))
				Expect(unmarshaledStats.HitRate).To(BeNumerically(">=", 0.0))
				Expect(unmarshaledStats.HitRate).To(BeNumerically("<=", 100.0))
			})
		})
	})
})
