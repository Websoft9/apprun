package metrics

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"

	"apprun/pkg/errors"
	"apprun/pkg/metrics/storage"
)

// Config holds the metrics module configuration.
// It can be loaded from config files and overridden by environment variables.
type Config struct {
	Storage StorageConfig `mapstructure:"storage"`
}

// StorageConfig holds storage backend configuration.
type StorageConfig struct {
	Backend  string                 `mapstructure:"backend"`
	Timeout  time.Duration          `mapstructure:"timeout"`
	Retry    RetryConfig            `mapstructure:"retry"`
	Settings map[string]interface{} `mapstructure:",remain"` // Catch-all for backend-specific settings
}

// RetryConfig defines retry behavior.
type RetryConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	MaxAttempts int    `mapstructure:"max_attempts"`
	Backoff     string `mapstructure:"backoff"`
}

// LoadConfig loads metrics configuration from file and environment variables.
// Environment variable METRICS_STORAGE_BACKEND overrides the config file backend.
//
// Example:
//
//	cfg, err := metrics.LoadConfig()
//	if err != nil {
//	    log.Fatal("Failed to load config", err)
//	}
//	storage, err := storage.NewStorage(cfg.Storage)
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set config file path
	v.SetConfigName("metrics")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")     // Current directory
	v.AddConfigPath("../config")    // Parent directory
	v.AddConfigPath("../../config") // Two levels up
	v.AddConfigPath("/etc/apprun")  // System-wide config

	// Set defaults
	v.SetDefault("storage.backend", "mock")
	v.SetDefault("storage.timeout", "5s")
	v.SetDefault("storage.retry.enabled", true)
	v.SetDefault("storage.retry.max_attempts", 3)
	v.SetDefault("storage.retry.backoff", "exponential")

	// Bind environment variables
	v.SetEnvPrefix("METRICS")
	v.AutomaticEnv()

	// Allow environment variable override for backend
	if backend := os.Getenv("METRICS_STORAGE_BACKEND"); backend != "" {
		v.Set("storage.backend", backend)
	}
	if timeout := os.Getenv("METRICS_STORAGE_TIMEOUT"); timeout != "" {
		v.Set("storage.timeout", timeout)
	}

	// Read config file (optional - uses defaults if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file found but has errors
			return nil, errors.Wrap(err, "METRICS_CONFIG_ERROR", "failed to read config file")
		}
		// Config file not found - use defaults (not an error)
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "METRICS_CONFIG_ERROR", "failed to unmarshal config")
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validateConfig checks if the configuration is valid.
func validateConfig(cfg *Config) error {
	// Validate backend
	validBackends := map[string]bool{
		"mock":       true,
		"badger":     true,
		"prometheus": true,
	}
	if !validBackends[cfg.Storage.Backend] {
		return storage.ErrInvalidConfig("backend", fmt.Sprintf("must be 'mock', 'badger', or 'prometheus', got '%s'", cfg.Storage.Backend))
	}

	// Validate timeout
	if cfg.Storage.Timeout < 0 {
		return storage.ErrInvalidConfig("timeout", "cannot be negative")
	}
	if cfg.Storage.Timeout == 0 {
		cfg.Storage.Timeout = 5 * time.Second // Default
	}

	// Validate retry config
	if cfg.Storage.Retry.Enabled {
		if cfg.Storage.Retry.MaxAttempts < 1 {
			return storage.ErrInvalidConfig("retry.max_attempts", "must be at least 1")
		}
		if cfg.Storage.Retry.Backoff != "exponential" && cfg.Storage.Retry.Backoff != "linear" {
			return storage.ErrInvalidConfig("retry.backoff", "must be 'exponential' or 'linear'")
		}
	}

	return nil
}

// ToStorageConfig converts Config to storage.Config for creating storage backends.
func (c *Config) ToStorageConfig() storage.Config {
	return storage.Config{
		Backend:  c.Storage.Backend,
		Settings: c.Storage.Settings,
		Timeout:  c.Storage.Timeout,
		Retry: storage.RetryConfig{
			Enabled:     c.Storage.Retry.Enabled,
			MaxAttempts: c.Storage.Retry.MaxAttempts,
			Backoff:     c.Storage.Retry.Backoff,
		},
	}
}
