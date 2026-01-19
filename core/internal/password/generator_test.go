package password

import (
	"testing"
)

func TestGenerateRandomPassword(t *testing.T) {
	tests := []struct {
		name      string
		length    int
		wantError bool
	}{
		{
			name:      "valid 16 char password",
			length:    16,
			wantError: false,
		},
		{
			name:      "minimum 8 char password",
			length:    8,
			wantError: false,
		},
		{
			name:      "long 32 char password",
			length:    32,
			wantError: false,
		},
		{
			name:      "too short - 7 chars",
			length:    7,
			wantError: true,
		},
		{
			name:      "too short - 0 chars",
			length:    0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GenerateRandomPassword(tt.length)

			if tt.wantError {
				if err == nil {
					t.Errorf("GenerateRandomPassword() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateRandomPassword() unexpected error: %v", err)
				return
			}

			// Verify length
			if len(password) != tt.length {
				t.Errorf("GenerateRandomPassword() length = %d, want %d", len(password), tt.length)
			}

			// Verify password meets validation rules (should pass Validate)
			if err := Validate(password); err != nil {
				t.Errorf("Generated password failed validation: %v, password: %s", err, password)
			}

			t.Logf("Generated password (length %d): %s", tt.length, password)
		})
	}
}

// TestGenerateRandomPasswordUniqueness verifies that multiple calls generate different passwords
func TestGenerateRandomPasswordUniqueness(t *testing.T) {
	passwords := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		pwd, err := GenerateRandomPassword(16)
		if err != nil {
			t.Fatalf("Failed to generate password: %v", err)
		}

		if passwords[pwd] {
			t.Errorf("Duplicate password generated: %s", pwd)
		}
		passwords[pwd] = true
	}

	if len(passwords) != iterations {
		t.Errorf("Expected %d unique passwords, got %d", iterations, len(passwords))
	}
}
