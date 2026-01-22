package storage

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadgerStorage_GCGoroutineLifecycle(t *testing.T) {
	tmpDir := t.TempDir()

	// Record initial goroutine count
	initialGoroutines := runtime.NumGoroutine()

	cfg := BadgerConfig{
		Path:       filepath.Join(tmpDir, "testdb"),
		TTL:        1 * time.Hour,
		GCInterval: 100 * time.Millisecond, // Short interval for testing
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)

	// Give GC goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Verify goroutine started (count increased)
	afterStartGoroutines := runtime.NumGoroutine()
	assert.Greater(t, afterStartGoroutines, initialGoroutines, "GC goroutine should be running")

	// Close storage
	err = storage.Close()
	require.NoError(t, err)

	// Give GC goroutine time to stop
	time.Sleep(200 * time.Millisecond)

	// Verify goroutine stopped (count decreased or back to initial)
	afterCloseGoroutines := runtime.NumGoroutine()
	assert.LessOrEqual(t, afterCloseGoroutines, afterStartGoroutines, "GC goroutine should have stopped")

	// Verify double-close is safe
	err = storage.Close()
	require.NoError(t, err, "Double close should not panic")
}

func TestBadgerStorage_GCStopsOnClose(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path:       filepath.Join(tmpDir, "testdb"),
		TTL:        1 * time.Hour,
		GCInterval: 50 * time.Millisecond,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)

	// Wait for at least one GC cycle
	time.Sleep(100 * time.Millisecond)

	// Close should stop GC immediately
	err = storage.Close()
	require.NoError(t, err)

	// Verify storage is closed
	ctx := context.Background()
	metric := Metric{
		Name:      "test",
		Value:     1.0,
		Timestamp: time.Now(),
	}

	err = storage.Store(ctx, metric)
	assert.Error(t, err, "Store should fail after close")
	assert.Contains(t, err.Error(), "unavailable")
}

func TestBadgerStorage_GCDoesNotBlockClose(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path:       filepath.Join(tmpDir, "testdb"),
		TTL:        1 * time.Hour,
		GCInterval: 10 * time.Second, // Long interval
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)

	// Close immediately without waiting for GC
	done := make(chan struct{})
	go func() {
		err := storage.Close()
		require.NoError(t, err)
		close(done)
	}()

	// Close should complete quickly (not wait for GC cycle)
	select {
	case <-done:
		// Success - close completed
	case <-time.After(1 * time.Second):
		t.Fatal("Close blocked for too long")
	}
}
