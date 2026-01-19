package jwt

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRefreshTestConfig configures viper with test JWT settings including refresh tokens
func setupRefreshTestConfig() {
	viper.Set("jwt.secret", "test-secret-key-minimum-32-chars-required-for-jwt")
	viper.Set("jwt.access_token_expiration", 24*time.Hour)
	viper.Set("jwt.refresh_token_expiration", 168*time.Hour)
	viper.Set("jwt.issuer", "test-issuer")
	viper.Set("jwt.audience", "test-audience")
}

func TestGenerateTokenPair(t *testing.T) {
	setupRefreshTestConfig()

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	// Test: Generate token pair
	accessToken, refreshToken, expiresAt, err := GenerateTokenPair(123, userClaims)

	// Assertions
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.True(t, expiresAt.After(time.Now()))
	assert.NotEqual(t, accessToken, refreshToken, "Access and refresh tokens should be different")
}

func TestGenerateTokenPair_TokenTypes(t *testing.T) {
	setupRefreshTestConfig()

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	// Generate tokens
	accessToken, refreshToken, _, err := GenerateTokenPair(123, userClaims)
	require.NoError(t, err)

	// Validate access token
	accessClaims, err := ValidateToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, "access", accessClaims.TokenType)
	assert.Equal(t, int64(123), accessClaims.UserID)
	assert.Equal(t, "testuser", accessClaims.Username)

	// Validate refresh token
	refreshClaims, err := ValidateToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, "refresh", refreshClaims.TokenType)
	assert.Equal(t, int64(123), refreshClaims.UserID)
}

func TestGenerateTokenPair_ExpirationDifferences(t *testing.T) {
	viper.Set("jwt.secret", "test-secret-key-minimum-32-chars-required-for-jwt")
	viper.Set("jwt.access_token_expiration", 1*time.Hour)
	viper.Set("jwt.refresh_token_expiration", 7*24*time.Hour)
	viper.Set("jwt.issuer", "test-issuer")
	viper.Set("jwt.audience", "test-audience")
	resetGlobalService()

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	// Generate tokens
	accessToken, refreshToken, _, err := GenerateTokenPair(123, userClaims)
	require.NoError(t, err)

	// Parse tokens to check expiration
	accessClaims, _ := ValidateToken(accessToken)
	refreshClaims, _ := ValidateToken(refreshToken)

	// Assert refresh token expires much later than access token
	accessExp := accessClaims.ExpiresAt.Time
	refreshExp := refreshClaims.ExpiresAt.Time
	assert.True(t, refreshExp.After(accessExp))

	// Check approximate durations (with 1 minute tolerance)
	accessDuration := time.Until(accessExp)
	refreshDuration := time.Until(refreshExp)
	assert.InDelta(t, 1*time.Hour, accessDuration, float64(1*time.Minute))
	assert.InDelta(t, 7*24*time.Hour, refreshDuration, float64(1*time.Minute))
}

func TestValidateRefreshToken_Success(t *testing.T) {
	setupRefreshTestConfig()

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	// Generate tokens
	_, refreshToken, _, err := GenerateTokenPair(123, userClaims)
	require.NoError(t, err)

	// Test: Validate refresh token
	claims, err := ValidateRefreshToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, "refresh", claims.TokenType)
	assert.Equal(t, int64(123), claims.UserID)
}

func TestValidateRefreshToken_WrongType(t *testing.T) {
	setupRefreshTestConfig()

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	// Generate tokens
	accessToken, _, _, err := GenerateTokenPair(123, userClaims)
	require.NoError(t, err)

	// Test: Try to validate access token as refresh token
	claims, err := ValidateRefreshToken(accessToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTokenType)
	assert.Nil(t, claims)
}

func TestValidateRefreshToken_Expired(t *testing.T) {
	viper.Set("jwt.secret", "test-secret-key-minimum-32-chars-required-for-jwt")
	viper.Set("jwt.access_token_expiration", 1*time.Second)
	viper.Set("jwt.refresh_token_expiration", 1*time.Nanosecond) // Expires immediately
	viper.Set("jwt.issuer", "test-issuer")
	viper.Set("jwt.audience", "test-audience")
	resetGlobalService()

	userClaims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	// Generate tokens
	_, refreshToken, _, err := GenerateTokenPair(123, userClaims)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Test: Validate expired token
	claims, err := ValidateRefreshToken(refreshToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenExpired)
	assert.Nil(t, claims)
}
