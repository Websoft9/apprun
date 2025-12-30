// Package logger provides a unified logging interface (Anti-Corruption Layer)
// that isolates third-party logging library dependencies.
package logger

import (
	"context"
)

// Logger defines the unified logging interface
type Logger interface {
	// Debug logs a debug-level message with optional structured fields
	Debug(msg string, fields ...Field)

	// Info logs an info-level message with optional structured fields
	Info(msg string, fields ...Field)

	// Warn logs a warning-level message with optional structured fields
	Warn(msg string, fields ...Field)

	// Error logs an error-level message with optional structured fields
	Error(msg string, fields ...Field)

	// Fatal logs a fatal-level message with optional structured fields and exits
	Fatal(msg string, fields ...Field)

	// With creates a child logger with fixed fields
	With(fields ...Field) Logger

	// WithContext creates a child logger with context (auto-injects request_id)
	WithContext(ctx context.Context) Logger

	// Close closes the logger and releases resources
	Close() error
}

// Field represents a structured log field (key-value pair)
type Field struct {
	Key   string
	Value interface{}
}

// Level represents log level
type Level string

const (
	// LevelDebug is for debugging information
	LevelDebug Level = "debug"

	// LevelInfo is for informational messages
	LevelInfo Level = "info"

	// LevelWarn is for warning messages
	LevelWarn Level = "warn"

	// LevelError is for error messages
	LevelError Level = "error"
)

// Config holds logger configuration
type Config struct {
	Level  Level        `json:"level"`
	Output OutputConfig `json:"output"`
}

// OutputConfig defines output targets
type OutputConfig struct {
	// Targets specifies where logs should be written
	// Supported formats:
	// - "stdout": standard output
	// - "stderr": standard error
	// - "file:/path/to/file.log": file path
	Targets []string `json:"targets"`
}

// Global logger instance
var defaultLogger Logger = &NopLogger{}

// SetLogger sets the global logger instance
func SetLogger(l Logger) {
	if l != nil {
		defaultLogger = l
	}
}

// L returns the current global logger instance
func L() Logger {
	return defaultLogger
}

// Debug logs a debug-level message using the global logger
func Debug(msg string, fields ...Field) {
	defaultLogger.Debug(msg, fields...)
}

// Info logs an info-level message using the global logger
func Info(msg string, fields ...Field) {
	defaultLogger.Info(msg, fields...)
}

// Warn logs a warning-level message using the global logger
func Warn(msg string, fields ...Field) {
	defaultLogger.Warn(msg, fields...)
}

// Error logs an error-level message using the global logger
func Error(msg string, fields ...Field) {
	defaultLogger.Error(msg, fields...)
}

// Fatal logs a fatal-level message using the global logger and exits
func Fatal(msg string, fields ...Field) {
	defaultLogger.Fatal(msg, fields...)
}
