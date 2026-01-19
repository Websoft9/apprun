package auth

import (
	"testing"
)

func TestIsReservedUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		// Reserved usernames (case-insensitive)
		{
			name:     "admin lowercase",
			username: "admin",
			expected: true,
		},
		{
			name:     "admin uppercase",
			username: "ADMIN",
			expected: true,
		},
		{
			name:     "admin mixed case",
			username: "AdMiN",
			expected: true,
		},
		{
			name:     "administrator",
			username: "administrator",
			expected: true,
		},
		{
			name:     "root",
			username: "root",
			expected: true,
		},
		{
			name:     "ROOT uppercase",
			username: "ROOT",
			expected: true,
		},
		{
			name:     "system",
			username: "system",
			expected: true,
		},
		{
			name:     "superuser",
			username: "superuser",
			expected: true,
		},
		{
			name:     "sysadmin",
			username: "sysadmin",
			expected: true,
		},
		// Non-reserved usernames
		{
			name:     "regular username",
			username: "john_doe",
			expected: false,
		},
		{
			name:     "username with admin suffix",
			username: "john_admin",
			expected: false,
		},
		{
			name:     "username with admin prefix",
			username: "adminuser",
			expected: false,
		},
		{
			name:     "empty username",
			username: "",
			expected: false,
		},
		{
			name:     "numeric username",
			username: "user123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReservedUsername(tt.username)
			if result != tt.expected {
				t.Errorf("IsReservedUsername(%q) = %v, want %v", tt.username, result, tt.expected)
			}
		})
	}
}

func TestReservedUsernamesList(t *testing.T) {
	// Verify that ReservedUsernames list contains expected values
	expectedUsernames := map[string]bool{
		"admin":         true,
		"administrator": true,
		"root":          true,
		"system":        true,
		"superuser":     true,
		"sysadmin":      true,
	}

	if len(ReservedUsernames) < len(expectedUsernames) {
		t.Errorf("ReservedUsernames has %d entries, expected at least %d",
			len(ReservedUsernames), len(expectedUsernames))
	}

	for _, reserved := range ReservedUsernames {
		if !expectedUsernames[reserved] {
			t.Logf("Info: ReservedUsernames contains additional entry: %q", reserved)
		}
	}

	for expected := range expectedUsernames {
		found := false
		for _, reserved := range ReservedUsernames {
			if reserved == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected reserved username %q not found in ReservedUsernames", expected)
		}
	}
}
