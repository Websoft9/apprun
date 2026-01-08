package password

import (
	"errors"
	"fmt"
	"regexp"
	"unicode"
)

const (
	// MinPasswordLength is the minimum required password length (OWASP recommendation)
	MinPasswordLength = 8
	// MaxPasswordLength prevents DoS attacks via bcrypt computational cost
	MaxPasswordLength = 72 // bcrypt internal limit
)

var (
	// ErrTooShort is returned when password length < MinPasswordLength
	ErrTooShort = errors.New("password must be at least 8 characters")
	// ErrTooLong is returned when password length > MaxPasswordLength
	ErrTooLong = errors.New("password must not exceed 72 characters")
	// ErrNoUppercase is returned when password lacks uppercase letters
	ErrNoUppercase = errors.New("password must contain at least one uppercase letter")
	// ErrNoLowercase is returned when password lacks lowercase letters
	ErrNoLowercase = errors.New("password must contain at least one lowercase letter")
	// ErrNoDigit is returned when password lacks numeric digits
	ErrNoDigit = errors.New("password must contain at least one number")
)

// Validate checks password strength against security requirements:
//   - Length: 8-72 characters
//   - At least 1 uppercase letter (A-Z)
//   - At least 1 lowercase letter (a-z)
//   - At least 1 digit (0-9)
//
// Note: Special characters are recommended but not enforced to avoid usability issues.
//
// Returns:
//   - error: specific validation error, or nil if password meets all requirements
//
// Example:
//
//	err := password.Validate("WeakPass")
//	if err != nil {
//	    return fmt.Errorf("invalid password: %w", err)
//	}
func Validate(password string) error {
	// Length check
	if len(password) < MinPasswordLength {
		return ErrTooShort
	}
	if len(password) > MaxPasswordLength {
		return ErrTooLong
	}

	// Character type requirements
	var (
		hasUpper = false
		hasLower = false
		hasDigit = false
	)

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper {
		return ErrNoUppercase
	}
	if !hasLower {
		return ErrNoLowercase
	}
	if !hasDigit {
		return ErrNoDigit
	}

	return nil
}

// ValidateUsername checks if a username meets the following requirements:
//   - Length: 3-64 characters
//   - Only alphanumeric characters and underscores (a-z, A-Z, 0-9, _)
//
// Returns:
//   - error: validation error with detailed message, or nil if valid
//
// Example:
//
//	err := password.ValidateUsername("john_doe123")
//	if err != nil {
//	    return fmt.Errorf("invalid username: %w", err)
//	}
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if len(username) > 64 {
		return errors.New("username must not exceed 64 characters")
	}

	// Only alphanumeric and underscore allowed
	matched, err := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if err != nil {
		return fmt.Errorf("username validation regex failed: %w", err)
	}
	if !matched {
		return errors.New("username can only contain letters, numbers and underscores")
	}

	return nil
}
