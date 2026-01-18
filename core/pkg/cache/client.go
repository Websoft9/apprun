// Package cache provides a unified Redis cache interface with fail-open design.
//
// This package implements an anti-corruption layer over go-redis/v9, providing:
//   - Key-value operations (Set/Get/Delete/Exists/TTL/Expire)
//   - Pub/Sub messaging (Publish/Subscribe with goroutine management)
//   - Connection pooling and health checks
//   - Graceful error handling with structured logging
//   - Environment-based configuration via REDIS_* env vars
//   - Fail-open strategy for business continuity
//
// Example usage:
//
//	cfg := cache.DefaultConfig()
//	client, err := cache.NewClient(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	ctx := context.Background()
//	err = client.Set(ctx, "user:123", "john@example.com", 5*time.Minute)
//	if err != nil {
//	    log.Printf("Cache error: %v", err)
//	}
//
//	val, err := client.Get(ctx, "user:123")
//	if err == cache.ErrKeyNotFound {
//	    log.Println("Key not found")
//	} else if err != nil {
//	    log.Printf("Cache error: %v", err)
//	} else {
//	    log.Printf("Value: %s", val)
//	}
//
// For testing, use the MockClient:
//
//	mock := cache.NewMockClient()
//	defer mock.Reset()
//	mock.Set(ctx, "key", "value", 0)
//
// See README.md for comprehensive documentation, best practices, and examples.
package cache

import (
	"context"
	"time"
)

// Client defines the cache interface for key-value and pub/sub operations.
// Implementations must handle errors gracefully with fail-open strategy by default.
type Client interface {
	// Key-Value Operations

	// Set stores a value with optional TTL (0 = no expiration).
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Get retrieves a value by key. Returns ErrKeyNotFound if key doesn't exist.
	Get(ctx context.Context, key string) (string, error)

	// Delete removes a key. Returns nil if key doesn't exist.
	Delete(ctx context.Context, keys ...string) error

	// Exists checks if a key exists.
	Exists(ctx context.Context, key string) (bool, error)

	// TTL returns the remaining TTL for a key (-2 = key doesn't exist, -1 = no expiration).
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Expire sets a new TTL for an existing key.
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// Pub/Sub Operations

	// Publish sends a message to a channel.
	Publish(ctx context.Context, channel string, message interface{}) error

	// Subscribe listens to one or more channels and invokes the callback for each message.
	// Returns a cancel function to stop the subscription.
	Subscribe(ctx context.Context, callback func(channel, message string), channels ...string) (cancel func() error, err error)

	// Health & Lifecycle

	// Ping checks if the cache is reachable.
	Ping(ctx context.Context) error

	// Close releases all resources.
	Close() error
}
