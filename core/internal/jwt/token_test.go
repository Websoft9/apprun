package jwt

import (
"context"
"testing"
"time"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
cfg := &RuntimeConfig{
Secret: "test-secret-key-with-32-chars!",
Expiry: 24 * time.Hour,
Issuer: "apprun-test",
}

token, err := GenerateToken(cfg, 123, "testuser", "test@example.com")
require.NoError(t, err)
assert.NotEmpty(t, token)
}

func TestGenerateTokenWithoutSecret(t *testing.T) {
cfg := &RuntimeConfig{
Secret: "",
Expiry: 24 * time.Hour,
Issuer: "apprun-test",
}

_, err := GenerateToken(cfg, 123, "testuser", "test@example.com")
assert.ErrorIs(t, err, ErrMissingSecret)
}

func TestValidateToken(t *testing.T) {
cfg := &RuntimeConfig{
Secret: "test-secret-key-with-32-chars!",
Expiry: 24 * time.Hour,
Issuer: "apprun-test",
}

token, err := GenerateToken(cfg, 123, "testuser", "test@example.com")
require.NoError(t, err)

claims, err := ValidateToken(cfg, token)
require.NoError(t, err)
assert.Equal(t, int64(123), claims.UserID)
assert.Equal(t, "testuser", claims.Username)
assert.Equal(t, "test@example.com", claims.Email)
assert.Equal(t, "apprun-test", claims.Issuer)
}

func TestValidateTokenExpired(t *testing.T) {
cfg := &RuntimeConfig{
Secret: "test-secret-key-with-32-chars!",
Expiry: -1 * time.Hour,
Issuer: "apprun-test",
}

token, err := GenerateToken(cfg, 123, "testuser", "test@example.com")
require.NoError(t, err)

_, err = ValidateToken(cfg, token)
assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestValidateTokenInvalid(t *testing.T) {
cfg := &RuntimeConfig{
Secret: "test-secret-key-with-32-chars!",
Expiry: 24 * time.Hour,
Issuer: "apprun-test",
}

_, err := ValidateToken(cfg, "invalid.token.here")
assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateTokenWrongSecret(t *testing.T) {
cfg1 := &RuntimeConfig{
Secret: "secret-key-1-with-32-characters!",
Expiry: 24 * time.Hour,
Issuer: "apprun-test",
}

cfg2 := &RuntimeConfig{
Secret: "secret-key-2-with-32-characters!",
Expiry: 24 * time.Hour,
Issuer: "apprun-test",
}

token, err := GenerateToken(cfg1, 123, "testuser", "test@example.com")
require.NoError(t, err)

_, err = ValidateToken(cfg2, token)
assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestContextGetters(t *testing.T) {
ctx := context.Background()
ctx = context.WithValue(ctx, UserIDKey, int64(123))
ctx = context.WithValue(ctx, UsernameKey, "testuser")
ctx = context.WithValue(ctx, EmailKey, "test@example.com")

assert.Equal(t, int64(123), GetUserID(ctx))
assert.Equal(t, "testuser", GetUsername(ctx))
assert.Equal(t, "test@example.com", GetEmail(ctx))
}

func TestContextGettersEmpty(t *testing.T) {
ctx := context.Background()

assert.Equal(t, int64(0), GetUserID(ctx))
assert.Equal(t, "", GetUsername(ctx))
assert.Equal(t, "", GetEmail(ctx))
}

func TestConfigToRuntimeConfig(t *testing.T) {
cfg := &Config{
Secret: "test-secret-32-characters-long!",
Expiry: "24h",
Issuer: "apprun-test",
}

runtimeCfg, err := cfg.ToRuntimeConfig()
require.NoError(t, err)
assert.Equal(t, "test-secret-32-characters-long!", runtimeCfg.Secret)
assert.Equal(t, 24*time.Hour, runtimeCfg.Expiry)
assert.Equal(t, "apprun-test", runtimeCfg.Issuer)
}

func TestConfigToRuntimeConfigInvalidDuration(t *testing.T) {
cfg := &Config{
Secret: "test-secret-32-characters-long!",
Expiry: "invalid",
Issuer: "apprun-test",
}

_, err := cfg.ToRuntimeConfig()
assert.Error(t, err)
}
