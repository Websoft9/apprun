// Package jwt provides JWT token generation, validation, and authentication utilities.
package jwt

import (
	"time"

	"apprun/pkg/env"

	"github.com/google/uuid"
)

// ============================================================================
// Module Constants
// ============================================================================

const (
	// MinSecretLength is the minimum required length for JWT secret (for security)
	MinSecretLength = 32

	// DefaultExpiry is the default token expiration duration
	DefaultExpiry = 24 * time.Hour

	// DefaultIssuer is the default JWT issuer identifier
	DefaultIssuer = "apprun-platform"

	// DefaultAudience is the default JWT audience identifier
	DefaultAudience = "apprun-api"
)

// DefaultWhitelistPaths are the default whitelist paths that bypass JWT authentication
var DefaultWhitelistPaths = []string{
	"/api/v1/auth/register",
	"/api/v1/auth/login",
	"/api/auth/login",
	"/api/auth/register",
	"/health",
	"/metrics",
}

// ============================================================================
// Configuration Structures
// ============================================================================

// Config defines the configuration for the JWT system.
// This struct follows the Configuration Center architecture (Story 10).
// It can be used independently by pkg/jwt or embedded by modules/auth.
type Config struct {
	// Secret is the signing key for JWT tokens (must be ≥32 characters).
	// MUST be set via environment variable JWT_SECRET.
	// db:"false" - Cannot be updated via config center API (security-critical).
	Secret string `mapstructure:"secret" json:"secret" default:"" db:"false" validate:"required,min=32"`

	// AccessTokenExpiration defines how long access tokens are valid.
	// Supports duration strings like "24h", "1h30m", "7d" (7*24h).
	// db:"true" - Can be updated dynamically via config center.
	AccessTokenExpiration time.Duration `mapstructure:"access_token_expiration" json:"access_token_expiration" default:"24h" db:"true" validate:"min=1h,max=168h"`

	// RefreshTokenExpiration defines how long refresh tokens are valid (for future use).
	// db:"true" - Can be updated dynamically via config center.
	RefreshTokenExpiration time.Duration `mapstructure:"refresh_token_expiration" json:"refresh_token_expiration" default:"168h" db:"true" validate:"min=24h,max=720h"`

	// Issuer is the JWT "iss" claim value.
	// db:"false" - Infrastructure config, not changeable at runtime.
	Issuer string `mapstructure:"issuer" json:"issuer" default:"apprun-platform" db:"false" validate:"required"`

	// Audience is the JWT "aud" claim value (optional, for token validation).
	// db:"false" - Infrastructure config, not changeable at runtime.
	Audience string `mapstructure:"audience" json:"audience" default:"apprun-api" db:"false" validate:"required"`

	// WhitelistPaths defines routes that bypass JWT authentication.
	// db:"true" - Can be updated to add/remove whitelisted routes at runtime.
	// Note: Paths should start with "/" (e.g., "/api/auth/login")
	// Validation: Only ensures the list is not empty (min=1)
	WhitelistPaths []string `mapstructure:"whitelist_paths" json:"whitelist_paths" db:"true" validate:"min=1"`
}

// DefaultConfig returns the default JWT configuration.
// Priority: JWT_SECRET env var > code defaults
// Note: Secret will be loaded from Viper config or environment variable
func DefaultConfig() *Config {
	// Check environment variable first
	secret := env.Get("JWT_SECRET", "")

	return &Config{
		Secret:                 secret, // Will be overridden by Viper if present in YAML
		AccessTokenExpiration:  DefaultExpiry,
		RefreshTokenExpiration: 7 * DefaultExpiry, // 7 days
		Issuer:                 DefaultIssuer,
		Audience:               DefaultAudience,
		WhitelistPaths:         DefaultWhitelistPaths,
	}
}

// ToRuntimeConfig converts Config to runtime configuration with optimized data structures.
// It parses duration strings and builds a whitelist map for O(1) path lookup.
func (c *Config) ToRuntimeConfig() (*RuntimeConfig, error) {
	// Build whitelist map for O(1) lookup
	whitelist := make(map[string]bool, len(c.WhitelistPaths))
	for _, path := range c.WhitelistPaths {
		whitelist[path] = true
	}

	return &RuntimeConfig{
		Secret:                 c.Secret,
		AccessTokenExpiration:  c.AccessTokenExpiration,
		RefreshTokenExpiration: c.RefreshTokenExpiration,
		Issuer:                 c.Issuer,
		Audience:               c.Audience,
		Whitelist:              whitelist,
	}, nil
}

// RuntimeConfig is the internal configuration used by JWT functions.
// This struct is optimized for runtime performance (e.g., map lookup instead of slice iteration).
type RuntimeConfig struct {
	Secret                 string
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
	Issuer                 string
	Audience               string
	Whitelist              map[string]bool // O(1) lookup for whitelist paths
}

// IsPathWhitelisted checks if a given path is in the whitelist.
func (r *RuntimeConfig) IsPathWhitelisted(path string) bool {
	return r.Whitelist[path]
}

// ============================================================================
// Backward Compatibility (Legacy API)
// ============================================================================

// LegacyConfig is the old config structure for backward compatibility.
//
// Deprecated: Use Config instead. Will be removed in v2.0.
type LegacyConfig struct {
	Secret         string   `mapstructure:"secret" json:"secret"`
	Expiry         string   `mapstructure:"expiry" json:"expiry"` // Duration string like "24h"
	Issuer         string   `mapstructure:"issuer" json:"issuer"`
	WhitelistPaths []string `mapstructure:"whitelist_paths" json:"whitelist_paths"`
}

// ToConfig converts LegacyConfig to the new Config structure.
//
// Deprecated: Use Config directly. Will be removed in v2.0.
func (lc *LegacyConfig) ToConfig() (*Config, error) {
	expiry, err := time.ParseDuration(lc.Expiry)
	if err != nil {
		expiry = DefaultExpiry
	}

	return &Config{
		Secret:                 lc.Secret,
		AccessTokenExpiration:  expiry,
		RefreshTokenExpiration: 7 * expiry,
		Issuer:                 lc.Issuer,
		Audience:               DefaultAudience,
		WhitelistPaths:         lc.WhitelistPaths,
	}, nil
}

// ============================================================================
// Utility Functions
// ============================================================================

// GenerateTokenID creates a unique identifier for token tracking (blacklist, audit logs).
func GenerateTokenID() string {
	return uuid.New().String()
}
