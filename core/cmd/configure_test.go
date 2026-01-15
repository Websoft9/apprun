package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestUserConfig_SaveAndLoad(t *testing.T) {
	// Create temporary directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Override getUserConfigPath for testing
	originalGetUserConfigPath := getUserConfigPath
	getUserConfigPath = func() (string, error) {
		return configPath, nil
	}
	defer func() {
		getUserConfigPath = originalGetUserConfigPath
	}()

	// Test data
	testConfig := &UserConfig{
		Endpoint:   "https://test.example.com",
		APIKey:     "test-api-key-12345",
		ConfigPath: "/test/config/path",
	}

	// Test saveUserConfig
	err := saveUserConfig(testConfig, configPath)
	require.NoError(t, err, "saveUserConfig should not return error")

	// Verify file exists
	_, err = os.Stat(configPath)
	require.NoError(t, err, "config file should exist")

	// Test loadUserConfig
	loadedConfig, loadedPath, err := loadUserConfig()
	require.NoError(t, err, "loadUserConfig should not return error")
	assert.Equal(t, configPath, loadedPath, "loaded path should match")
	assert.Equal(t, testConfig.Endpoint, loadedConfig.Endpoint)
	assert.Equal(t, testConfig.APIKey, loadedConfig.APIKey)
	assert.Equal(t, testConfig.ConfigPath, loadedConfig.ConfigPath)
}

func TestUserConfig_LoadNonExistent(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	// Override getUserConfigPath for testing
	originalGetUserConfigPath := getUserConfigPath
	getUserConfigPath = func() (string, error) {
		return configPath, nil
	}
	defer func() {
		getUserConfigPath = originalGetUserConfigPath
	}()

	// Test loadUserConfig with non-existent file
	config, path, err := loadUserConfig()
	assert.Error(t, err, "should return error for non-existent file")
	assert.True(t, os.IsNotExist(err), "should be os.IsNotExist error")
	assert.Nil(t, config, "config should be nil")
	assert.Equal(t, configPath, path, "path should still be returned")
}

func TestUserConfig_InvalidYAML(t *testing.T) {
	// Create temporary directory and invalid YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Use truly invalid YAML that yaml.v3 will reject
	invalidYAML := `
endpoint: https://test.example.com
api_key: [unclosed bracket
`
	err := os.WriteFile(configPath, []byte(invalidYAML), 0600)
	require.NoError(t, err)

	// Override getUserConfigPath for testing
	originalGetUserConfigPath := getUserConfigPath
	getUserConfigPath = func() (string, error) {
		return configPath, nil
	}
	defer func() {
		getUserConfigPath = originalGetUserConfigPath
	}()

	// Test loadUserConfig with invalid YAML
	config, _, err := loadUserConfig()
	assert.Error(t, err, "should return error for invalid YAML")
	assert.Nil(t, config, "config should be nil")
}

func TestGetUserConfig_DefaultWhenNotExists(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	// Override getUserConfigPath for testing
	originalGetUserConfigPath := getUserConfigPath
	getUserConfigPath = func() (string, error) {
		return configPath, nil
	}
	defer func() {
		getUserConfigPath = originalGetUserConfigPath
	}()

	// Test GetUserConfig when file doesn't exist
	config, err := GetUserConfig()
	require.NoError(t, err, "should return default config without error")
	assert.NotNil(t, config, "config should not be nil")
	assert.Equal(t, "./config/default.yaml", config.ConfigPath, "should have default config path")
	assert.Equal(t, "", config.Endpoint, "endpoint should be empty")
	assert.Equal(t, "", config.APIKey, "api_key should be empty")
}

func TestGetValueOrEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "(not set)",
		},
		{
			name:     "non-empty string",
			input:    "test-value",
			expected: "test-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getValueOrEmpty(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDisplayConfig(t *testing.T) {
	// This test just ensures displayConfig doesn't panic
	config := &UserConfig{
		Endpoint:   "https://test.example.com",
		APIKey:     "test-api-key-12345",
		ConfigPath: "/test/config/path",
	}

	// Test with masking
	t.Run("with masking", func(t *testing.T) {
		assert.NotPanics(t, func() {
			displayConfig(config, true)
		})
	})

	// Test without masking
	t.Run("without masking", func(t *testing.T) {
		assert.NotPanics(t, func() {
			displayConfig(config, false)
		})
	})

	// Test with empty config
	t.Run("empty config", func(t *testing.T) {
		emptyConfig := &UserConfig{}
		assert.NotPanics(t, func() {
			displayConfig(emptyConfig, true)
		})
	})
}

func TestUserConfig_FilePermissions(t *testing.T) {
	// Create temporary directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Override getUserConfigPath for testing
	originalGetUserConfigPath := getUserConfigPath
	getUserConfigPath = func() (string, error) {
		return configPath, nil
	}
	defer func() {
		getUserConfigPath = originalGetUserConfigPath
	}()

	testConfig := &UserConfig{
		APIKey: "sensitive-api-key",
	}

	// Save config
	err := saveUserConfig(testConfig, configPath)
	require.NoError(t, err)

	// Check file permissions (should be 0600 - user read/write only)
	info, err := os.Stat(configPath)
	require.NoError(t, err)

	// On Unix systems, verify permissions are restrictive
	if os.Getuid() != 0 { // Skip if running as root
		mode := info.Mode().Perm()
		assert.Equal(t, os.FileMode(0600), mode, "config file should have 0600 permissions")
	}
}

func TestUserConfig_DirectoryCreation(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "config.yaml")

	// Override getUserConfigPath for testing
	originalGetUserConfigPath := getUserConfigPath
	getUserConfigPath = func() (string, error) {
		return nestedPath, nil
	}
	defer func() {
		getUserConfigPath = originalGetUserConfigPath
	}()

	testConfig := &UserConfig{
		ConfigPath: "./config/default.yaml",
	}

	// Save config - should create nested directories
	err := saveUserConfig(testConfig, nestedPath)
	require.NoError(t, err)

	// Verify directory was created
	dirPath := filepath.Dir(nestedPath)
	info, err := os.Stat(dirPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), "directory should be created")

	// Verify file exists
	_, err = os.Stat(nestedPath)
	require.NoError(t, err, "config file should exist")
}

func TestUserConfig_YAMLFormat(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Override getUserConfigPath for testing
	originalGetUserConfigPath := getUserConfigPath
	getUserConfigPath = func() (string, error) {
		return configPath, nil
	}
	defer func() {
		getUserConfigPath = originalGetUserConfigPath
	}()

	testConfig := &UserConfig{
		Endpoint:   "https://api.example.com",
		APIKey:     "test-key-123",
		ConfigPath: "./config/production.yaml",
	}

	// Save config
	err := saveUserConfig(testConfig, configPath)
	require.NoError(t, err)

	// Read and verify YAML format
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var parsedConfig UserConfig
	err = yaml.Unmarshal(data, &parsedConfig)
	require.NoError(t, err)

	assert.Equal(t, testConfig.Endpoint, parsedConfig.Endpoint)
	assert.Equal(t, testConfig.APIKey, parsedConfig.APIKey)
	assert.Equal(t, testConfig.ConfigPath, parsedConfig.ConfigPath)
}

func TestGetUserConfig_DefaultWhenNotExists_NilSafe(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	// Override getUserConfigPath for testing
	originalGetUserConfigPath := getUserConfigPath
	getUserConfigPath = func() (string, error) {
		return configPath, nil
	}
	defer func() {
		getUserConfigPath = originalGetUserConfigPath
	}()

	// Test GetUserConfig when file doesn't exist - should not panic
	config, err := GetUserConfig()
	require.NoError(t, err, "should return default config without error")
	assert.NotNil(t, config, "config should not be nil")
	assert.Equal(t, "./config/default.yaml", config.ConfigPath, "should have default config path")

	// Verify we can safely access all fields without panic
	assert.NotPanics(t, func() {
		_ = config.Endpoint
		_ = config.APIKey
		_ = config.ConfigPath
	}, "accessing config fields should not panic")
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"Empty string is valid (optional)", "", true},
		{"Valid HTTP URL", "http://example.com", true},
		{"Valid HTTPS URL", "https://example.com", true},
		{"Valid URL with port", "https://example.com:8080", true},
		{"Valid URL with path", "https://api.example.com/v1", true},
		{"Invalid - no scheme", "example.com", false},
		{"Invalid - ftp scheme", "ftp://example.com", false},
		{"Invalid - malformed", "ht!tp://invalid", false},
		{"Invalid - only scheme", "https://", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidURL(tt.input)
			assert.Equal(t, tt.valid, result, "isValidURL(%q) = %v, want %v", tt.input, result, tt.valid)
		})
	}
}

func TestIsValidPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"Empty string is valid (optional)", "", true},
		{"Valid absolute path", "/etc/config/app.yaml", true},
		{"Valid relative path", "./config/default.yaml", true},
		{"Valid relative path with ..", "../config/app.yaml", true},
		{"Valid home dir path", "~/config.yaml", true},
		{"Valid current dir", ".", true},
		{"Valid filename only", "config.yaml", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPath(tt.input)
			assert.Equal(t, tt.valid, result, "isValidPath(%q) = %v, want %v", tt.input, result, tt.valid)
		})
	}
}
