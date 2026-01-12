package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	pkgerrors "apprun/pkg/errors"
	"apprun/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// redisClient implements the Client interface using go-redis/v9
type redisClient struct {
	client *redis.Client
	config *Config
	pubsub *redis.PubSub
}

// NewClient creates a new Redis cache client with the given configuration.
// It validates the config, establishes connection, and performs a health check.
func NewClient(cfg *Config) (Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Build Redis options
	opts := &redis.Options{
		Addr:         cfg.Address(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}

	// Add TLS if enabled
	if cfg.TLSEnabled {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	// Create Redis client
	rdb := redis.NewClient(opts)

	// Perform initial health check with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis",
			logger.Field{Key: "address", Value: cfg.Address()},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, pkgerrors.Wrap(err, CodeConnectFailed, "Redis connection failed")
	}

	logger.Info("Redis cache client connected",
		logger.Field{Key: "address", Value: cfg.Address()},
		logger.Field{Key: "db", Value: cfg.DB},
		logger.Field{Key: "pool_size", Value: cfg.PoolSize})

	return &redisClient{
		client: rdb,
		config: cfg,
	}, nil
}

// Set stores a value with optional TTL (0 = no expiration).
func (r *redisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Serialize value to JSON if not string
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		data = fmt.Sprintf("%d", v)
	default:
		// Serialize to JSON for complex types
		jsonData, err := json.Marshal(value)
		if err != nil {
			logger.Error("Failed to serialize value",
				logger.Field{Key: "key", Value: key},
				logger.Field{Key: "error", Value: err.Error()})
			return pkgerrors.Wrap(err, CodeSerializeFail, "value serialization failed")
		}
		data = string(jsonData)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		logger.Error("Cache SET operation failed",
			logger.Field{Key: "key", Value: key},
			logger.Field{Key: "error", Value: err.Error()})
		return pkgerrors.Wrap(err, CodeOperationFail, "SET operation failed")
	}

	logger.Debug("Cache SET successful",
		logger.Field{Key: "key", Value: key},
		logger.Field{Key: "ttl", Value: ttl.String()})
	return nil
}

// Get retrieves a value by key. Returns ErrKeyNotFound if key doesn't exist.
func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrKeyNotFound
	}
	if err != nil {
		logger.Error("Cache GET operation failed",
			logger.Field{Key: "key", Value: key},
			logger.Field{Key: "error", Value: err.Error()})
		return "", pkgerrors.Wrap(err, CodeOperationFail, "GET operation failed")
	}

	logger.Debug("Cache GET successful", logger.Field{Key: "key", Value: key})
	return val, nil
}

// Delete removes one or more keys. Returns nil if key doesn't exist.
func (r *redisClient) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		logger.Error("Cache DEL operation failed",
			logger.Field{Key: "keys", Value: keys},
			logger.Field{Key: "error", Value: err.Error()})
		return pkgerrors.Wrap(err, CodeOperationFail, "DEL operation failed")
	}

	logger.Debug("Cache DEL successful", logger.Field{Key: "keys", Value: keys})
	return nil
}

// Exists checks if a key exists.
func (r *redisClient) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		logger.Error("Cache EXISTS operation failed",
			logger.Field{Key: "key", Value: key},
			logger.Field{Key: "error", Value: err.Error()})
		return false, pkgerrors.Wrap(err, CodeOperationFail, "EXISTS operation failed")
	}

	logger.Debug("Cache EXISTS successful",
		logger.Field{Key: "key", Value: key},
		logger.Field{Key: "exists", Value: n > 0})
	return n > 0, nil
}

// TTL returns the remaining TTL for a key (-2 = key doesn't exist, -1 = no expiration).
func (r *redisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		logger.Error("Cache TTL operation failed",
			logger.Field{Key: "key", Value: key},
			logger.Field{Key: "error", Value: err.Error()})
		return 0, pkgerrors.Wrap(err, CodeOperationFail, "TTL operation failed")
	}

	logger.Debug("Cache TTL successful",
		logger.Field{Key: "key", Value: key},
		logger.Field{Key: "ttl", Value: ttl.String()})
	return ttl, nil
}

// Expire sets a new TTL for an existing key.
func (r *redisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		logger.Error("Cache EXPIRE operation failed",
			logger.Field{Key: "key", Value: key},
			logger.Field{Key: "ttl", Value: ttl.String()},
			logger.Field{Key: "error", Value: err.Error()})
		return pkgerrors.Wrap(err, CodeOperationFail, "EXPIRE operation failed")
	}

	logger.Debug("Cache EXPIRE successful",
		logger.Field{Key: "key", Value: key},
		logger.Field{Key: "ttl", Value: ttl.String()})
	return nil
}

// Publish sends a message to a channel.
func (r *redisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	// Serialize message to JSON if not string
	var data string
	switch v := message.(type) {
	case string:
		data = v
	default:
		jsonData, err := json.Marshal(message)
		if err != nil {
			logger.Error("Failed to serialize message",
				logger.Field{Key: "channel", Value: channel},
				logger.Field{Key: "error", Value: err.Error()})
			return pkgerrors.Wrap(err, CodeSerializeFail, "message serialization failed")
		}
		data = string(jsonData)
	}

	if err := r.client.Publish(ctx, channel, data).Err(); err != nil {
		logger.Error("Cache PUBLISH operation failed",
			logger.Field{Key: "channel", Value: channel},
			logger.Field{Key: "error", Value: err.Error()})
		return pkgerrors.Wrap(err, CodePubSubFailed, "PUBLISH operation failed")
	}

	logger.Debug("Cache PUBLISH successful", logger.Field{Key: "channel", Value: channel})
	return nil
}

// Subscribe listens to one or more channels and invokes the callback for each message.
// Returns a cancel function to stop the subscription.
//
// IMPORTANT: The returned cancel function MUST be called to prevent goroutine leaks.
// The goroutine will also exit automatically when ctx is canceled or the subscription closes.
//
// Example:
//
//	cancel, err := client.Subscribe(ctx, callback, "channel1")
//	if err != nil {
//	    return err
//	}
//	defer cancel() // CRITICAL: Always call cancel to clean up
func (r *redisClient) Subscribe(ctx context.Context, callback func(channel, message string), channels ...string) (cancel func() error, err error) {
	if len(channels) == 0 {
		return nil, pkgerrors.New(CodePubSubFailed, "no channels specified")
	}

	// Create a new pubsub instance
	pubsub := r.client.Subscribe(ctx, channels...)

	// Wait for subscription confirmation
	if _, err := pubsub.Receive(ctx); err != nil {
		logger.Error("Cache SUBSCRIBE operation failed",
			logger.Field{Key: "channels", Value: channels},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, pkgerrors.Wrap(err, CodePubSubFailed, "SUBSCRIBE operation failed")
	}

	logger.Info("Cache SUBSCRIBE successful", logger.Field{Key: "channels", Value: channels})

	// Start goroutine to receive messages
	ch := pubsub.Channel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Debug("Subscription context canceled", logger.Field{Key: "channels", Value: channels})
				return
			case msg, ok := <-ch:
				if !ok {
					logger.Debug("Subscription channel closed", logger.Field{Key: "channels", Value: channels})
					return
				}
				callback(msg.Channel, msg.Payload)
			}
		}
	}()

	// Return cancel function
	cancelFunc := func() error {
		if err := pubsub.Close(); err != nil {
			logger.Error("Failed to unsubscribe",
				logger.Field{Key: "channels", Value: channels},
				logger.Field{Key: "error", Value: err.Error()})
			return pkgerrors.Wrap(err, CodePubSubFailed, "unsubscribe failed")
		}
		logger.Info("Unsubscribed successfully", logger.Field{Key: "channels", Value: channels})
		return nil
	}

	return cancelFunc, nil
}

// Ping checks if the cache is reachable.
func (r *redisClient) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		logger.Error("Cache PING failed", logger.Field{Key: "error", Value: err.Error()})
		return pkgerrors.Wrap(err, CodeConnectFailed, "health check failed")
	}

	logger.Debug("Cache PING successful")
	return nil
}

// Close releases all resources.
func (r *redisClient) Close() error {
	// Defensive: check if already closed or never initialized
	if r.client == nil {
		return nil
	}

	if r.pubsub != nil {
		if err := r.pubsub.Close(); err != nil {
			logger.Warn("Failed to close pubsub", logger.Field{Key: "error", Value: err.Error()})
		}
	}

	if err := r.client.Close(); err != nil {
		logger.Error("Failed to close Redis client", logger.Field{Key: "error", Value: err.Error()})
		return pkgerrors.Wrap(err, CodeOperationFail, "close operation failed")
	}

	logger.Info("Redis cache client closed")
	return nil
}
