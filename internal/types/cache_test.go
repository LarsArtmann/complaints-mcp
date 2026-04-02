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
	name        string
	constructor func() (T, error)
	expected    T
	expectError bool
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.constructor()
			assertConstructorResult(t, tt.expectError, tt.expected, result, err)
		})
	}
}

func TestNewCacheSize(t *testing.T) {
	tests := []struct {
		name        string
		constructor func() (CacheSize, error)
		expected    CacheSize
		expectError bool
	}{
		{
			name:        "valid minimum size",
			constructor: func() (CacheSize, error) { return NewCacheSize(1) },
			expected:    1,
			expectError: false,
		},
		{
			name:        "valid medium size",
			constructor: func() (CacheSize, error) { return NewCacheSize(1000) },
			expected:    1000,
			expectError: false,
		},
		{
			name:        "valid maximum size",
			constructor: func() (CacheSize, error) { return NewCacheSize(100000) },
			expected:    100000,
			expectError: false,
		},
		{
			name:        "too small",
			constructor: func() (CacheSize, error) { return NewCacheSize(0) },
			expected:    MinCacheSize,
			expectError: true,
		},
		{
			name:        "too large",
			constructor: func() (CacheSize, error) { return NewCacheSize(100001) },
			expected:    MaxCacheSize,
			expectError: true,
		},
	}

	runConstructorTests(t, tests)
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
		name        string
		constructor func() (CacheEvictionPolicy, error)
		expected    CacheEvictionPolicy
		expectError bool
	}{
		{
			name:        "valid LRU",
			constructor: func() (CacheEvictionPolicy, error) { return NewEvictionPolicy("lru") },
			expected:    EvictionLRU,
			expectError: false,
		},
		{
			name:        "valid FIFO",
			constructor: func() (CacheEvictionPolicy, error) { return NewEvictionPolicy("fifo") },
			expected:    EvictionFIFO,
			expectError: false,
		},
		{
			name:        "valid none",
			constructor: func() (CacheEvictionPolicy, error) { return NewEvictionPolicy("none") },
			expected:    EvictionNone,
			expectError: false,
		},
		{
			name:        "empty defaults to LRU",
			constructor: func() (CacheEvictionPolicy, error) { return NewEvictionPolicy("") },
			expected:    EvictionLRU,
			expectError: false,
		},
		{
			name:        "invalid policy",
			constructor: func() (CacheEvictionPolicy, error) { return NewEvictionPolicy("invalid") },
			expected:    EvictionLRU,
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
