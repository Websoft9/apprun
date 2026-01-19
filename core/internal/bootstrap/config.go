// Package bootstrap provides application startup orchestration configuration.
package bootstrap

import (
	"apprun/pkg/env"
)

// Config holds bootstrap-specific configuration for platform initialization.
// This includes operational settings that control startup behavior.
// These are infrastructure/operational configs, NOT managed by config center (db:"false")
type Config struct {
	// AutoInit controls whether to automatically initialize the platform on first startup.
	//
	// Environment variable: AUTO_INIT
	// Default: false (production-safe)
	//
	// ⚠️  IMPORTANT: This is an operational setting and should NOT be configured
	// via YAML files to prevent accidental auto-initialization in production.
	// Use environment variables for explicit control in each deployment environment.
	//
	// Examples:
	//   Development:  export AUTO_INIT=true
	//   Production:   (unset, defaults to false)
	AutoInit bool `json:"auto_init" default:"false" db:"false" envonly:"true"`
}

// DefaultConfig returns default bootstrap configuration.
// Values are loaded from environment variables with fallback to code defaults.
func DefaultConfig() *Config {
	return &Config{
		AutoInit: env.GetBool("AUTO_INIT", false), // Default to false for safety
	}
}
