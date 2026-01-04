package logger

import (
	"context"
	"apprun/pkg/errors"
	"os"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger wraps zap.Logger to implement our Logger interface
type zapLogger struct {
	logger  *zap.Logger
	closers []func() error
}

// validateConfig validates logger configuration
func validateConfig(cfg Config) error {
	// Validate level - all levels are accepted, invalid ones degrade to Info
	// No validation needed here since parseLevel handles it

	// Check for duplicate targets
	seen := make(map[string]bool)
	for _, target := range cfg.Output.Targets {
		if seen[target] {
			return errors.New(errors.ErrCodeLogInvalidConfig, "Duplicate output target").WithContext("target", target)
		}
		seen[target] = true

		// Validate target format
		if !strings.HasPrefix(target, "file:") &&
			target != "stdout" &&
			target != "stderr" {
			return errors.New(errors.ErrCodeLogInvalidConfig, "Invalid output target").WithContext("target", target)
		}
	}

	return nil
}

// NewZapLogger creates a new zap-based logger with the given configuration
func NewZapLogger(cfg Config) (Logger, error) {
	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeLogInvalidConfig, "Invalid logger config")
	}

	// Parse log level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeLogLevelParseFailed, "Invalid log level")
	}

	// Create encoder config (JSON format)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(encoderCfg)

	// Parse output targets
	writeSyncers, closers, err := parseOutputTargets(cfg.Output.Targets)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeLogInvalidConfig, "Failed to parse output targets")
	}

	// Create multi-writer if multiple targets
	var writer zapcore.WriteSyncer
	if len(writeSyncers) == 1 {
		writer = writeSyncers[0]
	} else {
		writer = zapcore.NewMultiWriteSyncer(writeSyncers...)
	}

	// Create core
	core := zapcore.NewCore(encoder, writer, level)

	// Create logger
	zapLog := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// Store closers for cleanup
	return &zapLogger{
		logger:  zapLog,
		closers: closers,
	}, nil
}

// parseLevel converts Level to zapcore.Level
func parseLevel(level Level) (zapcore.Level, error) {
	switch level {
	case LevelDebug:
		return zapcore.DebugLevel, nil
	case LevelInfo:
		return zapcore.InfoLevel, nil
	case LevelWarn:
		return zapcore.WarnLevel, nil
	case LevelError:
		return zapcore.ErrorLevel, nil
	default:
		// Degrade to Info level for unknown levels
		return zapcore.InfoLevel, nil
	}
}

// parseOutputTargets parses target strings and returns WriteSyncers
func parseOutputTargets(targets []string) ([]zapcore.WriteSyncer, []func() error, error) {
	if len(targets) == 0 {
		// Default to stdout
		targets = []string{"stdout"}
	}

	var syncers []zapcore.WriteSyncer
	var closers []func() error

	for _, target := range targets {
		switch {
		case target == "stdout":
			syncers = append(syncers, zapcore.AddSync(os.Stdout))
		case target == "stderr":
			syncers = append(syncers, zapcore.AddSync(os.Stderr))
		case strings.HasPrefix(target, "file:"):
			filePath := strings.TrimPrefix(target, "file:")
			file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, nil, errors.Wrap(err, errors.ErrCodeLogFileOpenFailed, "Failed to open log file").WithContext("file", filePath)
			}
			syncers = append(syncers, zapcore.AddSync(file))
			closers = append(closers, file.Close)
		default:
			return nil, nil, errors.New(errors.ErrCodeLogInvalidConfig, "Unsupported output target").WithContext("target", target)
		}
	}

	return syncers, closers, nil
}

// fieldsToZap converts our Field type to zap.Field
func fieldsToZap(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}

// Debug logs a debug-level message
func (z *zapLogger) Debug(msg string, fields ...Field) {
	z.logger.Debug(msg, fieldsToZap(fields)...)
}

// Info logs an info-level message
func (z *zapLogger) Info(msg string, fields ...Field) {
	z.logger.Info(msg, fieldsToZap(fields)...)
}

// Warn logs a warning-level message
func (z *zapLogger) Warn(msg string, fields ...Field) {
	z.logger.Warn(msg, fieldsToZap(fields)...)
}

// Error logs an error-level message
func (z *zapLogger) Error(msg string, fields ...Field) {
	z.logger.Error(msg, fieldsToZap(fields)...)
}

// Fatal logs a fatal-level message and exits
func (z *zapLogger) Fatal(msg string, fields ...Field) {
	z.logger.Fatal(msg, fieldsToZap(fields)...)
}

// With creates a child logger with fixed fields
func (z *zapLogger) With(fields ...Field) Logger {
	childLogger := z.logger.With(fieldsToZap(fields)...)
	// Child loggers share the same closers (resources)
	return &zapLogger{
		logger:  childLogger,
		closers: z.closers,
	}
}

// WithContext creates a child logger with context, auto-injecting request_id
func (z *zapLogger) WithContext(ctx context.Context) Logger {
	// Handle nil context
	if ctx == nil {
		return z
	}

	// Extract request_id from context using chi middleware
	requestID := middleware.GetReqID(ctx)
	if requestID != "" {
		return z.With(Field{"request_id", requestID})
	}
	return z
}

// Close closes the logger and releases all resources
func (z *zapLogger) Close() error {
	// Sync zap logger first (ignore common errors for stdout/stderr)
	_ = z.logger.Sync()

	// Close all file handles
	var errs []error
	for _, closer := range z.closers {
		if err := closer(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.New(errors.ErrCodeLogOutputFailed, "Failed to close logger resources")
	}
	return nil
}
