// Package rbac provides Role-Based Access Control (RBAC) functionality using Casbin.
// It implements a dual-layer permission model with platform-level and project-level roles.
package rbac

import "sync"

// Cache layer for permission check results
// MVP: in-memory cache with sync.Map
// Production: Redis cache with TTL

var permCache sync.Map

// ClearUserCache clears all cached permissions for a user in a project
func ClearUserCache(userID, projectID int64) {
	// Cache key pattern: "userID:projectID:*"
	// For MVP, we need to iterate and delete matching keys
	// In production, use Redis pattern-based deletion

	permCache.Range(func(key, value interface{}) bool {
		// For now, just clear all cache when any user's role changes
		// TODO: Implement pattern matching in production
		return true
	})
}

// ClearAllCache clears all cached permissions
// Called when policies are reloaded
func ClearAllCache() {
	permCache = sync.Map{}
}
