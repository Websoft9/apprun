package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigToEnv(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "default.yaml")

	// Write test config file
	configContent := `
server:
  http_port: "9090"
  https_port: "9443"
  ssl_cert_file: "/path/to/cert"
  ssl_key_file: "/path/to/key"
  shutdown_timeout: "45s"
  enable_http_with_https: false

database:
  driver: mysql
  host: testhost
  port: 3306
  user: testuser
  password: testpass
  db_name: testdb
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Clear any existing env vars
	clearTestEnvVars()

	// Load config to env
	err = LoadConfigToEnv(tempDir)
	assert.NoError(t, err)

	// Verify server config was loaded with SERVER_ prefix
	assert.Equal(t, "9090", os.Getenv("SERVER_HTTP_PORT"))
	assert.Equal(t, "9443", os.Getenv("SERVER_HTTPS_PORT"))
	assert.Equal(t, "/path/to/cert", os.Getenv("SERVER_SSL_CERT_FILE"))
	assert.Equal(t, "/path/to/key", os.Getenv("SERVER_SSL_KEY_FILE"))
	assert.Equal(t, "45s", os.Getenv("SERVER_SHUTDOWN_TIMEOUT"))
	assert.Equal(t, "false", os.Getenv("SERVER_ENABLE_HTTP_WITH_HTTPS"))

	// Verify database config was loaded with DATABASE_ prefix
	assert.Equal(t, "mysql", os.Getenv("DATABASE_DRIVER"))
	assert.Equal(t, "testhost", os.Getenv("DATABASE_HOST"))
	assert.Equal(t, "3306", os.Getenv("DATABASE_PORT"))
	assert.Equal(t, "testuser", os.Getenv("DATABASE_USER"))
	assert.Equal(t, "testpass", os.Getenv("DATABASE_PASSWORD"))
	assert.Equal(t, "testdb", os.Getenv("DATABASE_DB_NAME"))

	// Cleanup
	clearTestEnvVars()
}

func TestLoadConfigToEnv_RespectExistingEnv(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "default.yaml")

	// Write test config file
	configContent := `
server:
  http_port: "9090"
database:
  host: testhost
  port: 3306
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Set environment variable before loading config (using new naming convention)
	os.Setenv("SERVER_HTTP_PORT", "8888")
	os.Setenv("DATABASE_HOST", "prodhost")

	// Load config to env
	err = LoadConfigToEnv(tempDir)
	assert.NoError(t, err)

	// Verify existing env vars were NOT overwritten
	assert.Equal(t, "8888", os.Getenv("SERVER_HTTP_PORT"), "Existing SERVER_HTTP_PORT should not be overwritten")
	assert.Equal(t, "prodhost", os.Getenv("DATABASE_HOST"), "Existing DATABASE_HOST should not be overwritten")

	// Verify other values from config file were loaded
	assert.Equal(t, "3306", os.Getenv("DATABASE_PORT"))

	// Cleanup
	clearTestEnvVars()
}

func TestLoadConfigToEnv_FileNotExists(t *testing.T) {
	// Use non-existent directory
	err := LoadConfigToEnv("/non/existent/dir")
	assert.NoError(t, err, "Should not error when config file doesn't exist")
}

func TestLoadConfigToEnv_InvalidYAML(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "default.yaml")

	// Write invalid YAML
	configContent := `
invalid yaml content
  no proper structure
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Load config should return error
	err = LoadConfigToEnv(tempDir)
	assert.Error(t, err)
}

func TestSetEnvIfNotExists(t *testing.T) {
	clearTestEnvVars()

	// Test setting new env var
	setEnvIfNotExists("TEST_VAR", "value1")
	assert.Equal(t, "value1", os.Getenv("TEST_VAR"))

	// Test not overwriting existing env var
	setEnvIfNotExists("TEST_VAR", "value2")
	assert.Equal(t, "value1", os.Getenv("TEST_VAR"), "Should not overwrite existing value")

	// Test not setting empty value
	setEnvIfNotExists("TEST_EMPTY", "")
	assert.Equal(t, "", os.Getenv("TEST_EMPTY"))

	// Cleanup
	os.Unsetenv("TEST_VAR")
	os.Unsetenv("TEST_EMPTY")
}

func clearTestEnvVars() {
	// Clear server env vars (new naming convention: SERVER_*)
	os.Unsetenv("SERVER_HTTP_PORT")
	os.Unsetenv("SERVER_HTTPS_PORT")
	os.Unsetenv("SERVER_SSL_CERT_FILE")
	os.Unsetenv("SERVER_SSL_KEY_FILE")
	os.Unsetenv("SERVER_SHUTDOWN_TIMEOUT")
	os.Unsetenv("SERVER_ENABLE_HTTP_WITH_HTTPS")

	// Clear database env vars (new naming convention: DATABASE_*)
	os.Unsetenv("DATABASE_DRIVER")
	os.Unsetenv("DATABASE_HOST")
	os.Unsetenv("DATABASE_PORT")
	os.Unsetenv("DATABASE_USER")
	os.Unsetenv("DATABASE_PASSWORD")
	os.Unsetenv("DATABASE_DB_NAME")
}
