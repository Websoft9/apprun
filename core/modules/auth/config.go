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

	// Platform Project Constants (Fixed Values)
	// The platform project uses a fixed UUID (all zeros) for easy identification
	PlatformProjectUUID        = "00000000-0000-0000-0000-000000000000"
	PlatformProjectName        = "Platform"
	PlatformProjectDescription = "Global platform-level resources and configuration"
)

// ReservedUsernames defines usernames that are reserved for system use.
// These usernames can only be used during platform initialization (super admin creation).
// Regular users cannot register with these names to prevent impersonation and confusion.
var ReservedUsernames = []string{
	"admin",
	"administrator",
	"root",
	"system",
	"superuser",
	"sysadmin",
}

// ============================================================================
// Configuration Structures (Config Center Managed)
// ============================================================================

// Config holds all authentication-related configuration.
// This struct follows the Configuration Center architecture defined in Story 10.
type Config struct {
	JWT      jwt.Config     `mapstructure:"jwt" json:"jwt" validate:"required"`           // JWT configuration (from pkg/jwt)
	Security SecurityConfig `mapstructure:"security" json:"security" validate:"required"` // Security settings specific to auth module
	Init     InitConfig     `mapstructure:"init" json:"init" validate:"required"`         // Platform initialization configuration (super admin account)
}

// InitConfig defines the platform initialization settings.
// These values are used during bootstrap to create the initial super admin account.
type InitConfig struct {
	// Username for the super admin account
	// Default: "admin"
	Username string `mapstructure:"username" json:"username" default:"admin"`

	// Email for the super admin account
	// Default: "admin@example.com"
	Email string `mapstructure:"email" json:"email" default:"admin@example.com"`

	// Password for the super admin account
	// If empty, a random password will be generated and logged during initialization
	// Default: "" (empty, will trigger random generation)
	Password string `mapstructure:"password" json:"password,omitempty" default:""`
}

// SecurityConfig defines security-related settings for authentication.
type SecurityConfig struct {
	// BcryptCost controls the computational cost of password hashing.
	// Higher = more secure but slower. Valid range: 4-31.
	// Recommended: 10 (default), 8 (high-traffic), 12 (maximum security).
	// db:"true" - Can be adjusted for performance tuning.
	BcryptCost int `mapstructure:"bcrypt_cost" json:"bcrypt_cost" default:"10" db:"true" validate:"omitempty,min=4,max=31"`

	// FailedLoginCacheEnabled enables in-memory caching of failed login attempts.
	// Prevents expensive bcrypt operations for repeated failed logins.
	// db:"true" - Can be toggled at runtime.
	FailedLoginCacheEnabled bool `mapstructure:"failed_login_cache_enabled" json:"failed_login_cache_enabled" default:"true" db:"true"`

	// FailedLoginCacheTTL defines how long to cache failed login attempts.
	// db:"true" - Can be adjusted for security/performance balance.
	FailedLoginCacheTTL time.Duration `mapstructure:"failed_login_cache_ttl" json:"failed_login_cache_ttl" default:"5m" db:"true" validate:"omitempty,min=1m,max=1h"`

	// MaxFailedAttempts defines the threshold for rate limiting.
	// If a user fails login this many times within the TTL window,
	// subsequent attempts are rejected without bcrypt verification.
	// db:"true" - Can be adjusted for security policy.
	MaxFailedAttempts int `mapstructure:"max_failed_attempts" json:"max_failed_attempts" default:"5" db:"true" validate:"omitempty,min=3,max=20"`
}

// DefaultConfig returns a Config with sensible defaults.
// This is used when no configuration file is present.
func DefaultConfig() *Config {
	return &Config{
		JWT:      *jwt.DefaultConfig(), // Use JWT package defaults
		Security: DefaultSecurityConfig(),
		Init:     DefaultInitConfig(),
	}
}

// DefaultInitConfig returns default platform initialization configuration.
func DefaultInitConfig() InitConfig {
	return InitConfig{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "", // Empty means random password will be generated
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

// IsReservedUsername checks if a username is reserved for system use.
// Reserved usernames can only be used during platform initialization.
func IsReservedUsername(username string) bool {
	if username == "" {
		return false
	}

	// Case-insensitive comparison
	lowerUsername := toLower(username)
	for _, reserved := range ReservedUsernames {
		if lowerUsername == reserved {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase (simple ASCII implementation)
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}
