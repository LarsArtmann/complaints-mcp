package types

import (
	"fmt"
)

// CacheSize is a type-safe cache size that prevents invalid values
type CacheSize uint32 // Max 4,294,967,295 - more than enough

const (
	MinCacheSize     CacheSize = 1
	MaxCacheSize     CacheSize = 100000
	DefaultCacheSize CacheSize = 1000
)

func NewCacheSize(size uint32) (CacheSize, error) {
	if size < uint32(MinCacheSize) {
		return MinCacheSize, fmt.Errorf("cache size must be >= %d", MinCacheSize)
	}
	if size > uint32(MaxCacheSize) {
		return MaxCacheSize, fmt.Errorf("cache size must be <= %d", MaxCacheSize)
	}
	return CacheSize(size), nil
}

func (cs CacheSize) Int() int {
	return int(cs)
}

func (cs CacheSize) Uint32() uint32 {
	return uint32(cs)
}

// CacheEvictionPolicy is a type-safe eviction policy
type CacheEvictionPolicy string

const (
	EvictionLRU  CacheEvictionPolicy = "lru"
	EvictionFIFO CacheEvictionPolicy = "fifo"
	EvictionNone CacheEvictionPolicy = "none"
)

var ValidEvictionPolicies = []CacheEvictionPolicy{
	EvictionLRU,
	EvictionFIFO,
	EvictionNone,
}

func NewEvictionPolicy(policy string) (CacheEvictionPolicy, error) {
	p := CacheEvictionPolicy(policy)
	for _, valid := range ValidEvictionPolicies {
		if p == valid {
			return p, nil
		}
	}
	return EvictionLRU, fmt.Errorf("invalid eviction policy: %s", policy)
}

func (cep CacheEvictionPolicy) String() string {
	return string(cep)
}

func (cep CacheEvictionPolicy) IsValid() bool {
	for _, valid := range ValidEvictionPolicies {
		if cep == valid {
			return true
		}
	}
	return false
}