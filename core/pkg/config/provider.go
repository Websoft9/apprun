// Package config provides configuration provider interfaces and implementations.
package config

import "time"

// Provider is the interface for accessing configuration values.
// This abstraction allows swapping configuration backends (Viper, environment, etc.)
// and simplifies testing with mock implementations.
type Provider interface {
	// Get retrieves a configuration value by key.
	// Returns nil if the key doesn't exist.
	Get(key string) interface{}

	// GetString retrieves a string configuration value.
	// Returns empty string if the key doesn't exist or value is not a string.
	GetString(key string) string

	// GetInt retrieves an integer configuration value.
	// Returns 0 if the key doesn't exist or value cannot be converted to int.
	GetInt(key string) int

	// GetInt64 retrieves an int64 configuration value.
	// Returns 0 if the key doesn't exist or value cannot be converted to int64.
	GetInt64(key string) int64

	// GetBool retrieves a boolean configuration value.
	// Returns false if the key doesn't exist or value cannot be converted to bool.
	GetBool(key string) bool

	// GetFloat64 retrieves a float64 configuration value.
	// Returns 0.0 if the key doesn't exist or value cannot be converted to float64.
	GetFloat64(key string) float64

	// GetDuration retrieves a time.Duration configuration value.
	// Supports duration strings like "300ms", "1.5h", "2h45m".
	// Returns 0 if the key doesn't exist or value cannot be parsed as duration.
	GetDuration(key string) time.Duration

	// GetStringSlice retrieves a string slice configuration value.
	// Returns empty slice if the key doesn't exist or value is not a string slice.
	GetStringSlice(key string) []string

	// IsSet checks if a configuration key exists.
	IsSet(key string) bool

	// Set sets a configuration value (for runtime updates).
	// Not all providers support this operation.
	Set(key string, value interface{})
}
