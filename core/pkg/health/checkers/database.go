// Package checkers provides health check implementations for various system components.
package checkers

import (
	"context"
	"fmt"
	"time"

	"apprun/ent"
	"apprun/pkg/health"
)

// DatabaseChecker checks database connectivity and health
type DatabaseChecker struct {
	client  *ent.Client
	timeout time.Duration
}

// NewDatabaseChecker creates a new database health checker
func NewDatabaseChecker(client *ent.Client) *DatabaseChecker {
	return &DatabaseChecker{
		client:  client,
		timeout: 2 * time.Second,
	}
}

// Name returns the component name
func (c *DatabaseChecker) Name() string {
	return "database"
}

// Check performs the database health check
func (c *DatabaseChecker) Check(ctx context.Context) health.CheckResult {
	// Create a timeout context for this check
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	start := time.Now()
	// Note: Using User.Query().Limit(1).Count() as a health check
	// Rationale:
	// 1. Ent doesn't expose raw DB ping (driver field is private)
	// 2. User table is a core system table guaranteed to exist after migrations
	// 3. Limit(1) minimizes query cost (no full table scan)
	// 4. This is better than User.Query().Count() which could scan large tables
	_, err := c.client.User.Query().Limit(1).Count(checkCtx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return health.CheckResult{
			Status:    health.StatusUnhealthy,
			LatencyMs: latency,
			Message:   fmt.Sprintf("Database ping failed: %v", err),
		}
	}

	return health.CheckResult{
		Status:    health.StatusHealthy,
		LatencyMs: latency,
		Message:   "Connected to PostgreSQL",
	}
}

// IsCritical indicates that database is critical for system operation
func (c *DatabaseChecker) IsCritical() bool {
	return true
}
