package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertTimeEqual checks if two times are equal within a tolerance
func AssertTimeEqual(t *testing.T, expected, actual time.Time, tolerance time.Duration) {
	t.Helper()

	diff := actual.Sub(expected)
	if diff < 0 {
		diff = -diff
	}

	assert.True(t, diff <= tolerance,
		"times not equal within tolerance. Expected: %v, Actual: %v, Diff: %v, Tolerance: %v",
		expected, actual, diff, tolerance)
}

// AssertTimeRecent checks if time is within last N seconds
func AssertTimeRecent(t *testing.T, timestamp time.Time, seconds int) {
	t.Helper()

	now := time.Now()
	diff := now.Sub(timestamp)

	assert.True(t, diff >= 0 && diff <= time.Duration(seconds)*time.Second,
		"time not recent. Timestamp: %v, Now: %v, Diff: %v",
		timestamp, now, diff)
}

// AssertContainsSubstring checks if string contains substring
func AssertContainsSubstring(t *testing.T, haystack, needle string) {
	t.Helper()
	assert.Contains(t, haystack, needle, "string does not contain expected substring")
}

// AssertNotContainsSubstring checks if string does not contain substring
func AssertNotContainsSubstring(t *testing.T, haystack, needle string) {
	t.Helper()
	assert.NotContains(t, haystack, needle, "string contains unexpected substring")
}

// AssertEmailFormat checks if string is valid email format
func AssertEmailFormat(t *testing.T, email string) {
	t.Helper()
	require.NotEmpty(t, email, "email is empty")
	require.Contains(t, email, "@", "email missing @ symbol")
	require.Contains(t, email, ".", "email missing domain")
}

// AssertUUIDFormat checks if string is valid UUID format
func AssertUUIDFormat(t *testing.T, uuid string) {
	t.Helper()
	require.Len(t, uuid, 36, "UUID should be 36 characters")
	require.Contains(t, uuid, "-", "UUID should contain dashes")
}

// AssertJWTFormat checks if string is valid JWT format (3 parts separated by dots)
func AssertJWTFormat(t *testing.T, token string) {
	t.Helper()
	require.NotEmpty(t, token, "JWT token is empty")

	// JWT has 3 parts: header.payload.signature
	parts := 0
	for _, c := range token {
		if c == '.' {
			parts++
		}
	}
	require.Equal(t, 2, parts, "JWT should have 3 parts (2 dots)")
}

// AssertPasswordHashed checks if password is hashed (bcrypt format)
func AssertPasswordHashed(t *testing.T, hash string) {
	t.Helper()
	require.NotEmpty(t, hash, "password hash is empty")
	require.True(t, len(hash) >= 60, "bcrypt hash should be at least 60 characters")
	require.True(t, hash[:4] == "$2a$" || hash[:4] == "$2b$" || hash[:4] == "$2y$",
		"bcrypt hash should start with $2a$, $2b$, or $2y$")
}

// AssertMapHasKeys checks if map contains all expected keys
func AssertMapHasKeys(t *testing.T, m map[string]interface{}, keys ...string) {
	t.Helper()
	for _, key := range keys {
		require.Contains(t, m, key, "map missing expected key: %s", key)
	}
}

// AssertSliceNotEmpty checks if slice is not empty
func AssertSliceNotEmpty(t *testing.T, slice interface{}) {
	t.Helper()
	require.NotNil(t, slice, "slice is nil")
	require.NotEmpty(t, slice, "slice is empty")
}
