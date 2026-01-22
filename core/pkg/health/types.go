// Package health provides health check functionality for the application.
// It includes types for health status, check results, and health reports.
package health

import "time"

// Status represents the health status of a component or system
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// CheckResult represents the result of a single health check
type CheckResult struct {
	Status    Status `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Message   string `json:"message"`
}

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Status    Status `json:"status" example:"healthy"`
	LatencyMs int64  `json:"latency_ms" example:"5"`
	Message   string `json:"message" example:"Connected to PostgreSQL"`
}

// HealthReport represents the overall system health report
type HealthReport struct {
	Status     Status                     `json:"status" example:"healthy"`
	Timestamp  time.Time                  `json:"timestamp" example:"2026-01-22T10:30:00Z"`
	Service    string                     `json:"service" example:"apprun"`
	Version    string                     `json:"version" example:"1.0.0"`
	Components map[string]ComponentHealth `json:"components"`
}

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}

// HTTPStatus returns the appropriate HTTP status code for the health status
func (s Status) HTTPStatus() int {
	switch s {
	case StatusHealthy:
		return 200
	case StatusUnhealthy:
		return 503
	case StatusDegraded:
		return 207
	default:
		return 500
	}
}
