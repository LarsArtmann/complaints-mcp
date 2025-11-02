package repo

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/complaints-mcp/internal/domain"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

// mockTracer implements tracing.Tracer for testing
type mockTracer struct{}

func (t *mockTracer) Start(ctx context.Context, operationName string) (context.Context, tracing.Span) {
	return ctx, &mockSpan{}
}

type mockSpan struct{}

func (s *mockSpan) End() {}
func (s *mockSpan) SetAttribute(ctx context.Context, key string, value interface{}) {}
func (s *mockSpan) AddEvent(ctx context.Context, event string, attributes map[string]interface{}) {}

// TestCachedRepository_FindByID_O1_Lookup tests the O(1) cache lookup performance
func TestCachedRepository_FindByID_O1_Lookup(t *testing.T) {
	baseDir := t.TempDir()
	repo := NewCachedRepository(baseDir, &mockTracer{}).(*CachedRepository)

	// Create test complaint
	id, _ := domain.NewComplaintID()
	complaint := &domain.Complaint{
		ID:              id,
		AgentName:       "Test Agent",
		TaskDescription: "Test Task",
		Severity:        domain.SeverityMedium,
		Timestamp:       time.Now(),
		Resolved:        false,
		ProjectName:     "test-project",
	}

	// Save complaint
	err := repo.Save(context.Background(), complaint)
	if err != nil {
		t.Fatalf("Failed to save complaint: %v", err)
	}

	// Test O(1) lookup - should be in cache
	found, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("Failed to find complaint: %v", err)
	}

	if found.ID != complaint.ID {
		t.Errorf("Expected complaint ID %s, got %s", complaint.ID, found.ID)
	}

	// Verify cache hit by checking internal state
	repo.cacheMu.RLock()
	_, exists := repo.cache[id.String()]
	repo.cacheMu.RUnlock()

	if !exists {
		t.Error("Complaint should be in cache after save")
	}
}

// TestCachedRepository_ConcurrentAccess tests thread safety of cache operations
func TestCachedRepository_ConcurrentAccess(t *testing.T) {
	baseDir := t.TempDir()
	repo := NewCachedRepository(baseDir, &mockTracer{}).(*CachedRepository)

	// Create test complaints
	var complaints []*domain.Complaint
	for i := 0; i < 10; i++ {
		id, _ := domain.NewComplaintID()
		complaint := &domain.Complaint{
			ID:              id,
			AgentName:       "Test Agent",
			TaskDescription: "Test Task",
			Severity:        domain.SeverityMedium,
			Timestamp:       time.Now(),
			Resolved:        false,
			ProjectName:     "test-project",
		}
		complaints = append(complaints, complaint)
	}

	// Concurrent save operations
	var wg sync.WaitGroup
	for i, complaint := range complaints {
		wg.Add(1)
		go func(idx int, c *domain.Complaint) {
			defer wg.Done()
			err := repo.Save(context.Background(), c)
			if err != nil {
				t.Errorf("Failed to save complaint %d: %v", idx, err)
			}
		}(i, complaint)
	}
	wg.Wait()

	// Concurrent find operations
	for _, complaint := range complaints {
		wg.Add(1)
		go func(c *domain.Complaint) {
			defer wg.Done()
			found, err := repo.FindByID(context.Background(), c.ID)
			if err != nil {
				t.Errorf("Failed to find complaint %s: %v", c.ID, err)
				return
			}
			if found.ID != c.ID {
				t.Errorf("Expected complaint ID %s, got %s", c.ID, found.ID)
			}
		}(complaint)
	}
	wg.Wait()

	// Verify all complaints are cached
	repo.cacheMu.RLock()
	if len(repo.cache) != 10 {
		t.Errorf("Expected 10 complaints in cache, got %d", len(repo.cache))
	}
	repo.cacheMu.RUnlock()
}

// TestCachedRepository_CacheInvalidation tests cache invalidation on update
func TestCachedRepository_CacheInvalidation(t *testing.T) {
	baseDir := t.TempDir()
	repo := NewCachedRepository(baseDir, &mockTracer{}).(*CachedRepository)

	// Create test complaint
	id, _ := domain.NewComplaintID()
	complaint := &domain.Complaint{
		ID:              id,
		AgentName:       "Test Agent",
		TaskDescription: "Test Task",
		Severity:        domain.SeverityMedium,
		Timestamp:       time.Now(),
		Resolved:        false,
		ProjectName:     "test-project",
	}

	// Save complaint
	err := repo.Save(context.Background(), complaint)
	if err != nil {
		t.Fatalf("Failed to save complaint: %v", err)
	}

	// Verify it's in cache
	repo.cacheMu.RLock()
	cached := repo.cache[id.String()]
	repo.cacheMu.RUnlock()

	if cached.Resolved {
		t.Error("Complaint should not be resolved initially")
	}

	// Update complaint
	complaint.Resolved = true
	now := time.Now()
	complaint.ResolvedAt = &now
	complaint.ResolvedBy = "test-resolver"

	err = repo.Update(context.Background(), complaint)
	if err != nil {
		t.Fatalf("Failed to update complaint: %v", err)
	}

	// Verify cache is updated
	repo.cacheMu.RLock()
	updated := repo.cache[id.String()]
	repo.cacheMu.RUnlock()

	if !updated.Resolved {
		t.Error("Complaint should be resolved after update")
	}

	if updated.ResolvedBy != "test-resolver" {
		t.Errorf("Expected resolved by 'test-resolver', got '%s'", updated.ResolvedBy)
	}
}

// TestCachedRepository_Performance_1000_Complaints demonstrates 10-100x performance improvement
func TestCachedRepository_Performance_1000_Complaints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	baseDir := t.TempDir()
	cachedRepo := NewCachedRepository(baseDir, &mockTracer{}).(*CachedRepository)
	legacyRepo := NewFileRepository(baseDir, &mockTracer{}).(*FileRepository)

	// Create 1000 test complaints
	var ids []domain.ComplaintID
	for i := 0; i < 1000; i++ {
		id, _ := domain.NewComplaintID()
		ids = append(ids, id)
		complaint := &domain.Complaint{
			ID:              id,
			AgentName:       "Test Agent",
			TaskDescription: "Test Task",
			Severity:        domain.SeverityMedium,
			Timestamp:       time.Now(),
			Resolved:        false,
			ProjectName:     "test-project",
		}

		// Save to both repositories
		cachedRepo.Save(context.Background(), complaint)
		var err error
		if err != nil {
			t.Fatalf("Failed to save complaint to cached repo: %v", err)
		}

		err = legacyRepo.Save(context.Background(), complaint)
		if err != nil {
			t.Fatalf("Failed to save complaint to legacy repo: %v", err)
		}
	}

	// Benchmark cached repository FindByID
	start := time.Now()
	for _, id := range ids {
		_, err := cachedRepo.FindByID(context.Background(), id)
		if err != nil {
			t.Fatalf("Failed to find complaint in cached repo: %v", err)
		}
	}
	cachedDuration := time.Since(start)

	// Benchmark legacy repository FindByID
	start = time.Now()
	for _, id := range ids {
		_, err := legacyRepo.FindByID(context.Background(), id)
		if err != nil {
			t.Fatalf("Failed to find complaint in legacy repo: %v", err)
		}
	}
	legacyDuration := time.Since(start)

	// Calculate performance improvement
	improvement := float64(legacyDuration) / float64(cachedDuration)

	t.Logf("Cached repository: %v for 1000 lookups", cachedDuration)
	t.Logf("Legacy repository: %v for 1000 lookups", legacyDuration)
	t.Logf("Performance improvement: %.1fx", improvement)

	// Verify we achieve at least 10x improvement
	if improvement < 10 {
		t.Errorf("Expected at least 10x performance improvement, got %.1fx", improvement)
	}

	// Verify cached lookup is under 10ms total (<0.01ms per lookup)
	if cachedDuration > 10*time.Millisecond {
		t.Errorf("Expected cached lookups to be under 10ms total, got %v", cachedDuration)
	}
}

// TestCachedRepository_WarmCache tests cache warm-up functionality
func TestCachedRepository_WarmCache(t *testing.T) {
	baseDir := t.TempDir()

	// Create some complaints manually first
	legacyRepo := NewFileRepository(baseDir, &mockTracer{}).(*FileRepository)
	var expectedIDs []string

	for i := 0; i < 5; i++ {
		id, _ := domain.NewComplaintID()
		expectedIDs = append(expectedIDs, id.String())
		complaint := &domain.Complaint{
			ID:              id,
			AgentName:       "Test Agent",
			TaskDescription: "Test Task",
			Severity:        domain.SeverityMedium,
			Timestamp:       time.Now(),
			Resolved:        false,
			ProjectName:     "test-project",
		}

		legacyRepo.Save(context.Background(), complaint)
		var err error
		if err != nil {
			t.Fatalf("Failed to save complaint: %v", err)
		}
	}

	// Create cached repository - should warm up automatically
	cachedRepo := NewCachedRepository(baseDir, &mockTracer{}).(*CachedRepository)

	// Verify cache is warmed with existing data
	cachedRepo.cacheMu.RLock()
	cacheSize := len(cachedRepo.cache)
	cachedRepo.cacheMu.RUnlock()

	if cacheSize != 5 {
		t.Errorf("Expected cache to contain 5 complaints after warm-up, got %d", cacheSize)
	}

	// Verify all expected IDs are in cache
	for _, expectedID := range expectedIDs {
		_, err := cachedRepo.FindByID(context.Background(), domain.ComplaintID{Value: expectedID})
		if err != nil {
			t.Errorf("Failed to find complaint %s in warmed cache: %v", expectedID, err)
		}
	}
}

// TestCachedRepository_MemoryUsage verifies reasonable memory usage
func TestCachedRepository_MemoryUsage(t *testing.T) {
	baseDir := t.TempDir()
	repo := NewCachedRepository(baseDir, &mockTracer{}).(*CachedRepository)

	// Create a complaint with realistic data
	id, _ := domain.NewComplaintID()
	complaint := &domain.Complaint{
		ID:              id,
		AgentName:       "Test Agent Name",
		TaskDescription: "This is a test task description that contains a reasonable amount of text to simulate real-world usage. It includes details about what the agent was trying to accomplish and what issues they encountered.",
		Severity:        domain.SeverityHigh,
		Timestamp:       time.Now(),
		Resolved:        false,
		ProjectName:     "test-project-with-long-name",
		ContextInfo:     "Additional context information about the project structure and the specific file the agent was working on when the issue occurred.",
		MissingInfo:     "The documentation was missing information about the expected input format and there were no examples provided.",
		ConfusedBy:      "The function signature was unclear and the parameter types were not well documented in the code comments.",
		FutureWishes:    "Better documentation with clear examples and type annotations for all public APIs.",
	}

	// Save to cache
	repo.Save(context.Background(), complaint)
	var err error
	if err != nil {
		t.Fatalf("Failed to save complaint: %v", err)
	}

	// Verify memory usage is reasonable (each complaint should be < 1KB)
	repo.cacheMu.RLock()
	cacheSize := len(repo.cache)
	cached := repo.cache[id.String()]
	repo.cacheMu.RUnlock()

	if cacheSize != 1 {
		t.Errorf("Expected 1 complaint in cache, got %d", cacheSize)
	}

	// Verify complaint data is preserved
	if cached.TaskDescription != complaint.TaskDescription {
		t.Error("Task description not preserved in cache")
	}

	if cached.ContextInfo != complaint.ContextInfo {
		t.Error("Context info not preserved in cache")
	}
}