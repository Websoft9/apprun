package logger

import (
	"context"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
)

// Test Logger interface compilation
func TestLoggerInterface(t *testing.T) {
	var _ Logger = (*NopLogger)(nil)
}

// TestGlobalLogger tests global singleton pattern
func TestGlobalLogger(t *testing.T) {
	// Should have default logger
	log := L()
	if log == nil {
		t.Fatal("L() should return non-nil logger")
	}

	// Test SetLogger
	custom := &NopLogger{}
	SetLogger(custom)
	if L() != custom {
		t.Error("SetLogger should update global logger")
	}

	// Reset for other tests
	SetLogger(&NopLogger{})
}

// TestNopLogger verifies NopLogger doesn't panic
func TestNopLogger(t *testing.T) {
	log := &NopLogger{}

	// Should not panic
	log.Debug("test", Field{"key", "value"})
	log.Info("test", Field{"key", "value"})
	log.Warn("test", Field{"key", "value"})
	log.Error("test", Field{"key", "value"})

	// Test With
	log2 := log.With(Field{"service", "test"})
	if log2 == nil {
		t.Error("With should return non-nil logger")
	}

	// Test WithContext
	ctx := context.Background()
	log3 := log.WithContext(ctx)
	if log3 == nil {
		t.Error("WithContext should return non-nil logger")
	}
}

// TestDefaultLogger tests behavior before initialization
func TestDefaultLogger(t *testing.T) {
	// Reset to default
	SetLogger(&NopLogger{})

	// Should not panic with default logger
	Debug("test")
	Info("test", Field{"key", "value"})
	Warn("test")
	Error("test")
}

// TestField tests Field structure
func TestField(t *testing.T) {
	f := Field{"name", "value"}
	if f.Key != "name" {
		t.Errorf("Expected key 'name', got '%s'", f.Key)
	}
	if f.Value != "value" {
		t.Errorf("Expected value 'value', got '%v'", f.Value)
	}
}

// TestLevel tests Level type and constants
func TestLevel(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "debug"},
		{LevelInfo, "info"},
		{LevelWarn, "warn"},
		{LevelError, "error"},
	}

	for _, tt := range tests {
		if string(tt.level) != tt.want {
			t.Errorf("Level %v should equal '%s'", tt.level, tt.want)
		}
	}
}

// TestConfig tests Config structure
func TestConfig(t *testing.T) {
	cfg := Config{
		Level: LevelInfo,
		Output: OutputConfig{
			Targets: []string{"stdout"},
		},
	}

	if cfg.Level != LevelInfo {
		t.Errorf("Expected level Info, got %v", cfg.Level)
	}

	if len(cfg.Output.Targets) != 1 || cfg.Output.Targets[0] != "stdout" {
		t.Errorf("Expected target stdout, got %v", cfg.Output.Targets)
	}
}

// TestWithContext_RequestID tests request_id extraction from context
func TestWithContext_RequestID(t *testing.T) {
	// Create context with request ID using chi middleware
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "test-request-123")

	log := &NopLogger{}
	logWithCtx := log.WithContext(ctx)

	// NopLogger should return itself but not panic
	if logWithCtx == nil {
		t.Error("WithContext should return non-nil logger")
	}
}
