// Package jwt provides JWT token generation, validation, and authentication utilities.
package jwt

import "time"

// Config defines the configuration for the JWT system
// Follows pkg/i18n/config.go pattern for consistency with config center registry
type Config struct {
	// Secret is the signing key for JWT tokens (must be ≥32 characters)
	Secret string `yaml:"secret" json:"secret" default:"" db:"false" validate:"required,min=32"`

	// Expiry is the token expiration duration (e.g., "24h", "7d")
	Expiry string `yaml:"expiry" json:"expiry" default:"24h" db:"false" validate:"required"`

	// Issuer is the token issuer identifier
	Issuer string `yaml:"issuer" json:"issuer" default:"apprun" db:"false" validate:"required"`

	// WhitelistPaths are the paths that bypass JWT authentication
	WhitelistPaths []string `yaml:"whitelist_paths" json:"whitelist_paths" default:"['/api/v1/auth/register','/api/v1/auth/login','/health','/metrics']" db:"false"`
}

// DefaultConfig returns the default JWT configuration
func DefaultConfig() Config {
	return Config{
		Secret:         "", // Must be set via environment variable
		Expiry:         "24h",
		Issuer:         "apprun",
		WhitelistPaths: []string{"/api/v1/auth/register", "/api/v1/auth/login", "/health", "/metrics"},
	}
}

// ToRuntimeConfig converts Config to runtime configuration with parsed duration
func (c *Config) ToRuntimeConfig() (*RuntimeConfig, error) {
	duration, err := time.ParseDuration(c.Expiry)
	if err != nil {
		return nil, err
	}

	// Build whitelist map for O(1) lookup
	whitelist := make(map[string]bool, len(c.WhitelistPaths))
	for _, path := range c.WhitelistPaths {
		whitelist[path] = true
	}

	return &RuntimeConfig{
		Secret:    c.Secret,
		Expiry:    duration,
		Issuer:    c.Issuer,
		Whitelist: whitelist,
	}, nil
}

// RuntimeConfig is the internal configuration used by JWT functions
type RuntimeConfig struct {
	Secret    string
	Expiry    time.Duration
	Issuer    string
	Whitelist map[string]bool // O(1) lookup for whitelist paths
}
