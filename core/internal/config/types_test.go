package config

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config",
			config: Config{
				App: struct {
					Name    string `validate:"required,min=1" default:"apprun" db:"false"`
					Version string `validate:"required" default:"1.0.0" db:"false"`
				}{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Database: struct {
					Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"`
					Host     string `validate:"required" default:"localhost" db:"false"`
					Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`
					User     string `validate:"required" default:"postgres" db:"false"`
					Password string `validate:"required,min=8" db:"false"`
					DBName   string `validate:"required" default:"apprun" db:"false"`
				}{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					User:     "testuser",
					Password: "testpassword",
					DBName:   "testdb",
				},
				POC: struct {
					Enabled  bool   `default:"true" db:"true"`
					Database string `validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"`
					APIKey   string `validate:"required,min=10" db:"true"`
				}{
					Enabled:  true,
					Database: "postgres://user:pass@localhost:5432/test",
					APIKey:   "test-api-key-12345",
				},
			},
			expectError: false,
		},
		{
			name: "missing app name",
			config: Config{
				App: struct {
					Name    string `validate:"required,min=1" default:"apprun" db:"false"`
					Version string `validate:"required" default:"1.0.0" db:"false"`
				}{
					Name:    "",
					Version: "1.0.0",
				},
				Database: struct {
					Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"`
					Host     string `validate:"required" default:"localhost" db:"false"`
					Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`
					User     string `validate:"required" default:"postgres" db:"false"`
					Password string `validate:"required,min=8" db:"false"`
					DBName   string `validate:"required" default:"apprun" db:"false"`
				}{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					User:     "testuser",
					Password: "testpassword",
					DBName:   "testdb",
				},
				POC: struct {
					Enabled  bool   `default:"true" db:"true"`
					Database string `validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"`
					APIKey   string `validate:"required,min=10" db:"true"`
				}{
					Enabled:  true,
					Database: "postgres://user:pass@localhost:5432/test",
					APIKey:   "test-api-key-12345",
				},
			},
			expectError: true,
		},
		{
			name: "invalid database driver",
			config: Config{
				App: struct {
					Name    string `validate:"required,min=1" default:"apprun" db:"false"`
					Version string `validate:"required" default:"1.0.0" db:"false"`
				}{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Database: struct {
					Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"`
					Host     string `validate:"required" default:"localhost" db:"false"`
					Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`
					User     string `validate:"required" default:"postgres" db:"false"`
					Password string `validate:"required,min=8" db:"false"`
					DBName   string `validate:"required" default:"apprun" db:"false"`
				}{
					Driver:   "invalid",
					Host:     "localhost",
					Port:     5432,
					User:     "testuser",
					Password: "testpassword",
					DBName:   "testdb",
				},
				POC: struct {
					Enabled  bool   `default:"true" db:"true"`
					Database string `validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"`
					APIKey   string `validate:"required,min=10" db:"true"`
				}{
					Enabled:  true,
					Database: "postgres://user:pass@localhost:5432/test",
					APIKey:   "test-api-key-12345",
				},
			},
			expectError: true,
		},
		{
			name: "invalid port range",
			config: Config{
				App: struct {
					Name    string `validate:"required,min=1" default:"apprun" db:"false"`
					Version string `validate:"required" default:"1.0.0" db:"false"`
				}{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Database: struct {
					Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"`
					Host     string `validate:"required" default:"localhost" db:"false"`
					Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`
					User     string `validate:"required" default:"postgres" db:"false"`
					Password string `validate:"required,min=8" db:"false"`
					DBName   string `validate:"required" default:"apprun" db:"false"`
				}{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     70000, // Invalid port
					User:     "testuser",
					Password: "testpassword",
					DBName:   "testdb",
				},
				POC: struct {
					Enabled  bool   `default:"true" db:"true"`
					Database string `validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"`
					APIKey   string `validate:"required,min=10" db:"true"`
				}{
					Enabled:  true,
					Database: "postgres://user:pass@localhost:5432/test",
					APIKey:   "test-api-key-12345",
				},
			},
			expectError: true,
		},
		{
			name: "weak password",
			config: Config{
				App: struct {
					Name    string `validate:"required,min=1" default:"apprun" db:"false"`
					Version string `validate:"required" default:"1.0.0" db:"false"`
				}{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Database: struct {
					Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"`
					Host     string `validate:"required" default:"localhost" db:"false"`
					Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`
					User     string `validate:"required" default:"postgres" db:"false"`
					Password string `validate:"required,min=8" db:"false"`
					DBName   string `validate:"required" default:"apprun" db:"false"`
				}{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					User:     "testuser",
					Password: "short", // Too short
					DBName:   "testdb",
				},
				POC: struct {
					Enabled  bool   `default:"true" db:"true"`
					Database string `validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"`
					APIKey   string `validate:"required,min=10" db:"true"`
				}{
					Enabled:  true,
					Database: "postgres://user:pass@localhost:5432/test",
					APIKey:   "test-api-key-12345",
				},
			},
			expectError: true,
		},
		{
			name: "invalid database URL",
			config: Config{
				App: struct {
					Name    string `validate:"required,min=1" default:"apprun" db:"false"`
					Version string `validate:"required" default:"1.0.0" db:"false"`
				}{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Database: struct {
					Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"`
					Host     string `validate:"required" default:"localhost" db:"false"`
					Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`
					User     string `validate:"required" default:"postgres" db:"false"`
					Password string `validate:"required,min=8" db:"false"`
					DBName   string `validate:"required" default:"apprun" db:"false"`
				}{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					User:     "testuser",
					Password: "testpassword",
					DBName:   "testdb",
				},
				POC: struct {
					Enabled  bool   `default:"true" db:"true"`
					Database string `validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"`
					APIKey   string `validate:"required,min=10" db:"true"`
				}{
					Enabled:  true,
					Database: "invalid-url", // Invalid URL
					APIKey:   "test-api-key-12345",
				},
			},
			expectError: true,
		},
		{
			name: "short API key",
			config: Config{
				App: struct {
					Name    string `validate:"required,min=1" default:"apprun" db:"false"`
					Version string `validate:"required" default:"1.0.0" db:"false"`
				}{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Database: struct {
					Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"`
					Host     string `validate:"required" default:"localhost" db:"false"`
					Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`
					User     string `validate:"required" default:"postgres" db:"false"`
					Password string `validate:"required,min=8" db:"false"`
					DBName   string `validate:"required" default:"apprun" db:"false"`
				}{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					User:     "testuser",
					Password: "testpassword",
					DBName:   "testdb",
				},
				POC: struct {
					Enabled  bool   `default:"true" db:"true"`
					Database string `validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"`
					APIKey   string `validate:"required,min=10" db:"true"`
				}{
					Enabled:  true,
					Database: "postgres://user:pass@localhost:5432/test",
					APIKey:   "short", // Too short
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.config)
			if tt.expectError {
				assert.Error(t, err, "Expected validation to fail for test case: %s", tt.name)
			} else {
				assert.NoError(t, err, "Expected validation to pass for test case: %s", tt.name)
			}
		})
	}
}

func TestConfigStructTags(t *testing.T) {
	// Test that struct tags are properly defined
	cfg := Config{}

	// This is more of a compile-time check, but we can verify the struct exists
	assert.NotNil(t, cfg)

	// Test that we can create a valid config
	validConfig := Config{
		App: struct {
			Name    string `validate:"required,min=1" default:"apprun" db:"false"`
			Version string `validate:"required" default:"1.0.0" db:"false"`
		}{
			Name:    "test-app",
			Version: "1.0.0",
		},
		Database: struct {
			Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"`
			Host     string `validate:"required" default:"localhost" db:"false"`
			Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`
			User     string `validate:"required" default:"postgres" db:"false"`
			Password string `validate:"required,min=8" db:"false"`
			DBName   string `validate:"required" default:"apprun" db:"false"`
		}{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "securepassword",
			DBName:   "apprun",
		},
		POC: struct {
			Enabled  bool   `default:"true" db:"true"`
			Database string `validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"`
			APIKey   string `validate:"required,min=10" db:"true"`
		}{
			Enabled:  true,
			Database: "postgres://user:pass@localhost:5432/apprun_poc",
			APIKey:   "secure-api-key-12345",
		},
	}

	assert.NotNil(t, validConfig)
	assert.Equal(t, "test-app", validConfig.App.Name)
	assert.Equal(t, "postgres", validConfig.Database.Driver)
	assert.True(t, validConfig.POC.Enabled)
}
