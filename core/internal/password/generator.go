// Package password provides password generation utilities.
package password

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const (
	// Character sets for password generation
	uppercaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowercaseLetters = "abcdefghijklmnopqrstuvwxyz"
	digits           = "0123456789"
	specialChars     = "!@#$%^&*"

	// DefaultPasswordLength is the default length for generated passwords
	DefaultPasswordLength = 16
)

var (
	// ErrInvalidLength is returned when password length is too short
	ErrInvalidLength = errors.New("password length must be at least 8 characters")
)

// GenerateRandomPassword generates a cryptographically secure random password.
// The password will contain at least one uppercase letter, one lowercase letter,
// one digit, and one special character.
//
// Parameters:
//   - length: desired password length (minimum 8)
//
// Returns:
//   - string: generated password
//   - error: ErrInvalidLength if length < 8, or crypto/rand error
//
// Example:
//
//	password, err := GenerateRandomPassword(16)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(password) // "aB3$xYz9PqR2sT4u"
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		return "", ErrInvalidLength
	}

	// Character pool (all possible characters)
	allChars := uppercaseLetters + lowercaseLetters + digits + specialChars

	// Ensure at least one character from each required set
	password := make([]byte, length)

	// First 4 characters: one from each required set (ensures validation passes)
	charSets := []string{uppercaseLetters, lowercaseLetters, digits, specialChars}
	for i := 0; i < 4; i++ {
		char, err := randomChar(charSets[i])
		if err != nil {
			return "", err
		}
		password[i] = char
	}

	// Fill remaining characters randomly from all character sets
	for i := 4; i < length; i++ {
		char, err := randomChar(allChars)
		if err != nil {
			return "", err
		}
		password[i] = char
	}

	// Shuffle the password to avoid predictable pattern (first 4 chars are always upper/lower/digit/special)
	if err := shuffle(password); err != nil {
		return "", err
	}

	return string(password), nil
}

// randomChar returns a random character from the given character set.
func randomChar(charset string) (byte, error) {
	maxVal := big.NewInt(int64(len(charset)))
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

// shuffle randomly shuffles a byte slice using Fisher-Yates algorithm.
func shuffle(data []byte) error {
	for i := len(data) - 1; i > 0; i-- {
		maxVal := big.NewInt(int64(i + 1))
		j, err := rand.Int(rand.Reader, maxVal)
		if err != nil {
			return err
		}
		jInt := j.Int64()
		data[i], data[jInt] = data[jInt], data[i]
	}
	return nil
}
