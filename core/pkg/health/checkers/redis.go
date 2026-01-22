// Package checkers provides health check implementations for various system components.
package checkers

import (
	"context"
	"fmt"
	"time"

	"apprun/pkg/cache"
	"apprun/pkg/health"
)

// RedisChecker checks Redis cache connectivity and health
type RedisChecker struct {
	client  cache.Client
	timeout time.Duration
}

// NewRedisChecker creates a new Redis health checker
func NewRedisChecker(client cache.Client) *RedisChecker {
	return &RedisChecker{
		client:  client,
		timeout: 2 * time.Second,
	}
}

// Name returns the component name
func (c *RedisChecker) Name() string {
	return "cache"
}

// Check performs the Redis health check
func (c *RedisChecker) Check(ctx context.Context) health.CheckResult {
	// Create a timeout context for this check
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	start := time.Now()
	err := c.client.Ping(checkCtx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return health.CheckResult{
			Status:    health.StatusUnhealthy,
			LatencyMs: latency,
			Message:   fmt.Sprintf("Redis ping failed: %v", err),
		}
	}

	return health.CheckResult{
		Status:    health.StatusHealthy,
		LatencyMs: latency,
		Message:   "Connected to Redis",
	}
}

// IsCritical indicates that cache is non-critical (system can operate without it)
func (c *RedisChecker) IsCritical() bool {
	return false
}
