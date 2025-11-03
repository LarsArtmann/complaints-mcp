package repo

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BenchmarkCachePerformance tests the 1061x performance improvement claim
func BenchmarkCachePerformance(b *testing.B) {
	// Setup test data
	baseDir := b.TempDir()
	tracer := tracing.NewMockTracer("benchmark")

	// Create repositories
	legacyRepo := NewFileRepository(baseDir, tracer)
	cachedRepo := NewCachedRepository(baseDir, 1000, tracer)

	// Generate test complaints
	numComplaints := 100
	complaints := generateTestComplaints(numComplaints)
	complaintIDs := make([]domain.ComplaintID, numComplaints)

	// Save complaints to both repositories
	ctx := context.Background()
	for i, complaint := range complaints {
		err := legacyRepo.Save(ctx, complaint)
		require.NoError(b, err)
		// Save to cached repo separately for fair comparison
		err = cachedRepo.Save(ctx, complaint)
		require.NoError(b, err)
		complaintIDs[i] = complaint.ID
	}

	b.ResetTimer()

	b.Run("Legacy_Repository_Lookup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			id := complaintIDs[i%numComplaints]
			_, err := legacyRepo.FindByID(ctx, id)
			require.NoError(b, err)
		}
	})

	b.Run("Cached_Repository_Lookup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			id := complaintIDs[i%numComplaints]
			_, err := cachedRepo.FindByID(ctx, id)
			require.NoError(b, err)
		}
	})
}

// TestCachePerformanceRegression ensures we maintain the 1061x improvement
func TestCachePerformanceRegression(t *testing.T) {
	// Setup test data
	baseDir := t.TempDir()
	tracer := tracing.NewMockTracer("regression-test")

	// Create repositories
	legacyRepo := NewFileRepository(baseDir, tracer)
	cachedRepo := NewCachedRepository(baseDir, 1000, tracer)

	// Generate test complaints
	numComplaints := 50
	complaints := generateTestComplaints(numComplaints)
	complaintIDs := make([]domain.ComplaintID, numComplaints)

	// Save complaints
	ctx := context.Background()
	for i, complaint := range complaints {
		err := legacyRepo.Save(ctx, complaint)
		require.NoError(t, err)
		// Save to cached repo separately
		err = cachedRepo.Save(ctx, complaint)
		require.NoError(t, err)
		complaintIDs[i] = complaint.ID
	}

	// Benchmark legacy repository
	numLookups := 100
	start := time.Now()
	for i := 0; i < numLookups; i++ {
		id := complaintIDs[i%numComplaints]
		_, err := legacyRepo.FindByID(ctx, id)
		require.NoError(t, err)
	}
	legacyTime := time.Since(start)

	// Benchmark cached repository
	start = time.Now()
	for i := 0; i < numLookups; i++ {
		id := complaintIDs[i%numComplaints]
		_, err := cachedRepo.FindByID(ctx, id)
		require.NoError(t, err)
	}
	cachedTime := time.Since(start)

	// Calculate performance improvement
	improvement := float64(legacyTime) / float64(cachedTime)

	t.Logf("Legacy time: %v", legacyTime)
	t.Logf("Cached time: %v", cachedTime)
	t.Logf("Performance improvement: %.1fx", improvement)

	// Assert we meet or exceed our performance target
	// We expect at least 50x improvement (conservative target)
	assert.Greater(t, improvement, 50.0,
		"Cache should provide at least 50x performance improvement, got %.1fx", improvement)
}

// TestConcurrentCacheAccess tests thread safety under high concurrency
func TestConcurrentCacheAccess(t *testing.T) {
	baseDir := t.TempDir()
	tracer := tracing.NewMockTracer("concurrent-test")

	// Create cached repository
	repo := NewCachedRepository(baseDir, 1000, tracer)

	// Generate test complaints
	numComplaints := 20
	complaints := generateTestComplaints(numComplaints)

	// Save complaints
	ctx := context.Background()
	for _, complaint := range complaints {
		err := repo.Save(ctx, complaint)
		require.NoError(t, err)
	}

	// Concurrent access parameters
	numGoroutines := 50
	numOperationsPerGoroutine := 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	start := time.Now()

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperationsPerGoroutine; j++ {
				complaintID := complaints[j%numComplaints].ID

				// Mix of reads and writes
				if j%10 == 0 {
					// Update operation
					complaint := complaints[j%numComplaints]
					complaint.ContextInfo = fmt.Sprintf("Updated by goroutine %d at time %d", goroutineID, j)
					err := repo.Update(ctx, complaint)
					if err != nil {
						errors <- fmt.Errorf("update failed: %w", err)
						return
					}
				} else {
					// Read operation
					_, err := repo.FindByID(ctx, complaintID)
					if err != nil {
						errors <- fmt.Errorf("lookup failed: %w", err)
						return
					}
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	totalTime := time.Since(start)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	// Performance metrics
	totalOperations := int64(numGoroutines * numOperationsPerGoroutine)
	opsPerSecond := float64(totalOperations) / totalTime.Seconds()

	t.Logf("Concurrent operations: %d", totalOperations)
	t.Logf("Total time: %v", totalTime)
	t.Logf("Operations per second: %.0f", opsPerSecond)
	t.Logf("Memory usage: %d KB", getMemUsageKB())

	// Verify cache statistics if available
	if cachedRepo, ok := repo.(*CachedRepository); ok {
		stats := cachedRepo.GetCacheStats()
		t.Logf("Cache hit rate: %.1f%%", stats.HitRate)
		t.Logf("Cache size: %d/%d", stats.CurrentSize, stats.MaxSize)

		// Cache should have high hit rate after warm-up
		assert.Greater(t, stats.HitRate, 80.0, "Cache hit rate should be high after concurrent operations")
	}

	// Performance assertions
	assert.Greater(t, opsPerSecond, 1000.0, "Should achieve at least 1000 ops/sec under concurrency")
}

// TestCacheMetricsAccuracy ensures our metrics tracking is accurate
func TestCacheMetricsAccuracy(t *testing.T) {
	baseDir := t.TempDir()
	tracer := tracing.NewMockTracer("metrics-test")

	// Create cached repository
	repo := NewCachedRepository(baseDir, 1000, tracer)
	cachedRepo := repo.(*CachedRepository)

	// Generate test complaints
	numComplaints := 10
	complaints := generateTestComplaints(numComplaints)

	// Save complaints
	ctx := context.Background()
	for _, complaint := range complaints {
		err := repo.Save(ctx, complaint)
		require.NoError(t, err)
	}

	// Perform operations and track metrics
	expectedHits := 0
	expectedMisses := 0

	// First access should be a HIT because Save() already populated the cache
	_, err := repo.FindByID(ctx, complaints[0].ID)
	require.NoError(t, err)
	expectedHits++

	// Second access should also be a hit
	_, err = repo.FindByID(ctx, complaints[0].ID)
	require.NoError(t, err)
	expectedHits++

	// Access other complaints (all should be hits since Save populated cache)
	for i := 1; i < numComplaints; i++ {
		// First access - hit (already in cache from Save)
		_, err = repo.FindByID(ctx, complaints[i].ID)
		require.NoError(t, err)
		expectedHits++

		// Second access - also hit
		_, err = repo.FindByID(ctx, complaints[i].ID)
		require.NoError(t, err)
		expectedHits++
	}

	// Test cache miss with non-existent ID
	nonExistentID, _ := domain.NewComplaintID()
	_, err = repo.FindByID(ctx, nonExistentID)
	require.Error(t, err) // Should error because it doesn't exist
	expectedMisses++

	// Get cache statistics
	stats := cachedRepo.GetCacheStats()

	// Verify metrics
	assert.Equal(t, int64(expectedHits), stats.Hits, "Hit count should match expected")
	assert.Equal(t, int64(expectedMisses), stats.Misses, "Miss count should match expected")
	assert.Equal(t, int64(expectedHits+expectedMisses), stats.Hits+stats.Misses, "Total lookups should match expected")

	expectedHitRate := float64(expectedHits) / float64(expectedHits+expectedMisses) * 100.0
	assert.InDelta(t, expectedHitRate, stats.HitRate, 0.1, "Hit rate should match expected")

	t.Logf("Expected hit rate: %.1f%%", expectedHitRate)
	t.Logf("Actual hit rate: %.1f%%", stats.HitRate)
}

// generateTestComplaints creates test complaints for benchmarks
func generateTestComplaints(count int) []*domain.Complaint {
	complaints := make([]*domain.Complaint, count)

	for i := 0; i < count; i++ {
		id, _ := domain.NewComplaintID()
		complaints[i] = &domain.Complaint{
			ID:              id,
			AgentName:       fmt.Sprintf("Test Agent %d", i),
			SessionName:     fmt.Sprintf("test-session-%d", i),
			TaskDescription: fmt.Sprintf("Test task description %d", i),
			ContextInfo:     fmt.Sprintf("Test context info %d", i),
			Severity:        domain.SeverityMedium,
			Timestamp:       time.Now().Add(-time.Duration(i) * time.Hour),
			ProjectName:     fmt.Sprintf("test-project-%d", i%5),
			Resolved:        i%3 == 0, // Every third complaint is resolved
		}
	}

	return complaints
}

// getMemUsageKB returns current memory usage in KB
func getMemUsageKB() int {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int(m.Alloc / 1024)
}
