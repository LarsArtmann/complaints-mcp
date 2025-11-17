package config_test

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/larsartmann/complaints-mcp/internal/config"
)

func TestConfig_Load(t *testing.T) {
	// Test that Load function can create a valid config
	// Note: This tests structure without requiring actual files
	ctx := context.Background()

	// Create a simple command with basic flags
	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("config", "", "config file")
	cmd.PersistentFlags().String("server.host", "", "server host")
	cmd.PersistentFlags().Int("server.port", 0, "server port")

	cfg, err := config.Load(ctx, cmd)
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestConfig_ServerConfig(t *testing.T) {
	// Test ServerConfig structure
	serverConfig := config.ServerConfig{
		Name: "test-server",
		Host: "localhost",
		Port: uint16(8080),
	}

	// Test Address method
	address := serverConfig.Address()
	require.Equal(t, "localhost:8080", address)

	// Test Address method with empty host
	serverConfig.Host = ""
	address = serverConfig.Address()
	require.Equal(t, ":8080", address)
}

func TestConfig_StorageConfig(t *testing.T) {
	// Test StorageConfig structure
	storageConfig := config.StorageConfig{
		BaseDir:    "/tmp/complaints",
		GlobalDir:  "/tmp/global",
		MaxSize:    uint64(1024 * 1024), // 1MB
		Retention:  uint(30),
		AutoBackup: true,
	}

	// Test that fields are set correctly
	require.Equal(t, "/tmp/complaints", storageConfig.BaseDir)
	require.Equal(t, "/tmp/global", storageConfig.GlobalDir)
	require.Equal(t, uint64(1024*1024), storageConfig.MaxSize)
	require.Equal(t, uint(30), storageConfig.Retention)
	require.True(t, storageConfig.AutoBackup)
}

func TestConfig_LogConfig(t *testing.T) {
	// Test LogConfig structure
	logConfig := config.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	// Test that fields are set correctly
	require.Equal(t, "info", logConfig.Level)
	require.Equal(t, "json", logConfig.Format)
	require.Equal(t, "stdout", logConfig.Output)
}

func TestConfig_CompleteConfig(t *testing.T) {
	// Test complete configuration structure
	cfg := config.Config{
		Server: config.ServerConfig{
			Name: "complaints-mcp",
			Host: "localhost",
			Port: uint16(8080),
		},
		Storage: config.StorageConfig{
			BaseDir:    "/data/complaints",
			GlobalDir:  "/data/global",
			MaxSize:    uint64(10 * 1024 * 1024), // 10MB
			Retention:  uint(90),
			AutoBackup: false,
		},
		Log: config.LogConfig{
			Level:  "debug",
			Format: "text",
			Output: "stderr",
		},
	}

	// Test all components
	require.Equal(t, "complaints-mcp", cfg.Server.Name)
	require.Equal(t, "localhost", cfg.Server.Host)
	require.Equal(t, uint16(8080), cfg.Server.Port)

	require.Equal(t, "/data/complaints", cfg.Storage.BaseDir)
	require.Equal(t, "/data/global", cfg.Storage.GlobalDir)
	require.Equal(t, uint64(10*1024*1024), cfg.Storage.MaxSize)
	require.Equal(t, uint(90), cfg.Storage.Retention)
	require.False(t, cfg.Storage.AutoBackup)

	require.Equal(t, "debug", cfg.Log.Level)
	require.Equal(t, "text", cfg.Log.Format)
	require.Equal(t, "stderr", cfg.Log.Output)
}

func TestConfig_ServerAddress(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		expected string
	}{
		{
			name:     "full address with host",
			host:     "localhost",
			port:     8080,
			expected: "localhost:8080",
		},
		{
			name:     "empty host",
			host:     "",
			port:     8080,
			expected: ":8080",
		},
		{
			name:     "different host",
			host:     "example.com",
			port:     9090,
			expected: "example.com:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverConfig := config.ServerConfig{
				Host: tt.host,
				Port: uint16(tt.port),
			}

			address := serverConfig.Address()
			require.Equal(t, tt.expected, address)
		})
	}
}
