package logger

import (
	"context"
)

// NopLogger is a no-operation logger implementation
// It safely discards all log messages and is useful for testing
type NopLogger struct{}

// Debug does nothing
func (n *NopLogger) Debug(msg string, fields ...Field) {}

// Info does nothing
func (n *NopLogger) Info(msg string, fields ...Field) {}

// Warn does nothing
func (n *NopLogger) Warn(msg string, fields ...Field) {}

// Error does nothing
func (n *NopLogger) Error(msg string, fields ...Field) {}

// Fatal does nothing (unlike real logger, doesn't exit)
func (n *NopLogger) Fatal(msg string, fields ...Field) {}

// With returns the same NopLogger
func (n *NopLogger) With(fields ...Field) Logger {
	return n
}

// WithContext returns the same NopLogger
func (n *NopLogger) WithContext(ctx context.Context) Logger {
	return n
}

// Close does nothing
func (n *NopLogger) Close() error {
	return nil
}
