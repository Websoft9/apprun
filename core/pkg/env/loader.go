package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// LoadConfigToEnv loads server and database configuration from default.yaml
// and sets them as environment variables (only if not already set).
// This allows configuration files to provide defaults while respecting
// existing environment variables.
//
// Environment variable naming convention:
// - Pattern: {GROUP}_UPPERCASE_{KEY}_UPPERCASE
// - Example: server.http_port → SERVER_HTTP_PORT
// - Example: database.host → DATABASE_HOST
//
// Priority order (highest to lowest):
// 1. Existing environment variables (runtime export, docker -e, .env file)
// 2. Configuration file (default.yaml)
// 3. Code defaults (in DefaultConfig() functions)
func LoadConfigToEnv(configDir string) error {
	configFile := filepath.Join(configDir, "default.yaml")

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// File doesn't exist, use pure environment variable mode
		return nil
	}

	// Setup viper to read the config file
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	// Load server configuration dynamically
	// Converts: server.http_port → SERVER_HTTP_PORT
	if err := loadSectionToEnv("server", viper.Sub("server")); err != nil {
		return fmt.Errorf("failed to load server config: %w", err)
	}

	// Load database configuration dynamically
	// Converts: database.host → DATABASE_HOST
	if err := loadSectionToEnv("database", viper.Sub("database")); err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}

	return nil
}

// loadSectionToEnv loads a configuration section and converts it to environment variables
// Naming convention: {GROUP_UPPERCASE}_{KEY_UPPERCASE}
// Example: "server" section with "http_port" key → SERVER_HTTP_PORT
func loadSectionToEnv(group string, section *viper.Viper) error {
	if section == nil {
		return nil // Section doesn't exist, skip
	}

	groupPrefix := toEnvPrefix(group)
	settings := section.AllSettings()

	for key, value := range settings {
		envKey := groupPrefix + "_" + toEnvKey(key)
		envValue := formatEnvValue(value)
		setEnvIfNotExists(envKey, envValue)
	}

	return nil
}

// toEnvPrefix converts a group name to environment variable prefix
// Example: "server" → "SERVER", "database" → "DATABASE"
func toEnvPrefix(group string) string {
	return strings.ToUpper(group)
}

// toEnvKey converts a config key to environment variable key format
// Example: "http_port" → "HTTP_PORT", "ssl_cert_file" → "SSL_CERT_FILE"
func toEnvKey(key string) string {
	return strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
}

// formatEnvValue formats a value for environment variable storage
// Converts various types to string representation
func formatEnvValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int64, int32:
		return fmt.Sprintf("%d", v)
	case float64, float32:
		return fmt.Sprintf("%v", v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// setEnvIfNotExists sets an environment variable only if it's not already set
// This respects the priority: runtime env > config file > code defaults
func setEnvIfNotExists(key, value string) {
	if value == "" {
		return // Don't set empty values
	}
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
	// If env var already exists, skip (higher priority)
}
