package metrics

import "time"

// SnapshotResponse represents the aggregated metrics snapshot
type SnapshotResponse struct {
	Scope     string                 `json:"scope"`   // "system", "users", "all"
	Metrics   map[string]interface{} `json:"metrics"` // Dynamic metrics based on scope
	Timestamp time.Time              `json:"timestamp"`
	CacheHit  bool                   `json:"cache_hit,omitempty"`
}

// MetricKey represents metadata about a single metric
type MetricKey struct {
	Name        string `json:"name"`        // e.g., "user_count_total"
	Source      string `json:"source"`      // "system" or "user"
	Description string `json:"description"` // Human-readable description
	Unit        string `json:"unit"`        // e.g., "count", "percent", "bytes"
}

// KeysResponse lists all available metric keys
type KeysResponse struct {
	Keys  []MetricKey `json:"keys"`
	Count int         `json:"count"`
}

// ScopeDefinition represents a snapshot scope configuration
type ScopeDefinition struct {
	Scope       string   `json:"scope"`       // e.g., "system", "users"
	Metrics     []string `json:"metrics"`     // Metric names included in this scope
	Description string   `json:"description"` // Human-readable description
}

// ScopesResponse lists all available snapshot scopes
type ScopesResponse struct {
	Scopes []ScopeDefinition `json:"scopes"`
	Count  int               `json:"count"`
}

// MetricsResponse - Generic metrics response wrapper (DEPRECATED - use SnapshotResponse)
type MetricsResponse struct {
	Metrics   map[string]interface{} `json:"metrics"`
	Timestamp time.Time              `json:"timestamp"`
	CacheHit  bool                   `json:"cache_hit,omitempty"`
}

// UserMetrics - User-related metrics (used internally by snapshot system)
type UserMetrics struct {
	TotalUsers                 int       `json:"total_users"`
	ActiveUsers                int       `json:"active_users"`
	AdminUsers                 int       `json:"admin_users"`
	BannedUsers                int       `json:"banned_users"`
	NewUsersToday              int       `json:"new_users_today"`
	UserRegistrationsLast7Days int       `json:"user_registrations_last_7_days"`
	Timestamp                  time.Time `json:"timestamp"`
	CacheHit                   bool      `json:"cache_hit,omitempty"`
}

// SystemMetrics - System health metrics (used internally by snapshot system)
type SystemMetrics struct {
	UptimeSeconds    int64     `json:"uptime_seconds"`
	MemoryUsageMB    uint64    `json:"memory_usage_mb"`
	CPUUsagePercent  float64   `json:"cpu_usage_percent"`
	DiskUsagePercent float64   `json:"disk_usage_percent"`
	Goroutines       int       `json:"goroutines"`
	Timestamp        time.Time `json:"timestamp"`
}

// PerformanceMetrics - API performance metrics (used internally by snapshot system)
type PerformanceMetrics struct {
	APIRequestsTotal         int64     `json:"api_requests_total"`
	APIResponseTimeP95       int       `json:"api_response_time_p95"`       // milliseconds
	APIErrorRate             float64   `json:"api_error_rate"`              // percentage
	DatabaseQueryDurationAvg float64   `json:"database_query_duration_avg"` // milliseconds
	Timestamp                time.Time `json:"timestamp"`
	CacheHit                 bool      `json:"cache_hit,omitempty"`
}

// AuthMetrics - Authentication metrics (used internally by snapshot system)
type AuthMetrics struct {
	LoginAttemptsTotal  int64     `json:"login_attempts_total"`
	LoginSuccessRate    float64   `json:"login_success_rate"` // percentage
	FailedLoginAttempts int64     `json:"failed_login_attempts"`
	TokenIssuedTotal    int64     `json:"token_issued_total"`
	Timestamp           time.Time `json:"timestamp"`
	CacheHit            bool      `json:"cache_hit,omitempty"`
}

// AllMetrics - Aggregated metrics response (DEPRECATED - use SnapshotResponse)
type AllMetrics struct {
	TotalUsers                 int       `json:"total_users"`
	ActiveUsers                int       `json:"active_users"`
	AdminUsers                 int       `json:"admin_users"`
	BannedUsers                int       `json:"banned_users"`
	NewUsersToday              int       `json:"new_users_today"`
	UserRegistrationsLast7Days int       `json:"user_registrations_last_7_days"`
	LoginAttemptsTotal         int64     `json:"login_attempts_total"`
	LoginSuccessRate           float64   `json:"login_success_rate"`
	FailedLoginAttempts        int64     `json:"failed_login_attempts"`
	TokenIssuedTotal           int64     `json:"token_issued_total"`
	APIRequestsTotal           int64     `json:"api_requests_total"`
	APIResponseTimeP95         int       `json:"api_response_time_p95"`
	APIErrorRate               float64   `json:"api_error_rate"`
	UptimeSeconds              int64     `json:"uptime_seconds"`
	MemoryUsageMB              uint64    `json:"memory_usage_mb"`
	CPUUsagePercent            float64   `json:"cpu_usage_percent"`
	DiskUsagePercent           float64   `json:"disk_usage_percent"`
	Timestamp                  time.Time `json:"timestamp"`
	CacheHit                   bool      `json:"cache_hit,omitempty"`
}

// MetricPoint represents a single metric data point in time series
type MetricPoint struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// HistoryResponse represents the response for historical metrics query
type HistoryResponse struct {
	Metrics []MetricPoint `json:"metrics"`
	Count   int           `json:"count"`
	Start   time.Time     `json:"start"`
	End     time.Time     `json:"end"`
	HasMore bool          `json:"has_more"`
}

// IngestRequest represents a single metric ingestion request
type IngestRequest struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
	Timestamp *time.Time        `json:"timestamp,omitempty"`
}

// IngestBatchResponse represents the response for batch metric ingestion
type IngestBatchResponse struct {
	Success       bool     `json:"success"`
	Ingested      int      `json:"ingested"`
	Failed        int      `json:"failed"`
	FailedMetrics []string `json:"failed_metrics,omitempty"`
	Message       string   `json:"message"`
}
