package obs

import "time"

// Cache configuration
const (
	MetricsCacheTTL       = 5 * time.Minute
	MetricsCacheKeyPrefix = "metrics:"
)

// Cache keys
const (
	CacheKeyAllMetrics    = "metrics:all"
	CacheKeyUserMetrics   = "metrics:users"
	CacheKeySystemMetrics = "metrics:system"
	CacheKeyPerfMetrics   = "metrics:performance"
	CacheKeyAuthMetrics   = "metrics:auth"
)

// Rate limiting configuration
const (
	// MetricsRateLimitRequests is the maximum number of requests allowed
	MetricsRateLimitRequests = 100
	// MetricsRateLimitWindow is the time window for rate limiting
	MetricsRateLimitWindow = 1 * time.Minute
)
