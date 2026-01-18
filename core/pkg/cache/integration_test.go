//go:build integration
// +build integration

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests require a running Redis instance on localhost:6379
// Run with: go test -tags=integration ./pkg/cache/...

func TestRedisClient_Integration(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     "6379",
		DB:       1, // Use DB 1 for testing to avoid conflicts
		PoolSize: 5,
		Timeout:  2 * time.Second,
		FailOpen: true,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err, "Failed to connect to Redis")
	defer client.Close()

	ctx := context.Background()

	// Test Ping
	err = client.Ping(ctx)
	require.NoError(t, err)

	// Clean up test keys
	defer client.Delete(ctx, "test:key1", "test:key2", "test:key3")

	// Test Set and Get
	err = client.Set(ctx, "test:key1", "value1", 0)
	require.NoError(t, err)

	val, err := client.Get(ctx, "test:key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Test Get non-existent key
	_, err = client.Get(ctx, "test:nonexistent")
	assert.Equal(t, ErrKeyNotFound, err)

	// Test Set with TTL
	err = client.Set(ctx, "test:key2", "value2", 1*time.Second)
	require.NoError(t, err)

	val, err = client.Get(ctx, "test:key2")
	require.NoError(t, err)
	assert.Equal(t, "value2", val)

	// Wait for expiration
	time.Sleep(1100 * time.Millisecond)
	_, err = client.Get(ctx, "test:key2")
	assert.Equal(t, ErrKeyNotFound, err)

	// Test Exists
	err = client.Set(ctx, "test:key3", "value3", 0)
	require.NoError(t, err)

	exists, err := client.Exists(ctx, "test:key3")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = client.Exists(ctx, "test:nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test Delete
	err = client.Delete(ctx, "test:key3")
	require.NoError(t, err)

	exists, err = client.Exists(ctx, "test:key3")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test TTL
	err = client.Set(ctx, "test:key4", "value4", 10*time.Second)
	require.NoError(t, err)

	ttl, err := client.TTL(ctx, "test:key4")
	require.NoError(t, err)
	assert.Greater(t, ttl, 5*time.Second)
	assert.LessOrEqual(t, ttl, 10*time.Second)

	// Test Expire
	err = client.Expire(ctx, "test:key4", 20*time.Second)
	require.NoError(t, err)

	ttl, err = client.TTL(ctx, "test:key4")
	require.NoError(t, err)
	assert.Greater(t, ttl, 15*time.Second)

	// Clean up
	client.Delete(ctx, "test:key4")
}

func TestRedisClient_PubSub_Integration(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     "6379",
		DB:       1,
		PoolSize: 5,
		Timeout:  2 * time.Second,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Track received messages
	messages := make([]string, 0)
	callback := func(channel, message string) {
		messages = append(messages, message)
	}

	// Subscribe to channel
	unsubscribe, err := client.Subscribe(ctx, callback, "test-channel")
	require.NoError(t, err)
	defer unsubscribe()

	// Give subscription time to establish
	time.Sleep(100 * time.Millisecond)

	// Publish messages
	err = client.Publish(ctx, "test-channel", "message1")
	require.NoError(t, err)

	err = client.Publish(ctx, "test-channel", "message2")
	require.NoError(t, err)

	// Give messages time to be received
	time.Sleep(100 * time.Millisecond)

	// Verify messages received
	assert.GreaterOrEqual(t, len(messages), 2)
}

func TestRedisClient_JSON_Integration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DB = 1

	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	defer client.Delete(ctx, "test:json")

	// Test storing complex struct
	type TestData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	data := TestData{ID: 123, Name: "test"}
	err = client.Set(ctx, "test:json", data, 0)
	require.NoError(t, err)

	// Get raw JSON string
	val, err := client.Get(ctx, "test:json")
	require.NoError(t, err)
	assert.Contains(t, val, `"id":123`)
	assert.Contains(t, val, `"name":"test"`)
}
