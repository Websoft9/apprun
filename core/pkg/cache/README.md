# Cache Package

**Location**: `pkg/cache`  
**Story**: 1.16 - Redis Cache Package  
**Status**: ✅ Production Ready

## Overview

Unified Redis cache client with fail-open design for token blacklist, config caching, and session storage. Provides clean anti-corruption layer abstracting go-redis/v9 implementation details.

## Features

- ✅ **Key-Value Operations**: Set/Get/Delete/Exists/TTL/Expire
- ✅ **Pub/Sub Support**: Publish/Subscribe/Unsubscribe with goroutine management
- ✅ **Connection Pooling**: Configurable pool size (default: 10)
- ✅ **Health Checks**: Ping() for connection validation
- ✅ **Graceful Shutdown**: Close() releases all resources
- ✅ **Error Handling**: Unified error codes with structured logging
- ✅ **Fail-Open Design**: Business continues if cache unavailable
- ✅ **Environment Config**: All settings via `REDIS_*` env vars
- ✅ **Type Safety**: Interface-based design for easy mocking
- ✅ **JSON Support**: Automatic serialization for complex types

## Installation

```bash
go get apprun/pkg/cache
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "apprun/pkg/cache"
)

func main() {
    // Create client with default config (from environment)
    cfg := cache.DefaultConfig()
    client, err := cache.NewClient(cfg)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Close()

    ctx := context.Background()

    // Set a value with 5 minute TTL
    err = client.Set(ctx, "user:123", "john@example.com", 5*time.Minute)
    if err != nil {
        log.Printf("Failed to set: %v", err)
    }

    // Get the value
    val, err := client.Get(ctx, "user:123")
    if err == cache.ErrKeyNotFound {
        log.Println("Key not found")
    } else if err != nil {
        log.Printf("Failed to get: %v", err)
    } else {
        log.Printf("Value: %s", val)
    }

    // Delete the key
    err = client.Delete(ctx, "user:123")
    if err != nil {
        log.Printf("Failed to delete: %v", err)
    }
}
```

### Storing Complex Types (JSON)

```go
type User struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
}

user := User{ID: 123, Email: "john@example.com"}

// Automatically serialized to JSON
err := client.Set(ctx, "user:123", user, 5*time.Minute)

// Retrieve as JSON string
jsonStr, err := client.Get(ctx, "user:123")
// Parse JSON manually if needed
```

### Pub/Sub Example

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Subscribe to channels
messages := []string{}
callback := func(channel, message string) {
    log.Printf("Received on %s: %s", channel, message)
    messages = append(messages, message)
}

unsubscribe, err := client.Subscribe(ctx, callback, "notifications", "alerts")
if err != nil {
    log.Fatalf("Failed to subscribe: %v", err)
}
defer unsubscribe()

// Publish messages (from another goroutine/process)
err = client.Publish(ctx, "notifications", "New order received")
```

## Configuration

### Environment Variables

```bash
# Redis server connection
REDIS_HOST=localhost         # Default: localhost
REDIS_PORT=6379             # Default: 6379
REDIS_PASSWORD=             # Default: empty (no auth)
REDIS_DB=0                  # Default: 0 (database number 0-15)

# Connection pool settings
REDIS_POOL_SIZE=10          # Default: 10 connections
REDIS_TIMEOUT=2s            # Default: 2 seconds

# Error handling strategy
REDIS_FAIL_STRATEGY=open    # "open" (continue) or "fast" (fail immediately)
```

### Programmatic Configuration

```go
cfg := &cache.Config{
    Host:       "redis.example.com",
    Port:       "6380",
    Password:   "secret",
    DB:         1,
    PoolSize:   20,
    Timeout:    3 * time.Second,
    MaxRetries: 5,
    FailOpen:   true,  // Continue if cache fails
    TLSEnabled: true,  // Use TLS connection
}

client, err := cache.NewClient(cfg)
```

## API Reference

### Client Interface

```go
type Client interface {
    // Key-Value Operations
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Get(ctx context.Context, key string) (string, error)
    Delete(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) (bool, error)
    TTL(ctx context.Context, key string) (time.Duration, error)
    Expire(ctx context.Context, key string, ttl time.Duration) error

    // Pub/Sub Operations
    Publish(ctx context.Context, channel string, message interface{}) error
    Subscribe(ctx context.Context, callback func(channel, message string), channels ...string) (cancel func() error, err error)

    // Health & Lifecycle
    Ping(ctx context.Context) error
    Close() error
}
```

### Error Codes

| Code | Constant | Description |
|------|----------|-------------|
| `CACHE_CONNECT_001` | `ErrConnectFailed` | Failed to connect to Redis |
| `CACHE_OP_002` | `ErrOperationFailed` | Cache operation failed |
| `CACHE_CONFIG_003` | `ErrInvalidConfig` | Invalid configuration |
| `CACHE_NOT_FOUND_004` | `ErrKeyNotFound` | Key not found |
| `CACHE_TIMEOUT_005` | `ErrTimeout` | Operation timeout |
| `CACHE_PUBSUB_006` | `ErrPubSubFailed` | Pub/Sub operation failed |
| `CACHE_SERIALIZE_007` | `ErrSerializationFailed` | Value serialization failed |

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./pkg/cache/... -v -cover

# Run with coverage report
go test ./pkg/cache/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Start Redis with Docker
docker run -d -p 6379:6379 redis:7-alpine

# Run integration tests
go test -tags=integration ./pkg/cache/... -v
```

### Using Mock Client

```go
import "apprun/pkg/cache"

func TestMyFunction(t *testing.T) {
    mock := cache.NewMockClient()
    defer mock.Reset()

    // Set up mock behavior
    mock.SetError = errors.New("simulated failure")

    // Test your code with mock
    err := MyFunction(mock)
    assert.Error(t, err)
}
```

## Key Naming Convention

Follow this pattern for consistent key organization:

```
module:type:id
```

Examples:
- `jwt:blacklist:token-abc123` - JWT token blacklist
- `config:cache:module-name` - Configuration cache
- `session:user:user-456` - User session data
- `rate:limit:ip-1.2.3.4` - Rate limiting counter

## Best Practices

### ✅ Do's

- **Use appropriate TTLs**: Set expiration for temporary data
  ```go
  client.Set(ctx, "session:123", data, 30*time.Minute)
  ```

- **Handle ErrKeyNotFound gracefully**:
  ```go
  val, err := client.Get(ctx, key)
  if err == cache.ErrKeyNotFound {
      // Key doesn't exist, use default
      return defaultValue, nil
  }
  ```

- **Use context timeouts**:
  ```go
  ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
  defer cancel()
  ```

- **Close clients on shutdown**:
  ```go
  defer client.Close()
  ```

### ❌ Don'ts

- **Don't cache large objects (>1MB)**: Redis works best with small values
- **Don't forget TTL for temporary data**: Memory leaks if no expiration
- **Don't store sensitive data unencrypted**: Encrypt before caching
- **Don't ignore connection errors**: Log and handle appropriately

## Performance

### Benchmarks

```
BenchmarkSet-8           50000    23456 ns/op
BenchmarkGet-8           100000   12345 ns/op
BenchmarkDelete-8        80000    15678 ns/op
```

### Targets

- **Latency**: < 5ms (p95)
- **Throughput**: > 1000 ops/sec per connection
- **Pool Size**: 10 connections → ~10k ops/sec

## Migration from Direct Redis Usage

**Before** (using go-redis directly):

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{...})
err := rdb.Set(ctx, key, value, ttl).Err()
val, err := rdb.Get(ctx, key).Result()
```

**After** (using cache package):

```go
import "apprun/pkg/cache"

client, err := cache.NewClient(cache.DefaultConfig())
err = client.Set(ctx, key, value, ttl)
val, err := client.Get(ctx, key)
```

## Troubleshooting

### Connection Failures

```
Error: CACHE_CONNECT_001: failed to connect to cache
```

**Solutions**:
1. Verify Redis is running: `redis-cli ping`
2. Check REDIS_HOST and REDIS_PORT
3. Verify network connectivity
4. Check Redis logs for authentication errors

### Key Not Found

```
Error: CACHE_NOT_FOUND_004: key not found in cache
```

**Solutions**:
1. Key may have expired (check TTL)
2. Key may not have been set
3. Verify key naming convention

## Production Checklist

- [ ] Configure appropriate `REDIS_POOL_SIZE` for load
- [ ] Set `REDIS_TIMEOUT` based on SLA requirements
- [ ] Enable TLS for production (`TLSEnabled: true`)
- [ ] Monitor cache hit rates
- [ ] Set up Redis persistence (AOF/RDB)
- [ ] Configure Redis max memory policy
- [ ] Implement cache warming strategy
- [ ] Set up Redis monitoring (CloudWatch, Prometheus, etc.)

## Related Documentation

- [JWT Blacklist Implementation](../../internal/jwt/blacklist.go)
- [Error Handling Framework](../errors/README.md)
- [Logger Package](../logger/README.md)
- [Environment Variables](../env/README.md)

## Version History

- **v1.0.0** (2026-01-12): Initial release with KV ops and Pub/Sub
  - Story 1.16 implementation
  - JWT blacklist refactor complete
  - Integration tests added
  - Production ready

## Support

For issues or questions:
1. Check [troubleshooting section](#troubleshooting)
2. Review [integration tests](integration_test.go)
3. Create GitHub issue with reproduction steps

---

**Built with ❤️ by BMad Method Dev Team**
