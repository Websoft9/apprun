// Package mock provides mock implementations for testing.
package mock

import (
	"sync"
	"time"

	"apprun/pkg/config"
)

// ConfigProvider is a mock implementation of config.Provider for testing.
// It stores configuration values in memory and is safe for concurrent use.
type ConfigProvider struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewConfigProvider creates a new mock config provider with the given data.
func NewConfigProvider(data map[string]interface{}) *ConfigProvider {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &ConfigProvider{
		data: data,
	}
}

// Get retrieves a configuration value by key.
func (m *ConfigProvider) Get(key string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key]
}

// GetString retrieves a string configuration value.
func (m *ConfigProvider) GetString(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key].(string); ok {
		return val
	}
	return ""
}

// GetInt retrieves an integer configuration value.
func (m *ConfigProvider) GetInt(key string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key].(int); ok {
		return val
	}
	return 0
}

// GetInt64 retrieves an int64 configuration value.
func (m *ConfigProvider) GetInt64(key string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key].(int64); ok {
		return val
	}
	if val, ok := m.data[key].(int); ok {
		return int64(val)
	}
	return 0
}

// GetBool retrieves a boolean configuration value.
func (m *ConfigProvider) GetBool(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key].(bool); ok {
		return val
	}
	return false
}

// GetFloat64 retrieves a float64 configuration value.
func (m *ConfigProvider) GetFloat64(key string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key].(float64); ok {
		return val
	}
	return 0.0
}

// GetDuration retrieves a time.Duration configuration value.
func (m *ConfigProvider) GetDuration(key string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key].(time.Duration); ok {
		return val
	}
	// Support string duration like "1h30m"
	if val, ok := m.data[key].(string); ok {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return 0
}

// GetStringSlice retrieves a string slice configuration value.
func (m *ConfigProvider) GetStringSlice(key string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key].([]string); ok {
		return val
	}
	return nil
}

// IsSet checks if a configuration key exists.
func (m *ConfigProvider) IsSet(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok
}

// Set sets a configuration value.
func (m *ConfigProvider) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// Verify that ConfigProvider implements config.Provider interface at compile time.
var _ config.Provider = (*ConfigProvider)(nil)
