package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCacheSize(t *testing.T) {
	tests := []struct {
		name        string
		size        uint32
		expected    CacheSize
		expectError bool
	}{
		{
			name:        "valid minimum size",
			size:        1,
			expected:    1,
			expectError: false,
		},
		{
			name:        "valid medium size",
			size:        1000,
			expected:    1000,
			expectError: false,
		},
		{
			name:        "valid maximum size",
			size:        100000,
			expected:    100000,
			expectError: false,
		},
		{
			name:        "too small",
			size:        0,
			expected:    MinCacheSize,
			expectError: true,
		},
		{
			name:        "too large",
			size:        100001,
			expected:    MaxCacheSize,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewCacheSize(tt.size)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, result) // Should return fallback value
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMustNewCacheSize(t *testing.T) {
	assert.NotPanics(t, func() {
		result := MustNewCacheSize(1000)
		assert.Equal(t, CacheSize(1000), result)
	})

	assert.Panics(t, func() {
		MustNewCacheSize(0) // Invalid size should panic
	})
}

func TestCacheSizeMethods(t *testing.T) {
	cs := CacheSize(1000)

	assert.Equal(t, 1000, cs.Int())
	assert.Equal(t, uint32(1000), cs.Uint32())
}

func TestNewEvictionPolicy(t *testing.T) {
	tests := []struct {
		name        string
		policy      string
		expected    CacheEvictionPolicy
		expectError bool
	}{
		{
			name:        "valid LRU",
			policy:      "lru",
			expected:    EvictionLRU,
			expectError: false,
		},
		{
			name:        "valid FIFO",
			policy:      "fifo",
			expected:    EvictionFIFO,
			expectError: false,
		},
		{
			name:        "valid none",
			policy:      "none",
			expected:    EvictionNone,
			expectError: false,
		},
		{
			name:        "empty defaults to LRU",
			policy:      "",
			expected:    EvictionLRU,
			expectError: false,
		},
		{
			name:        "invalid policy",
			policy:      "invalid",
			expected:    EvictionLRU, // Fallback
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewEvictionPolicy(tt.policy)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, result) // Should return fallback value
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCacheEvictionPolicyMethods(t *testing.T) {
	tests := []struct {
		name     string
		policy   CacheEvictionPolicy
		expected string
		valid    bool
	}{
		{
			name:     "LRU policy",
			policy:   EvictionLRU,
			expected: "lru",
			valid:    true,
		},
		{
			name:     "FIFO policy",
			policy:   EvictionFIFO,
			expected: "fifo",
			valid:    true,
		},
		{
			name:     "None policy",
			policy:   EvictionNone,
			expected: "none",
			valid:    true,
		},
		{
			name:     "invalid policy",
			policy:   CacheEvictionPolicy("invalid"),
			expected: "invalid",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.policy.String())
			assert.Equal(t, tt.valid, tt.policy.IsValid())
		})
	}
}
