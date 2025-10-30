package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		Host string `mapstructure:"host" validate:"required,hostname" json:"host"`
		Port int    `mapstructure:"port" validate:"required,min=1024,max=65535" json:"port"`
	}
	
	Database struct {
		URL             string        `mapstructure:"url" validate:"required,url" json:"url"`
		MaxConnections  int           `mapstructure:"max_connections" validate:"min=1,max=100" json:"max_connections"`
		Timeout         time.Duration `mapstructure:"timeout" validate:"required,min=1s" json:"timeout"`
	}
	
	Logging struct {
		Level      string `mapstructure:"level" validate:"required,oneof=debug info warn error fatal" json:"level"`
		Format     string `mapstructure:"format" validate:"required,oneof=json console" json:"format"`
		OutputPath string `mapstructure:"output_path" json:"output_path"`
	}
	
	Complaints struct {
		ProjectName      string `mapstructure:"project_name" validate:"required" json:"project_name"`
		StorageLocation  string `mapstructure:"storage_location" validate:"required,oneof=local global both" json:"storage_location"`
		RetentionDays   int    `mapstructure:"retention_days" validate:"min=1,max=365" json:"retention_days"`
	}
	
	Security struct {
		EnableAuth bool   `mapstructure:"enable_auth" json:"enable_auth"`
		SecretKey  string `mapstructure:"secret_key" json:"secret_key"`
		JWTSecret  string `mapstructure:"jwt_secret" json:"jwt_secret"`
		TokenExpiry int    `mapstructure:"token_expiry" json:"token_expiry"`
	}
}

// Load loads configuration from various sources
func Load() (*Config, error) {
	v := viper.New()
	
	// Set configuration file path and name
	v.SetConfigName("complaints-mcp")
	v.SetConfigType("yaml")
	
	// Add configuration search paths
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.config/complaints-mcp")
	v.AddConfigPath("/etc/complaints-mcp")
	
	// Read environment variables
	v.AutomaticEnv()
	
	// Set default values
	setDefaults(v)
	
	// Read configuration
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found is OK, use defaults
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

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)
	
	// Database defaults
	v.SetDefault("database.url", "file:complaints.db")
	v.SetDefault("database.max_connections", 10)
	v.SetDefault("database.timeout", "30s")
	
	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output_path", "/var/log/complaints-mcp.log")
	
	// Complaints defaults
	v.SetDefault("complaints.project_name", "")
	v.SetDefault("complaints.storage_location", "local")
	v.SetDefault("complaints.retention_days", 30)
	
	// Security defaults
	v.SetDefault("security.enable_auth", false)
	v.SetDefault("security.secret_key", "")
	v.SetDefault("security.jwt_secret", "")
	v.SetDefault("security.token_expiry", 86400) // 24 hours
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1024 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1024 and 65535")
	}
	
	if c.Database.MaxConnections < 1 || c.Database.MaxConnections > 100 {
		return fmt.Errorf("database max connections must be between 1 and 100")
	}
	
	if c.Database.Timeout < time.Second {
		return fmt.Errorf("database timeout must be at least 1 second")
	}
	
	validStorage := map[string]bool{
		"local":  true,
		"global": true,
		"both":   true,
	}
	
	if !validStorage[c.Complaints.StorageLocation] {
		return fmt.Errorf("storage location must be one of: local, global, both")
	}
	
	if c.Complaints.RetentionDays < 1 || c.Complaints.RetentionDays > 365 {
		return fmt.Errorf("retention days must be between 1 and 365")
	}
	
	return nil
}

// GetAddress returns the server address
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsDebug returns true if debug mode is enabled
func (c *Config) IsDebug() bool {
	return c.Logging.Level == "debug"
}