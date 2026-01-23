package metrics

import (
	"fmt"
	"strings"
	"time"
)

// Config holds the metrics module configuration.
// This module handles metric collection, caching, and HTTP API exposure.
// Note: Storage configuration is managed separately by pkg/metricstore
type Config struct {
	// Collection configuration
	Collection CollectionConfig `mapstructure:"collection" json:"collection" yaml:"collection"`

	// Cache configuration
	Cache CacheConfig `mapstructure:"cache" json:"cache" yaml:"cache"`

	// Export configuration
	Export ExportConfig `mapstructure:"export" json:"export" yaml:"export"`

	// Rate limiting configuration
	RateLimit RateLimitConfig `mapstructure:"rate_limit" json:"rate_limit" yaml:"rate_limit"`

	// Tags configuration (Story 9.1)
	Tags TagsConfig `mapstructure:"tags" json:"tags" yaml:"tags"`
}

// CollectionConfig defines metric collection behavior.
type CollectionConfig struct {
	// Enabled determines if metric collection is active
	Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled" default:"true" db:"true"`

	// Interval between metric collections (default: 15s)
	Interval time.Duration `mapstructure:"interval" json:"interval" yaml:"interval" default:"15s" db:"true" validate:"gte=1s"`

	// IncludeRuntime includes Go runtime metrics (goroutines, memory, GC)
	IncludeRuntime bool `mapstructure:"include_runtime" json:"include_runtime" yaml:"include_runtime" default:"true" db:"true"`
}

// CacheConfig defines metric caching behavior for performance optimization.
type CacheConfig struct {
	// Enabled determines if caching is active
	Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled" default:"true" db:"true"`

	// TTL is cache expiration time (default: 1m)
	TTL time.Duration `mapstructure:"ttl" json:"ttl" yaml:"ttl" default:"1m" db:"true" validate:"gte=0"`

	// KeyPrefix for cache keys (default: "metrics:")
	KeyPrefix string `mapstructure:"key_prefix" json:"key_prefix" yaml:"key_prefix" default:"metrics:" db:"false"`
}

// ExportConfig defines Prometheus export settings.
type ExportConfig struct {
	// Enabled determines if Prometheus export endpoint is active
	Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled" default:"true" db:"true"`

	// Path is the HTTP endpoint for Prometheus scraping (default: "/metrics")
	Path string `mapstructure:"path" json:"path" yaml:"path" default:"/metrics" db:"true" validate:"required,startswith=/"`
}

// RateLimitConfig defines API rate limiting for metrics endpoints.
type RateLimitConfig struct {
	// Enabled determines if rate limiting is active
	Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled" default:"true" db:"true"`

	// Requests is max requests per window (default: 100)
	Requests int `mapstructure:"requests" json:"requests" yaml:"requests" default:"100" db:"true" validate:"gte=1"`

	// Window is the time window for rate limiting (default: 1m)
	Window time.Duration `mapstructure:"window" json:"window" yaml:"window" default:"1m" db:"true" validate:"gte=1s"`
}

// TagsConfig defines global tags for all metrics (Story 9.1).
type TagsConfig struct {
	// Enabled determines if tags collection is active
	Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled" default:"true" db:"true"`

	// GlobalTags are automatically added to all metrics
	GlobalTags map[string]string `mapstructure:"global_tags" json:"global_tags" yaml:"global_tags"`

	// Blacklist contains tag keys that should never be recorded (for security)
	Blacklist []string `mapstructure:"blacklist" json:"blacklist" yaml:"blacklist"`
}

// DefaultConfig returns the default metrics configuration.
func DefaultConfig() *Config {
	return &Config{
		Collection: CollectionConfig{
			Enabled:        true,
			Interval:       15 * time.Second,
			IncludeRuntime: true,
		},
		Cache: CacheConfig{
			Enabled:   true,
			TTL:       1 * time.Minute,
			KeyPrefix: "metrics:",
		},
		Export: ExportConfig{
			Enabled: true,
			Path:    "/metrics",
		},
		RateLimit: RateLimitConfig{
			Enabled:  true,
			Requests: 100,
			Window:   1 * time.Minute,
		},
		Tags: TagsConfig{
			Enabled:    true,
			GlobalTags: map[string]string{},
			Blacklist:  []string{"password", "token", "secret", "key"},
		},
	}
}

// Validate validates the metrics configuration
func (c *Config) Validate() error {
	if c.Collection.Interval < 1*time.Second {
		return fmt.Errorf("collection interval must be >= 1s")
	}
	if c.Cache.TTL < 0 {
		return fmt.Errorf("cache TTL cannot be negative")
	}
	if c.RateLimit.Requests < 1 {
		return fmt.Errorf("rate limit requests must be >= 1")
	}
	if c.RateLimit.Window < 1*time.Second {
		return fmt.Errorf("rate limit window must be >= 1s")
	}
	if c.Export.Path == "" || !strings.HasPrefix(c.Export.Path, "/") {
		return fmt.Errorf("export path must start with /")
	}
	return nil
}

// Cache configuration (DEPRECATED: use Config.Cache instead)
const (
	MetricsCacheTTL       = 1 * time.Minute // Story 9.1: Balance real-time vs performance
	MetricsCacheKeyPrefix = "metrics:"
)

// Storage configuration (DEPRECATED: moved to pkg/metricstore/config.go)
const (
	// MetricsStorageRetentionBadger is the maximum retention period for BadgerDB storage
	// After this period, metrics are automatically purged to prevent disk space issues
	MetricsStorageRetentionBadger = 24 * time.Hour

	// MetricsStorageRetentionPrometheus is the default retention for Prometheus (configured in Prometheus itself)
	MetricsStorageRetentionPrometheus = 15 * 24 * time.Hour // 15 days
)

// Cache keys
const (
	CacheKeyAllMetrics    = "metrics:all"
	CacheKeyUserMetrics   = "metrics:users"
	CacheKeySystemMetrics = "metrics:system"
	CacheKeyPerfMetrics   = "metrics:performance"
	CacheKeyAuthMetrics   = "metrics:auth"
)

// Rate limiting configuration (DEPRECATED: use Config.RateLimit instead)
const (
	// MetricsRateLimitRequests is the maximum number of requests allowed
	MetricsRateLimitRequests = 100
	// MetricsRateLimitWindow is the time window for rate limiting
	MetricsRateLimitWindow = 1 * time.Minute
)

// Standard metric names (Story 9.1 - Storage integration)
const (
	MetricNameUserTotal    = "user_count_total"
	MetricNameUserActive   = "user_count_active"
	MetricNameUserAdmin    = "user_count_admin"
	MetricNameUserBanned   = "user_count_banned"
	MetricNameUserNewToday = "user_count_new_today"
	MetricNameAPIRequests  = "api_requests_total"
	MetricNameSystemMemory = "system_memory_mb"
	MetricNameSystemCPU    = "system_cpu_percent"
	MetricNameSystemDisk   = "system_disk_percent"
	MetricNameSystemUptime = "system_uptime_seconds"
)
