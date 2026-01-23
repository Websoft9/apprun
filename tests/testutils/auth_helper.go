package testutils

import (
"testing"
"time"

"github.com/golang-jwt/jwt/v5"
)

// Auth Test Helpers for ATDD Tests
// Simplified version without ent dependency

// CreateAdminToken creates a JWT token for platform_admin role
// Used in metrics tests to verify admin-only access
func CreateAdminToken(t *testing.T, userID int, email string) string {
t.Helper()

// Generate JWT token
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"user_id": userID,
"email":   email,
"role":    "platform_admin",
"exp":     time.Now().Add(1 * time.Hour).Unix(),
})

// Sign token (use test secret)
tokenString, err := token.SignedString([]byte("test_jwt_secret"))
if err != nil {
t.Fatalf("failed to create admin token: %v", err)
}

return tokenString
}

// CreateUserToken creates a JWT token for regular user role
// Used to test permission denial scenarios
func CreateUserToken(t *testing.T, userID int, email string, role string) string {
t.Helper()

if role == "" {
role = "user"
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"user_id": userID,
"email":   email,
"role":    role,
"exp":     time.Now().Add(1 * time.Hour).Unix(),
})

tokenString, err := token.SignedString([]byte("test_jwt_secret"))
if err != nil {
t.Fatalf("failed to create user token: %v", err)
}

return tokenString
}

// CreateCustomRoleToken creates a token with custom role
func CreateCustomRoleToken(t *testing.T, userID int, email string, role string) string {
t.Helper()
return CreateUserToken(t, userID, email, role)
}
