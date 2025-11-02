package repo

import (
	"container/list"
	"sync"

	"github.com/larsartmann/complaints-mcp/internal/domain"
)

// LRUCache implements a thread-safe Least Recently Used cache with O(1) operations
type LRUCache struct {
	maxSize  int
	mu       sync.RWMutex
	items    map[string]*list.Element // key -> list element (contains cacheEntry)
	lruList  *list.List               // doubly-linked list for LRU tracking
	metrics  *CacheMetrics
}

// cacheEntry represents a single cache entry with key and value
type cacheEntry struct {
	key   string
	value *domain.Complaint
}

// NewLRUCache creates a new LRU cache with the specified maximum size
func NewLRUCache(maxSize int) *LRUCache {
	return &LRUCache{
		maxSize:  maxSize,
		items:    make(map[string]*list.Element),
		lruList:  list.New(),
		metrics:  NewCacheMetrics(int64(maxSize)),
	}
}

// Get retrieves a value from the cache and marks it as recently used
func (c *LRUCache) Get(key string) (*domain.Complaint, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.items[key]
	if !exists {
		c.metrics.RecordMiss()
		return nil, false
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(element)
	entry := element.Value.(*cacheEntry)
	c.metrics.RecordHit()
	return entry.value, true
}

// Put adds or updates a value in the cache
func (c *LRUCache) Put(key string, value *domain.Complaint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if element, exists := c.items[key]; exists {
		// Update existing entry and move to front
		c.lruList.MoveToFront(element)
		entry := element.Value.(*cacheEntry)
		entry.value = value
		return
	}

	// Add new entry
	entry := &cacheEntry{
		key:   key,
		value: value,
	}
	element := c.lruList.PushFront(entry)
	c.items[key] = element
	c.metrics.IncrementSize()

	// Evict least recently used if cache is full
	if c.lruList.Len() > c.maxSize {
		c.evictLRU()
	}
}

// evictLRU removes the least recently used item from the cache
// Must be called with lock held
func (c *LRUCache) evictLRU() {
	if c.lruList.Len() == 0 {
		return
	}

	// Get least recently used (back of list)
	element := c.lruList.Back()
	if element != nil {
		entry := element.Value.(*cacheEntry)
		delete(c.items, entry.key)
		c.lruList.Remove(element)
		c.metrics.RecordEviction()
		c.metrics.DecrementSize()
	}
}

// Delete removes a key from the cache
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, exists := c.items[key]; exists {
		c.lruList.Remove(element)
		delete(c.items, key)
		c.metrics.DecrementSize()
	}
}

// GetAll returns all values in the cache (for iteration)
// Values are returned in arbitrary order (map iteration order)
func (c *LRUCache) GetAll() []*domain.Complaint {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values := make([]*domain.Complaint, 0, len(c.items))
	for _, element := range c.items {
		entry := element.Value.(*cacheEntry)
		values = append(values, entry.value)
	}
	return values
}

// Len returns the current number of items in the cache
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lruList.Len()
}

// GetStats returns current cache statistics
func (c *LRUCache) GetStats() CacheStats {
	return c.metrics.GetStats()
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.lruList.Init()
	// Note: We don't reset metrics as they're useful for monitoring over time
}
