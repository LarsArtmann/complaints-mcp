package types

import (
	"fmt"
)

// CacheSize provides type-safe cache size that prevents invalid states
type CacheSize uint32

const (
	MinCacheSize     CacheSize = 1
	MaxCacheSize     CacheSize = 100000
	DefaultCacheSize CacheSize = 1000
)

// NewCacheSize creates a validated cache size
func NewCacheSize(size uint32) (CacheSize, error) {
	if size < uint32(MinCacheSize) {
		return MinCacheSize, fmt.Errorf("cache size must be >= %d", MinCacheSize)
	}
	if size > uint32(MaxCacheSize) {
		return MaxCacheSize, fmt.Errorf("cache size must be <= %d", MaxCacheSize)
	}
	return CacheSize(size), nil
}

// MustNewCacheSize creates a cache size or panics (for constants)
func MustNewCacheSize(size uint32) CacheSize {
	cs, err := NewCacheSize(size)
	if err != nil {
		panic(err) // Used for compile-time constants only
	}
	return cs
}

// Int returns the cache size as int for compatibility
func (cs CacheSize) Int() int {
	return int(cs)
}

// Uint32 returns the cache size as uint32
func (cs CacheSize) Uint32() uint32 {
	return uint32(cs)
}

// CacheEvictionPolicy provides type-safe eviction policy
type CacheEvictionPolicy string

const (
	EvictionLRU  CacheEvictionPolicy = "lru"
	EvictionFIFO CacheEvictionPolicy = "fifo"
	EvictionNone CacheEvictionPolicy = "none"
)

// NewEvictionPolicy creates a validated eviction policy
func NewEvictionPolicy(policy string) (CacheEvictionPolicy, error) {
	if policy == "" {
		return EvictionLRU, nil // Default to LRU
	}
	p := CacheEvictionPolicy(policy)
	switch p {
	case EvictionLRU, EvictionFIFO, EvictionNone:
		return p, nil
	default:
		return EvictionLRU, fmt.Errorf("invalid eviction policy: %s (must be lru, fifo, or none)", policy)
	}
}

// String returns the eviction policy as string
func (cep CacheEvictionPolicy) String() string {
	return string(cep)
}

// IsValid returns true if the eviction policy is valid
func (cep CacheEvictionPolicy) IsValid() bool {
	switch cep {
	case EvictionLRU, EvictionFIFO, EvictionNone:
		return true
	default:
		return false
	}
}