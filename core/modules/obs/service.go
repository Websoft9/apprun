package obs

import (
	"context"
	"encoding/json"
	"time"

	"apprun/ent"
	"apprun/pkg/cache"
	"apprun/pkg/errors"
	"apprun/pkg/metrics"
)

// MetricsService provides metrics with caching support
type MetricsService struct {
	collector *MetricsCollector
	cache     cache.Client
	repo      *metrics.Repository // Story 9.1: For history queries
}

// NewMetricsService creates a new metrics service instance
func NewMetricsService(entClient *ent.Client, cacheClient cache.Client, repo *metrics.Repository) *MetricsService {
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
func (s *MetricsService) GetMetricsHistory(ctx context.Context, name string, start, end time.Time, limit int) (*HistoryResponse, error) {
	// Query storage layer (BadgerDB)
	if s.repo == nil {
		return nil, errors.New(errors.ErrCodeInternalError, "metrics repository not initialized")
	}

	results, err := s.repo.GetMetricsByRange(ctx, name, start, end, limit)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to query metrics history")
	}

	// Convert storage.Metric to MetricPoint
	points := make([]MetricPoint, len(results))
	for i, m := range results {
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
