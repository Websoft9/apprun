package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if password matches hash
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidatePasswordStrength checks password meets security requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	if len(password) > 128 {
		return ErrPasswordTooLong
	}

	// Check for at least one number, one letter, one special char
	hasNumber := false
	hasLetter := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= '0' && char <= '9':
			hasNumber = true
		case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'):
			hasLetter = true
		default:
			hasSpecial = true
		}
	}

	if !hasNumber {
		return ErrPasswordNeedsNumber
	}
	if !hasLetter {
		return ErrPasswordNeedsLetter
	}
	if !hasSpecial {
		return ErrPasswordNeedsSpecial
	}

	return nil
}

// Test-specific errors
var (
	ErrPasswordTooShort     = assert.AnError
	ErrPasswordTooLong      = assert.AnError
	ErrPasswordNeedsNumber  = assert.AnError
	ErrPasswordNeedsLetter  = assert.AnError
	ErrPasswordNeedsSpecial = assert.AnError
)

func TestHashPassword(t *testing.T) {
	t.Run("valid password hashing", func(t *testing.T) {
		password := "SecurePass123!"

		hash, err := HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		// Verify hash format (bcrypt starts with $2a$, $2b$, or $2y$)
		assert.True(t, hash[:4] == "$2a$" || hash[:4] == "$2b$" || hash[:4] == "$2y$",
			"hash should start with $2a$, $2b$, or $2y$")

		// Verify hash length (bcrypt is 60 characters)
		assert.Equal(t, 60, len(hash), "bcrypt hash should be 60 characters")

		// Verify password not in hash
		assert.NotContains(t, hash, password, "hash should not contain plaintext password")
	})

	t.Run("different hashes for same password", func(t *testing.T) {
		password := "SecurePass123!"

		hash1, err := HashPassword(password)
		require.NoError(t, err)

		hash2, err := HashPassword(password)
		require.NoError(t, err)

		// Same password should produce different hashes (salt)
		assert.NotEqual(t, hash1, hash2, "same password should produce different hashes")
	})

	t.Run("verify bcrypt cost factor", func(t *testing.T) {
		password := "SecurePass123!"

		hash, err := HashPassword(password)
		require.NoError(t, err)

		// Extract cost factor from hash
		cost, err := bcrypt.Cost([]byte(hash))
		require.NoError(t, err)

		// Verify cost is at least 12 (security requirement)
		assert.GreaterOrEqual(t, cost, 12, "bcrypt cost should be at least 12")
	})
}

func TestVerifyPassword(t *testing.T) {
	t.Run("correct password verification", func(t *testing.T) {
		password := "SecurePass123!"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		err = VerifyPassword(password, hash)
		assert.NoError(t, err, "correct password should verify successfully")
	})

	t.Run("incorrect password rejected", func(t *testing.T) {
		password := "SecurePass123!"
		wrongPassword := "WrongPass456!"

		hash, err := HashPassword(password)
		require.NoError(t, err)

		err = VerifyPassword(wrongPassword, hash)
		assert.Error(t, err, "incorrect password should fail verification")
		assert.Equal(t, bcrypt.ErrMismatchedHashAndPassword, err)
	})

	t.Run("empty password rejected", func(t *testing.T) {
		hash, err := HashPassword("ValidPass123!")
		require.NoError(t, err)

		err = VerifyPassword("", hash)
		assert.Error(t, err, "empty password should fail verification")
	})
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid strong password",
			password:    "SecurePass123!",
			expectError: false,
		},
		{
			name:        "too short",
			password:    "Pass1!",
			expectError: true,
			errorType:   ErrPasswordTooShort,
		},
		{
			name:        "no special character",
			password:    "Password123",
			expectError: true,
			errorType:   ErrPasswordNeedsSpecial,
		},
		{
			name:        "no number",
			password:    "Password!@#",
			expectError: true,
			errorType:   ErrPasswordNeedsNumber,
		},
		{
			name:        "no letter",
			password:    "12345678!@#",
			expectError: true,
			errorType:   ErrPasswordNeedsLetter,
		},
		{
			name:        "exactly 8 characters valid",
			password:    "Pass123!",
			expectError: false,
		},
		{
			name:        "very long password valid",
			password:    "VeryLongSecurePassword123!WithManyCharacters",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark password hashing performance
func BenchmarkHashPassword(b *testing.B) {
	password := "SecurePass123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "SecurePass123!"
	hash, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyPassword(password, hash)
	}
}
