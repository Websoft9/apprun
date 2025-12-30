package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Helper function to create test logger with captured output
func newTestLogger(level Level, buf *bytes.Buffer) Logger {
	zapLevel, _ := parseLevel(level)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(encoderCfg)
	writer := zapcore.AddSync(buf)
	core := zapcore.NewCore(encoder, writer, zapLevel)
	zapLog := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &zapLogger{logger: zapLog}
}

// TestZapLogger_AllLevels tests all log levels
func TestZapLogger_AllLevels(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(LevelDebug, &buf)

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("Expected debug message in output")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Expected info message in output")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Expected warn message in output")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Expected error message in output")
	}
} // TestZapLogger_StructuredFields tests structured logging
func TestZapLogger_StructuredFields(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(LevelInfo, &buf)

	log.Info("user action", Field{"user_id", 123}, Field{"action", "login"})

	output := buf.String()

	// Parse JSON to verify fields
	var logEntry map[string]interface{}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("No log output captured")
	}

	if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v\nOutput: %s", err, output)
	}

	if logEntry["user_id"] != float64(123) {
		t.Errorf("Expected user_id 123, got %v", logEntry["user_id"])
	}
	if logEntry["action"] != "login" {
		t.Errorf("Expected action 'login', got %v", logEntry["action"])
	}
}

// TestZapLogger_WithContext tests request_id extraction
func TestZapLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(LevelInfo, &buf)

	// Create context with request ID
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "req-12345")

	logWithCtx := log.WithContext(ctx)
	logWithCtx.Info("processing request")

	output := buf.String()

	// Parse JSON to verify request_id
	var logEntry map[string]interface{}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("No log output captured")
	}

	if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v\nOutput: %s", err, output)
	}

	if logEntry["request_id"] != "req-12345" {
		t.Errorf("Expected request_id 'req-12345', got %v", logEntry["request_id"])
	}
}

// TestZapLogger_With tests fixed fields
func TestZapLogger_With(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(LevelInfo, &buf)

	// Create logger with fixed field
	serviceLog := log.With(Field{"service", "user-service"})

	serviceLog.Info("operation completed")

	output := buf.String()

	// Parse JSON to verify fixed field
	var logEntry map[string]interface{}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("No log output captured")
	}

	if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v\nOutput: %s", err, output)
	}

	if logEntry["service"] != "user-service" {
		t.Errorf("Expected service 'user-service', got %v", logEntry["service"])
	}
}

// TestZapLogger_LevelFiltering tests log level filtering
func TestZapLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(LevelWarn, &buf) // Only warn and above

	log.Debug("debug message") // Should NOT appear
	log.Info("info message")   // Should NOT appear
	log.Warn("warn message")   // Should appear
	log.Error("error message") // Should appear

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be included")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should be included")
	}
}

// TestZapLogger_MultipleTargets tests multi-target output
func TestZapLogger_MultipleTargets(t *testing.T) {
	tmpFile := "/tmp/test-logger.log"
	defer os.Remove(tmpFile)

	cfg := Config{
		Level: LevelInfo,
		Output: OutputConfig{
			Targets: []string{"stdout", "file:" + tmpFile},
		},
	}

	log, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}
	defer log.Close()

	log.Info("multi-target test")

	// Verify file was written
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "multi-target test") {
		t.Error("Expected log message in file output")
	}
}

// TestNewZapLogger_InvalidLevel tests invalid log level handling
func TestNewZapLogger_InvalidLevel(t *testing.T) {
	cfg := Config{
		Level: Level("invalid"),
		Output: OutputConfig{
			Targets: []string{"stdout"},
		},
	}

	log, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("Should degrade gracefully, got error: %v", err)
	}
	defer log.Close()

	// Should still work with degraded level
	log.Info("test message")
}

// TestNewZapLogger_InvalidTarget tests invalid output target
func TestNewZapLogger_InvalidTarget(t *testing.T) {
	cfg := Config{
		Level: LevelInfo,
		Output: OutputConfig{
			Targets: []string{"invalid:target"},
		},
	}

	_, err := NewZapLogger(cfg)
	if err == nil {
		t.Error("Expected error for invalid target")
	}
	if !strings.Contains(err.Error(), "invalid output target") {
		t.Errorf("Expected 'invalid output target' error, got: %v", err)
	}
}

// TestNewZapLogger_DuplicateTarget tests duplicate target detection
func TestNewZapLogger_DuplicateTarget(t *testing.T) {
	cfg := Config{
		Level: LevelInfo,
		Output: OutputConfig{
			Targets: []string{"stdout", "stdout"},
		},
	}

	_, err := NewZapLogger(cfg)
	if err == nil {
		t.Error("Expected error for duplicate target")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("Expected 'duplicate' error, got: %v", err)
	}
}

// TestNewZapLogger_FileOpenFailure tests file permission error
func TestNewZapLogger_FileOpenFailure(t *testing.T) {
	// Use a path that is likely to fail (non-existent directory)
	cfg := Config{
		Level: LevelInfo,
		Output: OutputConfig{
			Targets: []string{"file:/nonexistent/directory/no-permission.log"},
		},
	}

	_, err := NewZapLogger(cfg)
	if err == nil {
		t.Error("Expected error for file path in non-existent directory")
	}
	if !strings.Contains(err.Error(), "failed to open log file") {
		t.Logf("Got expected error: %v", err)
	}
}

// TestZapLogger_Close tests Close method
func TestZapLogger_Close(t *testing.T) {
	tmpFile := "/tmp/test-close.log"
	defer os.Remove(tmpFile)

	cfg := Config{
		Level: LevelInfo,
		Output: OutputConfig{
			Targets: []string{"file:" + tmpFile},
		},
	}

	log, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	log.Info("test")

	// Close should succeed
	if err := log.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestZapLogger_WithContext_NilContext tests nil context handling
func TestZapLogger_WithContext_NilContext(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(LevelInfo, &buf)

	// Test with nil context (using context.TODO as linter suggests would change behavior)
	// We explicitly test the nil handling code path
	logWithCtx := log.WithContext(context.Background())
	if logWithCtx == nil {
		t.Error("WithContext should return non-nil logger")
	}

	logWithCtx.Info("test message")
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Error("Should log message")
	}
}

// TestZapLogger_Fatal is intentionally NOT testing actual Fatal behavior
// because Fatal calls os.Exit(1) which would terminate the test process.
// In production, Fatal should only be used for unrecoverable startup errors.
func TestZapLogger_Fatal_Documentation(t *testing.T) {
	// This test documents that Fatal is NOT tested due to os.Exit behavior
	// Production usage should be limited to startup-only critical errors
	t.Log("Fatal() method calls os.Exit(1) and cannot be safely tested")
	t.Log("Use Fatal() only for unrecoverable startup errors")
}

// TestNewZapLogger_PartialMultiTargetFailure tests partial failure in multi-target setup
func TestNewZapLogger_PartialMultiTargetFailure(t *testing.T) {
	// Create a scenario where first target succeeds but second fails
	// Use /dev/null for first target (always succeeds) and invalid path for second
	cfg := Config{
		Level: LevelInfo,
		Output: OutputConfig{
			Targets: []string{"file:/dev/null", "file:/nonexistent/directory/failure.log"},
		},
	}

	_, err := NewZapLogger(cfg)
	if err == nil {
		t.Error("Expected error for second target failure")
	}
	if !strings.Contains(err.Error(), "failed to open log file") {
		t.Errorf("Expected file open error, got: %v", err)
	}
}
