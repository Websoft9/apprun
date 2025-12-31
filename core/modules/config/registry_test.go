package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockModuleConfig struct {
	Value string `yaml:"value" db:"true"`
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, 0, registry.Count())
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	// Test successful registration
	cfg := &MockModuleConfig{Value: "test"}
	err := registry.Register("mock", cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, registry.Count())

	// Test duplicate registration
	err = registry.Register("mock", cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test empty namespace
	err = registry.Register("", cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "namespace cannot be empty")

	// Test nil config
	err = registry.Register("nil", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configStruct cannot be nil")
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	cfg := &MockModuleConfig{Value: "test"}

	// Register config
	err := registry.Register("mock", cfg)
	require.NoError(t, err)

	// Test get existing
	retrieved, exists := registry.Get("mock")
	assert.True(t, exists)
	assert.Equal(t, cfg, retrieved)

	// Test get non-existing
	_, exists = registry.Get("nonexistent")
	assert.False(t, exists)
}

func TestRegistry_GetAll(t *testing.T) {
	registry := NewRegistry()

	cfg1 := &MockModuleConfig{Value: "test1"}
	cfg2 := &MockModuleConfig{Value: "test2"}

	err := registry.Register("mock1", cfg1)
	require.NoError(t, err)
	err = registry.Register("mock2", cfg2)
	require.NoError(t, err)

	all := registry.GetAll()
	assert.Equal(t, 2, len(all))
	assert.Equal(t, cfg1, all["mock1"])
	assert.Equal(t, cfg2, all["mock2"])
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry()
	cfg := &MockModuleConfig{Value: "test"}

	assert.False(t, registry.Has("mock"))

	err := registry.Register("mock", cfg)
	require.NoError(t, err)

	assert.True(t, registry.Has("mock"))
	assert.False(t, registry.Has("nonexistent"))
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()
	assert.Equal(t, 0, registry.Count())

	cfg1 := &MockModuleConfig{Value: "test1"}
	cfg2 := &MockModuleConfig{Value: "test2"}

	err := registry.Register("mock1", cfg1)
	require.NoError(t, err)
	assert.Equal(t, 1, registry.Count())

	err = registry.Register("mock2", cfg2)
	require.NoError(t, err)
	assert.Equal(t, 2, registry.Count())
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()
	cfg := &MockModuleConfig{Value: "test"}

	// Register
	err := registry.Register("mock", cfg)
	require.NoError(t, err)

	// Concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, exists := registry.Get("mock")
			assert.True(t, exists)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
