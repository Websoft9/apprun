package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// InitializeGlobalViper loads default.yaml into the global Viper instance.
// This is required for JWT TokenService to read jwt.secret from configuration.
// Should be called during application bootstrap (Phase 0).
func InitializeGlobalViper(configDir string) error {
	configFile := filepath.Join(configDir, "default.yaml")

	// Set config file path
	viper.SetConfigFile(configFile)

	// Enable automatic environment variable reading
	viper.AutomaticEnv()
	// Replace dots with underscores in environment variable names
	// Example: jwt.secret -> JWT_SECRET
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	// Manually bind environment variables for nested keys that Viper doesn't auto-detect
	// This is necessary because Viper's AutomaticEnv() doesn't work for nested keys by default
	_ = viper.BindEnv("auth.jwt.secret", "JWT_SECRET")
	_ = viper.BindEnv("auth.jwt.access_token_expiration", "JWT_ACCESS_TOKEN_EXPIRATION")
	_ = viper.BindEnv("auth.jwt.issuer", "JWT_ISSUER")
	_ = viper.BindEnv("auth.jwt.audience", "JWT_AUDIENCE")

	// Create aliases for JWT config to support both paths:
	// - auth.jwt.* (YAML config path)
	// - jwt.* (legacy TokenService path)
	viper.RegisterAlias("jwt.secret", "auth.jwt.secret")
	viper.RegisterAlias("jwt.access_token_expiration", "auth.jwt.access_token_expiration")
	viper.RegisterAlias("jwt.issuer", "auth.jwt.issuer")
	viper.RegisterAlias("jwt.audience", "auth.jwt.audience")
	viper.RegisterAlias("jwt.blacklist_enabled", "auth.jwt.blacklist_enabled")

	return nil
} // ViperProvider wraps Viper to implement the Provider interface.
// This adapter allows Viper to be used interchangeably with other config providers.
type ViperProvider struct {
	v *viper.Viper
}

// NewViperProvider creates a new Viper-based config provider.
// If v is nil, uses the global Viper instance.
func NewViperProvider(v *viper.Viper) *ViperProvider {
	if v == nil {
		// Use global Viper instance for backward compatibility
		return &ViperProvider{v: viper.GetViper()}
	}
	return &ViperProvider{v: v}
}

// Get retrieves a configuration value by key.
func (p *ViperProvider) Get(key string) interface{} {
	return p.v.Get(key)
}

// GetString retrieves a string configuration value.
func (p *ViperProvider) GetString(key string) string {
	return p.v.GetString(key)
}

// GetInt retrieves an integer configuration value.
func (p *ViperProvider) GetInt(key string) int {
	return p.v.GetInt(key)
}

// GetInt64 retrieves an int64 configuration value.
func (p *ViperProvider) GetInt64(key string) int64 {
	return p.v.GetInt64(key)
}

// GetBool retrieves a boolean configuration value.
func (p *ViperProvider) GetBool(key string) bool {
	return p.v.GetBool(key)
}

// GetFloat64 retrieves a float64 configuration value.
func (p *ViperProvider) GetFloat64(key string) float64 {
	return p.v.GetFloat64(key)
}

// GetDuration retrieves a time.Duration configuration value.
func (p *ViperProvider) GetDuration(key string) time.Duration {
	return p.v.GetDuration(key)
}

// GetStringSlice retrieves a string slice configuration value.
func (p *ViperProvider) GetStringSlice(key string) []string {
	return p.v.GetStringSlice(key)
}

// IsSet checks if a configuration key exists.
func (p *ViperProvider) IsSet(key string) bool {
	return p.v.IsSet(key)
}

// Set sets a configuration value.
func (p *ViperProvider) Set(key string, value interface{}) {
	p.v.Set(key, value)
}

// Verify that ViperProvider implements Provider interface at compile time.
var _ Provider = (*ViperProvider)(nil)
