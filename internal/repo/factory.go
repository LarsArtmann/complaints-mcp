package repo

import (
	"fmt"

	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
	"github.com/larsartmann/complaints-mcp/internal/types"
)

// RepositoryConfig holds configuration for repository creation
type RepositoryConfig struct {
	Type          string
	BaseDir       string
	StorageConfig config.StorageConfig
}

// Normalize populates type-safe fields from storage config
func (cfg *RepositoryConfig) Normalize() error {
	// Populate type-safe cache configuration
	cacheSize, err := types.NewCacheSize(uint32(cfg.StorageConfig.CacheMaxSize))
	if err != nil {
		return fmt.Errorf("invalid cache size: %w", err)
	}
	cfg.StorageConfig.CacheSize = cacheSize
	
	evictionPolicy, err := types.NewEvictionPolicy(cfg.StorageConfig.CacheEviction)
	if err != nil {
		return fmt.Errorf("invalid eviction policy: %w", err)
	}
	cfg.StorageConfig.EvictionPolicy = evictionPolicy
	
	return nil
}

// NewRepository creates a repository based on configuration
func NewRepository(cfg RepositoryConfig) Repository {
	// Determine if caching should be enabled
	cacheEnabled := cfg.StorageConfig.CacheEnabled
	if cfg.Type == "file" {
		cacheEnabled = false // Force legacy for explicit type
	}

	// Create tracer
	tracerConfig := tracing.DefaultTracerConfig()
	tracer := tracing.NewTracer(tracerConfig)

	if cacheEnabled {
		// Use type-safe cache size from StorageConfig
		cacheSize := cfg.StorageConfig.CacheSize
		return NewCachedRepository(cfg.BaseDir, cacheSize.Int(), tracer)
	}

	// Default to file repository
	return NewFileRepository(cfg.BaseDir, tracer)
}

// NewRepositoryFromConfig creates repository from full config
func NewRepositoryFromConfig(cfg *config.Config) Repository {
	repoConfig := RepositoryConfig{
		BaseDir:       cfg.Storage.BaseDir,
		StorageConfig: cfg.Storage,
		Type:          "cached", // Default to cached
	}

	// Use cache disabled if explicitly set
	if !cfg.Storage.CacheEnabled {
		repoConfig.Type = "file"
	}

	// Normalize configuration (populates type-safe fields)
	if err := repoConfig.Normalize(); err != nil {
		// This shouldn't happen if config was properly loaded
		// Fall back to file repository with default cache size
		repoConfig.Type = "file"
		repoConfig.StorageConfig.CacheSize = types.DefaultCacheSize
		repoConfig.StorageConfig.EvictionPolicy = types.EvictionLRU
	}

	return NewRepository(repoConfig)
}
