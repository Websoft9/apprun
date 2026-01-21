package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"apprun/pkg/logger"
	"apprun/pkg/metrics/storage"
)

// BadgerExporter exports OTEL metrics to BadgerDB storage.
// It implements the metric.Exporter interface.
type BadgerExporter struct {
	storage storage.Storage
	logger  logger.Logger
}

// NewBadgerExporter creates a new OTEL exporter that writes to storage.
func NewBadgerExporter(storage storage.Storage) (*BadgerExporter, error) {
	return &BadgerExporter{
		storage: storage,
		logger:  logger.L(),
	}, nil
}

// Export exports OTEL metrics to BadgerDB storage.
func (e *BadgerExporter) Export(ctx context.Context, rm *metricdata.ResourceMetrics) error {
	if rm == nil {
		return nil
	}

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if err := e.exportMetric(ctx, m); err != nil {
				e.logger.Error("Failed to export metric",
					logger.Field{Key: "metric", Value: m.Name},
					logger.Field{Key: "error", Value: err},
				)
				// Continue with other metrics
			}
		}
	}

	return nil
}

// exportMetric converts and stores a single OTEL metric.
func (e *BadgerExporter) exportMetric(ctx context.Context, m metricdata.Metrics) error {
	switch data := m.Data.(type) {
	case metricdata.Sum[int64]:
		return e.exportSum(ctx, m.Name, data)
	case metricdata.Sum[float64]:
		return e.exportSumFloat(ctx, m.Name, data)
	case metricdata.Gauge[int64]:
		return e.exportGauge(ctx, m.Name, data)
	case metricdata.Gauge[float64]:
		return e.exportGaugeFloat(ctx, m.Name, data)
	case metricdata.Histogram[int64]:
		return e.exportHistogram(ctx, m.Name, data)
	case metricdata.Histogram[float64]:
		return e.exportHistogramFloat(ctx, m.Name, data)
	default:
		e.logger.Warn("Unsupported metric type",
			logger.Field{Key: "metric", Value: m.Name},
			logger.Field{Key: "type", Value: data},
		)
		return nil
	}
}

// exportSum exports Sum metrics (int64).
func (e *BadgerExporter) exportSum(ctx context.Context, name string, sum metricdata.Sum[int64]) error {
	for _, dp := range sum.DataPoints {
		metric := storage.Metric{
			Name:      name,
			Value:     float64(dp.Value),
			Tags:      attributesToTags(dp.Attributes),
			Timestamp: dp.Time,
		}
		if err := e.storage.Store(ctx, metric); err != nil {
			return err
		}
	}
	return nil
}

// exportSumFloat exports Sum metrics (float64).
func (e *BadgerExporter) exportSumFloat(ctx context.Context, name string, sum metricdata.Sum[float64]) error {
	for _, dp := range sum.DataPoints {
		metric := storage.Metric{
			Name:      name,
			Value:     dp.Value,
			Tags:      attributesToTags(dp.Attributes),
			Timestamp: dp.Time,
		}
		if err := e.storage.Store(ctx, metric); err != nil {
			return err
		}
	}
	return nil
}

// exportGauge exports Gauge metrics (int64).
func (e *BadgerExporter) exportGauge(ctx context.Context, name string, gauge metricdata.Gauge[int64]) error {
	for _, dp := range gauge.DataPoints {
		metric := storage.Metric{
			Name:      name,
			Value:     float64(dp.Value),
			Tags:      attributesToTags(dp.Attributes),
			Timestamp: dp.Time,
		}
		if err := e.storage.Store(ctx, metric); err != nil {
			return err
		}
	}
	return nil
}

// exportGaugeFloat exports Gauge metrics (float64).
func (e *BadgerExporter) exportGaugeFloat(ctx context.Context, name string, gauge metricdata.Gauge[float64]) error {
	for _, dp := range gauge.DataPoints {
		metric := storage.Metric{
			Name:      name,
			Value:     dp.Value,
			Tags:      attributesToTags(dp.Attributes),
			Timestamp: dp.Time,
		}
		if err := e.storage.Store(ctx, metric); err != nil {
			return err
		}
	}
	return nil
}

// exportHistogram exports Histogram metrics (int64).
func (e *BadgerExporter) exportHistogram(ctx context.Context, name string, hist metricdata.Histogram[int64]) error {
	for _, dp := range hist.DataPoints {
		// Store histogram summary metrics
		tags := attributesToTags(dp.Attributes)

		// Store count
		countMetric := storage.Metric{
			Name:      name + "_count",
			Value:     float64(dp.Count),
			Tags:      tags,
			Timestamp: dp.Time,
		}
		if err := e.storage.Store(ctx, countMetric); err != nil {
			return err
		}

		// Store sum
		sumMetric := storage.Metric{
			Name:      name + "_sum",
			Value:     float64(dp.Sum),
			Tags:      tags,
			Timestamp: dp.Time,
		}
		if err := e.storage.Store(ctx, sumMetric); err != nil {
			return err
		}
	}
	return nil
}

// exportHistogramFloat exports Histogram metrics (float64).
func (e *BadgerExporter) exportHistogramFloat(ctx context.Context, name string, hist metricdata.Histogram[float64]) error {
	for _, dp := range hist.DataPoints {
		// Store histogram summary metrics
		tags := attributesToTags(dp.Attributes)

		// Store count
		countMetric := storage.Metric{
			Name:      name + "_count",
			Value:     float64(dp.Count),
			Tags:      tags,
			Timestamp: dp.Time,
		}
		if err := e.storage.Store(ctx, countMetric); err != nil {
			return err
		}

		// Store sum
		sumMetric := storage.Metric{
			Name:      name + "_sum",
			Value:     dp.Sum,
			Tags:      tags,
			Timestamp: dp.Time,
		}
		if err := e.storage.Store(ctx, sumMetric); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown gracefully shuts down the exporter.
func (e *BadgerExporter) Shutdown(ctx context.Context) error {
	e.logger.Info("Shutting down BadgerDB exporter")
	return e.storage.Close()
}

// ForceFlush is a no-op for BadgerDB exporter (writes are immediate).
func (e *BadgerExporter) ForceFlush(ctx context.Context) error {
	return nil
}

// Temporality returns the temporality for the exporter.
func (e *BadgerExporter) Temporality(k metric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

// Aggregation returns the aggregation for the exporter.
func (e *BadgerExporter) Aggregation(k metric.InstrumentKind) metric.Aggregation {
	return metric.DefaultAggregationSelector(k)
}

// attributesToTags converts OTEL attributes to storage tags.
func attributesToTags(attrs attribute.Set) map[string]string {
	tags := make(map[string]string, attrs.Len())
	iter := attrs.Iter()
	for iter.Next() {
		kv := iter.Attribute()
		tags[string(kv.Key)] = kv.Value.AsString()
	}
	return tags
}

// SetupOTEL initializes the OpenTelemetry SDK with BadgerDB exporter.
func SetupOTEL(storage storage.Storage) (*metric.MeterProvider, error) {
	exporter, err := NewBadgerExporter(storage)
	if err != nil {
		return nil, err
	}

	// Create periodic reader
	reader := metric.NewPeriodicReader(exporter,
		metric.WithInterval(10*time.Second),
	)

	// Create meter provider
	provider := metric.NewMeterProvider(
		metric.WithReader(reader),
	)

	logger.L().Info("OpenTelemetry SDK initialized with BadgerDB exporter")

	return provider, nil
}
