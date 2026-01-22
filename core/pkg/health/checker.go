// Package health provides health check functionality for the application.
// It includes interfaces for component health checkers and the core health checking logic.
package health

import "context"

// ComponentChecker defines the interface for checking component health
type ComponentChecker interface {
	// Name returns the component name (e.g., "database", "cache")
	Name() string

	// Check performs the health check and returns the result
	Check(ctx context.Context) CheckResult

	// IsCritical indicates whether this component is critical for system operation
	// Critical components cause overall status to be "unhealthy" if they fail
	// Non-critical components cause overall status to be "degraded" if they fail
	IsCritical() bool
}
