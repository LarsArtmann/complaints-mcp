package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigFromFiles tests real file-based configuration loading
func TestConfigFromFiles(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		configFile  string
		content     string
		expectError bool
		expected    config.StorageConfig
	}{
		{
			name:       "valid cache configuration",
			configFile: "valid-cache.toml",
			content: `
[storage]
base_dir = "/tmp/complaints"
cache_enabled = true
cache_max_size = 500
cache_eviction = "lru"

[tracing]
enabled = true
jaeger_endpoint = "http://localhost:14268/api/traces"
service_name = "test-service"
`,
			expectError: false,
			expected: config.StorageConfig{
				BaseDir:       "/tmp/complaints",
				CacheEnabled:  true,
				CacheMaxSize:  500,
				CacheEviction: "lru",
			},
		},
		{
			name:       "cache disabled",
			configFile: "no-cache.toml",
			content: `
[storage]
base_dir = "/tmp/complaints"
cache_enabled = false
cache_max_size = 0
cache_eviction = "lru"

[tracing]
enabled = false
`,
			expectError: false,
			expected: config.StorageConfig{
				BaseDir:       "/tmp/complaints",
				CacheEnabled:  false,
				CacheMaxSize:  0,
				CacheEviction: "lru",
			},
		},
		{
			name:       "invalid cache size too large",
			configFile: "invalid-large-cache.toml",
			content: `
[storage]
base_dir = "/tmp/complaints"
cache_enabled = true
cache_max_size = 200000
cache_eviction = "lru"
`,
			expectError: true,
		},
		{
			name:       "invalid eviction policy",
			configFile: "invalid-eviction.toml",
			content: `
[storage]
base_dir = "/tmp/complaints"
cache_enabled = true
cache_max_size = 100
cache_eviction = "invalid_policy"
`,
			expectError: true,
		},
		{
			name:       "maximum valid cache size",
			configFile: "max-cache.toml",
			content: `
[storage]
base_dir = "/tmp/complaints"
cache_enabled = true
cache_max_size = 100000
cache_eviction = "fifo"
`,
			expectError: false,
			expected: config.StorageConfig{
				BaseDir:       "/tmp/complaints",
				CacheEnabled:  true,
				CacheMaxSize:  100000,
				CacheEviction: "fifo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write config file
			configPath := filepath.Join(tempDir, tt.configFile)
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Load configuration
			cfg, err := config.LoadFromFile(configPath)

			if tt.expectError {
				assert.Error(t, err, "Expected configuration loading to fail")
				assert.Contains(t, err.Error(), "validation failed")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.BaseDir, cfg.Storage.BaseDir)
			assert.Equal(t, tt.expected.CacheEnabled, cfg.Storage.CacheEnabled)
			assert.Equal(t, tt.expected.CacheMaxSize, cfg.Storage.CacheMaxSize)
			assert.Equal(t, tt.expected.CacheEviction, cfg.Storage.CacheEviction)

			// Verify type-safe fields are populated
			assert.NotZero(t, cfg.Storage.CacheSize, "CacheSize should be populated when cache is enabled")
			assert.NotZero(t, cfg.Storage.EvictionPolicy, "EvictionPolicy should be populated")

			if cfg.Storage.CacheEnabled {
				assert.Equal(t, uint32(tt.expected.CacheMaxSize), cfg.Storage.CacheSize.Int())
			}
		})
	}
}

// TestConfigEnvironmentVariables tests environment variable configuration
func TestConfigEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedCache  bool
		expectedSize   int64
		expectedPolicy string
		expectError    bool
	}{
		{
			name: "environment variables override",
			envVars: map[string]string{
				"COMPLAINTS_STORAGE_CACHE_ENABLED":   "true",
				"COMPLAINTS_STORAGE_CACHE_MAX_SIZE":  "2000",
				"COMPLAINTS_STORAGE_CACHE_EVICTION": "fifo",
			},
			expectedCache:  true,
			expectedSize:   2000,
			expectedPolicy: "fifo",
			expectError:    false,
		},
		{
			name: "invalid environment cache size",
			envVars: map[string]string{
				"COMPLAINTS_STORAGE_CACHE_ENABLED":   "true",
				"COMPLAINTS_STORAGE_CACHE_MAX_SIZE":  "999999",
				"COMPLAINTS_STORAGE_CACHE_EVICTION": "lru",
			},
			expectError: true,
		},
		{
			name: "invalid environment eviction policy",
			envVars: map[string]string{
				"COMPLAINTS_STORAGE_CACHE_ENABLED":   "true",
				"COMPLAINTS_STORAGE_CACHE_MAX_SIZE":  "1000",
				"COMPLAINTS_STORAGE_CACHE_EVICTION": "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				// Clean up environment variables
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			// Load configuration
			cfg, err := config.LoadFromEnvironment()

			if tt.expectError {
				assert.Error(t, err, "Expected environment configuration to fail")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCache, cfg.Storage.CacheEnabled)
			assert.Equal(t, tt.expectedSize, cfg.Storage.CacheMaxSize)
			assert.Equal(t, tt.expectedPolicy, cfg.Storage.CacheEviction)

			// Verify type-safe fields
			if tt.expectedCache {
				assert.NotZero(t, cfg.Storage.CacheSize)
				assert.NotZero(t, cfg.Storage.EvictionPolicy)
			}
		})
	}
}

// TestConfigValidationBoundaryConditions tests edge cases
func TestConfigValidationBoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		cacheSize   int64
		expectError bool
		description string
	}{
		{
			name:        "minimum valid cache size",
			cacheSize:   1,
			expectError: false,
			description: "Should accept cache size of 1",
		},
		{
			name:        "zero cache size",
			cacheSize:   0,
			expectError: true,
			description: "Should reject cache size of 0",
		},
		{
			name:        "negative cache size",
			cacheSize:   -1,
			expectError: true,
			description: "Should reject negative cache size",
		},
		{
			name:        "maximum valid cache size",
			cacheSize:   100000,
			expectError: false,
			description: "Should accept maximum cache size of 100,000",
		},
		{
			name:        "just over maximum",
			cacheSize:   100001,
			expectError: true,
			description: "Should reject cache size over 100,000",
		},
		{
			name:        "int32 overflow size",
			cacheSize:   2147483648, // 2^31
			expectError: true,
			description: "Should reject int32 overflow values",
		},
		{
			name:        "int64 overflow risk",
			cacheSize:   9223372036854775807, // max int64
			expectError: true,
			description: "Should reject values that would overflow int32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.StorageConfig{
				CacheMaxSize:  tt.cacheSize,
				CacheEnabled:  true,
				CacheEviction: "lru",
			}

			err := config.Validate()

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), "validation failed")
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestConfigTypeSafetyConsistency ensures type-safe fields stay consistent
func TestConfigTypeSafetyConsistency(t *testing.T) {
	validSizes := []int64{1, 10, 100, 1000, 10000, 100000}
	validPolicies := []string{"lru", "fifo", "none"}

	for _, size := range validSizes {
		for _, policy := range validPolicies {
			t.Run(fmt.Sprintf("size_%d_policy_%s", size, policy), func(t *testing.T) {
				config := config.StorageConfig{
					BaseDir:       "/tmp/test",
					CacheEnabled:  true,
					CacheMaxSize:  size,
					CacheEviction: policy,
				}

				// Validate configuration
				err := config.Validate()
				require.NoError(t, err)

				// Check type-safe consistency
				assert.Equal(t, uint32(size), config.CacheSize.Int(), 
					"CacheSize should match CacheMaxSize as uint32")
				
				assert.Equal(t, policy, string(config.EvictionPolicy),
					"EvictionPolicy should match CacheEviction string")
			})
		}
	}
}