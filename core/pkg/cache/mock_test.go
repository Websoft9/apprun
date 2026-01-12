package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockClient_SetGet(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Test Set and Get
	err := mock.Set(ctx, "key1", "value1", 0)
	require.NoError(t, err)

	val, err := mock.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestMockClient_GetNotFound(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	val, err := mock.Get(ctx, "nonexistent")
	assert.Equal(t, ErrKeyNotFound, err)
	assert.Empty(t, val)
}

func TestMockClient_Delete(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Set a key
	err := mock.Set(ctx, "key1", "value1", 0)
	require.NoError(t, err)

	// Delete it
	err = mock.Delete(ctx, "key1")
	require.NoError(t, err)

	// Verify it's gone
	_, err = mock.Get(ctx, "key1")
	assert.Equal(t, ErrKeyNotFound, err)
}

func TestMockClient_Exists(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Key doesn't exist initially
	exists, err := mock.Exists(ctx, "key1")
	require.NoError(t, err)
	assert.False(t, exists)

	// Set the key
	err = mock.Set(ctx, "key1", "value1", 0)
	require.NoError(t, err)

	// Now it exists
	exists, err = mock.Exists(ctx, "key1")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMockClient_TTL(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Key doesn't exist
	ttl, err := mock.TTL(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, -2*time.Second, ttl)

	// Key with no expiration
	err = mock.Set(ctx, "key1", "value1", 0)
	require.NoError(t, err)
	ttl, err = mock.TTL(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1), ttl)

	// Key with TTL
	err = mock.Set(ctx, "key2", "value2", 5*time.Second)
	require.NoError(t, err)
	ttl, err = mock.TTL(ctx, "key2")
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, 5*time.Second)
}

func TestMockClient_Expire(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Set a key with no expiration
	err := mock.Set(ctx, "key1", "value1", 0)
	require.NoError(t, err)

	// Set expiration
	err = mock.Expire(ctx, "key1", 10*time.Second)
	require.NoError(t, err)

	// Verify TTL is set
	ttl, err := mock.TTL(ctx, "key1")
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, 10*time.Second)
}

func TestMockClient_PubSub(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Track received messages
	messages := []string{}
	callback := func(channel, message string) {
		messages = append(messages, message)
	}

	// Subscribe to a channel
	cancel, err := mock.Subscribe(ctx, callback, "test-channel")
	require.NoError(t, err)
	require.NotNil(t, cancel)
	defer cancel()

	// Publish a message
	err = mock.Publish(ctx, "test-channel", "hello")
	require.NoError(t, err)

	// Verify message received
	assert.Len(t, messages, 1)
	assert.Equal(t, "hello", messages[0])
}

func TestMockClient_ErrorInjection(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Inject Set error
	mock.SetError = assert.AnError
	err := mock.Set(ctx, "key1", "value1", 0)
	assert.Error(t, err)

	// Reset error
	mock.SetError = nil
	err = mock.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)

	// Inject Get error
	mock.GetError = assert.AnError
	_, err = mock.Get(ctx, "key1")
	assert.Error(t, err)
}

func TestMockClient_Ping(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Normal ping
	err := mock.Ping(ctx)
	assert.NoError(t, err)

	// Inject error
	mock.PingError = assert.AnError
	err = mock.Ping(ctx)
	assert.Error(t, err)
}

func TestMockClient_Close(t *testing.T) {
	mock := NewMockClient()

	// Normal close
	err := mock.Close()
	assert.NoError(t, err)

	// Inject error
	mock.CloseError = assert.AnError
	err = mock.Close()
	assert.Error(t, err)
}

func TestMockClient_Reset(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Add some data
	err := mock.Set(ctx, "key1", "value1", 0)
	require.NoError(t, err)

	// Inject errors
	mock.SetError = assert.AnError
	mock.GetError = assert.AnError

	// Reset
	mock.Reset()

	// Verify store is empty
	_, err = mock.Get(ctx, "key1")
	assert.Equal(t, ErrKeyNotFound, err)

	// Verify errors cleared
	assert.Nil(t, mock.SetError)
	assert.Nil(t, mock.GetError)
}
