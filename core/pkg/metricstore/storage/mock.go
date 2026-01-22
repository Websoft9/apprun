package storage

import (
	"context"
	"sort"
	"sync"
	"time"
)

// MockStorage is an in-memory implementation of the Storage interface.
// It's thread-safe and primarily used for testing.
// Data is stored in a map keyed by metric name, with values as time-series slices.
type MockStorage struct {
	mu      sync.RWMutex
	data    map[string][]Metric // key: metric name, value: time-series data
	closed  bool
	healthy bool
}

// newMockStorage creates a new in-memory mock storage backend.
// It's initialized in a healthy state with no data.
func newMockStorage() *MockStorage {
	return &MockStorage{
		data:    make(map[string][]Metric),
		healthy: true,
	}
}

// Store adds a metric to the in-memory storage.
// Metrics are appended to the time-series for their name.
func (m *MockStorage) Store(ctx context.Context, metric Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrUnavailable("mock", nil)
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ErrTimeout("Store", "context canceled")
	default:
	}

	// Append metric to time-series
	m.data[metric.Name] = append(m.data[metric.Name], metric)

	return nil
}

// Query retrieves metrics by name within a time range.
// Results are sorted by timestamp in ascending order.
// limit: Maximum number of results (0 or negative = default 1000, max 10000)
func (m *MockStorage) Query(ctx context.Context, name string, start, end time.Time, limit int) ([]Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrUnavailable("mock", nil)
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ErrTimeout("Query", "context canceled")
	default:
	}

	// Apply default and max limit
	if limit <= 0 {
		limit = 1000
	} else if limit > 10000 {
		limit = 10000
	}

	// Check if metric exists
	series, exists := m.data[name]
	if !exists {
		return nil, ErrNotFound(name)
	}

	// Filter by time range with limit
	var results []Metric
	for _, metric := range series {
		if len(results) >= limit {
			break
		}
		if (metric.Timestamp.Equal(start) || metric.Timestamp.After(start)) &&
			(metric.Timestamp.Equal(end) || metric.Timestamp.Before(end)) {
			results = append(results, metric)
		}
	}

	// Sort by timestamp (ascending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.Before(results[j].Timestamp)
	})

	return results, nil
}

// Delete removes metrics older than the specified time.
// Returns nil on success (mock doesn't return count of deleted items).
func (m *MockStorage) Delete(ctx context.Context, before time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrUnavailable("mock", nil)
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ErrTimeout("Delete", "context canceled")
	default:
	}

	// Filter out metrics before the cutoff time
	for name, series := range m.data {
		var remaining []Metric
		for _, metric := range series {
			if metric.Timestamp.Equal(before) || metric.Timestamp.After(before) {
				remaining = append(remaining, metric)
			}
		}

		if len(remaining) == 0 {
			// Remove empty series
			delete(m.data, name)
		} else {
			m.data[name] = remaining
		}
	}

	return nil
}

// Health checks if the mock storage is available.
// Returns an error if the storage is closed or marked unhealthy.
func (m *MockStorage) Health(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return ErrUnavailable("mock", nil)
	}

	if !m.healthy {
		return ErrUnavailable("mock", nil)
	}

	return nil
}

// Close marks the storage as closed.
// After calling Close, other methods will return errors.
func (m *MockStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

// Reset clears all data from the mock storage.
// This is useful for cleaning up between tests.
// Not part of the Storage interface - only for testing.
func (m *MockStorage) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string][]Metric)
	m.closed = false
	m.healthy = true
}

// SetHealthy controls the health status of the mock storage.
// This is useful for testing error conditions.
// Not part of the Storage interface - only for testing.
func (m *MockStorage) SetHealthy(healthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.healthy = healthy
}

// Count returns the total number of metrics stored.
// Not part of the Storage interface - only for testing.
func (m *MockStorage) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, series := range m.data {
		count += len(series)
	}
	return count
}
