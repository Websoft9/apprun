// Package metricstore provides metrics storage and retrieval functionality for the application.
// It supports multiple storage backends and includes configuration management for metrics collection.
package metricstore

import (
	stdErrors "errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"

	"apprun/pkg/errors"
	"apprun/pkg/metricstore/storage"
)

// Config holds the metricstore module configuration.
// It can be loaded from config files and overridden by environment variables.
type Config struct {
	// Storage backend configuration
	Storage StorageConfig `mapstructure:"storage" json:"storage" yaml:"storage"`
}

// StorageConfig holds storage backend configuration for metricstore.
// This is infrastructure-level configuration that defines where and how metrics are persisted.
type StorageConfig struct {
	// Backend type: "mock", "badger", or "prometheus"
	Backend string `mapstructure:"backend" json:"backend" yaml:"backend" default:"mock" db:"false" validate:"oneof=mock badger prometheus"`

	// Timeout for storage operations (default: 5s)
	Timeout time.Duration `mapstructure:"timeout" json:"timeout" yaml:"timeout" default:"5s" db:"true" validate:"gte=0"`

	// Retry configuration for failed operations
	Retry RetryConfig `mapstructure:"retry" json:"retry" yaml:"retry"`

	// Retention period for BadgerDB (default: 24h)
	// Metrics older than this will be automatically purged
	Retention time.Duration `mapstructure:"retention" json:"retention" yaml:"retention" default:"24h" db:"true" validate:"gte=0"`

	// Settings contains backend-specific configuration
	// For BadgerDB: {"path": "/data/metrics"}
	// For Prometheus: {"url": "http://prometheus:9090"}
	Settings map[string]interface{} `mapstructure:",remain" json:"settings,omitempty" yaml:",inline"`
}

// RetryConfig defines retry behavior for failed storage operations.
type RetryConfig struct {
	// Enabled determines if retry is enabled (default: true)
	Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled" default:"true" db:"true"`

	// MaxAttempts is the maximum number of retry attempts (default: 3)
	MaxAttempts int `mapstructure:"max_attempts" json:"max_attempts" yaml:"max_attempts" default:"3" db:"true" validate:"gte=1"`

	// Backoff strategy: "exponential" or "linear" (default: exponential)
	Backoff string `mapstructure:"backoff" json:"backoff" yaml:"backoff" default:"exponential" db:"true" validate:"oneof=exponential linear"`
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
	v.SetDefault("storage.retention", "24h") // BadgerDB retention period
	v.SetDefault("storage.retry.enabled", true)
	v.SetDefault("storage.retry.max_attempts", 3)
	v.SetDefault("storage.retry.backoff", "exponential")

	// Bind environment variables
	v.SetEnvPrefix("METRICSTORE")
	v.AutomaticEnv()

	// Allow environment variable override for backend
	if backend := os.Getenv("METRICSTORE_STORAGE_BACKEND"); backend != "" {
		v.Set("storage.backend", backend)
	}
	if timeout := os.Getenv("METRICSTORE_STORAGE_TIMEOUT"); timeout != "" {
		v.Set("storage.timeout", timeout)
	}
	if retention := os.Getenv("METRICSTORE_STORAGE_RETENTION"); retention != "" {
		v.Set("storage.retention", retention)
	}

	// Read config file (optional - uses defaults if not found)
	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !stdErrors.As(err, &notFoundErr) {
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

	// Validate retention (for BadgerDB)
	if cfg.Storage.Backend == "badger" {
		if cfg.Storage.Retention < 0 {
			return storage.ErrInvalidConfig("retention", "cannot be negative")
		}
		if cfg.Storage.Retention == 0 {
			cfg.Storage.Retention = 24 * time.Hour // Default 24 hours
		}
		// Warn if retention is too long (may cause disk space issues)
		if cfg.Storage.Retention > 7*24*time.Hour {
			// Log warning but allow it
			fmt.Printf("Warning: BadgerDB retention period is set to %v, this may consume significant disk space\n", cfg.Storage.Retention)
		}
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
	// Add retention to settings for BadgerDB
	settings := c.Storage.Settings
	if settings == nil {
		settings = make(map[string]interface{})
	}
	if c.Storage.Backend == "badger" && c.Storage.Retention > 0 {
		settings["retention"] = c.Storage.Retention
	}

	return storage.Config{
		Backend:  c.Storage.Backend,
		Settings: settings,
		Timeout:  c.Storage.Timeout,
		Retry: storage.RetryConfig{
			Enabled:     c.Storage.Retry.Enabled,
			MaxAttempts: c.Storage.Retry.MaxAttempts,
			Backoff:     c.Storage.Retry.Backoff,
		},
	}
}
