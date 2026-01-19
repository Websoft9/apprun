// Package validator provides common validation functions.
package validator

import (
	"net/url"
	"strings"

	"apprun/pkg/errors"
)

var (
	// ErrFieldTooShort is returned when field is below min length
	ErrFieldTooShort = errors.New(errors.ErrCodeInvalidParam, "Field value is too short")
	// ErrFieldTooLong is returned when field exceeds max length
	ErrFieldTooLong = errors.New(errors.ErrCodeInvalidParam, "Field value exceeds maximum length")
	// ErrInvalidURL is returned when URL format is invalid
	ErrInvalidURL = errors.New(errors.ErrCodeInvalidParam, "Invalid URL format")
	// ErrNotHTTPS is returned when URL doesn't use HTTPS
	ErrNotHTTPS = errors.New(errors.ErrCodeInvalidParam, "URL must use HTTPS protocol")
)

// ValidateName validates a display name field.
// Rules: minLen-maxLen characters, allows letters, numbers, spaces, Unicode characters.
func ValidateName(name string, minLen, maxLen int) error {
	trimmed := strings.TrimSpace(name)
	length := len([]rune(trimmed)) // Count Unicode characters, not bytes

	if length < minLen {
		return ErrFieldTooShort.WithContext("field", "name").WithContext("min_length", minLen)
	}
	if length > maxLen {
		return ErrFieldTooLong.WithContext("field", "name").WithContext("max_length", maxLen)
	}

	return nil
}

// ValidateHTTPSURL validates a URL that must use HTTPS protocol.
// Rules: HTTPS protocol required, max maxLen characters.
func ValidateHTTPSURL(urlStr string, maxLen int, fieldName string) error {
	trimmed := strings.TrimSpace(urlStr)

	// Check length
	if len(trimmed) > maxLen {
		return ErrFieldTooLong.WithContext("field", fieldName).WithContext("max_length", maxLen)
	}

	// Parse URL
	u, err := url.Parse(trimmed)
	if err != nil {
		return ErrInvalidURL.WithContext("field", fieldName)
	}

	// Check HTTPS protocol
	if u.Scheme != "https" {
		return ErrNotHTTPS.WithContext("field", fieldName)
	}

	return nil
}

// ValidateText validates a text field with maximum length.
// Rules: max maxLen characters (counts Unicode characters properly).
func ValidateText(text string, maxLen int, fieldName string) error {
	length := len([]rune(text)) // Count Unicode characters

	if length > maxLen {
		return ErrFieldTooLong.WithContext("field", fieldName).WithContext("max_length", maxLen)
	}

	return nil
}
