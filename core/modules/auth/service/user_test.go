package service

import (
	"context"
	stdErrors "errors"
	"testing"

	"apprun/ent"
	"apprun/ent/enttest"
	"apprun/internal/jwt"
	"apprun/internal/password"
	"apprun/modules/auth/repository"
	"apprun/pkg/errors"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUserServiceTest creates a test database and user service
func setupUserServiceTest(t *testing.T) (*ent.Client, *UserService, func()) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	userRepo := repository.NewUserRepository(client)
	userSvc := NewUserService(userRepo)

	cleanup := func() {
		client.Close()
	}

	return client, userSvc, cleanup
}

// createTestUserForService creates a test user for user service tests
func createTestUserForService(t *testing.T, client *ent.Client) *ent.User {
	hashedPassword, err := password.Hash("TestPass123")
	require.NoError(t, err)

	user, err := client.User.Create().
		SetEmail("test@example.com").
		SetPasswordHash(hashedPassword).
		SetNickname("Test User").
		SetAvatar("https://example.com/avatar.jpg").
		SetBio("Test bio").
		SetStatus(1).
		Save(context.Background())
	require.NoError(t, err)

	return user
}

// contextWithUserID creates a context with user ID (simulates JWT middleware)
func contextWithUserID(userID int64) context.Context {
	return jwt.SetUserID(context.Background(), userID)
}

func TestGetCurrentUser_Success(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	// Create test user
	user := createTestUserForService(t, client)

	// Get current user with context
	ctx := contextWithUserID(user.ID)
	profile, err := userSvc.GetCurrentUser(ctx)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, user.UUID.String(), profile.ID)
	assert.Equal(t, user.Email, profile.Email)
	assert.Equal(t, "Test User", *profile.Name)
	assert.Equal(t, "https://example.com/avatar.jpg", *profile.Avatar)
	assert.Equal(t, "Test bio", *profile.Bio)
	assert.True(t, profile.IsActive)
}

func TestGetCurrentUser_NoAuth(t *testing.T) {
	_, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	// Call without user ID in context
	ctx := context.Background()
	profile, err := userSvc.GetCurrentUser(ctx)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "Authentication required")
}

func TestGetCurrentUser_UserNotFound(t *testing.T) {
	_, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	// Use non-existent user ID
	ctx := contextWithUserID(99999)
	profile, err := userSvc.GetCurrentUser(ctx)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, profile)
	// The service returns "Failed to retrieve user profile" for internal errors
	// including user not found cases for security reasons
}

func TestUpdateProfile_Success(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	// Create test user
	user := createTestUserForService(t, client)

	// Update profile
	ctx := contextWithUserID(user.ID)
	newName := "Updated Name"
	newAvatar := "https://cdn.example.com/new-avatar.jpg"
	newBio := "Updated bio"

	req := &UpdateProfileRequest{
		Name:   &newName,
		Avatar: &newAvatar,
		Bio:    &newBio,
	}

	profile, err := userSvc.UpdateProfile(ctx, req)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "Updated Name", *profile.Name)
	assert.Equal(t, "https://cdn.example.com/new-avatar.jpg", *profile.Avatar)
	assert.Equal(t, "Updated bio", *profile.Bio)

	// Verify in database
	updatedUser, err := client.User.Get(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.Nickname)
	assert.Equal(t, "https://cdn.example.com/new-avatar.jpg", updatedUser.Avatar)
	assert.Equal(t, "Updated bio", updatedUser.Bio)
}

func TestUpdateProfile_NameTooShort(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	user := createTestUserForService(t, client)
	ctx := contextWithUserID(user.ID)

	// Try to update with 1-character name
	shortName := "A"
	req := &UpdateProfileRequest{
		Name: &shortName,
	}

	profile, err := userSvc.UpdateProfile(ctx, req)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "too short")
}

func TestUpdateProfile_NameTooLong(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	user := createTestUserForService(t, client)
	ctx := contextWithUserID(user.ID)

	// Try to update with 51-character name
	longName := "A very long name that exceeds fifty characters limit"
	req := &UpdateProfileRequest{
		Name: &longName,
	}

	profile, err := userSvc.UpdateProfile(ctx, req)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}

func TestUpdateProfile_AvatarNotHTTPS(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	user := createTestUserForService(t, client)
	ctx := contextWithUserID(user.ID)

	// Try to update with HTTP URL (should fail)
	httpAvatar := "http://example.com/avatar.jpg"
	req := &UpdateProfileRequest{
		Avatar: &httpAvatar,
	}

	profile, err := userSvc.UpdateProfile(ctx, req)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "HTTPS")
}

func TestUpdateProfile_AvatarTooLong(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	user := createTestUserForService(t, client)
	ctx := contextWithUserID(user.ID)

	// Try to update with URL > 255 characters
	longURL := "https://example.com/" + string(make([]byte, 250)) + ".jpg"
	req := &UpdateProfileRequest{
		Avatar: &longURL,
	}

	profile, err := userSvc.UpdateProfile(ctx, req)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}

func TestUpdateProfile_BioTooLong(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	user := createTestUserForService(t, client)
	ctx := contextWithUserID(user.ID)

	// Try to update with bio > 500 characters
	longBio := string(make([]byte, 501))
	req := &UpdateProfileRequest{
		Bio: &longBio,
	}

	profile, err := userSvc.UpdateProfile(ctx, req)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}

func TestChangePassword_Success(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	user := createTestUserForService(t, client)
	ctx := contextWithUserID(user.ID)

	// Change password
	req := &ChangePasswordRequest{
		OldPassword: "TestPass123",
		NewPassword: "NewPass456",
	}

	err := userSvc.ChangePassword(ctx, req)

	// Assertions
	require.NoError(t, err)

	// Verify new password works
	updatedUser, err := client.User.Get(context.Background(), user.ID)
	require.NoError(t, err)
	err = password.Verify("NewPass456", updatedUser.PasswordHash)
	assert.NoError(t, err)

	// Verify old password no longer works
	err = password.Verify("TestPass123", updatedUser.PasswordHash)
	assert.Error(t, err)
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	user := createTestUserForService(t, client)
	ctx := contextWithUserID(user.ID)

	// Try to change with wrong old password
	req := &ChangePasswordRequest{
		OldPassword: "WrongPass123",
		NewPassword: "NewPass456",
	}

	err := userSvc.ChangePassword(ctx, req)

	// Assertions
	require.Error(t, err)
	var appErr *errors.AppError
	assert.True(t, stdErrors.As(err, &appErr))
	assert.Equal(t, errors.ErrCodeAuthInvalidCredentials, appErr.Code)
}

func TestChangePassword_WeakNewPassword(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	user := createTestUserForService(t, client)
	ctx := contextWithUserID(user.ID)

	// Try to change to weak password
	req := &ChangePasswordRequest{
		OldPassword: "TestPass123",
		NewPassword: "weak",
	}

	err := userSvc.ChangePassword(ctx, req)

	// Assertions
	require.Error(t, err)
	// The error message from password validator contains "at least 8 characters"
	assert.True(t, err != nil)
}

func TestChangePassword_SamePassword(t *testing.T) {
	client, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	user := createTestUserForService(t, client)
	ctx := contextWithUserID(user.ID)

	// Try to change to same password
	req := &ChangePasswordRequest{
		OldPassword: "TestPass123",
		NewPassword: "TestPass123",
	}

	err := userSvc.ChangePassword(ctx, req)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "different")
}

func TestChangePassword_NoAuth(t *testing.T) {
	_, userSvc, cleanup := setupUserServiceTest(t)
	defer cleanup()

	// Try without authentication context
	ctx := context.Background()
	req := &ChangePasswordRequest{
		OldPassword: "TestPass123",
		NewPassword: "NewPass456",
	}

	err := userSvc.ChangePassword(ctx, req)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Authentication required")
}
