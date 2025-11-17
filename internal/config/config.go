package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Storage StorageConfig `mapstructure:"storage"`
	Log     LogConfig     `mapstructure:"log"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Name string `mapstructure:"name" validate:"required"`
	Host string `mapstructure:"host"`
	Port uint16 `mapstructure:"port" validate:"min=1,max=65535"` // uint16: ports are 0-65535
}

// Address returns the full server address
func (s ServerConfig) Address() string {
	if s.Host == "" {
		return fmt.Sprintf(":%d", s.Port)
	}
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	BaseDir    string `mapstructure:"base_dir" validate:"required"`
	GlobalDir  string `mapstructure:"global_dir"`
	MaxSize    uint64 `mapstructure:"max_size" validate:"min=1024"` // uint64: file sizes cannot be negative
	Retention  uint   `mapstructure:"retention_days"` // 0 = infinite retention
	AutoBackup bool   `mapstructure:"auto_backup"`

	// Documentation storage configuration
	DocsDir     string `mapstructure:"docs_dir"`
	DocsEnabled bool   `mapstructure:"docs_enabled"`
	DocsFormat  string `mapstructure:"docs_format"` // "markdown", "html", "text"

	// Cache configuration - JSON fields for Viper
	CacheEnabled  bool   `mapstructure:"cache_enabled"`
	CacheMaxSize  uint32 `mapstructure:"cache_max_size" validate:"min=1,max=100000"` // uint32: cache sizes cannot be negative
	CacheEviction string `mapstructure:"cache_eviction"` // "lru", "fifo", "none"

	// Type-safe fields for internal use (populated in postProcessConfig)
	CacheSize      types.CacheSize           `mapstructure:"-"` // derived from CacheMaxSize
	EvictionPolicy types.CacheEvictionPolicy `mapstructure:"-"` // derived from CacheEviction
}

// LogConfig represents logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// Load loads configuration from various sources
func Load(ctx context.Context, cmd *cobra.Command) (*Config, error) {
	logger := log.FromContext(ctx)

	// Initialize viper
	v := viper.New()

	// Set configuration file path
	if configFile, _ := cmd.Flags().GetString("config"); configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// Search for config in XDG locations
		configFile, err := xdg.SearchConfigFile("complaints-mcp/config.yaml")
		if err == nil {
			v.SetConfigFile(configFile)
			logger.Info("Using XDG config file", "config_file", configFile)
		} else {
			logger.Debug("No XDG config file found, using defaults", "error", err)
		}
	}

	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.complaints-mcp")
	v.AddConfigPath("/etc/complaints-mcp")

	// Environment variables
	v.SetEnvPrefix("COMPLAINTS_MCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Bind command-line flags
	if err := v.BindPFlags(cmd.PersistentFlags()); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Read configuration
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Info("Config file not found, using defaults")
		} else {
			logger.Warn("Failed to read config file", "error", err, "config_file", v.ConfigFileUsed())
		}
	}

	// Unmarshal configuration
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Post-processing
	if err := postProcessConfig(&cfg); err != nil {
		return nil, fmt.Errorf("failed to post-process configuration: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	logger.Info("Configuration loaded successfully", "config_file", v.ConfigFileUsed())

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.name", "complaints-mcp")
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)

	// Storage defaults using XDG
	v.SetDefault("storage.base_dir", filepath.Join(xdg.DataHome, "complaints"))
	v.SetDefault("storage.global_dir", filepath.Join(xdg.DataHome, "complaints"))
	v.SetDefault("storage.max_size", 10485760)      // 10MB
	v.SetDefault("storage.retention_days", uint(0)) // 0 = infinite retention
	v.SetDefault("storage.auto_backup", true)

	// Documentation storage defaults
	v.SetDefault("storage.docs_dir", "docs/complaints") // Relative to project root
	v.SetDefault("storage.docs_enabled", true)          // Enable markdown export by default
	v.SetDefault("storage.docs_format", "markdown")     // Default to markdown format

	// Cache defaults
	v.SetDefault("storage.cache_enabled", true)
	v.SetDefault("storage.cache_max_size", 1000) // Maximum complaints to cache
	v.SetDefault("storage.cache_eviction", "lru")

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text") // text, json, logfmt
	v.SetDefault("log.output", "stdout")
}

func postProcessConfig(cfg *Config) error {
	// Expand home directory in paths
	if strings.HasPrefix(cfg.Storage.BaseDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		cfg.Storage.BaseDir = filepath.Join(home, cfg.Storage.BaseDir[2:])
	}

	if strings.HasPrefix(cfg.Storage.GlobalDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		cfg.Storage.GlobalDir = filepath.Join(home, cfg.Storage.GlobalDir[2:])
	}

	// Ensure directories exist
	for _, dir := range []string{cfg.Storage.BaseDir, cfg.Storage.GlobalDir} {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func validateConfig(cfg *Config) error {
	// Basic validation
	if cfg.Server.Name == "" {
		return fmt.Errorf("server.name is required")
	}

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}

	if cfg.Storage.BaseDir == "" {
		return fmt.Errorf("storage.base_dir is required")
	}

	if cfg.Storage.MaxSize <= 0 {
		return fmt.Errorf("storage.max_size must be positive")
	}

	// Storage configuration validation (0 = infinite retention is valid)
	// No validation needed for Retention as 0 is allowed for infinite retention

	// Cache configuration validation
	validEvictionPolicies := []string{"lru", "fifo", "none"}
	if cfg.Storage.CacheEviction != "" {
		if !slices.Contains(validEvictionPolicies, cfg.Storage.CacheEviction) {
			return fmt.Errorf("invalid cache eviction policy: %s", cfg.Storage.CacheEviction)
		}
	}

	if cfg.Storage.CacheMaxSize <= 0 {
		return fmt.Errorf("storage.cache_max_size must be positive")
	}
	if cfg.Storage.CacheMaxSize > 100000 {
		return fmt.Errorf("storage.cache_max_size must be <= 100000")
	}

	// Populate type-safe cache configuration
	cfg.Storage.CacheSize = types.MustNewCacheSize(uint32(cfg.Storage.CacheMaxSize))

	evictionPolicy, err := types.NewEvictionPolicy(cfg.Storage.CacheEviction)
	if err != nil {
		return fmt.Errorf("invalid cache eviction policy: %w", err)
	}
	cfg.Storage.EvictionPolicy = evictionPolicy

	// Log level validation
	validLogLevels := []string{"trace", "debug", "info", "warn", "error"}
	if cfg.Log.Level != "" {
		if !slices.Contains(validLogLevels, cfg.Log.Level) {
			return fmt.Errorf("invalid log level: %s", cfg.Log.Level)
		}
	}

	// Log format validation
	validLogFormats := []string{"text", "json", "logfmt"}
	if cfg.Log.Format != "" {
		if !slices.Contains(validLogFormats, cfg.Log.Format) {
			return fmt.Errorf("invalid log format: %s", cfg.Log.Format)
		}
	}

	return nil
}
