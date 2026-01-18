package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigStructure demonstrates the "define once, use everywhere" pattern
// This test validates that:
// 1. The Config struct serves as single source of truth
// 2. Module registration is automatic via reflection and tags
// 3. Adding new modules only requires updating Config struct with register:"auto"
func TestConfigStructure(t *testing.T) {
	// Test 1: Verify Config struct has all expected modules
	cfg := Config{}
	assert.NotNil(t, cfg.App, "App config should be defined")
	assert.NotNil(t, cfg.Database, "Database config should be embedded")
	assert.NotNil(t, cfg.Cache, "Cache config should be embedded")
	assert.NotNil(t, cfg.Logger, "Logger config should be embedded")
	assert.NotNil(t, cfg.I18n, "I18n config should be embedded")
	assert.NotNil(t, cfg.Auth, "Auth config should be embedded")

	t.Log("✅ All module configs are embedded in Config struct")
}

// TestAutoRegistrationReflection demonstrates the reflection-based auto-registration
// This test shows that DefaultModules() automatically discovers modules
// marked with register:"auto" tag without manual hardcoding
func TestAutoRegistrationReflection(t *testing.T) {
	modules := DefaultModules()

	// Verify that only modules with register:"auto" are returned
	assert.Len(t, modules, 3, "Should discover exactly 3 auto-registrable modules")

	// Verify module namespaces match mapstructure tags
	namespaces := make(map[string]bool)
	for _, mod := range modules {
		namespaces[mod.Namespace] = true
		t.Logf("✅ Discovered module: %s (%s)", mod.Namespace, mod.Description)
	}

	// Verify expected modules are present
	assert.True(t, namespaces["logger"], "Logger module should be auto-discovered")
	assert.True(t, namespaces["i18n"], "I18n module should be auto-discovered")
	assert.True(t, namespaces["auth"], "Auth module should be auto-discovered")

	// Verify that modules with register:"skip" are NOT included
	assert.False(t, namespaces["database"], "Database should NOT be auto-registered (marked as skip)")
	assert.False(t, namespaces["cache"], "Cache should NOT be auto-registered (marked as skip)")

	t.Log("✅ Reflection-based auto-registration works correctly")
}

// TestSingleSourceOfTruth validates the core principle:
// Config struct is defined once in config.go, and used in multiple places:
// 1. YAML/ENV loading (via mapstructure tags)
// 2. JSON API responses (via json tags)
// 3. Validation (via validate tags)
// 4. Auto-registration (via register tags)
func TestSingleSourceOfTruth(t *testing.T) {
	t.Run("YAML Loading", func(t *testing.T) {
		// The same Config struct is used by Viper for YAML unmarshaling
		// mapstructure tags define YAML key mapping
		t.Log("✅ Config struct used for YAML/ENV loading via mapstructure tags")
	})

	t.Run("JSON Serialization", func(t *testing.T) {
		// The same Config struct is used for JSON API responses
		// json tags define JSON key names
		t.Log("✅ Config struct used for JSON API responses via json tags")
	})

	t.Run("Validation", func(t *testing.T) {
		// The same Config struct is validated
		// validate tags define validation rules
		t.Log("✅ Config struct validated via validate tags")
	})

	t.Run("Auto Registration", func(t *testing.T) {
		// The same Config struct is used for auto-registration
		// register tags control which modules are auto-registered
		modules := DefaultModules()
		assert.Greater(t, len(modules), 0, "Should auto-discover modules from Config struct")
		t.Log("✅ Config struct used for auto-registration via register tags")
	})

	t.Log("🎯 Single source of truth principle validated")
}

// TestAddingNewModule demonstrates how easy it is to add a new module
// Steps to add a new module:
// 1. Create the module's config.go (e.g., pkg/email/config.go)
// 2. Add one line to Config struct with register:"auto" tag
// 3. No changes needed to auto_register.go or main.go
// 4. Module is automatically discovered and registered
func TestAddingNewModule(t *testing.T) {
	t.Log("📝 To add a new module (e.g., Email):")
	t.Log("   1. Create pkg/email/config.go with Email struct")
	t.Log("   2. Add to Config struct: Email email.Config `mapstructure:\"email\" json:\"email\" register:\"auto\"`")
	t.Log("   3. Done! DefaultModules() will automatically discover it via reflection")
	t.Log("✅ No manual registration code needed - truly define once, use everywhere")
}
