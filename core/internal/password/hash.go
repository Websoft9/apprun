// Package password provides password hashing and validation utilities using bcrypt.
package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost factor (12 = ~250ms per hash)
	// Balances security and performance for production use
	DefaultCost = 12
)

var (
	// ErrEmptyPassword is returned when an empty password is provided
	ErrEmptyPassword = errors.New("password cannot be empty")
	// ErrHashGeneration is returned when bcrypt fails to generate a hash
	ErrHashGeneration = errors.New("failed to generate password hash")
	// ErrInvalidHash is returned when password verification fails
	ErrInvalidHash = errors.New("invalid password or hash")
)

// Hash generates a bcrypt hash from a plaintext password.
// Uses DefaultCost (12) for consistent security/performance balance.
//
// Returns:
//   - string: bcrypt hash (60 bytes, base64-encoded)
//   - error: ErrEmptyPassword if password is empty, ErrHashGeneration on bcrypt failure
//
// Example:
//
//	hash, err := password.Hash("SecurePass123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(hash) // $2a$12$...
func Hash(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", ErrHashGeneration
	}

	return string(hash), nil
}

// Verify checks if a plaintext password matches a bcrypt hash.
//
// Parameters:
//   - password: plaintext password to verify
//   - hash: bcrypt hash to compare against
//
// Returns:
//   - error: nil if password matches, ErrInvalidHash if mismatch or invalid
//
// Example:
//
//	err := password.Verify("SecurePass123", storedHash)
//	if err != nil {
//	    // Authentication failed
//	    return errors.New("invalid credentials")
//	}
func Verify(password, hash string) error {
	if password == "" || hash == "" {
		return ErrInvalidHash
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return ErrInvalidHash
	}

	return nil
}
