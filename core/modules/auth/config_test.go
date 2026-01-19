package auth

import (
	"testing"
)

func TestPlatformConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant interface{}
		expected interface{}
	}{
		{
			name:     "PlatformProjectUUID is all zeros",
			constant: PlatformProjectUUID,
			expected: "00000000-0000-0000-0000-000000000000",
		},
		{
			name:     "PlatformProjectName is Platform",
			constant: PlatformProjectName,
			expected: "Platform",
		},
		{
			name:     "PlatformProjectDescription is set",
			constant: PlatformProjectDescription,
			expected: "Global platform-level resources and configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("got %v, want %v", tt.constant, tt.expected)
			}
		})
	}
}

func TestInitConfig(t *testing.T) {
	config := DefaultInitConfig()

	if config.Username != "admin" {
		t.Errorf("InitConfig.Username = %q, want %q", config.Username, "admin")
	}

	if config.Email != "admin@example.com" {
		t.Errorf("InitConfig.Email = %q, want %q", config.Email, "admin@example.com")
	}

	if config.Password != "" {
		t.Errorf("InitConfig.Password should be empty by default, got %q", config.Password)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Init.Username != "admin" {
		t.Errorf("Config.Init.Username = %q, want %q", config.Init.Username, "admin")
	}

	if config.Init.Email != "admin@example.com" {
		t.Errorf("Config.Init.Email = %q, want %q", config.Init.Email, "admin@example.com")
	}

	if config.Security.BcryptCost != DefaultBcryptCost {
		t.Errorf("Config.Security.BcryptCost = %d, want %d", config.Security.BcryptCost, DefaultBcryptCost)
	}
}
