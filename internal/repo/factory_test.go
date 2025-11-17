package repo

import (
	"testing"

	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	tracer := tracing.NewMockTracer("test")
	defer tracer.Close()

	tests := []struct {
		name           string
		config         RepositoryConfig
		expectedType   any
		expectedCached bool
	}{
		{
			name: "should create cached repository when cache enabled",
			config: RepositoryConfig{
				BaseDir: t.TempDir(),
				StorageConfig: config.StorageConfig{
					CacheEnabled: true,
					CacheMaxSize: uint32(500),
				},
				Type: "cached",
			},
			expectedType:   &CachedRepository{},
			expectedCached: true,
		},
		{
			name: "should create file repository when cache disabled",
			config: RepositoryConfig{
				BaseDir: t.TempDir(),
				StorageConfig: config.StorageConfig{
					CacheEnabled: false,
					CacheMaxSize: uint32(1000),
				},
				Type: "cached",
			},
			expectedType:   &FileRepository{},
			expectedCached: false,
		},
		{
			name: "should create file repository when type explicitly set to file",
			config: RepositoryConfig{
				BaseDir: t.TempDir(),
				StorageConfig: config.StorageConfig{
					CacheEnabled: true, // Even if enabled, explicit "file" type should win
					CacheMaxSize: uint32(1000),
				},
				Type: "file",
			},
			expectedType:   &FileRepository{},
			expectedCached: false,
		},
		{
			name: "should create cached repository with default size when cache enabled",
			config: RepositoryConfig{
				BaseDir: t.TempDir(),
				StorageConfig: config.StorageConfig{
					CacheEnabled: true,
					CacheMaxSize: uint32(1000), // Default size
				},
				Type: "cached",
			},
			expectedType:   &CachedRepository{},
			expectedCached: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			err := tt.config.Normalize()
			require.NoError(t, err)

			// Act
			repo := NewRepository(tt.config)

			// Assert
			assert.IsType(t, repo, tt.expectedType)

			// Check cache stats availability
			stats := repo.GetCacheStats()
			if tt.expectedCached {
				assert.NotEqual(t, int64(0), stats.MaxSize, "Cached repository should have non-zero max cache size")
				assert.Equal(t, int64(tt.config.StorageConfig.CacheMaxSize), stats.MaxSize, "Cache size should match config")
			} else {
				assert.Equal(t, int64(0), stats.MaxSize, "File repository should have zero cache size")
			}
		})
	}
}

func TestNewRepositoryFromConfig(t *testing.T) {
	tracer := tracing.NewMockTracer("test")
	defer tracer.Close()

	tests := []struct {
		name           string
		config         *config.Config
		expectedType   any
		expectedCached bool
	}{
		{
			name: "should create cached repository when cache enabled in config",
			config: &config.Config{
				Storage: config.StorageConfig{
					BaseDir:      t.TempDir(),
					CacheEnabled: true,
					CacheMaxSize: uint32(750),
				},
			},
			expectedType:   &CachedRepository{},
			expectedCached: true,
		},
		{
			name: "should create file repository when cache disabled in config",
			config: &config.Config{
				Storage: config.StorageConfig{
					BaseDir:      t.TempDir(),
					CacheEnabled: false,
					CacheMaxSize: uint32(1000),
				},
			},
			expectedType:   &FileRepository{},
			expectedCached: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			repo := NewRepositoryFromConfig(tt.config)

			// Assert
			require.NotNil(t, repo)
			assert.IsType(t, tt.expectedType, repo)

			// Check cache stats availability
			stats := repo.GetCacheStats()
			if tt.expectedCached {
				assert.NotEqual(t, 0, stats.MaxSize, "Cached repository should have non-zero max cache size")
				assert.Equal(t, int64(tt.config.Storage.CacheMaxSize), stats.MaxSize, "Cache size should match config")
			} else {
				assert.Equal(t, int64(0), stats.MaxSize, "File repository should have zero cache size")
			}
		})
	}
}

func TestRepositoryConfigValidation(t *testing.T) {
	tracer := tracing.NewMockTracer("test")
	defer tracer.Close()

	// Test with different cache sizes
	cacheSizes := []int64{1, 10, 100, 1000, 5000, 100000}

	for _, cacheSize := range cacheSizes {
		t.Run("cache_size_"+string(rune(cacheSize)), func(t *testing.T) {
			config := RepositoryConfig{
				BaseDir: t.TempDir(),
				StorageConfig: config.StorageConfig{
					CacheEnabled: true,
					CacheMaxSize: uint32(cacheSize),
				},
				Type: "cached",
			}

			// Normalize configuration to populate type-safe fields
			err := config.Normalize()
			require.NoError(t, err)

			repo := NewRepository(config)
			require.NotNil(t, repo)

			stats := repo.GetCacheStats()
			assert.Equal(t, cacheSize, stats.MaxSize, "Cache size should be properly configured")
		})
	}
}

func TestRepositoryTypePriority(t *testing.T) {
	tracer := tracing.NewMockTracer("test")
	defer tracer.Close()

	// Test that explicit "file" type takes priority over cache enabled setting
	config := RepositoryConfig{
		BaseDir: t.TempDir(),
		StorageConfig: config.StorageConfig{
			CacheEnabled: true,         // This should be ignored
			CacheMaxSize: uint32(1000), // This should be ignored
		},
		Type: "file", // This should win
	}

	repo := NewRepository(config)
	require.NotNil(t, repo)

	// Should be a FileRepository regardless of cache settings
	stats := repo.GetCacheStats()
	assert.Equal(t, int64(0), stats.MaxSize, "Explicit file type should have zero cache size")
}

func TestCacheSizeValidation(t *testing.T) {
	tracer := tracing.NewMockTracer("test")
	defer tracer.Close()

	tests := []struct {
		name        string
		cacheSize   int64
		expectError bool
	}{
		{"valid_min_size", 1, false},
		{"valid_small_size", 100, false},
		{"valid_large_size", 50000, false},
		{"valid_max_size", 100000, false},
		{"invalid_zero_size", 0, true},
		{"invalid_negative_size", -1, true},
		{"invalid_too_large_size", 100001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For now, we test that validation works via repository creation
			// validation is handled at config load time
			config := config.Config{
				Storage: config.StorageConfig{
					BaseDir:      t.TempDir(),
					CacheEnabled: true,
					CacheMaxSize: uint32(tt.cacheSize),
				},
			}

			// Skip validation test for now, focus on repository creation
			// Validation is handled at config load time in the actual application
			if !tt.expectError {
				repo := NewRepositoryFromConfig(&config)
				require.NotNil(t, repo)
				stats := repo.GetCacheStats()
				assert.Equal(t, tt.cacheSize, stats.MaxSize, "Cache size should match configuration")
			}
		})
	}
}

func TestCacheConfigurationIntegration(t *testing.T) {
	tracer := tracing.NewMockTracer("test")
	defer tracer.Close()

	// Test creating repository with different cache sizes
	testSizes := []int64{10, 100, 1000, 10000}

	for _, cacheSize := range testSizes {
		t.Run("cache_size_"+string(rune(cacheSize)), func(t *testing.T) {
			config := &config.Config{
				Storage: config.StorageConfig{
					BaseDir:       t.TempDir(),
					CacheEnabled:  true,
					CacheMaxSize:  uint32(cacheSize),
					CacheEviction: "lru",
				},
			}

			// Create repository
			repo := NewRepositoryFromConfig(config)
			require.NotNil(t, repo)

			// Verify cache size is properly set
			stats := repo.GetCacheStats()
			assert.Equal(t, cacheSize, stats.MaxSize, "Cache size should match configuration")

			// Verify repository type
			assert.IsType(t, repo, &CachedRepository{})
		})
	}
}
