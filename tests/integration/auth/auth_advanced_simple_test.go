package auth
// Package auth_test provides black-box integration tests for advanced auth scenarios
package auth_test

import (
	"testing"
)

// TestAuthINT007_PasswordResetRequest tests password reset request flow (P1)
func TestAuthINT007_PasswordResetRequest(t *testing.T) {
	t.Skip("Placeholder - server needs password reset endpoint")
}

// TestAuthINT010_ConcurrentLoginLimit tests concurrent login session limits (P2)
func TestAuthINT010_ConcurrentLoginLimit(t *testing.T) {
	t.Skip("Placeholder - rate limiting not yet implemented")
}

// TestAuthINT011_JWTExpirationHandling tests JWT expiration handling (P0)
func TestAuthINT011_JWTExpirationHandling(t *testing.T) {
	t.Skip("Placeholder - requires running server")
}
