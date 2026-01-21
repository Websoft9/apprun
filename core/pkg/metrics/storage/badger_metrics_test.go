package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// TestBadgerStorage_MetricsInstrumentation verifies that BadgerDB operations
// emit OpenTelemetry metrics for monitoring performance.
func TestBadgerStorage_MetricsInstrumentation(t *testing.T) {
	// Setup OTEL metric reader to capture metrics
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))
	otel.SetMeterProvider(provider)

	// Re-initialize meter with test provider
	meter = provider.Meter("apprun.storage.badger")

	// Re-create metrics instruments with test meter
	var err error
	writeLatencyMs, err = meter.Float64Histogram("badger.write.latency_ms")
	require.NoError(t, err)

	queryLatencyMs, err = meter.Float64Histogram("badger.query.latency_ms")
	require.NoError(t, err)

	deleteLatencyMs, err = meter.Float64Histogram("badger.delete.latency_ms")
	require.NoError(t, err)

	// Create storage
	cfg := BadgerConfig{
		Path:       t.TempDir() + "/test_metrics",
		TTL:        1 * time.Hour,
		GCInterval: 100 * time.Millisecond,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	// Perform operations
	testMetric := Metric{
		Name:      "test_metric_instrumented",
		Value:     100.0,
		Timestamp: now,
		Tags:      map[string]string{"env": "test"},
	}

	// Store operation
	err = storage.Store(ctx, testMetric)
	require.NoError(t, err)

	// Query operation
	metrics, err := storage.Query(ctx, "test_metric_instrumented", now.Add(-1*time.Minute), now.Add(1*time.Minute), 100)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)

	// Delete operation
	err = storage.Delete(ctx, now.Add(1*time.Hour))
	require.NoError(t, err)

	// Wait for GC tick to update disk usage (happens periodically)
	time.Sleep(150 * time.Millisecond)

	// Collect metrics
	var rm metricdata.ResourceMetrics
	err = reader.Collect(ctx, &rm)
	require.NoError(t, err)

	// Verify metrics were recorded
	foundWrite := false
	foundQuery := false
	foundDelete := false

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch m.Name {
			case "badger.write.latency_ms":
				foundWrite = true
				// Verify it's a histogram with data
				hist, ok := m.Data.(metricdata.Histogram[float64])
				assert.True(t, ok, "write metric should be a histogram")
				assert.NotEmpty(t, hist.DataPoints, "write histogram should have data points")
				if len(hist.DataPoints) > 0 {
					assert.Greater(t, hist.DataPoints[0].Count, uint64(0), "write should have been recorded")
				}

			case "badger.query.latency_ms":
				foundQuery = true
				hist, ok := m.Data.(metricdata.Histogram[float64])
				assert.True(t, ok, "query metric should be a histogram")
				assert.NotEmpty(t, hist.DataPoints, "query histogram should have data points")

			case "badger.delete.latency_ms":
				foundDelete = true
				hist, ok := m.Data.(metricdata.Histogram[float64])
				assert.True(t, ok, "delete metric should be a histogram")
				assert.NotEmpty(t, hist.DataPoints, "delete histogram should have data points")

			case "badger.disk_usage_bytes":
				// Disk usage is an observable gauge, updated by GC goroutine
				gauge, ok := m.Data.(metricdata.Gauge[int64])
				assert.True(t, ok, "disk usage should be a gauge")
				if len(gauge.DataPoints) > 0 {
					assert.GreaterOrEqual(t, gauge.DataPoints[0].Value, int64(0), "disk usage should be non-negative")
				}
			}
		}
	}

	assert.True(t, foundWrite, "badger.write.latency_ms metric should be recorded")
	assert.True(t, foundQuery, "badger.query.latency_ms metric should be recorded")
	assert.True(t, foundDelete, "badger.delete.latency_ms metric should be recorded")
}
