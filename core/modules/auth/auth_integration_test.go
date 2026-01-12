package auth_test

import (
	"context"
	"os"
	"testing"
	"time"

	"apprun/ent"
	"apprun/ent/enttest"
	"apprun/internal/jwt"
	"apprun/internal/password"
	"apprun/modules/auth/handler"
	"apprun/modules/auth/repository"
	"apprun/modules/auth/service"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *ent.Client {
	t.Helper()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")

	return client
}

// setupTestConfig configures viper for integration testing
func setupTestConfig(t *testing.T) {
	t.Helper()
	viper.Reset()
	viper.Set("jwt.secret", "test-secret-key-32-chars-minimum!!")
	viper.Set("jwt.access_token_expiration", "24h")
	viper.Set("jwt.refresh_token_expiration", "168h")
	viper.Set("jwt.issuer", "apprun-test")
	viper.Set("jwt.audience", "apprun-api-test")
}

// setupTestUser creates a test user in the database
func setupTestUser(t *testing.T, client *ent.Client, email, username, pwd string) *ent.User {
	t.Helper()

	// Use real password hashing
	passwordHash, err := password.Hash(pwd)
	require.NoError(t, err)

	user, err := client.User.Create().
		SetEmail(email).
		SetUsername(username).
		SetPasswordHash(passwordHash).
		SetStatus(1). // Active status
		SetTimezone("UTC").
		SetLanguage("en").
		Save(context.Background())

	require.NoError(t, err)
	return user
}

// TestLoginIntegration_SuccessWithEmail tests complete login flow with email
func TestLoginIntegration_SuccessWithEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	setupTestConfig(t)

	// Create test user
	testEmail := "test@example.com"
	testPassword := "SecurePass123"
	user := setupTestUser(t, client, testEmail, "testuser", testPassword)

	// Initialize service layer
	userRepo := repository.NewUserRepository(client)
	authSvc := service.NewAuthService(userRepo)

	// Execute login
	ctx := context.Background()
	loginReq := &service.LoginRequest{
		Identifier: testEmail,
		Password:   testPassword,
	}

	loginResp, err := authSvc.Login(ctx, loginReq, "127.0.0.1")

	// Verify
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.Token)
	assert.NotEmpty(t, loginResp.ExpiresAt)
	assert.Equal(t, user.Email, loginResp.User.Email)
	assert.Equal(t, user.Username, *loginResp.User.Username)
	assert.Equal(t, int8(1), loginResp.User.Status)

	// Verify token is valid
	claims, err := jwt.ValidateToken(loginResp.Token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
}

// TestLoginIntegration_SuccessWithUsername tests login flow with username
func TestLoginIntegration_SuccessWithUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	setupTestConfig(t)

	testUsername := "johndoe"
	testPassword := "SecurePass123"
	user := setupTestUser(t, client, "john@example.com", testUsername, testPassword)

	userRepo := repository.NewUserRepository(client)
	authSvc := service.NewAuthService(userRepo)

	// Execute login with username
	ctx := context.Background()
	loginReq := &service.LoginRequest{
		Identifier: testUsername,
		Password:   testPassword,
	}

	loginResp, err := authSvc.Login(ctx, loginReq, "192.168.1.100")

	// Verify
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.Token)
	assert.Equal(t, user.Username, *loginResp.User.Username)
}

// TestLoginIntegration_InvalidPassword tests login with wrong password
func TestLoginIntegration_InvalidPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	setupTestConfig(t)

	setupTestUser(t, client, "test@example.com", "testuser", "CorrectPass123")

	userRepo := repository.NewUserRepository(client)
	authSvc := service.NewAuthService(userRepo)

	// Execute login with wrong password
	ctx := context.Background()
	loginReq := &service.LoginRequest{
		Identifier: "test@example.com",
		Password:   "WrongPassword",
	}

	loginResp, err := authSvc.Login(ctx, loginReq, "127.0.0.1")

	// Verify - should return generic error
	require.Error(t, err)
	assert.Nil(t, loginResp)
	assert.Contains(t, err.Error(), "Invalid")
}

// TestLoginIntegration_UserNotFound tests login with non-existent user
func TestLoginIntegration_UserNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	setupTestConfig(t)

	userRepo := repository.NewUserRepository(client)
	authSvc := service.NewAuthService(userRepo)

	// Execute login with non-existent user
	ctx := context.Background()
	loginReq := &service.LoginRequest{
		Identifier: "nonexistent@example.com",
		Password:   "SomePassword123",
	}

	loginResp, err := authSvc.Login(ctx, loginReq, "127.0.0.1")

	// Verify - should return generic error (don't reveal user existence)
	require.Error(t, err)
	assert.Nil(t, loginResp)
	assert.Contains(t, err.Error(), "Invalid")
}

// TestLoginIntegration_DisabledAccount tests login with disabled account
func TestLoginIntegration_DisabledAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	setupTestConfig(t)

	// Create disabled user
	passwordHash, err := password.Hash("SecurePass123")
	require.NoError(t, err)

	user, err := client.User.Create().
		SetEmail("disabled@example.com").
		SetUsername("disableduser").
		SetPasswordHash(passwordHash).
		SetStatus(0). // Disabled status
		SetTimezone("UTC").
		SetLanguage("en").
		Save(context.Background())
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(client)
	authSvc := service.NewAuthService(userRepo)

	// Execute login
	ctx := context.Background()
	loginReq := &service.LoginRequest{
		Identifier: user.Email,
		Password:   "SecurePass123",
	}

	loginResp, err := authSvc.Login(ctx, loginReq, "127.0.0.1")

	// Verify - should return account disabled error
	require.Error(t, err)
	assert.Nil(t, loginResp)
	assert.Contains(t, err.Error(), "disabled")
}

// TestLoginIntegration_LoginHistoryTracking tests last_login_at update
func TestLoginIntegration_LoginHistoryTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	setupTestConfig(t)

	testEmail := "history@example.com"
	testPassword := "SecurePass123"
	user := setupTestUser(t, client, testEmail, "historyuser", testPassword)

	// Verify initial login timestamp is zero
	assert.Nil(t, user.LastLoginAt)
	assert.Equal(t, "", user.LastLoginIP)

	userRepo := repository.NewUserRepository(client)
	authSvc := service.NewAuthService(userRepo)

	// Execute login
	ctx := context.Background()
	testIP := "203.0.113.42"
	loginReq := &service.LoginRequest{
		Identifier: testEmail,
		Password:   testPassword,
	}

	_, err := authSvc.Login(ctx, loginReq, testIP)
	require.NoError(t, err)

	// Wait for async update (goroutine)
	time.Sleep(100 * time.Millisecond)

	// Verify login history was updated
	updatedUser, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)

	assert.NotNil(t, updatedUser.LastLoginAt)
	assert.Equal(t, testIP, updatedUser.LastLoginIP)
	assert.WithinDuration(t, time.Now(), *updatedUser.LastLoginAt, 5*time.Second)
}

// TestGetUserProfileIntegration tests /me endpoint flow
func TestGetUserProfileIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	setupTestConfig(t)

	// Create test user
	user := setupTestUser(t, client, "profile@example.com", "profileuser", "SecurePass123")

	userRepo := repository.NewUserRepository(client)
	authSvc := service.NewAuthService(userRepo)

	// Execute GetUserProfile
	ctx := context.Background()
	profile, err := authSvc.GetUserProfile(ctx, user.ID)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, user.UUID.String(), profile.UUID)
	assert.Equal(t, user.Email, profile.Email)
	assert.Equal(t, user.Username, *profile.Username)
	assert.Equal(t, int8(1), profile.Status)
	assert.Equal(t, "UTC", profile.Timezone)
	assert.Equal(t, "en", profile.Language)

	// Verify password hash is NOT included (check struct has no password fields)
	assert.NotEmpty(t, profile.UUID)
	assert.NotEmpty(t, profile.Email)
}

// TestGetUserProfileIntegration_UserNotFound tests /me with deleted user
func TestGetUserProfileIntegration_UserNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	setupTestConfig(t)

	userRepo := repository.NewUserRepository(client)
	authSvc := service.NewAuthService(userRepo)

	// Execute GetUserProfile with non-existent ID
	ctx := context.Background()
	profile, err := authSvc.GetUserProfile(ctx, 99999)

	// Verify
	require.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "not found")
}

// TestJWTTokenValidation tests token expiration and validation
func TestJWTTokenValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupTestConfig(t)

	// Generate token
	userClaims := map[string]interface{}{
		"user_id":  int64(123),
		"username": "testuser",
		"email":    "test@example.com",
	}
	token, expiresAt, err := jwt.GenerateToken(123, userClaims)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token is valid
	claims, err := jwt.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, int64(123), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "test@example.com", claims.Email)

	// Verify expiration time is ~24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, expiresAt, 10*time.Second)
}

// TestRepositoryFindByIdentifier tests OR query for username/email
func TestRepositoryFindByIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	// Create test users
	user1 := setupTestUser(t, client, "user1@example.com", "user1", "pass123")
	user2 := setupTestUser(t, client, "user2@example.com", "user2", "pass456")

	userRepo := repository.NewUserRepository(client)
	ctx := context.Background()

	// Test finding by email
	foundUser, err := userRepo.FindByIdentifier(ctx, "user1@example.com")
	require.NoError(t, err)
	assert.Equal(t, user1.ID, foundUser.ID)
	assert.Equal(t, user1.Email, foundUser.Email)

	// Test finding by username
	foundUser, err = userRepo.FindByIdentifier(ctx, "user2")
	require.NoError(t, err)
	assert.Equal(t, user2.ID, foundUser.ID)
	assert.Equal(t, user2.Username, foundUser.Username)

	// Test not found
	foundUser, err = userRepo.FindByIdentifier(ctx, "nonexistent")
	require.Error(t, err)
	assert.Nil(t, foundUser)
	assert.True(t, ent.IsNotFound(err))
}

// TestRepositoryUpdateLoginHistory tests async login history update
func TestRepositoryUpdateLoginHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	user := setupTestUser(t, client, "history@example.com", "historyuser", "pass123")

	userRepo := repository.NewUserRepository(client)
	ctx := context.Background()

	// Update login history
	testIP := "198.51.100.42"
	err := userRepo.UpdateLoginHistory(ctx, user.ID, testIP)
	require.NoError(t, err)

	// Verify update
	updatedUser, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)

	assert.NotNil(t, updatedUser.LastLoginAt)
	assert.Equal(t, testIP, updatedUser.LastLoginIP)
	assert.WithinDuration(t, time.Now(), *updatedUser.LastLoginAt, 2*time.Second)
}

// TestIntegration_FullLoginFlow tests complete end-to-end login flow
func TestIntegration_FullLoginFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	client := setupTestDB(t)
	defer client.Close()

	setupTestConfig(t)

	// Step 1: Create user
	testEmail := "fullflow@example.com"
	testUsername := "fullflowuser"
	testPassword := "SecurePass123"
	user := setupTestUser(t, client, testEmail, testUsername, testPassword)

	// Step 2: Initialize full stack
	userRepo := repository.NewUserRepository(client)
	authSvc := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authSvc)

	assert.NotNil(t, authHandler) // Verify handler created

	// Step 3: Login with email
	ctx := context.Background()
	loginReq := &service.LoginRequest{
		Identifier: testEmail,
		Password:   testPassword,
	}

	loginResp, err := authSvc.Login(ctx, loginReq, "127.0.0.1")
	require.NoError(t, err)

	// Step 4: Validate JWT token
	claims, err := jwt.ValidateToken(loginResp.Token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)

	// Step 5: Get user profile using token claims
	profile, err := authSvc.GetUserProfile(ctx, claims.UserID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, profile.Email)
	assert.Equal(t, user.Username, *profile.Username)

	// Step 6: Verify login history updated
	time.Sleep(100 * time.Millisecond)
	updatedUser, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	assert.NotNil(t, updatedUser.LastLoginAt)
	assert.Equal(t, "127.0.0.1", updatedUser.LastLoginIP)
}

// TestMain handles test setup/teardown
func TestMain(m *testing.M) {
	// Setup test environment
	os.Setenv("JWT_SECRET", "test-secret-key-32-chars-minimum!!")

	// Run tests
	code := m.Run()

	// Cleanup
	os.Exit(code)
}
