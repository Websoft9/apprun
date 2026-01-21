package obs

import "time"

// Cache configuration
const (
	MetricsCacheTTL       = 1 * time.Minute // Story 9.1: Balance real-time vs performance
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
