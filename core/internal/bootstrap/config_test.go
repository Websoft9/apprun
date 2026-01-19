package bootstrap

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	// Test default values (no environment variables set)
	cfg := DefaultConfig()

	if cfg.AutoInit != false {
		t.Errorf("Expected AutoInit=false by default, got %v", cfg.AutoInit)
	}
}

func TestDefaultConfig_WithEnvVars(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		expectedAuto bool
	}{
		{
			name:         "AUTO_INIT=true",
			envValue:     "true",
			expectedAuto: true,
		},
		{
			name:         "AUTO_INIT=false",
			envValue:     "false",
			expectedAuto: false,
		},
		{
			name:         "AUTO_INIT=1",
			envValue:     "1",
			expectedAuto: true,
		},
		{
			name:         "AUTO_INIT=0",
			envValue:     "0",
			expectedAuto: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("AUTO_INIT", tt.envValue)
			defer os.Unsetenv("AUTO_INIT")

			cfg := DefaultConfig()

			if cfg.AutoInit != tt.expectedAuto {
				t.Errorf("Expected AutoInit=%v, got %v", tt.expectedAuto, cfg.AutoInit)
			}
		})
	}
}

func TestConfig_SafeDefaults(t *testing.T) {
	// Ensure production-safe defaults
	cfg := DefaultConfig()

	// Critical: AUTO_INIT must default to false for production safety
	if cfg.AutoInit {
		t.Error("SECURITY: AutoInit must default to false to prevent accidental auto-initialization in production")
	}
}
