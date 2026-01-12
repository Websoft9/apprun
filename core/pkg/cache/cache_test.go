package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultConfig verifies default configuration values
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, DefaultHost, cfg.Host)
	assert.Equal(t, DefaultPort, cfg.Port)
	assert.Equal(t, DefaultDB, cfg.DB)
	assert.Equal(t, DefaultPoolSize, cfg.PoolSize)
	assert.Equal(t, DefaultTimeout, cfg.Timeout)
	assert.Equal(t, DefaultMaxRetries, cfg.MaxRetries)
	assert.True(t, cfg.FailOpen) // Default is fail-open
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Host:     "localhost",
				Port:     "6379",
				DB:       0,
				PoolSize: 10,
				Timeout:  2 * time.Second,
			},
			wantError: false,
		},
		{
			name: "empty host",
			config: &Config{
				Host:     "",
				Port:     "6379",
				DB:       0,
				PoolSize: 10,
				Timeout:  2 * time.Second,
			},
			wantError: true,
			errorMsg:  "host cannot be empty",
		},
		{
			name: "empty port",
			config: &Config{
				Host:     "localhost",
				Port:     "",
				DB:       0,
				PoolSize: 10,
				Timeout:  2 * time.Second,
			},
			wantError: true,
			errorMsg:  "port cannot be empty",
		},
		{
			name: "invalid DB negative",
			config: &Config{
				Host:     "localhost",
				Port:     "6379",
				DB:       -1,
				PoolSize: 10,
				Timeout:  2 * time.Second,
			},
			wantError: true,
			errorMsg:  "db must be between 0 and 15",
		},
		{
			name: "invalid DB too large",
			config: &Config{
				Host:     "localhost",
				Port:     "6379",
				DB:       16,
				PoolSize: 10,
				Timeout:  2 * time.Second,
			},
			wantError: true,
			errorMsg:  "db must be between 0 and 15",
		},
		{
			name: "invalid pool size",
			config: &Config{
				Host:     "localhost",
				Port:     "6379",
				DB:       0,
				PoolSize: 0,
				Timeout:  2 * time.Second,
			},
			wantError: true,
			errorMsg:  "pool_size must be at least 1",
		},
		{
			name: "invalid timeout",
			config: &Config{
				Host:     "localhost",
				Port:     "6379",
				DB:       0,
				PoolSize: 10,
				Timeout:  0,
			},
			wantError: true,
			errorMsg:  "timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestConfigAddress tests the Address() method
func TestConfigAddress(t *testing.T) {
	cfg := &Config{
		Host: "redis.example.com",
		Port: "6380",
	}

	assert.Equal(t, "redis.example.com:6380", cfg.Address())
}

// TestErrorCodes verifies all error codes are defined
func TestErrorCodes(t *testing.T) {
	assert.Equal(t, "CACHE_CONNECT_001", CodeConnectFailed)
	assert.Equal(t, "CACHE_OP_002", CodeOperationFail)
	assert.Equal(t, "CACHE_CONFIG_003", CodeInvalidConfig)
	assert.Equal(t, "CACHE_NOT_FOUND_004", CodeKeyNotFound)
	assert.Equal(t, "CACHE_TIMEOUT_005", CodeTimeout)
	assert.Equal(t, "CACHE_PUBSUB_006", CodePubSubFailed)
	assert.Equal(t, "CACHE_SERIALIZE_007", CodeSerializeFail)
}

// TestPredefinedErrors verifies all predefined errors
func TestPredefinedErrors(t *testing.T) {
	assert.NotNil(t, ErrConnectFailed)
	assert.NotNil(t, ErrOperationFailed)
	assert.NotNil(t, ErrInvalidConfig)
	assert.NotNil(t, ErrKeyNotFound)
	assert.NotNil(t, ErrTimeout)
	assert.NotNil(t, ErrPubSubFailed)
	assert.NotNil(t, ErrSerializationFailed)

	// Verify error codes match
	assert.Equal(t, CodeConnectFailed, ErrConnectFailed.Code)
	assert.Equal(t, CodeOperationFail, ErrOperationFailed.Code)
	assert.Equal(t, CodeInvalidConfig, ErrInvalidConfig.Code)
	assert.Equal(t, CodeKeyNotFound, ErrKeyNotFound.Code)
	assert.Equal(t, CodeTimeout, ErrTimeout.Code)
	assert.Equal(t, CodePubSubFailed, ErrPubSubFailed.Code)
	assert.Equal(t, CodeSerializeFail, ErrSerializationFailed.Code)
}
