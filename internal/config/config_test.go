package config

import (
	"os"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Server: struct {
					Host         string        `mapstructure:"host" validate:"required,hostname" json:"host"`
					Port         int           `mapstructure:"port" validate:"required,min=1024,max=65535" json:"port"`
					ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required,min=1s" json:"read_timeout"`
					WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required,min=1s" json:"write_timeout"`
				}{
					Host:         "localhost",
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Database: struct {
					URL            string        `mapstructure:"url" validate:"required,url" json:"url"`
					MaxConnections int           `mapstructure:"max_connections" validate:"required,min=1,max=100" json:"max_connections"`
					Timeout        time.Duration `mapstructure:"timeout" validate:"required,min=1s" json:"timeout"`
					SSLMode        string        `mapstructure:"ssl_mode" validate:"required,oneof=disable require" json:"ssl_mode"`
				}{
					URL:            "file:test.db",
					MaxConnections: 10,
					Timeout:        30 * time.Second,
					SSLMode:        "disable",
				},
				Logging: struct {
					Level      string `mapstructure:"level" validate:"required,oneof=debug info warn error" json:"level"`
					Format     string `mapstructure:"format" validate:"required,oneof=json console" json:"format"`
					OutputPath string `mapstructure:"output_path" validate:"required" json:"output_path"`
					Enable     *bool  `mapstructure:"enable" json:"enable"`
				}{
					Level:      "info",
					Format:     "json",
					OutputPath: "/var/log/test.log",
				},
				Complaints: struct {
					StorageLocation StorageLocation `mapstructure:"storage_location" validate:"required,oneof=local global both" json:"storage_location"`
					RetentionDays   int             `mapstructure:"retention_days" validate:"min=1,max=365" json:"retention_days"`
					ProjectName     string          `mapstructure:"project_name" json:"project_name"`
					AutoResolve     *bool           `mapstructure:"auto_resolve" json:"auto_resolve"`
					MaxFileSize     int64           `mapstructure:"max_file_size" validate:"min=1024,max=1048576" json:"max_file_size"`
				}{
					StorageLocation: StorageLocal,
					RetentionDays:   90,
					ProjectName:     "test-project",
					MaxFileSize:     1048576,
				},
				Security: struct {
					EnableAuth  bool   `mapstructure:"enable_auth" json:"enable_auth"`
					JWTSecret   string `mapstructure:"jwt_secret" json:"jwt_secret"`
					TokenExpiry int    `mapstructure:"token_expiry" validate:"min=60" json:"token_expiry"`
					BasicAuth   struct {
						Username string `mapstructure:"username" json:"username"`
						Password string `mapstructure:"password" json:"password"`
					} `mapstructure:"basic_auth" json:"basic_auth"`
				}{
					EnableAuth:  false,
					TokenExpiry: 86400,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: Config{
				Server: struct {
					Host         string        `mapstructure:"host" validate:"required,hostname" json:"host"`
					Port         int           `mapstructure:"port" validate:"required,min=1024,max=65535" json:"port"`
					ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required,min=1s" json:"read_timeout"`
					WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required,min=1s" json:"write_timeout"`
				}{
					Host:         "localhost",
					Port:         80, // Invalid port
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid storage location",
			config: Config{
				Complaints: struct {
					StorageLocation StorageLocation `mapstructure:"storage_location" validate:"required,oneof=local global both" json:"storage_location"`
					RetentionDays   int             `mapstructure:"retention_days" validate:"min=1,max=365" json:"retention_days"`
					ProjectName     string          `mapstructure:"project_name" json:"project_name"`
					AutoResolve     *bool           `mapstructure:"auto_resolve" json:"auto_resolve"`
					MaxFileSize     int64           `mapstructure:"max_file_size" validate:"min=1024,max=1048576" json:"max_file_size"`
				}{
					StorageLocation: StorageLocation("invalid"), // Invalid location
					RetentionDays:   90,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GetServerAddress(t *testing.T) {
	config := &Config{
		Server: struct {
			Host         string        `mapstructure:"host" validate:"required,hostname" json:"host"`
			Port         int           `mapstructure:"port" validate:"required,min=1024,max=65535" json:"port"`
			ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required,min=1s" json:"read_timeout"`
			WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required,min=1s" json:"write_timeout"`
		}{
			Host: "localhost",
			Port: 8080,
		},
	}

	want := "localhost:8080"
	got := config.GetServerAddress()

	if got != want {
		t.Errorf("Config.GetServerAddress() = %v, want %v", got, want)
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name: "production config",
			config: &Config{
				Logging: struct {
					Level      string `mapstructure:"level" validate:"required,oneof=debug info warn error" json:"level"`
					Format     string `mapstructure:"format" validate:"required,oneof=json console" json:"format"`
					OutputPath string `mapstructure:"output_path" validate:"required" json:"output_path"`
					Enable     *bool  `mapstructure:"enable" json:"enable"`
				}{
					Level:  "info",
					Format: "json",
				},
				Database: struct {
					URL            string        `mapstructure:"url" validate:"required,url" json:"url"`
					MaxConnections int           `mapstructure:"max_connections" validate:"required,min=1,max=100" json:"max_connections"`
					Timeout        time.Duration `mapstructure:"timeout" validate:"required,min=1s" json:"timeout"`
					SSLMode        string        `mapstructure:"ssl_mode" validate:"required,oneof=disable require" json:"ssl_mode"`
				}{
					SSLMode: "require",
				},
			},
			want: true,
		},
		{
			name: "debug config",
			config: &Config{
				Logging: struct {
					Level      string `mapstructure:"level" validate:"required,oneof=debug info warn error" json:"level"`
					Format     string `mapstructure:"format" validate:"required,oneof=json console" json:"format"`
					OutputPath string `mapstructure:"output_path" validate:"required" json:"output_path"`
					Enable     *bool  `mapstructure:"enable" json:"enable"`
				}{
					Level:  "debug",
					Format: "json",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsProduction()
			if got != tt.want {
				t.Errorf("Config.IsProduction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectProjectName(t *testing.T) {
	// Test with environment variable
	os.Setenv("COMPLAINTS_MCP_PROJECT_NAME", "test-from-env")
	defer os.Unsetenv("COMPLAINTS_MCP_PROJECT_NAME")

	got := detectProjectName()
	want := "test-from-env"

	if got != want {
		t.Errorf("detectProjectName() = %v, want %v", got, want)
	}
}
