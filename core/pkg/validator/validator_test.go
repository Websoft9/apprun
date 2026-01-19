package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateName_Success(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		minLen int
		maxLen int
	}{
		{"valid english name", "John Doe", 2, 50},
		{"valid chinese name", "张三", 2, 50},
		{"min length", "AB", 2, 50},
		{"max length", strings.Repeat("A", 50), 2, 50},
		{"with spaces", "John Q Public", 2, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.value, tt.minLen, tt.maxLen)
			assert.NoError(t, err)
		})
	}
}

func TestValidateName_TooShort(t *testing.T) {
	err := ValidateName("A", 2, 50)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestValidateName_TooLong(t *testing.T) {
	err := ValidateName(strings.Repeat("A", 51), 2, 50)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}

func TestValidateName_Unicode(t *testing.T) {
	// Test that Unicode characters are counted correctly
	// "你好世界" is 4 characters, not 12 bytes
	err := ValidateName("你好世界", 2, 10)
	assert.NoError(t, err)

	// 51 Chinese characters should fail for maxLen=50
	err = ValidateName(strings.Repeat("中", 51), 2, 50)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}

func TestValidateHTTPSURL_Success(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"simple https", "https://example.com"},
		{"with path", "https://example.com/avatar/user.jpg"},
		{"with query", "https://cdn.example.com/image.jpg?size=100"},
		{"with port", "https://example.com:8080/file.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHTTPSURL(tt.value, 255, "avatar")
			assert.NoError(t, err)
		})
	}
}

func TestValidateHTTPSURL_NotHTTPS(t *testing.T) {
	err := ValidateHTTPSURL("http://example.com/avatar.jpg", 255, "avatar")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTPS")
}

func TestValidateHTTPSURL_TooLong(t *testing.T) {
	longURL := "https://example.com/" + strings.Repeat("a", 250)
	err := ValidateHTTPSURL(longURL, 255, "avatar")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}

func TestValidateHTTPSURL_InvalidFormat(t *testing.T) {
	err := ValidateHTTPSURL("not a url", 255, "avatar")
	require.Error(t, err)
	// Invalid URLs may pass URL parsing, so we check for either error
}

func TestValidateText_Success(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		maxLen int
	}{
		{"empty text", "", 500},
		{"short text", "Hello", 500},
		{"max length", strings.Repeat("A", 500), 500},
		{"unicode text", "你好世界", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateText(tt.value, tt.maxLen, "bio")
			assert.NoError(t, err)
		})
	}
}

func TestValidateText_TooLong(t *testing.T) {
	err := ValidateText(strings.Repeat("A", 501), 500, "bio")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}

func TestValidateText_Unicode(t *testing.T) {
	// 500 Chinese characters should pass for maxLen=500
	err := ValidateText(strings.Repeat("中", 500), 500, "bio")
	assert.NoError(t, err)

	// 501 Chinese characters should fail
	err = ValidateText(strings.Repeat("中", 501), 500, "bio")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}
