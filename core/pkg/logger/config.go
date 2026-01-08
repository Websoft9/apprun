// Package logger provides a unified logging interface (Anti-Corruption Layer)
// that isolates third-party logging library dependencies.
package logger

// Config holds logger configuration
// Follows internal/config/types.go standards for consistency
type Config struct {
	Level  Level        `yaml:"level" default:"info" db:"true" validate:"oneof=debug info warn error"`
	Output OutputConfig `yaml:"output"`
}

// OutputConfig defines output targets
type OutputConfig struct {
	// Targets specifies where logs should be written
	// Supported formats:
	// - "stdout": standard output
	// - "stderr": standard error
	// - "file:/path/to/file.log": file path
	Targets []string `yaml:"targets" default:"stdout" db:"true" validate:"min=1,dive,oneof=stdout stderr file"`
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() Config {
	return Config{
		Level: LevelInfo,
		Output: OutputConfig{
			Targets: []string{"stdout"},
		},
	}
}

// NewLoggerFromConfig creates a logger instance from configuration
// This is the factory connector pattern required by Story 10a
// Returns error if configuration is invalid
func NewLoggerFromConfig(cfg *Config) (Logger, error) {
	if cfg == nil {
		defaultCfg := DefaultConfig()
		cfg = &defaultCfg
	}

	// TODO: Implement actual logger creation based on config
	// For now, this is a placeholder that would integrate with
	// the actual logger implementation (e.g., zap, logrus)

	// Validation example:
	if len(cfg.Output.Targets) == 0 {
		return nil, ErrInvalidConfig("at least one output target is required")
	}

	// Return default logger for now
	// In production, this would create a configured logger instance
	return &NopLogger{}, nil
}

// ErrInvalidConfig creates a configuration error
func ErrInvalidConfig(msg string) error {
	return &ConfigError{Message: msg}
}

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return "logger config error: " + e.Message
}
