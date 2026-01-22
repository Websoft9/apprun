package metrics

import (
	"context"
	"encoding/json"
	"time"

	"apprun/ent"
	"apprun/pkg/cache"
	"apprun/pkg/errors"
	"apprun/pkg/metricstore"
	"apprun/pkg/metricstore/storage"
)

// MetricsService provides metrics with caching support
type MetricsService struct {
	collector *MetricsCollector
	cache     cache.Client
	repo      *metricstore.Repository // Story 9.1: For history queries
}

// NewMetricsService creates a new metrics service instance
func NewMetricsService(entClient *ent.Client, cacheClient cache.Client, repo *metricstore.Repository) *MetricsService {
	return &MetricsService{
		collector: NewMetricsCollector(entClient, repo),
		cache:     cacheClient,
		repo:      repo,
	}
}

// GetUserMetrics retrieves user metrics with caching
func (s *MetricsService) GetUserMetrics(ctx context.Context) (*UserMetrics, error) {
	// Try cache first
	if cached, err := s.cache.Get(ctx, CacheKeyUserMetrics); err == nil && cached != "" {
		var metrics UserMetrics
		if err := json.Unmarshal([]byte(cached), &metrics); err == nil {
			metrics.CacheHit = true
			return &metrics, nil
		}
	}

	// Cache miss - collect fresh data
	metrics, err := s.collector.CollectUserMetrics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to collect user metrics")
	}

	// Store in cache
	if data, marshalErr := json.Marshal(metrics); marshalErr == nil {
		// Ignore cache write errors - cache is optional
		_ = s.cache.Set(ctx, CacheKeyUserMetrics, string(data), MetricsCacheTTL) //nolint:errcheck // Cache write is best-effort
	}

	metrics.CacheHit = false
	return metrics, nil
}

// GetSystemMetrics retrieves system health metrics (no caching for real-time data)
func (s *MetricsService) GetSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	return s.collector.CollectSystemMetrics(ctx)
}

// GetAuthMetrics retrieves authentication metrics with caching
func (s *MetricsService) GetAuthMetrics(ctx context.Context) (*AuthMetrics, error) {
	// Try cache first
	if cached, err := s.cache.Get(ctx, CacheKeyAuthMetrics); err == nil && cached != "" {
		var metrics AuthMetrics
		if err := json.Unmarshal([]byte(cached), &metrics); err == nil {
			metrics.CacheHit = true
			return &metrics, nil
		}
	}

	// Cache miss - collect fresh data
	metrics, err := s.collector.CollectAuthMetrics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to collect auth metrics")
	}

	// Store in cache
	if data, marshalErr := json.Marshal(metrics); marshalErr == nil {
		_ = s.cache.Set(ctx, CacheKeyAuthMetrics, string(data), MetricsCacheTTL) //nolint:errcheck // Cache write is best-effort
	}

	metrics.CacheHit = false
	return metrics, nil
}

// GetPerformanceMetrics retrieves performance metrics with caching
func (s *MetricsService) GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	// Try cache first
	if cached, err := s.cache.Get(ctx, CacheKeyPerfMetrics); err == nil && cached != "" {
		var metrics PerformanceMetrics
		if err := json.Unmarshal([]byte(cached), &metrics); err == nil {
			metrics.CacheHit = true
			return &metrics, nil
		}
	}

	// Cache miss - collect fresh data
	metrics, err := s.collector.CollectPerformanceMetrics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to collect performance metrics")
	}

	// Store in cache
	if data, marshalErr := json.Marshal(metrics); marshalErr == nil {
		_ = s.cache.Set(ctx, CacheKeyPerfMetrics, string(data), MetricsCacheTTL) //nolint:errcheck // Cache write is best-effort
	}

	metrics.CacheHit = false
	return metrics, nil
}

// GetAllMetrics retrieves all metrics with caching
func (s *MetricsService) GetAllMetrics(ctx context.Context) (*AllMetrics, error) {
	// Try cache first
	if cached, err := s.cache.Get(ctx, CacheKeyAllMetrics); err == nil && cached != "" {
		var metrics AllMetrics
		if err := json.Unmarshal([]byte(cached), &metrics); err == nil {
			metrics.CacheHit = true
			return &metrics, nil
		}
	}

	// Cache miss - collect fresh data
	metrics, err := s.collector.CollectAllMetrics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to collect all metrics")
	}

	// Store in cache
	if data, marshalErr := json.Marshal(metrics); marshalErr == nil {
		_ = s.cache.Set(ctx, CacheKeyAllMetrics, string(data), MetricsCacheTTL) //nolint:errcheck // Cache write is best-effort
	}

	metrics.CacheHit = false
	return metrics, nil
}

// GetMetricsHistory retrieves historical metrics from storage (Story 9.1)
func (s *MetricsService) GetMetricsHistory(ctx context.Context, name string, start, end time.Time, limit int, tags map[string]string) (*HistoryResponse, error) {
	// Query storage layer (BadgerDB)
	if s.repo == nil {
		return nil, errors.New(errors.ErrCodeInternalError, "metrics repository not initialized")
	}

	results, err := s.repo.GetMetricsByRange(ctx, name, start, end, limit)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to query metrics history")
	}

	// Filter by tags if provided
	var filtered []storage.Metric
	for _, m := range results {
		if matchesTags(m.Tags, tags) {
			filtered = append(filtered, m)
		}
	}

	// Convert storage.Metric to MetricPoint
	points := make([]MetricPoint, len(filtered))
	for i, m := range filtered {
		points[i] = MetricPoint{
			Name:      m.Name,
			Value:     m.Value,
			Tags:      m.Tags,
			Timestamp: m.Timestamp,
		}
	}

	return &HistoryResponse{
		Metrics: points,
		Count:   len(points),
		Start:   start,
		End:     end,
		HasMore: len(points) >= limit, // Simple pagination indicator
	}, nil
}

// matchesTags checks if metric tags match the filter tags (all filter tags must be present)
func matchesTags(metricTags, filterTags map[string]string) bool {
	if len(filterTags) == 0 {
		return true // No filter, match all
	}
	for key, value := range filterTags {
		if metricTags[key] != value {
			return false
		}
	}
	return true
}

// GetSnapshot retrieves aggregated metrics snapshot based on scope
func (s *MetricsService) GetSnapshot(ctx context.Context, scope string) (*SnapshotResponse, error) {
	timestamp := time.Now()
	metrics := make(map[string]interface{})
	var cacheHit bool

	switch scope {
	case "system":
		systemMetrics, err := s.GetSystemMetrics(ctx)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to get system metrics")
		}
		metrics["system"] = systemMetrics
		cacheHit = false // System metrics are never cached

	case "users":
		userMetrics, err := s.GetUserMetrics(ctx)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to get user metrics")
		}
		metrics["users"] = userMetrics
		cacheHit = userMetrics.CacheHit

	case "all":
		// Get all metrics
		userMetrics, userErr := s.GetUserMetrics(ctx)
		systemMetrics, sysErr := s.GetSystemMetrics(ctx)
		authMetrics, authErr := s.GetAuthMetrics(ctx)
		perfMetrics, perfErr := s.GetPerformanceMetrics(ctx)

		if userErr != nil || sysErr != nil || authErr != nil || perfErr != nil {
			return nil, errors.New(errors.ErrCodeInternalError, "failed to collect all metrics")
		}

		metrics["users"] = userMetrics
		metrics["system"] = systemMetrics
		metrics["auth"] = authMetrics
		metrics["performance"] = perfMetrics
		cacheHit = userMetrics.CacheHit || authMetrics.CacheHit || perfMetrics.CacheHit

	default:
		return nil, errors.New(errors.ErrCodeInvalidParam, "invalid scope")
	}

	return &SnapshotResponse{
		Scope:     scope,
		Metrics:   metrics,
		Timestamp: timestamp,
		CacheHit:  cacheHit,
	}, nil
}

// GetMetricKeys returns all available metric keys for history queries
func (s *MetricsService) GetMetricKeys(ctx context.Context, source string) (*KeysResponse, error) {
	// Predefined system metrics
	systemKeys := []MetricKey{
		{Name: "user_count_total", Source: "system", Description: "Total user count", Unit: "count"},
		{Name: "user_count_active", Source: "system", Description: "Active user count", Unit: "count"},
		{Name: "user_count_admin", Source: "system", Description: "Admin user count", Unit: "count"},
		{Name: "user_count_banned", Source: "system", Description: "Banned user count", Unit: "count"},
		{Name: "system_cpu_percent", Source: "system", Description: "CPU usage", Unit: "percent"},
		{Name: "system_memory_mb", Source: "system", Description: "Memory usage", Unit: "megabytes"},
		{Name: "system_disk_percent", Source: "system", Description: "Disk usage", Unit: "percent"},
		{Name: "system_goroutines", Source: "system", Description: "Number of goroutines", Unit: "count"},
		{Name: "api_requests_total", Source: "system", Description: "Total API requests", Unit: "count"},
		{Name: "api_response_time_p95", Source: "system", Description: "API response time P95", Unit: "milliseconds"},
		{Name: "api_error_rate", Source: "system", Description: "API error rate", Unit: "percent"},
		{Name: "auth_login_attempts", Source: "system", Description: "Login attempts", Unit: "count"},
		{Name: "auth_login_success_rate", Source: "system", Description: "Login success rate", Unit: "percent"},
		{Name: "auth_failed_logins", Source: "system", Description: "Failed login attempts", Unit: "count"},
		{Name: "auth_tokens_issued", Source: "system", Description: "JWT tokens issued", Unit: "count"},
	}

	// Filter by source if specified
	var filteredKeys []MetricKey
	if source == "" {
		filteredKeys = systemKeys
	} else {
		for _, key := range systemKeys {
			if key.Source == source {
				filteredKeys = append(filteredKeys, key)
			}
		}
	}

	// TODO: In the future, query user-defined metrics from storage
	// if source == "user" || source == "" {
	//   userKeys := s.repo.GetUserDefinedMetrics(ctx)
	//   filteredKeys = append(filteredKeys, userKeys...)
	// }

	return &KeysResponse{
		Keys:  filteredKeys,
		Count: len(filteredKeys),
	}, nil
}

// GetScopes returns all available snapshot scopes
func (s *MetricsService) GetScopes(ctx context.Context) (*ScopesResponse, error) {
	scopes := []ScopeDefinition{
		{
			Scope:       "users",
			Metrics:     []string{"total_users", "active_users", "admin_users", "banned_users", "new_users_today", "registrations_last_7d"},
			Description: "User-related statistics and demographics",
		},
		{
			Scope:       "system",
			Metrics:     []string{"uptime_seconds", "memory_usage_mb", "cpu_usage_percent", "disk_usage_percent", "goroutines"},
			Description: "System resource metrics (CPU, memory, disk, goroutines)",
		},
		{
			Scope:       "auth",
			Metrics:     []string{"login_attempts_total", "login_success_rate", "failed_login_attempts", "token_issued_total"},
			Description: "Authentication and security metrics",
		},
		{
			Scope:       "performance",
			Metrics:     []string{"api_requests_total", "api_response_time_p95", "api_error_rate", "database_query_duration_avg"},
			Description: "API performance and response time metrics",
		},
	}

	return &ScopesResponse{
		Scopes: scopes,
		Count:  len(scopes),
	}, nil
}

// IngestBatch processes a batch of metric ingestion requests
func (s *MetricsService) IngestBatch(ctx context.Context, requests []IngestRequest) (*IngestBatchResponse, error) {
	if s.repo == nil {
		return nil, errors.New(errors.ErrCodeInternalError, "metrics repository not initialized")
	}

	ingested := 0
	failed := 0
	var failedMetrics []string

	for _, req := range requests {
		// Validate metric
		if req.Name == "" {
			failed++
			failedMetrics = append(failedMetrics, "unnamed_metric")
			continue
		}

		// Ensure tags map exists
		if req.Tags == nil {
			req.Tags = make(map[string]string)
		}

		// Record metric
		err := s.repo.RecordMetric(ctx, req.Name, req.Value, req.Tags)
		if err != nil {
			failed++
			failedMetrics = append(failedMetrics, req.Name)
			continue
		}

		ingested++
	}

	message := "All metrics ingested successfully"
	if failed > 0 {
		message = "Partial success: some metrics failed to ingest"
	}

	return &IngestBatchResponse{
		Success:       failed == 0,
		Ingested:      ingested,
		Failed:        failed,
		FailedMetrics: failedMetrics,
		Message:       message,
	}, nil
}
