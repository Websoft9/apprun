package database

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	// Set environment variables (new naming convention: DATABASE_*)
	_ = os.Setenv("DATABASE_DRIVER", "postgres")
	_ = os.Setenv("DATABASE_HOST", "testhost")
	_ = os.Setenv("DATABASE_PORT", "5433")
	_ = os.Setenv("DATABASE_USER", "testuser")
	_ = os.Setenv("DATABASE_PASSWORD", "testpassword123")
	_ = os.Setenv("DATABASE_DB_NAME", "testdb")
	defer func() {
		_ = os.Unsetenv("DATABASE_DRIVER")
		_ = os.Unsetenv("DATABASE_HOST")
		_ = os.Unsetenv("DATABASE_PORT")
		_ = os.Unsetenv("DATABASE_USER")
		_ = os.Unsetenv("DATABASE_PASSWORD")
		_ = os.Unsetenv("DATABASE_DB_NAME")
	}()

	cfg := DefaultConfig()

	assert.Equal(t, "postgres", cfg.Driver)
	assert.Equal(t, "testhost", cfg.Host)
	assert.Equal(t, 5433, cfg.Port)
	assert.Equal(t, "testuser", cfg.User)
	assert.Equal(t, "testpassword123", cfg.Password)
	assert.Equal(t, "testdb", cfg.DBName)
}

func TestDefaultConfig_Defaults(t *testing.T) {
	// Set only required password (new naming convention: DATABASE_PASSWORD)
	_ = os.Setenv("DATABASE_PASSWORD", "required123")
	defer func() {
		_ = os.Unsetenv("DATABASE_PASSWORD")
	}()

	cfg := DefaultConfig()

	assert.Equal(t, "postgres", cfg.Driver)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "postgres", cfg.User)
	assert.Equal(t, "required123", cfg.Password)
	assert.Equal(t, "apprun", cfg.DBName)
}

func TestConnect_InvalidConfig(t *testing.T) {
	ctx := context.Background()

	cfg := &Config{
		Driver:   "postgres",
		Host:     "invalid-host-that-does-not-exist",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "testdb",
	}

	_, err := Connect(ctx, cfg)
	assert.Error(t, err)
	// Error can be either from connection or schema creation
	assert.True(t, err != nil)
} // Note: TestConnect_Success and TestClient_Ping require a real database connection
// These should be run as integration tests with a test database
