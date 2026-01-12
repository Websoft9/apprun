package password

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestHash tests password hashing functionality
func TestHash(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     error
		description string
	}{
		{
			name:        "valid_password",
			password:    "SecurePass123",
			wantErr:     nil,
			description: "should successfully hash a valid password",
		},
		{
			name:        "empty_password",
			password:    "",
			wantErr:     ErrEmptyPassword,
			description: "should reject empty password",
		},
		{
			name:        "long_password",
			password:    strings.Repeat("a", 72),
			wantErr:     nil,
			description: "should handle max length password (72 chars)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := Hash(tt.password)

			// Check error expectation
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Hash() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			// Success case validations
			if err != nil {
				t.Errorf("Hash() unexpected error = %v", err)
				return
			}

			// Verify hash format (bcrypt prefix with current cost)
			cost := GetCost()
			expectedPrefix := fmt.Sprintf("$2a$%02d$", cost)
			if !strings.HasPrefix(hash, expectedPrefix) && !strings.HasPrefix(hash, fmt.Sprintf("$2b$%02d$", cost)) {
				t.Errorf("Hash() invalid bcrypt format: got %s, want prefix $2a$%02d$ or $2b$%02d$", hash, cost, cost)
			}

			// Verify hash length (60 bytes)
			if len(hash) != 60 {
				t.Errorf("Hash() invalid length = %d, want 60", len(hash))
			}

			// Verify password can be verified against hash
			if err := Verify(tt.password, hash); err != nil {
				t.Errorf("Hash() generated hash cannot be verified: %v", err)
			}
		})
	}
}

// TestVerify tests password verification functionality
func TestVerify(t *testing.T) {
	// Generate a valid hash for testing
	validPassword := "SecurePass123"
	validHash, _ := Hash(validPassword)

	tests := []struct {
		name        string
		password    string
		hash        string
		wantErr     error
		description string
	}{
		{
			name:        "valid_credentials",
			password:    validPassword,
			hash:        validHash,
			wantErr:     nil,
			description: "should verify correct password",
		},
		{
			name:        "wrong_password",
			password:    "WrongPassword",
			hash:        validHash,
			wantErr:     ErrInvalidHash,
			description: "should reject incorrect password",
		},
		{
			name:        "empty_password",
			password:    "",
			hash:        validHash,
			wantErr:     ErrInvalidHash,
			description: "should reject empty password",
		},
		{
			name:        "empty_hash",
			password:    validPassword,
			hash:        "",
			wantErr:     ErrInvalidHash,
			description: "should reject empty hash",
		},
		{
			name:        "invalid_hash_format",
			password:    validPassword,
			hash:        "not-a-bcrypt-hash",
			wantErr:     ErrInvalidHash,
			description: "should reject malformed hash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Verify(tt.password, tt.hash)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Verify() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidate tests password strength validation
func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid_strong_password",
			password: "SecurePass123",
			wantErr:  nil,
		},
		{
			name:     "too_short",
			password: "Short1A",
			wantErr:  ErrTooShort,
		},
		{
			name:     "too_long",
			password: strings.Repeat("a", 73) + "A1",
			wantErr:  ErrTooLong,
		},
		{
			name:     "no_uppercase",
			password: "alllowercase123",
			wantErr:  ErrNoUppercase,
		},
		{
			name:     "no_lowercase",
			password: "ALLUPPERCASE123",
			wantErr:  ErrNoLowercase,
		},
		{
			name:     "no_digit",
			password: "NoDigitsHere",
			wantErr:  ErrNoDigit,
		},
		{
			name:     "with_special_chars",
			password: "Secure@Pass123!",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.password)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateUsername tests username validation
func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid_username",
			username: "john_doe123",
			wantErr:  false,
		},
		{
			name:     "too_short",
			username: "ab",
			wantErr:  true,
			errMsg:   "at least 3 characters",
		},
		{
			name:     "too_long",
			username: strings.Repeat("a", 65),
			wantErr:  true,
			errMsg:   "not exceed 64 characters",
		},
		{
			name:     "invalid_special_chars",
			username: "john@doe",
			wantErr:  true,
			errMsg:   "letters, numbers and underscores",
		},
		{
			name:     "valid_with_numbers",
			username: "user123",
			wantErr:  false,
		},
		{
			name:     "valid_with_underscore",
			username: "user_name",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateUsername() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateUsername() error = %v, should contain %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateUsername() unexpected error = %v", err)
				}
			}
		})
	}
}

// BenchmarkPasswordHash benchmarks bcrypt hashing performance
func BenchmarkPasswordHash(b *testing.B) {
	password := "SecurePass123"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Hash(password)
	}
}

// BenchmarkPasswordVerify benchmarks bcrypt verification performance
func BenchmarkPasswordVerify(b *testing.B) {
	password := "SecurePass123"
	hash, _ := Hash(password)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Verify(password, hash)
	}
}

// TestHashConsistency ensures different hashes are generated for the same password (salt)
func TestHashConsistency(t *testing.T) {
	password := "SecurePass123"

	hash1, _ := Hash(password)
	hash2, _ := Hash(password)

	if hash1 == hash2 {
		t.Error("Hash() should generate different hashes for the same password (due to salt)")
	}

	// Both hashes should verify the original password
	if err := Verify(password, hash1); err != nil {
		t.Errorf("First hash verification failed: %v", err)
	}
	if err := Verify(password, hash2); err != nil {
		t.Errorf("Second hash verification failed: %v", err)
	}
}

// TestBcryptCostFactor verifies the configured cost factor
func TestBcryptCostFactor(t *testing.T) {
	password := "SecurePass123"
	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	// Extract cost from hash
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("Failed to extract cost: %v", err)
	}

	expectedCost := GetCost()
	if cost != expectedCost {
		t.Errorf("Hash cost = %d, want %d", cost, expectedCost)
	}
}

// TestSetCost verifies dynamic cost configuration
func TestSetCost(t *testing.T) {
	// Save original cost
	originalCost := GetCost()
	defer func() {
		// Restore original cost after test
		_ = SetCost(originalCost)
	}()

	tests := []struct {
		name     string
		cost     int
		wantErr  bool
		wantCost int
		testHash bool // Whether to test actual hashing (slow for high costs)
	}{
		{
			name:     "valid cost 8",
			cost:     8,
			wantErr:  false,
			wantCost: 8,
			testHash: true,
		},
		{
			name:     "valid cost 12",
			cost:     12,
			wantErr:  false,
			wantCost: 12,
			testHash: true,
		},
		{
			name:     "minimum cost 4",
			cost:     4,
			wantErr:  false,
			wantCost: 4,
			testHash: true,
		},
		{
			name:     "maximum cost 31",
			cost:     31,
			wantErr:  false,
			wantCost: 31,
			testHash: false, // Skip hashing test (too slow: ~8-10 minutes)
		},
		{
			name:    "invalid cost too low",
			cost:    3,
			wantErr: true,
		},
		{
			name:    "invalid cost too high",
			cost:    32,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetCost(tt.cost)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SetCost(%d) expected error, got nil", tt.cost)
				}
				if !errors.Is(err, ErrInvalidCost) {
					t.Errorf("SetCost(%d) error = %v, want ErrInvalidCost", tt.cost, err)
				}
				return
			}

			if err != nil {
				t.Errorf("SetCost(%d) unexpected error: %v", tt.cost, err)
				return
			}

			actualCost := GetCost()
			if actualCost != tt.wantCost {
				t.Errorf("GetCost() = %d, want %d", actualCost, tt.wantCost)
			}

			// Verify Hash() uses new cost (only for low costs to keep tests fast)
			if tt.testHash {
				hash, err := Hash("test123")
				if err != nil {
					t.Errorf("Hash() failed with cost %d: %v", tt.cost, err)
					return
				}

				hashCost, err := bcrypt.Cost([]byte(hash))
				if err != nil {
					t.Errorf("Failed to extract cost from hash: %v", err)
					return
				}

				if hashCost != tt.wantCost {
					t.Errorf("Hash uses cost %d, want %d", hashCost, tt.wantCost)
				}
			}
		})
	}
}
