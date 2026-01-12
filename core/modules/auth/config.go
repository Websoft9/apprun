// Package auth provides authentication configuration management.
package auth

import (
	"time"

	"apprun/pkg/jwt"
)

// ============================================================================
// Module Constants (Validation Rules)
// ============================================================================

const (
	// Password validation rules
	MinPasswordLength = 8
	MaxPasswordLength = 128

	// Username validation rules
	MinUsernameLength = 3
	MaxUsernameLength = 64

	// Account status codes
	StatusActive   int8 = 1
	StatusDisabled int8 = 2
	StatusPending  int8 = 3

	// Gender codes
	GenderUnknown int8 = 0
	GenderMale    int8 = 1
	GenderFemale  int8 = 2

	// Default values for security settings
	DefaultBcryptCost          = 10
	DefaultMaxFailedAttempts   = 5
	DefaultFailedLoginCacheTTL = 5 * time.Minute
)

// ============================================================================
// Configuration Structures (Config Center Managed)
// ============================================================================

// Config holds all authentication-related configuration.
// This struct follows the Configuration Center architecture defined in Story 10.
type Config struct {
	JWT      jwt.Config     `yaml:"jwt"`      // JWT configuration (from pkg/jwt)
	Security SecurityConfig `yaml:"security"` // Security settings specific to auth module
}

// SecurityConfig defines security-related settings for authentication.
type SecurityConfig struct {
	// BcryptCost controls the computational cost of password hashing.
	// Higher = more secure but slower. Valid range: 4-31.
	// Recommended: 10 (default), 8 (high-traffic), 12 (maximum security).
	// db:"true" - Can be adjusted for performance tuning.
	BcryptCost int `yaml:"bcrypt_cost" default:"10" db:"true" validate:"min=4,max=31"`

	// FailedLoginCacheEnabled enables in-memory caching of failed login attempts.
	// Prevents expensive bcrypt operations for repeated failed logins.
	// db:"true" - Can be toggled at runtime.
	FailedLoginCacheEnabled bool `yaml:"failed_login_cache_enabled" default:"true" db:"true"`

	// FailedLoginCacheTTL defines how long to cache failed login attempts.
	// db:"true" - Can be adjusted for security/performance balance.
	FailedLoginCacheTTL time.Duration `yaml:"failed_login_cache_ttl" default:"5m" db:"true" validate:"min=1m,max=1h"`

	// MaxFailedAttempts defines the threshold for rate limiting.
	// If a user fails login this many times within the TTL window,
	// subsequent attempts are rejected without bcrypt verification.
	// db:"true" - Can be adjusted for security policy.
	MaxFailedAttempts int `yaml:"max_failed_attempts" default:"5" db:"true" validate:"min=3,max=20"`
}

// DefaultConfig returns a Config with sensible defaults.
// This is used when no configuration file is present.
func DefaultConfig() *Config {
	return &Config{
		JWT:      *jwt.DefaultConfig(), // Use JWT package defaults
		Security: DefaultSecurityConfig(),
	}
}

// DefaultSecurityConfig returns default security configuration.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		BcryptCost:              DefaultBcryptCost,
		FailedLoginCacheEnabled: true,
		FailedLoginCacheTTL:     DefaultFailedLoginCacheTTL,
		MaxFailedAttempts:       DefaultMaxFailedAttempts,
	}
}
