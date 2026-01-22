// Package health provides health check functionality for the application.
// It manages multiple component health checks and aggregates their results.
package health

import (
	"context"
	"sync"
	"time"
)

// HealthChecker manages multiple component health checks
type HealthChecker struct {
	checkers []ComponentChecker
	service  string
	version  string
}

// NewHealthChecker creates a new health checker with the given components
func NewHealthChecker(service, version string, checkers ...ComponentChecker) *HealthChecker {
	return &HealthChecker{
		checkers: checkers,
		service:  service,
		version:  version,
	}
}

// CheckAll performs health checks on all registered components in parallel
func (h *HealthChecker) CheckAll(ctx context.Context) HealthReport {
	report := HealthReport{
		Timestamp:  time.Now(),
		Service:    h.service,
		Version:    h.version,
		Components: make(map[string]ComponentHealth),
	}

	// Use channel to collect results from parallel checks
	type result struct {
		name     string
		health   ComponentHealth
		critical bool
	}

	results := make(chan result, len(h.checkers))
	var wg sync.WaitGroup

	// Launch parallel health checks
	for _, checker := range h.checkers {
		wg.Add(1)
		go func(c ComponentChecker) {
			defer wg.Done()

			checkResult := c.Check(ctx)
			results <- result{
				name:     c.Name(),
				health:   ComponentHealth(checkResult),
				critical: c.IsCritical(),
			}
		}(checker)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	hasCriticalFailure := false
	hasNonCriticalFailure := false

	for res := range results {
		report.Components[res.name] = res.health

		if res.health.Status == StatusUnhealthy {
			if res.critical {
				hasCriticalFailure = true
			} else {
				hasNonCriticalFailure = true
			}
		}
	}

	// Calculate overall status
	report.Status = calculateOverallStatus(hasCriticalFailure, hasNonCriticalFailure)

	return report
}

// calculateOverallStatus determines the overall system status based on component failures
func calculateOverallStatus(hasCriticalFailure, hasNonCriticalFailure bool) Status {
	if hasCriticalFailure {
		return StatusUnhealthy
	}
	if hasNonCriticalFailure {
		return StatusDegraded
	}
	return StatusHealthy
}
