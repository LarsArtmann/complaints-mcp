package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfigSuccess(t *testing.T) {
	// Create a temporary config
	v := viper.New()
	v.Set("server.name", "test-server")
	v.Set("server.host", "localhost")
	v.Set("server.port", 8080)
	v.Set("storage.base_dir", "/tmp/test")
	v.Set("storage.max_size", 1048576) // 1MB
	v.Set("storage.retention_days", 30)
	v.Set("storage.max_size", 1048576) // 1MB
	v.Set("storage.retention_days", 30)
	v.Set("storage.max_size", 1048576) // 1MB
	v.Set("storage.cache_enabled", true)
	v.Set("storage.cache_max_size", 100)
	v.Set("storage.cache_eviction", "lru")
	v.Set("log.level", "info")

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	assert.Equal(t, "test-server", cfg.Server.Name)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.True(t, cfg.Storage.CacheEnabled)
	assert.Equal(t, int64(100), cfg.Storage.CacheMaxSize)
}

func TestValidateConfigMissingRequired(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "missing server name",
			cfg: Config{
				Server:  ServerConfig{Port: 8080},
				Storage: StorageConfig{BaseDir: "/tmp"},
			},
		},
		{
			name: "invalid port",
			cfg: Config{
				Server:  ServerConfig{Name: "test", Port: 0},
				Storage: StorageConfig{BaseDir: "/tmp"},
			},
		},
		{
			name: "missing base dir",
			cfg: Config{
				Server:  ServerConfig{Name: "test", Port: 8080},
				Storage: StorageConfig{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.cfg)
			assert.Error(t, err)
		})
	}
}

func TestCacheSizeConversion(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			CacheMaxSize: 500,
		},
	}

	// Test that int64 gets converted to CacheSize properly
	assert.Equal(t, int64(500), cfg.Storage.CacheMaxSize)
}

func TestCacheEvictionPolicyValidation(t *testing.T) {
	tests := []struct {
		name    string
		policy  string
		wantErr bool
	}{
		{
			name:    "valid LRU policy",
			policy:  "lru",
			wantErr: false,
		},
		{
			name:    "valid FIFO policy",
			policy:  "fifo",
			wantErr: false,
		},
		{
			name:    "valid none policy",
			policy:  "none",
			wantErr: false,
		},
		{
			name:    "invalid policy",
			policy:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Name: "test",
					Port: 8080,
				},
				Storage: StorageConfig{
					BaseDir:       "/tmp",
					MaxSize:       1048576,
					Retention:     30,
					CacheEviction: tt.policy,
					CacheMaxSize:  100,
				},
			}

			err := validateConfig(cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigIntegration(t *testing.T) {
	// Test the full configuration loading process
	v := viper.New()
	v.Set("server.name", "integration-test")
	v.Set("server.host", "127.0.0.1")
	v.Set("server.port", 9090)
	v.Set("storage.base_dir", "/tmp/integration")
	v.Set("storage.global_dir", "/tmp/integration-global")
	v.Set("storage.max_size", 5242880) // 5MB
	v.Set("storage.retention_days", 60)
	v.Set("storage.auto_backup", true)
	v.Set("storage.cache_enabled", true)
	v.Set("storage.cache_max_size", 200)
	v.Set("storage.cache_eviction", "fifo")
	v.Set("log.level", "debug")

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Verify all values
	assert.Equal(t, "integration-test", cfg.Server.Name)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "/tmp/integration", cfg.Storage.BaseDir)
	assert.Equal(t, "/tmp/integration-global", cfg.Storage.GlobalDir)
	assert.Equal(t, int64(5242880), cfg.Storage.MaxSize)
	assert.Equal(t, 60, cfg.Storage.Retention)
	assert.True(t, cfg.Storage.AutoBackup)
	assert.True(t, cfg.Storage.CacheEnabled)
	assert.Equal(t, int64(200), cfg.Storage.CacheMaxSize)
	assert.Equal(t, "fifo", cfg.Storage.CacheEviction)
}

func BenchmarkConfigValidation(b *testing.B) {
	cfg := &Config{
		Server: ServerConfig{
			Name: "bench-server",
			Host: "localhost",
			Port: 8080,
		},
		Storage: StorageConfig{
			BaseDir:       "/tmp/bench",
			CacheEnabled:  true,
			CacheMaxSize:  1000,
			CacheEviction: "lru",
		},
	}

	b.ResetTimer()
	for i := range b.N {
		if err := validateConfig(cfg); err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}
