package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

// Severity levels for complaints
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// StorageLocation represents where complaints are stored
type StorageLocation string

const (
	StorageLocal    StorageLocation = "local"
	StorageGlobal   StorageLocation = "global"
	StorageBoth     StorageLocation = "both"
)

// Config represents the complete application configuration
type Config struct {
	Server struct {
		Host         string        `mapstructure:"host" validate:"required,hostname" json:"host"`
		Port         int           `mapstructure:"port" validate:"required,min=1024,max=65535" json:"port"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required,min=1s" json:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required,min=1s" json:"write_timeout"`
	}
	
	Database struct {
		URL             string        `mapstructure:"url" validate:"required,url" json:"url"`
		MaxConnections  int           `mapstructure:"max_connections" validate:"required,min=1,max=100" json:"max_connections"`
		Timeout         time.Duration `mapstructure:"timeout" validate:"required,min=1s" json:"timeout"`
		SSLMode         string        `mapstructure:"ssl_mode" validate:"required,oneof=disable require" json:"ssl_mode"`
	}
	
	Logging struct {
		Level      string `mapstructure:"level" validate:"required,oneof=debug info warn error" json:"level"`
		Format     string `mapstructure:"format" validate:"required,oneof=json console" json:"format"`
		OutputPath string `mapstructure:"output_path" validate:"required" json:"output_path"`
		Enable     *bool  `mapstructure:"enable" json:"enable"`
	}
	
	Complaints struct {
		StorageLocation  StorageLocation `mapstructure:"storage_location" validate:"required,oneof=local global both" json:"storage_location"`
		RetentionDays   int             `mapstructure:"retention_days" validate:"min=1,max=365" json:"retention_days"`
		ProjectName      string          `mapstructure:"project_name" json:"project_name"`
		AutoResolve     *bool           `mapstructure:"auto_resolve" json:"auto_resolve"`
		MaxFileSize     int64           `mapstructure:"max_file_size" validate:"min=1024,max=1048576" json:"max_file_size"`
	}
	
	Security struct {
		EnableAuth   bool   `mapstructure:"enable_auth" json:"enable_auth"`
		JWTSecret   string `mapstructure:"jwt_secret" json:"jwt_secret"`
		TokenExpiry int    `mapstructure:"token_expiry" validate:"min=60" json:"token_expiry"`
		BasicAuth struct {
			Username string `mapstructure:"username" json:"username"`
			Password string `mapstructure:"password" json:"password"`
		} `mapstructure:"basic_auth" json:"basic_auth"`
	}
	
	Observability struct {
		EnableTracing  bool `mapstructure:"enable_tracing" json:"enable_tracing"`
		EnableMetrics  bool `mapstructure:"enable_metrics" json:"enable_metrics"`
		ServiceName  string `mapstructure:"service_name" json:"service_name"`
		EnableHealthChecks bool `mapstructure:"enable_health_checks" json:"enable_health_checks"`
	}
}

// Load loads configuration from multiple sources with proper validation
func Load() (*Config, error) {
	v := viper.New()
	
	// Set configuration name and type
	v.SetConfigName("complaints-mcp")
	v.SetConfigType("yaml")
	
	// Add configuration search paths with priority order
	v.AddConfigPath("./config")        // Highest priority: local config
	v.AddConfigPath("$HOME/.config/complaints-mcp") // Medium priority: user home
	v.AddConfigPath("/etc/complaints-mcp")     // Low priority: system-wide
	
	// Enable environment variable support
	v.AutomaticEnv()
	
	// Set environment variable prefix
	v.SetEnvPrefix("COMPLAINTS_MCP")
	
	// Set all defaults with comprehensive configuration
	setDefaults(v)
	
	// Read configuration with error handling
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found is acceptable
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &config, nil
}

// Validate validates the entire configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Port < 1024 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1024 and 65535")
	}
	
	if c.Server.ReadTimeout < time.Second || c.Server.WriteTimeout < time.Second {
		return fmt.Errorf("timeouts must be at least 1 second")
	}
	
	// Validate database configuration
	if c.Database.MaxConnections < 1 || c.Database.MaxConnections > 100 {
		return fmt.Errorf("database max connections must be between 1 and 100")
	}
	
	if c.Database.Timeout < time.Second {
		return fmt.Errorf("database timeout must be at least 1 second")
	}
	
	validSSLModes := map[string]bool{
		"disable":  true,
		"require": true,
		"prefer":  true,
	}
	
	if !validSSLModes[c.Database.SSLMode] {
		return fmt.Errorf("invalid SSL mode: %s", c.Database.SSLMode)
	}
	
	// Validate logging configuration
	validLogLevels := map[string]bool{
		"debug":  true,
		"info":   true,
		"warn":   true,
		"error":  true,
	}
	
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}
	
	validLogFormats := map[string]bool{
		"json":    true,
		"console": true,
	}
	
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s", c.Logging.Format)
	}
	
	// Validate complaints configuration
	validStorageLocations := map[string]bool{
		"local":   true,
		"global":   true,
		"both":    true,
	}
	
	if !validStorageLocations[c.Complaints.StorageLocation] {
		return fmt.Errorf("invalid storage location: %s", c.Complaints.StorageLocation)
	}
	
	if c.Complaints.RetentionDays < 1 || c.Complaints.RetentionDays > 365 {
		return fmt.Errorf("retention days must be between 1 and 365")
	}
	
	if c.Complaints.MaxFileSize < 1024 {
		return fmt.Errorf("max file size must be at least 1024 bytes")
	}
	
	if c.Complaints.ProjectName == "" {
		c.Complaints.ProjectName = detectProjectName()
	}
	
	// Validate security configuration
	if c.Security.EnableAuth && c.Security.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required when auth is enabled")
	}
	
	if c.Security.JWTSecret != "" && len(c.Security.JWTSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}
	
	if c.Security.TokenExpiry < 60 {
		return fmt.Errorf("token expiry must be at least 60 seconds")
	}
	
	return nil
}

// setDefaults sets all default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	
	// Database defaults
	v.SetDefault("database.url", "file:complaints.db")
	v.SetDefault("database.max_connections", 10)
	v.SetDefault("database.timeout", "30s")
	v.SetDefault("database.ssl_mode", "disable")
	
	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output_path", "/var/log/complaints-mcp.log")
	v.SetDefault("logging.enable", true)
	
	// Complaints defaults
	v.SetDefault("complaints.storage_location", "local")
	v.SetDefault("complaints.retention_days", 90)
	v.SetDefault("complaints.project_name", "")
	v.SetDefault("complaints.auto_resolve", false)
	v.SetDefault("complaints.max_file_size", 1048576) // 1MB
	
	// Security defaults
	v.SetDefault("security.enable_auth", false)
	v.SetDefault("security.jwt_secret", "")
	v.SetDefault("security.token_expiry", 86400) // 24 hours
	
	// Observability defaults
	v.SetDefault("observability.enable_tracing", false)
	v.SetDefault("observability.enable_metrics", false)
	v.SetDefault("observability.service_name", "complaints-mcp")
	v.SetDefault("observability.enable_health_checks", false)
}

// detectProjectName attempts to detect the project name from git or environment
func detectProjectName() string {
	// Try to get from environment
	if name := os.Getenv("COMPLAINTS_MCP_PROJECT_NAME"); name != "" {
		return name
	}
	
	// Try to get from git remote (simplified)
	return "complaints-mcp"
}

// GetServerAddress returns the formatted server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Logging.Level == "info" && c.Database.SSLMode == "require"
}

// IsDebug returns true if debug mode is enabled
func (c *Config) IsDebug() bool {
	return c.Logging.Level == "debug"
}