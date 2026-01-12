package config

import (
	"time"

	"github.com/spf13/viper"
)

// ViperProvider wraps Viper to implement the Provider interface.
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
