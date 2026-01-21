package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	bopt "github.com/dgraph-io/badger/v3/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"apprun/pkg/errors"
	"apprun/pkg/logger"
)

// Internal metrics for monitoring BadgerDB performance
var (
	meterProvider = otel.GetMeterProvider()
	meter         = meterProvider.Meter("apprun.storage.badger")

	// Operation latency histograms
	writeLatencyMs  metric.Float64Histogram
	queryLatencyMs  metric.Float64Histogram
	deleteLatencyMs metric.Float64Histogram

	// Resource usage gauges
	diskUsageBytes metric.Int64ObservableGauge
	dbSize         *int64 // Pointer to current DB size for callback
	dbSizeMu       sync.RWMutex
)

func init() {
	var err error

	// Initialize latency histograms
	writeLatencyMs, err = meter.Float64Histogram(
		"badger.write.latency_ms",
		metric.WithDescription("Latency of BadgerDB write operations in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create write latency metric: %v", err))
	}

	queryLatencyMs, err = meter.Float64Histogram(
		"badger.query.latency_ms",
		metric.WithDescription("Latency of BadgerDB query operations in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create query latency metric: %v", err))
	}

	deleteLatencyMs, err = meter.Float64Histogram(
		"badger.delete.latency_ms",
		metric.WithDescription("Latency of BadgerDB delete operations in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create delete latency metric: %v", err))
	}

	// Initialize disk usage gauge (updated via callback)
	size := int64(0)
	dbSize = &size
	diskUsageBytes, err = meter.Int64ObservableGauge(
		"badger.disk_usage_bytes",
		metric.WithDescription("Current disk usage of BadgerDB in bytes"),
		metric.WithUnit("By"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			dbSizeMu.RLock()
			defer dbSizeMu.RUnlock()
			o.Observe(*dbSize)
			return nil
		}),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create disk usage metric: %v", err))
	}
}

// BadgerStorage implements the Storage interface using BadgerDB.
// It provides persistent, embedded key-value storage with TTL support.
type BadgerStorage struct {
	db     *badger.DB
	ttl    time.Duration
	cfg    BadgerConfig
	logger logger.Logger
	closed bool
	mu     sync.RWMutex
	stopGC chan struct{}
}

// BadgerConfig holds configuration for BadgerDB storage.
type BadgerConfig struct {
	Path        string        // Database file path
	TTL         time.Duration // Metric TTL (default: 24h)
	Compression bool          // Enable ZSTD compression
	GCInterval  time.Duration // Garbage collection interval (default: 5m)
}

// NewBadgerStorage creates a new BadgerDB storage backend.
//
// Example:
//
//	cfg := BadgerConfig{
//	    Path: "/var/lib/apprun/metrics",
//	    TTL: 24 * time.Hour,
//	    Compression: true,
//	}
//	storage, err := NewBadgerStorage(cfg)
func NewBadgerStorage(cfg BadgerConfig) (*BadgerStorage, error) {
	// Apply defaults
	if cfg.TTL == 0 {
		cfg.TTL = 24 * time.Hour
	}
	if cfg.GCInterval == 0 {
		cfg.GCInterval = 5 * time.Minute
	}

	// Configure BadgerDB options
	opts := badger.DefaultOptions(cfg.Path)
	if cfg.Compression {
		opts = opts.WithCompression(bopt.Snappy)
	}
	opts = opts.
		WithLoggingLevel(badger.WARNING).
		WithNumVersionsToKeep(1) // Only keep latest version

	// Open database
	db, err := badger.Open(opts)
	if err != nil {
		return nil, ErrUnavailable("badger_open", err)
	}

	// Initialize storage struct BEFORE starting goroutine
	// This ensures stopGC channel is ready if we need to abort
	storage := &BadgerStorage{
		db:     db,
		ttl:    cfg.TTL,
		cfg:    cfg,
		logger: logger.L(),
		stopGC: make(chan struct{}),
		closed: false,
	}

	// Start background garbage collection
	// GC goroutine will stop when stopGC channel is closed or storage is closed
	go storage.runGC()

	storage.logger.Info("BadgerDB storage initialized",
		logger.Field{Key: "path", Value: cfg.Path},
		logger.Field{Key: "ttl", Value: cfg.TTL},
		logger.Field{Key: "compression", Value: cfg.Compression},
	)

	return storage, nil
}

// Store writes a metric to BadgerDB with TTL.
// Key format: {metric_name}:{timestamp_ns}
func (s *BadgerStorage) Store(ctx context.Context, metric Metric) error {
	start := time.Now()
	defer func() {
		latency := float64(time.Since(start).Milliseconds())
		writeLatencyMs.Record(ctx, latency)
	}()

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return ErrUnavailable("store", nil)
	}
	s.mu.RUnlock()

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ErrTimeout("store", "context cancelled")
	default:
	}

	// Generate key: {name}:{timestamp_ns}
	key := fmt.Sprintf("%s:%d", metric.Name, metric.Timestamp.UnixNano())

	// Marshal metric to JSON
	value, err := json.Marshal(metric)
	if err != nil {
		return errors.Wrap(err, "marshal_metric", "failed to marshal metric")
	}

	// Write with TTL
	err = s.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), value).WithTTL(s.ttl)
		return txn.SetEntry(e)
	})

	if err != nil {
		s.logger.Error("Failed to store metric",
			logger.Field{Key: "metric", Value: metric.Name},
			logger.Field{Key: "error", Value: err},
		)
		return ErrUnavailable("store", err)
	}

	return nil
}

// Query retrieves metrics by name within a time range.
// Returns metrics sorted by timestamp (ascending).
// limit: Maximum number of results (0 or negative = default 1000, max 10000)
func (s *BadgerStorage) Query(ctx context.Context, name string, start, end time.Time, limit int) ([]Metric, error) {
	startTime := time.Now()
	defer func() {
		latency := float64(time.Since(startTime).Milliseconds())
		queryLatencyMs.Record(ctx, latency)
	}()

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, ErrUnavailable("query", nil)
	}
	s.mu.RUnlock()

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ErrTimeout("query", "context cancelled")
	default:
	}

	// Apply default and max limit
	if limit <= 0 {
		limit = 1000 // default
	} else if limit > 10000 {
		limit = 10000 // max to prevent OOM
	}

	var metrics []Metric

	err := s.db.View(func(txn *badger.Txn) error {
		// Set up iterator with prefix
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(name + ":")

		it := txn.NewIterator(opts)
		defer it.Close()

		// Key range
		startKey := []byte(fmt.Sprintf("%s:%d", name, start.UnixNano()))
		endKey := []byte(fmt.Sprintf("%s:%d", name, end.UnixNano()))

		// Iterate through keys with limit
		for it.Seek(startKey); it.Valid() && len(metrics) < limit; it.Next() {
			item := it.Item()
			key := item.Key()

			// Check if beyond end range
			if string(key) > string(endKey) {
				break
			}

			// Check context cancellation
			select {
			case <-ctx.Done():
				return ErrTimeout("query", "context cancelled during iteration")
			default:
			}

			// Read value
			err := item.Value(func(val []byte) error {
				var metric Metric
				if err := json.Unmarshal(val, &metric); err != nil {
					s.logger.Warn("Failed to unmarshal metric",
						logger.Field{Key: "key", Value: string(key)},
						logger.Field{Key: "error", Value: err},
					)
					return nil // Skip corrupted entry
				}
				metrics = append(metrics, metric)
				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to query metrics",
			logger.Field{Key: "metric", Value: name},
			logger.Field{Key: "start", Value: start},
			logger.Field{Key: "end", Value: end},
			logger.Field{Key: "error", Value: err},
		)
		return nil, ErrUnavailable("query", err)
	}

	// Sort by timestamp (ascending) - BadgerDB iterator should already be sorted
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Timestamp.Before(metrics[j].Timestamp)
	})

	return metrics, nil
}

// Delete removes metrics older than the specified time.
// This is primarily for manual cleanup; TTL handles automatic expiration.
func (s *BadgerStorage) Delete(ctx context.Context, before time.Time) error {
	start := time.Now()
	defer func() {
		latency := float64(time.Since(start).Milliseconds())
		deleteLatencyMs.Record(ctx, latency)
	}()

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return ErrUnavailable("delete", nil)
	}
	s.mu.RUnlock()

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ErrTimeout("delete", "context cancelled")
	default:
	}

	// Collect keys to delete
	var keysToDelete [][]byte

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need keys

		it := txn.NewIterator(opts)
		defer it.Close()

		beforeNs := before.UnixNano()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())

			// Extract timestamp from key: {name}:{timestamp_ns}
			parts := strings.Split(key, ":")
			if len(parts) != 2 {
				continue
			}

			var ts int64
			if _, err := fmt.Sscanf(parts[1], "%d", &ts); err != nil {
				continue
			}

			if ts < beforeNs {
				keysToDelete = append(keysToDelete, item.KeyCopy(nil))
			}

			// Check context cancellation
			select {
			case <-ctx.Done():
				return ErrTimeout("delete", "context cancelled during scan")
			default:
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Delete in batches
	const batchSize = 1000
	for i := 0; i < len(keysToDelete); i += batchSize {
		end := i + batchSize
		if end > len(keysToDelete) {
			end = len(keysToDelete)
		}

		batch := keysToDelete[i:end]

		err := s.db.Update(func(txn *badger.Txn) error {
			for _, key := range batch {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			s.logger.Error("Failed to delete batch",
				logger.Field{Key: "batch_start", Value: i},
				logger.Field{Key: "error", Value: err},
			)
			return ErrUnavailable("delete", err)
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ErrTimeout("delete", "context cancelled during batch delete")
		default:
		}
	}

	s.logger.Info("Deleted old metrics",
		logger.Field{Key: "before", Value: before},
		logger.Field{Key: "count", Value: len(keysToDelete)},
	)

	return nil
}

// Health checks if BadgerDB is operational.
func (s *BadgerStorage) Health(ctx context.Context) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return ErrUnavailable("health", nil)
	}
	s.mu.RUnlock()

	// Try a simple read operation
	err := s.db.View(func(txn *badger.Txn) error {
		return nil
	})

	if err != nil {
		return ErrUnavailable("health", err)
	}

	return nil
}

// Close gracefully shuts down BadgerDB.
func (s *BadgerStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil // Already closed
	}

	s.closed = true
	close(s.stopGC) // Stop GC goroutine

	s.logger.Info("Closing BadgerDB storage")

	return s.db.Close()
}

// runGC performs periodic garbage collection on BadgerDB.
func (s *BadgerStorage) runGC() {
	ticker := time.NewTicker(s.cfg.GCInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.RLock()
			closed := s.closed
			s.mu.RUnlock()

			if closed {
				return
			}

			// Update disk usage metrics
			lsm, vlog := s.db.Size()
			totalSize := lsm + vlog
			dbSizeMu.Lock()
			*dbSize = totalSize
			dbSizeMu.Unlock()

			// Run value log GC (0.5 = discard 50% of value log)
			err := s.db.RunValueLogGC(0.5)
			if err != nil && err != badger.ErrNoRewrite {
				s.logger.Warn("GC error",
					logger.Field{Key: "error", Value: err},
				)
			} else if err == nil {
				s.logger.Debug("GC completed successfully")
			}

		case <-s.stopGC:
			s.logger.Info("GC goroutine stopped")
			return
		}
	}
}

// GetStats returns BadgerDB statistics.
func (s *BadgerStorage) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return map[string]interface{}{"status": "closed"}
	}

	lsm, vlog := s.db.Size()

	return map[string]interface{}{
		"lsm_size_bytes":   lsm,
		"vlog_size_bytes":  vlog,
		"total_size_bytes": lsm + vlog,
		"path":             s.cfg.Path,
		"ttl_seconds":      s.ttl.Seconds(),
	}
}
