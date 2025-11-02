package repo

import (
	"sync/atomic"
)

// CacheMetrics tracks cache performance statistics
type CacheMetrics struct {
	hits         int64 // Number of cache hits
	misses       int64 // Number of cache misses
	evictions    int64 // Number of cache evictions
	currentSize  int64 // Current number of entries in cache
	maxSize      int64 // Maximum allowed cache size
	totalLookups int64 // Total number of lookup operations
}

// NewCacheMetrics creates a new cache metrics tracker
func NewCacheMetrics(maxSize int64) *CacheMetrics {
	return &CacheMetrics{
		maxSize: maxSize,
	}
}

// RecordHit records a cache hit
func (m *CacheMetrics) RecordHit() {
	atomic.AddInt64(&m.hits, 1)
	atomic.AddInt64(&m.totalLookups, 1)
}

// RecordMiss records a cache miss
func (m *CacheMetrics) RecordMiss() {
	atomic.AddInt64(&m.misses, 1)
	atomic.AddInt64(&m.totalLookups, 1)
}

// RecordEviction records a cache eviction
func (m *CacheMetrics) RecordEviction() {
	atomic.AddInt64(&m.evictions, 1)
}

// IncrementSize increments current cache size
func (m *CacheMetrics) IncrementSize() {
	atomic.AddInt64(&m.currentSize, 1)
}

// DecrementSize decrements current cache size
func (m *CacheMetrics) DecrementSize() {
	atomic.AddInt64(&m.currentSize, -1)
}

// GetHitRate returns cache hit rate as a percentage
func (m *CacheMetrics) GetHitRate() float64 {
	total := atomic.LoadInt64(&m.totalLookups)
	if total == 0 {
		return 0.0
	}
	hits := atomic.LoadInt64(&m.hits)
	return float64(hits) / float64(total) * 100.0
}

// GetStats returns a snapshot of current cache statistics
func (m *CacheMetrics) GetStats() CacheStats {
	return CacheStats{
		Hits:        atomic.LoadInt64(&m.hits),
		Misses:      atomic.LoadInt64(&m.misses),
		Evictions:   atomic.LoadInt64(&m.evictions),
		CurrentSize: atomic.LoadInt64(&m.currentSize),
		MaxSize:     atomic.LoadInt64(&m.maxSize),
		HitRate:     m.GetHitRate(),
	}
}

// CacheStats represents a snapshot of cache statistics
type CacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Evictions   int64   `json:"evictions"`
	CurrentSize int64   `json:"current_size"`
	MaxSize     int64   `json:"max_size"`
	HitRate     float64 `json:"hit_rate_percent"`
}
