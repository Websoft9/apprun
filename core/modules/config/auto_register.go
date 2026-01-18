package config

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// ModuleRegistration defines module registration information
type ModuleRegistration struct {
	Namespace    string      // Module namespace (e.g., "logger", "i18n", "auth")
	ConfigStruct interface{} // Pointer to config struct (e.g., &logger.Config{})
	Description  string      // Optional: description for logging
}

// DefaultModules returns the list of default modules to register.
// It uses reflection to read the Config struct and automatically discovers
// all fields marked with `register:"auto"` tag.
//
// This achieves "define once, use everywhere":
// - Single source of truth: Config struct in config.go
// - Automatic discovery: No manual list maintenance
// - Tag-driven: Simple to add/remove modules
func DefaultModules() []ModuleRegistration {
	var modules []ModuleRegistration

	// Create a Config instance to inspect its structure
	cfg := Config{}
	cfgType := reflect.TypeOf(cfg)

	// Iterate through all fields in the Config struct
	for i := 0; i < cfgType.NumField(); i++ {
		field := cfgType.Field(i)

		// Check if field has register:"auto" tag
		registerTag := field.Tag.Get("register")
		if registerTag != "auto" {
			continue // Skip fields without auto-registration
		}

		// Get the namespace from mapstructure tag (fallback to lowercase field name)
		namespace := field.Tag.Get("mapstructure")
		if namespace == "" {
			namespace = strings.ToLower(field.Name)
		}

		// Get the description from description tag
		description := field.Tag.Get("description")

		// Get the field value and create a new pointer to its type
		fieldValue := reflect.ValueOf(cfg).Field(i)
		configStruct := reflect.New(fieldValue.Type()).Interface()

		modules = append(modules, ModuleRegistration{
			Namespace:    namespace,
			ConfigStruct: configStruct,
			Description:  description,
		})
	}

	return modules
}

// RegisterDefaultModules registers all default modules to the registry
// This is the recommended way to register modules in production
func RegisterDefaultModules(registry *ConfigRegistry) error {
	modules := DefaultModules()

	for _, mod := range modules {
		if err := registry.Register(mod.Namespace, mod.ConfigStruct); err != nil {
			return fmt.Errorf("failed to register module '%s': %w", mod.Namespace, err)
		}
		if mod.Description != "" {
			log.Printf("✅ %s registered", mod.Description)
		} else {
			log.Printf("✅ %s module registered with config center", mod.Namespace)
		}
	}

	return nil
}

// RegisterModules registers a custom list of modules to the registry
// This allows for flexible module registration in tests or custom scenarios
func RegisterModules(registry *ConfigRegistry, modules []ModuleRegistration) error {
	for _, mod := range modules {
		if err := registry.Register(mod.Namespace, mod.ConfigStruct); err != nil {
			return fmt.Errorf("failed to register module '%s': %w", mod.Namespace, err)
		}
	}
	return nil
}
