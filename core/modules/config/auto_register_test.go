package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultModules(t *testing.T) {
	modules := DefaultModules()

	assert.Len(t, modules, 4, "Should have 4 default modules")

	// Verify module names
	namespaces := make(map[string]bool)
	for _, mod := range modules {
		namespaces[mod.Namespace] = true
	}

	assert.True(t, namespaces["logger"], "Should include logger module")
	assert.True(t, namespaces["i18n"], "Should include i18n module")
	assert.True(t, namespaces["auth"], "Should include auth module")
	assert.True(t, namespaces["audit"], "Should include audit module")

	// Verify all modules have config structs
	for _, mod := range modules {
		assert.NotNil(t, mod.ConfigStruct, "Module %s should have config struct", mod.Namespace)
		assert.NotEmpty(t, mod.Description, "Module %s should have description", mod.Namespace)
	}
}

func TestRegisterDefaultModules(t *testing.T) {
	registry := NewRegistry()

	err := RegisterDefaultModules(registry)
	require.NoError(t, err)

	// Verify all modules are registered
	assert.Equal(t, 4, registry.Count(), "Should register 4 modules")
	assert.True(t, registry.Has("logger"))
	assert.True(t, registry.Has("i18n"))
	assert.True(t, registry.Has("auth"))
	assert.True(t, registry.Has("audit"))
}

func TestRegisterDefaultModules_Duplicate(t *testing.T) {
	registry := NewRegistry()

	// First registration should succeed
	err := RegisterDefaultModules(registry)
	require.NoError(t, err)

	// Second registration should fail (duplicate)
	err = RegisterDefaultModules(registry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegisterModules(t *testing.T) {
	type TestConfig struct {
		Value string `mapstructure:"value"`
	}

	registry := NewRegistry()

	customModules := []ModuleRegistration{
		{
			Namespace:    "test1",
			ConfigStruct: &TestConfig{},
			Description:  "Test module 1",
		},
		{
			Namespace:    "test2",
			ConfigStruct: &TestConfig{},
			Description:  "Test module 2",
		},
	}

	err := RegisterModules(registry, customModules)
	require.NoError(t, err)

	assert.Equal(t, 2, registry.Count())
	assert.True(t, registry.Has("test1"))
	assert.True(t, registry.Has("test2"))
}

func TestRegisterModules_Empty(t *testing.T) {
	registry := NewRegistry()

	err := RegisterModules(registry, []ModuleRegistration{})
	require.NoError(t, err)

	assert.Equal(t, 0, registry.Count())
}

func TestRegisterModules_WithDefaultModules(t *testing.T) {
	type CustomConfig struct {
		Setting string `mapstructure:"setting"`
	}

	registry := NewRegistry()

	// Register default modules first
	err := RegisterDefaultModules(registry)
	require.NoError(t, err)

	// Then register custom modules
	customModules := []ModuleRegistration{
		{
			Namespace:    "custom",
			ConfigStruct: &CustomConfig{},
			Description:  "Custom module",
		},
	}

	err = RegisterModules(registry, customModules)
	require.NoError(t, err)

	// Should have all modules (4 default + 1 custom = 5)
	assert.Equal(t, 5, registry.Count())
	assert.True(t, registry.Has("logger"))
	assert.True(t, registry.Has("i18n"))
	assert.True(t, registry.Has("auth"))
	assert.True(t, registry.Has("audit"))
	assert.True(t, registry.Has("custom"))
}
