package jwt

import (
	"sync"
	"testing"
	"time"

	pkgconfig "apprun/pkg/config"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestConfig configures viper with test JWT settings
func setupTestConfig() {
	viper.Set("jwt.secret", "test-secret-key-minimum-32-chars-required!")
	viper.Set("jwt.access_token_expiration", 24*time.Hour)
	viper.Set("jwt.issuer", "apprun-platform-test")
	viper.Set("jwt.audience", "apprun-api-test")

	// Reset the global token service to pick up new config
	resetGlobalService()
}

// resetGlobalService resets the global token service for testing
func resetGlobalService() {
	once = sync.Once{}
	globalTokenService = NewTokenService(pkgconfig.NewViperProvider(nil))
}

func TestGenerateToken(t *testing.T) {
	setupTestConfig()

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	token, expiresAt, err := GenerateToken(123, userClaims)

	require.NoError(t, err, "GenerateToken should not return error")
	assert.NotEmpty(t, token, "Generated token should not be empty")
	assert.True(t, expiresAt.After(time.Now()), "Token expiration should be in the future")
	assert.True(t, expiresAt.Before(time.Now().Add(25*time.Hour)), "Token should expire within 25 hours")
}

func TestValidateToken_Success(t *testing.T) {
	setupTestConfig()

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	token, _, err := GenerateToken(123, userClaims)
	require.NoError(t, err)

	claims, err := ValidateToken(token)

	require.NoError(t, err, "ValidateToken should not return error for valid token")
	assert.Equal(t, int64(123), claims.UserID, "UserID should match")
	assert.Equal(t, "testuser", claims.Username, "Username should match")
	assert.Equal(t, "test@example.com", claims.Email, "Email should match")
	assert.Equal(t, "apprun-platform-test", claims.Issuer, "Issuer should match")
	assert.Contains(t, claims.Audience, "apprun-api-test", "Audience should match")
}

func TestValidateToken_Expired(t *testing.T) {
	viper.Set("jwt.secret", "test-secret-key-minimum-32-chars-required!")
	viper.Set("jwt.access_token_expiration", -1*time.Hour) // Expired token
	viper.Set("jwt.issuer", "apprun-platform-test")
	viper.Set("jwt.audience", "apprun-api-test")

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	token, _, err := GenerateToken(123, userClaims)
	require.NoError(t, err)

	// Reset to normal expiration for validation
	viper.Set("jwt.access_token_expiration", 24*time.Hour)

	_, err = ValidateToken(token)

	assert.ErrorIs(t, err, ErrTokenExpired, "ValidateToken should return ErrTokenExpired for expired token")
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	setupTestConfig()

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	token, _, err := GenerateToken(123, userClaims)
	require.NoError(t, err)

	// Change the secret key
	viper.Set("jwt.secret", "different-secret-key-32-chars-min!!")

	_, err = ValidateToken(token)

	assert.ErrorIs(t, err, ErrInvalidToken, "ValidateToken should return ErrInvalidToken for wrong signature")
}

func TestGenerateToken_MissingSecret(t *testing.T) {
	viper.Set("jwt.secret", "")
	viper.Set("jwt.access_token_expiration", 24*time.Hour)
	viper.Set("jwt.issuer", "apprun-platform-test")
	viper.Set("jwt.audience", "apprun-api-test")

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	_, _, err := GenerateToken(123, userClaims)

	assert.ErrorIs(t, err, ErrMissingSecret, "GenerateToken should return ErrMissingSecret when secret is empty")
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	setupTestConfig()

	_, err := ValidateToken("not.a.valid.jwt.token")

	assert.ErrorIs(t, err, ErrInvalidToken, "ValidateToken should return ErrInvalidToken for malformed token")
}

func TestValidateToken_EmptyToken(t *testing.T) {
	setupTestConfig()

	_, err := ValidateToken("")

	assert.ErrorIs(t, err, ErrInvalidToken, "ValidateToken should return ErrInvalidToken for empty token")
}
