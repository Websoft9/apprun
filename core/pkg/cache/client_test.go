package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClientWithNilConfig tests creating client with nil config uses defaults
func TestNewClientWithNilConfig(t *testing.T) {
	// This test will fail if Redis is not running, which is expected for unit tests
	// We're testing that it attempts to use default config
	_, err := NewClient(nil)
	// We expect either success (if Redis running) or connection error
	if err != nil {
		assert.Contains(t, err.Error(), "Redis connection failed")
	}
}

// TestNewClientWithInvalidConfig tests validation is called
func TestNewClientWithInvalidConfig(t *testing.T) {
	cfg := &Config{
		Host:     "", // Invalid: empty host
		Port:     "6379",
		DB:       0,
		PoolSize: 10,
		Timeout:  2 * time.Second,
	}

	_, err := NewClient(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "host cannot be empty")
}

// TestClientInterface verifies redisClient implements Client interface
func TestClientInterface(t *testing.T) {
	var _ Client = (*redisClient)(nil)
}

// TestMockClientInterface verifies MockClient implements Client interface
func TestMockClientInterface(t *testing.T) {
	var _ Client = (*MockClient)(nil)
}

// TestClient_SetGetDelete_String tests string operations using mock
func TestClient_SetGetDelete_String(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Set string value
	err := mock.Set(ctx, "user:1", "john", 0)
	require.NoError(t, err)

	// Get value
	val, err := mock.Get(ctx, "user:1")
	require.NoError(t, err)
	assert.Equal(t, "john", val)

	// Delete
	err = mock.Delete(ctx, "user:1")
	require.NoError(t, err)

	// Verify deleted
	_, err = mock.Get(ctx, "user:1")
	assert.Equal(t, ErrKeyNotFound, err)
}

// TestClient_SetWithTTL tests TTL functionality
func TestClient_SetWithTTL(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Set with 1 second TTL
	err := mock.Set(ctx, "temp:key", "value", 1*time.Second)
	require.NoError(t, err)

	// Should exist immediately
	exists, err := mock.Exists(ctx, "temp:key")
	require.NoError(t, err)
	assert.True(t, exists)

	// Check TTL
	ttl, err := mock.TTL(ctx, "temp:key")
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
}

// TestClient_ExpireExistingKey tests setting expiration on existing key
func TestClient_ExpireExistingKey(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Set key without expiration
	err := mock.Set(ctx, "persistent:key", "value", 0)
	require.NoError(t, err)

	// Verify no TTL
	ttl, err := mock.TTL(ctx, "persistent:key")
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1), ttl)

	// Set expiration
	err = mock.Expire(ctx, "persistent:key", 10*time.Second)
	require.NoError(t, err)

	// Verify TTL is set
	ttl, err = mock.TTL(ctx, "persistent:key")
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
}

// TestClient_DeleteMultipleKeys tests deleting multiple keys at once
func TestClient_DeleteMultipleKeys(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Set multiple keys
	mock.Set(ctx, "key1", "val1", 0)
	mock.Set(ctx, "key2", "val2", 0)
	mock.Set(ctx, "key3", "val3", 0)

	// Delete all at once
	err := mock.Delete(ctx, "key1", "key2", "key3")
	require.NoError(t, err)

	// Verify all deleted
	_, err = mock.Get(ctx, "key1")
	assert.Equal(t, ErrKeyNotFound, err)
	_, err = mock.Get(ctx, "key2")
	assert.Equal(t, ErrKeyNotFound, err)
	_, err = mock.Get(ctx, "key3")
	assert.Equal(t, ErrKeyNotFound, err)
}

// TestClient_PubSubMultipleChannels tests subscribing to multiple channels
func TestClient_PubSubMultipleChannels(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	receivedMessages := make(map[string][]string)
	callback := func(channel, message string) {
		receivedMessages[channel] = append(receivedMessages[channel], message)
	}

	// Subscribe to multiple channels
	cancel, err := mock.Subscribe(ctx, callback, "channel1", "channel2")
	require.NoError(t, err)
	defer cancel()

	// Publish to different channels
	mock.Publish(ctx, "channel1", "msg1")
	mock.Publish(ctx, "channel2", "msg2")
	mock.Publish(ctx, "channel1", "msg3")

	// Verify messages received on correct channels
	assert.Len(t, receivedMessages["channel1"], 2)
	assert.Len(t, receivedMessages["channel2"], 1)
	assert.Equal(t, "msg1", receivedMessages["channel1"][0])
	assert.Equal(t, "msg2", receivedMessages["channel2"][0])
	assert.Equal(t, "msg3", receivedMessages["channel1"][1])
}

// TestClient_PubSubNoChannels tests error when no channels provided
func TestClient_PubSubNoChannels(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	callback := func(channel, message string) {}

	_, err := mock.Subscribe(ctx, callback)
	require.Error(t, err)
	// Mock returns ErrPubSubFailed
	assert.Equal(t, ErrPubSubFailed, err)
}

// TestClient_TTLNonExistentKey tests TTL on non-existent key
func TestClient_TTLNonExistentKey(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	ttl, err := mock.TTL(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, -2*time.Second, ttl)
}

// TestClient_ExpireNonExistentKey tests expiring non-existent key (should not error)
func TestClient_ExpireNonExistentKey(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	err := mock.Expire(ctx, "nonexistent", 10*time.Second)
	assert.NoError(t, err) // Should not error on non-existent key
}

// TestClient_SetIntegerValue tests storing integer values
func TestClient_SetIntegerValue(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Set integer value
	err := mock.Set(ctx, "counter", 42, 0)
	require.NoError(t, err)

	// Get value - mock returns "mocked" for non-string values
	val, err := mock.Get(ctx, "counter")
	require.NoError(t, err)
	assert.NotEmpty(t, val)
}

// TestClient_PublishToNonExistentChannel tests publishing to channel with no subscribers
func TestClient_PublishToNonExistentChannel(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Should not error even if no subscribers
	err := mock.Publish(ctx, "empty-channel", "message")
	assert.NoError(t, err)
}

// TestMockClient_MultipleSubscribers tests multiple subscribers to same channel
func TestMockClient_MultipleSubscribers(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	messages1 := []string{}
	messages2 := []string{}

	callback1 := func(channel, message string) {
		messages1 = append(messages1, message)
	}
	callback2 := func(channel, message string) {
		messages2 = append(messages2, message)
	}

	// Both subscribe to same channel
	cancel1, err := mock.Subscribe(ctx, callback1, "test-channel")
	require.NoError(t, err)
	defer cancel1()

	cancel2, err := mock.Subscribe(ctx, callback2, "test-channel")
	require.NoError(t, err)
	defer cancel2()

	// Publish message
	mock.Publish(ctx, "test-channel", "broadcast")

	// Both should receive it
	assert.Contains(t, messages1, "broadcast")
	assert.Contains(t, messages2, "broadcast")
}
