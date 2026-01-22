// Package checkers provides health check implementations for various system components.
package checkers

import (
	"context"
	"time"

	"apprun/pkg/health"
)

// MetricsChecker checks metrics storage health
type MetricsChecker struct {
	enabled bool
	timeout time.Duration
}

// NewMetricsChecker creates a new metrics health checker
func NewMetricsChecker(enabled bool) *MetricsChecker {
	return &MetricsChecker{
		enabled: enabled,
		timeout: 2 * time.Second,
	}
}

// Name returns the component name
func (c *MetricsChecker) Name() string {
	return "metrics_storage"
}

// Check performs the metrics storage health check
func (c *MetricsChecker) Check(ctx context.Context) health.CheckResult {
	if !c.enabled {
		return health.CheckResult{
			Status:    health.StatusHealthy,
			LatencyMs: 0,
			Message:   "Metrics storage disabled",
		}
	}

	// Create a timeout context for this check
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	start := time.Now()

	// TODO: Implement actual metrics storage check when metrics module is available
	// For now, return healthy with a note that actual check is pending
	_ = checkCtx // Use context to avoid lint warning
	latency := time.Since(start).Milliseconds()

	return health.CheckResult{
		Status:    health.StatusHealthy,
		LatencyMs: latency,
		Message:   "Metrics storage enabled (detailed check pending)",
	}
}

// IsCritical indicates that metrics storage is non-critical
func (c *MetricsChecker) IsCritical() bool {
	return false
}
