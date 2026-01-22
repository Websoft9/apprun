package metricstore

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("METRICS_STORAGE_BACKEND")
	os.Unsetenv("METRICS_STORAGE_TIMEOUT")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify defaults
	assert.Equal(t, "mock", cfg.Storage.Backend)
	assert.Equal(t, 5*time.Second, cfg.Storage.Timeout)
	assert.True(t, cfg.Storage.Retry.Enabled)
	assert.Equal(t, 3, cfg.Storage.Retry.MaxAttempts)
	assert.Equal(t, "exponential", cfg.Storage.Retry.Backoff)
}

func TestLoadConfig_EnvironmentOverride(t *testing.T) {
	// Set environment variables
	os.Setenv("METRICS_STORAGE_BACKEND", "badger")
	os.Setenv("METRICS_STORAGE_TIMEOUT", "10s")
	defer func() {
		os.Unsetenv("METRICS_STORAGE_BACKEND")
		os.Unsetenv("METRICS_STORAGE_TIMEOUT")
	}()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify environment override
	assert.Equal(t, "badger", cfg.Storage.Backend)
	assert.Equal(t, 10*time.Second, cfg.Storage.Timeout)
}

func TestLoadConfig_InvalidBackend(t *testing.T) {
	os.Setenv("METRICS_STORAGE_BACKEND", "invalid")
	defer os.Unsetenv("METRICS_STORAGE_BACKEND")

	_, err := LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be 'mock', 'badger', or 'prometheus'")
}

func TestLoadConfig_ToStorageConfig(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
			Retry: RetryConfig{
				Enabled:     true,
				MaxAttempts: 3,
				Backoff:     "exponential",
			},
		},
	}

	storageCfg := cfg.ToStorageConfig()
	assert.Equal(t, "mock", storageCfg.Backend)
	assert.Equal(t, 5*time.Second, storageCfg.Timeout)
	assert.True(t, storageCfg.Retry.Enabled)
	assert.Equal(t, 3, storageCfg.Retry.MaxAttempts)
	assert.Equal(t, "exponential", storageCfg.Retry.Backoff)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: &Config{
				Storage: StorageConfig{
					Backend: "mock",
					Timeout: 5 * time.Second,
					Retry: RetryConfig{
						Enabled:     true,
						MaxAttempts: 3,
						Backoff:     "exponential",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid backend",
			cfg: &Config{
				Storage: StorageConfig{
					Backend: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "must be 'mock', 'badger', or 'prometheus'",
		},
		{
			name: "invalid retry max_attempts",
			cfg: &Config{
				Storage: StorageConfig{
					Backend: "mock",
					Retry: RetryConfig{
						Enabled:     true,
						MaxAttempts: 0,
					},
				},
			},
			wantErr: true,
			errMsg:  "must be at least 1",
		},
		{
			name: "invalid retry backoff",
			cfg: &Config{
				Storage: StorageConfig{
					Backend: "mock",
					Retry: RetryConfig{
						Enabled:     true,
						MaxAttempts: 3,
						Backoff:     "invalid",
					},
				},
			},
			wantErr: true,
			errMsg:  "must be 'exponential' or 'linear'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
