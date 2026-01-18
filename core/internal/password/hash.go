// Package password provides password hashing and validation utilities using bcrypt.
package password

import (
	"errors"
	"sync/atomic"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost factor (10 = ~560ms per hash on typical CPU)
	// Balances security and performance for production use.
	// OWASP recommends cost 10-12 for web applications.
	DefaultCost = 10

	// MinCost is the minimum allowed bcrypt cost (security floor)
	MinCost = 4

	// MaxCost is the maximum allowed bcrypt cost (performance ceiling)
	MaxCost = 31
)

var (
	// ErrEmptyPassword is returned when an empty password is provided
	ErrEmptyPassword = errors.New("password cannot be empty")
	// ErrHashGeneration is returned when bcrypt fails to generate a hash
	ErrHashGeneration = errors.New("failed to generate password hash")
	// ErrInvalidHash is returned when password verification fails
	ErrInvalidHash = errors.New("invalid password or hash")
	// ErrInvalidCost is returned when bcrypt cost is out of valid range
	ErrInvalidCost = errors.New("bcrypt cost must be between 4 and 31")
)

// currentCost stores the active bcrypt cost factor.
// Can be updated at runtime via SetCost() for performance tuning.
// Uses atomic operations for thread-safe updates.
var currentCost int32 = DefaultCost

// GetCost returns the current bcrypt cost factor.
func GetCost() int {
	return int(atomic.LoadInt32(&currentCost))
}

// SetCost updates the bcrypt cost factor for all subsequent Hash() calls.
// Valid range: 4-31 (lower = faster but less secure).
//
// Cost recommendations:
//   - 10: Default (560ms, balanced)
//   - 8:  High-traffic scenarios (140ms, still secure)
//   - 12: Maximum security (2.2s, suitable for admin accounts)
//
// Returns error if cost is out of valid range.
func SetCost(cost int) error {
	if cost < MinCost || cost > MaxCost {
		return ErrInvalidCost
	}
	atomic.StoreInt32(&currentCost, int32(cost))
	return nil
}

// Hash generates a bcrypt hash from a plaintext password.
// Uses the current cost factor (default 10, configurable via SetCost).
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
//	fmt.Println(hash) // $2a$10$...
func Hash(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	cost := GetCost()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
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
