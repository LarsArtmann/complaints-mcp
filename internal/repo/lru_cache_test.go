package repo

import (
	"testing"
	"time"

	"github.com/larsartmann/complaints-mcp/internal/domain"
)

// mustNewComplaintID is a test helper that panics on error
func mustNewComplaintID() domain.ComplaintID {
	id, err := domain.NewComplaintID()
	if err != nil {
		panic(err)
	}
	return id
}

func TestLRUCache_BasicOperations(t *testing.T) {
	cache := NewLRUCache(3)

	// Test Put and Get
	c1 := &domain.Complaint{
		ID:              mustNewComplaintID(),
		AgentName:       "Agent1",
		TaskDescription: "Task1",
		Timestamp:       time.Now(),
	}
	cache.Put("c1", c1)

	got, exists := cache.Get("c1")
	if !exists {
		t.Fatal("Expected complaint to exist in cache")
	}
	if got.AgentName != "Agent1" {
		t.Errorf("Expected AgentName=Agent1, got %s", got.AgentName)
	}

	// Test cache miss
	_, exists = cache.Get("nonexistent")
	if exists {
		t.Error("Expected cache miss for nonexistent key")
	}

	stats := cache.GetStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
}

func TestLRUCache_EvictionOnMaxSize(t *testing.T) {
	cache := NewLRUCache(3) // Max size = 3

	c1 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent1", TaskDescription: "Task1", Timestamp: time.Now()}
	c2 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent2", TaskDescription: "Task2", Timestamp: time.Now()}
	c3 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent3", TaskDescription: "Task3", Timestamp: time.Now()}
	c4 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent4", TaskDescription: "Task4", Timestamp: time.Now()}

	// Add 3 items (at max capacity)
	cache.Put("c1", c1)
	cache.Put("c2", c2)
	cache.Put("c3", c3)

	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Len())
	}

	// Verify all 3 exist
	if _, exists := cache.Get("c1"); !exists {
		t.Error("c1 should exist")
	}
	if _, exists := cache.Get("c2"); !exists {
		t.Error("c2 should exist")
	}
	if _, exists := cache.Get("c3"); !exists {
		t.Error("c3 should exist")
	}

	// Add 4th item - should evict c1 (least recently used)
	cache.Put("c4", c4)

	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3 after eviction, got %d", cache.Len())
	}

	// c1 should be evicted (it was least recently used)
	if _, exists := cache.Get("c1"); exists {
		t.Error("c1 should have been evicted")
	}

	// c2, c3, c4 should still exist
	if _, exists := cache.Get("c2"); !exists {
		t.Error("c2 should still exist")
	}
	if _, exists := cache.Get("c3"); !exists {
		t.Error("c3 should still exist")
	}
	if _, exists := cache.Get("c4"); !exists {
		t.Error("c4 should exist")
	}

	// Check eviction metrics
	stats := cache.GetStats()
	if stats.Evictions != 1 {
		t.Errorf("Expected 1 eviction, got %d", stats.Evictions)
	}
}

func TestLRUCache_LRUOrdering(t *testing.T) {
	cache := NewLRUCache(3)

	c1 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent1", TaskDescription: "Task1", Timestamp: time.Now()}
	c2 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent2", TaskDescription: "Task2", Timestamp: time.Now()}
	c3 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent3", TaskDescription: "Task3", Timestamp: time.Now()}
	c4 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent4", TaskDescription: "Task4", Timestamp: time.Now()}

	// Add c1, c2, c3
	cache.Put("c1", c1)
	cache.Put("c2", c2)
	cache.Put("c3", c3)

	// Access c1 (moves it to front - most recently used)
	cache.Get("c1")

	// Now order is: c1 (most recent), c3, c2 (least recent)
	// Add c4 - should evict c2
	cache.Put("c4", c4)

	// c2 should be evicted
	if _, exists := cache.Get("c2"); exists {
		t.Error("c2 should have been evicted (was least recently used)")
	}

	// c1, c3, c4 should exist
	if _, exists := cache.Get("c1"); !exists {
		t.Error("c1 should exist (was accessed recently)")
	}
	if _, exists := cache.Get("c3"); !exists {
		t.Error("c3 should exist")
	}
	if _, exists := cache.Get("c4"); !exists {
		t.Error("c4 should exist")
	}
}

func TestLRUCache_Update(t *testing.T) {
	cache := NewLRUCache(3)

	c1 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent1", TaskDescription: "Task1", Timestamp: time.Now()}

	cache.Put("c1", c1)

	// Update c1
	c1Updated := &domain.Complaint{ID: c1.ID, AgentName: "Agent1Updated", TaskDescription: "Task1Updated", Timestamp: time.Now()}
	cache.Put("c1", c1Updated)

	got, exists := cache.Get("c1")
	if !exists {
		t.Fatal("c1 should exist")
	}
	if got.AgentName != "Agent1Updated" {
		t.Errorf("Expected AgentName=Agent1Updated, got %s", got.AgentName)
	}

	// Should still be size 1 (update, not add)
	if cache.Len() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Len())
	}
}

func TestLRUCache_Delete(t *testing.T) {
	cache := NewLRUCache(3)

	c1 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent1", TaskDescription: "Task1", Timestamp: time.Now()}
	cache.Put("c1", c1)

	if cache.Len() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Len())
	}

	cache.Delete("c1")

	if cache.Len() != 0 {
		t.Errorf("Expected cache size 0 after delete, got %d", cache.Len())
	}

	if _, exists := cache.Get("c1"); exists {
		t.Error("c1 should not exist after delete")
	}
}

func TestLRUCache_GetAll(t *testing.T) {
	cache := NewLRUCache(10)

	c1 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent1", TaskDescription: "Task1", Timestamp: time.Now()}
	c2 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent2", TaskDescription: "Task2", Timestamp: time.Now()}
	c3 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent3", TaskDescription: "Task3", Timestamp: time.Now()}

	cache.Put("c1", c1)
	cache.Put("c2", c2)
	cache.Put("c3", c3)

	all := cache.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 complaints, got %d", len(all))
	}

	// Verify all complaints are present (order doesn't matter for GetAll)
	found := make(map[string]bool)
	for _, c := range all {
		found[c.AgentName] = true
	}

	if !found["Agent1"] || !found["Agent2"] || !found["Agent3"] {
		t.Error("Not all complaints found in GetAll()")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(10)

	c1 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent1", TaskDescription: "Task1", Timestamp: time.Now()}
	c2 := &domain.Complaint{ID: mustNewComplaintID(), AgentName: "Agent2", TaskDescription: "Task2", Timestamp: time.Now()}

	cache.Put("c1", c1)
	cache.Put("c2", c2)

	if cache.Len() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.Len())
	}

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Len())
	}

	if _, exists := cache.Get("c1"); exists {
		t.Error("c1 should not exist after clear")
	}
	if _, exists := cache.Get("c2"); exists {
		t.Error("c2 should not exist after clear")
	}
}

func TestLRUCache_ConcurrentAccess(t *testing.T) {
	cache := NewLRUCache(100)

	// Test concurrent puts and gets
	done := make(chan bool)

	// Writer goroutines
	for i := range 10 {
		go func(id int) {
			for range 100 {
				c := &domain.Complaint{
					ID:              mustNewComplaintID(),
					AgentName:       "Agent",
					TaskDescription: "Task",
					Timestamp:       time.Now(),
				}
				cache.Put(c.ID.String(), c)
			}
			done <- true
		}(i)
	}

	// Reader goroutines
	for range 10 {
		go func() {
			for range 100 {
				cache.GetAll()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 20 {
		<-done
	}

	// If we get here without deadlock or race conditions, test passes
	t.Log("Concurrent access test passed")
}
