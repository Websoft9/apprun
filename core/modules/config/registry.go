package config

import (
	"fmt"
	"sync"
)

// ConfigRegistry manages module config registration
// Allows business modules to register their config structs independently
type ConfigRegistry struct {
	mu      sync.RWMutex
	modules map[string]interface{} // namespace -> config struct pointer
}

// NewRegistry creates a new config registry
func NewRegistry() *ConfigRegistry {
	return &ConfigRegistry{
		modules: make(map[string]interface{}),
	}
}

// Register registers a module's config struct
// namespace: module name (e.g., "logger", "user", "project")
// configStruct: pointer to config struct (e.g., &logger.Config{})
func (r *ConfigRegistry) Register(namespace string, configStruct interface{}) error {
	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	if configStruct == nil {
		return fmt.Errorf("configStruct cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.modules[namespace]; exists {
		return fmt.Errorf("module '%s' already registered", namespace)
	}

	r.modules[namespace] = configStruct
	return nil
}

// Get retrieves a registered module's config struct
func (r *ConfigRegistry) Get(namespace string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, exists := r.modules[namespace]
	return cfg, exists
}

// GetAll returns all registered module configs
func (r *ConfigRegistry) GetAll() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]interface{}, len(r.modules))
	for k, v := range r.modules {
		result[k] = v
	}
	return result
}

// Has checks if a namespace is registered
func (r *ConfigRegistry) Has(namespace string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.modules[namespace]
	return exists
}

// Count returns the number of registered modules
func (r *ConfigRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.modules)
}
