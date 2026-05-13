package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertConstructorResult[T any](t *testing.T, expectError bool, expected, result T, err error) {
	if expectError {
		assert.Error(t, err)
		assert.Equal(t, expected, result)
	} else {
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	}
}

func runConstructorTests[T any](t *testing.T, tests []struct {
	expected    T
	constructor func() (T, error)
	name        string
	expectError bool
},
) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.constructor()
			assertConstructorResult(t, tt.expectError, tt.expected, result, err)
		})
	}
}

func TestNewCacheSize(t *testing.T) {
	tests := []struct {
		constructor func() (CacheSize, error)
		name        string
		expected    CacheSize
		expectError bool
	}{
		{
			expected:    1,
			name:        "valid minimum size",
			constructor: func() (CacheSize, error) { return NewCacheSize(1) },
			expectError: false,
		},
		{
			expected:    1000,
			name:        "valid medium size",
			constructor: func() (CacheSize, error) { return NewCacheSize(1000) },
			expectError: false,
		},
		{
			expected:    100000,
			name:        "valid maximum size",
			constructor: func() (CacheSize, error) { return NewCacheSize(100000) },
			expectError: false,
		},
		{
			expected:    MinCacheSize,
			name:        "too small",
			constructor: func() (CacheSize, error) { return NewCacheSize(0) },
			expectError: true,
		},
		{
			expected:    MaxCacheSize,
			name:        "too large",
			constructor: func() (CacheSize, error) { return NewCacheSize(100001) },
			expectError: true,
		},
	}

	runConstructorTests[CacheSize](t, tests)
}

func TestMustNewCacheSize(t *testing.T) {
	assert.NotPanics(t, func() {
		result := MustNewCacheSize(1000)
		assert.Equal(t, CacheSize(1000), result)
	})

	assert.Panics(t, func() {
		MustNewCacheSize(0)
	})
}

func TestCacheSizeMethods(t *testing.T) {
	cs := CacheSize(1000)

	assert.Equal(t, 1000, cs.Int())
	assert.Equal(t, uint32(1000), cs.Uint32())
}

func TestNewEvictionPolicy(t *testing.T) {
	tests := []struct {
		expected    CacheEvictionPolicy
		constructor func() (CacheEvictionPolicy, error)
		name        string
		expectError bool
	}{
		{
			expected:    EvictionLRU,
			name:        "valid LRU",
			constructor: func() (CacheEvictionPolicy, error) { return NewEvictionPolicy("lru") },
			expectError: false,
		},
		{
			expected:    EvictionFIFO,
			name:        "valid FIFO",
			constructor: func() (CacheEvictionPolicy, error) { return NewEvictionPolicy("fifo") },
			expectError: false,
		},
		{
			expected:    EvictionNone,
			name:        "valid none",
			constructor: func() (CacheEvictionPolicy, error) { return NewEvictionPolicy("none") },
			expectError: false,
		},
		{
			expected:    EvictionLRU,
			name:        "empty defaults to LRU",
			constructor: func() (CacheEvictionPolicy, error) { return NewEvictionPolicy("") },
			expectError: false,
		},
		{
			expected:    EvictionLRU,
			name:        "invalid policy",
			constructor: func() (CacheEvictionPolicy, error) { return NewEvictionPolicy("invalid") },
			expectError: true,
		},
	}

	runConstructorTests(t, tests)
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
