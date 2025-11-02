package repo

import (
	"github.com/larsartmann/complaints-mcp/internal/config"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
)

// RepositoryConfig holds configuration for repository creation
type RepositoryConfig struct {
	Type        string
	BaseDir     string
	StorageConfig config.StorageConfig
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
		return NewCachedRepository(cfg.BaseDir, tracer)
	}

	// Default to file repository
	return NewFileRepository(cfg.BaseDir, tracer)
}

// NewRepositoryFromConfig creates repository from full config
func NewRepositoryFromConfig(cfg *config.Config) Repository {
	repoConfig := RepositoryConfig{
		BaseDir:     cfg.Storage.BaseDir,
		StorageConfig: cfg.Storage,
		Type:        "cached", // Default to cached
	}

	// Use cache disabled if explicitly set
	if !cfg.Storage.CacheEnabled {
		repoConfig.Type = "file"
	}

	return NewRepository(repoConfig)
}