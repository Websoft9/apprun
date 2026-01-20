package obs

import "time"

// MetricsResponse - Generic metrics response wrapper
type MetricsResponse struct {
	Metrics   map[string]interface{} `json:"metrics"`
	Timestamp time.Time              `json:"timestamp"`
	CacheHit  bool                   `json:"cache_hit,omitempty"`
}

// UserMetrics - User-related metrics
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

// SystemMetrics - System health metrics
type SystemMetrics struct {
	UptimeSeconds    int64     `json:"uptime_seconds"`
	MemoryUsageMB    uint64    `json:"memory_usage_mb"`
	CPUUsagePercent  float64   `json:"cpu_usage_percent"`
	DiskUsagePercent float64   `json:"disk_usage_percent"`
	Goroutines       int       `json:"goroutines"`
	Timestamp        time.Time `json:"timestamp"`
}

// PerformanceMetrics - API performance metrics
type PerformanceMetrics struct {
	APIRequestsTotal         int64     `json:"api_requests_total"`
	APIResponseTimeP95       int       `json:"api_response_time_p95"`       // milliseconds
	APIErrorRate             float64   `json:"api_error_rate"`              // percentage
	DatabaseQueryDurationAvg float64   `json:"database_query_duration_avg"` // milliseconds
	Timestamp                time.Time `json:"timestamp"`
	CacheHit                 bool      `json:"cache_hit,omitempty"`
}

// AuthMetrics - Authentication metrics
type AuthMetrics struct {
	LoginAttemptsTotal  int64     `json:"login_attempts_total"`
	LoginSuccessRate    float64   `json:"login_success_rate"` // percentage
	FailedLoginAttempts int64     `json:"failed_login_attempts"`
	TokenIssuedTotal    int64     `json:"token_issued_total"`
	Timestamp           time.Time `json:"timestamp"`
	CacheHit            bool      `json:"cache_hit,omitempty"`
}

// AllMetrics - Aggregated metrics response
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
