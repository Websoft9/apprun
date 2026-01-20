package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"apprun/ent"
	"apprun/ent/enttest"
	"apprun/modules/auth/repository"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *ent.Client {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	return client
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOfSubstring(s, substr) >= 0
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// TestEnsureSystemUser_FirstTime tests creating system user for the first time
func TestEnsureSystemUser_FirstTime(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	// Call EnsureSystemUser
	user, err := service.EnsureSystemUser(ctx)
	if err != nil {
		t.Fatalf("EnsureSystemUser() error = %v, want nil", err)
	}

	// Verify user properties
	if user == nil {
		t.Fatal("EnsureSystemUser() returned nil user")
	}

	if user.UUID != SystemUserUUID {
		t.Errorf("UUID = %v, want %v", user.UUID, SystemUserUUID)
	}

	if user.Username != "system" {
		t.Errorf("Username = %v, want 'system'", user.Username)
	}

	if user.Email != "system@internal" {
		t.Errorf("Email = %v, want 'system@internal'", user.Email)
	}

	if user.PasswordHash != "" {
		t.Errorf("PasswordHash should be empty, got %v", user.PasswordHash)
	}

	if !user.IsSystem {
		t.Error("IsSystem should be true")
	}

	if user.Role != "platform_user" {
		t.Errorf("Role = %v, want 'platform_user'", user.Role)
	}
}

// TestEnsureSystemUser_Idempotent tests calling EnsureSystemUser multiple times
func TestEnsureSystemUser_Idempotent(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	// First call - creates user
	user1, err := service.EnsureSystemUser(ctx)
	if err != nil {
		t.Fatalf("First EnsureSystemUser() error = %v", err)
	}

	// Second call - should return same user
	user2, err := service.EnsureSystemUser(ctx)
	if err != nil {
		t.Fatalf("Second EnsureSystemUser() error = %v", err)
	}

	// Verify both calls returned same user
	if user1.ID != user2.ID {
		t.Errorf("User IDs don't match: %d vs %d", user1.ID, user2.ID)
	}

	if user1.UUID != user2.UUID {
		t.Errorf("User UUIDs don't match: %v vs %v", user1.UUID, user2.UUID)
	}

	// Verify only one user exists in database
	count, err := client.User.Query().Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 user in database, got %d", count)
	}
}

// TestEnsureSystemUser_ConsistencyCheck tests that inconsistent data is detected
func TestEnsureSystemUser_ConsistencyCheck(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	// Create a user named "system" but with wrong properties (simulating data corruption)
	username := "system"
	_, err := userRepo.CreateUser(ctx, &repository.CreateUserParams{
		Email:        "system@internal",
		PasswordHash: "",
		Username:     &username,
	})
	if err != nil {
		t.Fatalf("Failed to create corrupt system user: %v", err)
	}

	// Call EnsureSystemUser - should detect inconsistency
	_, err = service.EnsureSystemUser(ctx)
	if err == nil {
		t.Error("EnsureSystemUser() should return error for inconsistent data")
	}

	// Verify it's the corruption error
	if !errors.Is(err, ErrSystemUserCorruption) {
		t.Errorf("Expected ErrSystemUserCorruption, got %v", err)
	}
}

// TestEnsureAdminUser_FirstTime tests creating admin user for the first time
func TestEnsureAdminUser_FirstTime(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	email := "admin@example.com"
	password := "AdminPass123"

	// Call EnsureAdminUser
	user, err := service.EnsureAdminUser(ctx, email, password)
	if err != nil {
		t.Fatalf("EnsureAdminUser() error = %v, want nil", err)
	}

	// Verify user properties
	if user == nil {
		t.Fatal("EnsureAdminUser() returned nil user")
	}

	if user.Email != email {
		t.Errorf("Email = %v, want %v", user.Email, email)
	}

	if user.PasswordHash == "" {
		t.Error("PasswordHash should not be empty")
	}

	if user.IsSystem {
		t.Error("IsSystem should be false for admin user")
	}

	if user.Role != "platform_admin" {
		t.Errorf("Role = %v, want 'platform_admin'", user.Role)
	}

	// Username should be extracted from email
	expectedUsername := "admin"
	if user.Username != expectedUsername {
		t.Errorf("Username = %v, want %v", user.Username, expectedUsername)
	}
}

// TestEnsureAdminUser_Idempotent tests calling EnsureAdminUser multiple times doesn't change password
func TestEnsureAdminUser_Idempotent(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	email := "admin@example.com"
	password1 := "FirstPass123"
	password2 := "SecondPass456"

	// First call - creates user with password1
	user1, err := service.EnsureAdminUser(ctx, email, password1)
	if err != nil {
		t.Fatalf("First EnsureAdminUser() error = %v", err)
	}

	originalPasswordHash := user1.PasswordHash

	// Second call with different password - should NOT update password
	user2, err := service.EnsureAdminUser(ctx, email, password2)
	if err != nil {
		t.Fatalf("Second EnsureAdminUser() error = %v", err)
	}

	// Verify password hash wasn't changed
	if user2.PasswordHash != originalPasswordHash {
		t.Error("Password hash should not change on second call")
	}

	// Verify same user was returned
	if user1.ID != user2.ID {
		t.Errorf("User IDs don't match: %d vs %d", user1.ID, user2.ID)
	}

	// Verify only one user exists
	count, err := client.User.Query().Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 user in database, got %d", count)
	}
}

// TestEnsureAdminUser_InvalidEmail tests email validation
func TestEnsureAdminUser_InvalidEmail(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	tests := []struct {
		name  string
		email string
	}{
		{"empty email", ""},
		{"invalid format", "not-an-email"},
		{"no domain", "test@"},
		{"no @", "testexample.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.EnsureAdminUser(ctx, tt.email, "ValidPass123")
			if err == nil {
				t.Error("EnsureAdminUser() should return error for invalid email")
			}

			if !errors.Is(err, ErrInvalidEmail) {
				t.Errorf("Expected ErrInvalidEmail, got %v", err)
			}
		})
	}
}

// TestEnsureAdminUser_WeakPassword tests password strength validation
func TestEnsureAdminUser_WeakPassword(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	tests := []struct {
		name     string
		password string
	}{
		{"empty password", ""},
		{"too short", "Short1"},
		{"no uppercase", "lowercase123"},
		{"no lowercase", "UPPERCASE123"},
		{"no digit", "NoDigitPass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.EnsureAdminUser(ctx, "admin@example.com", tt.password)
			if err == nil {
				t.Error("EnsureAdminUser() should return error for weak password")
			}

			// For non-empty passwords, check it's a password-related error
			if tt.password != "" && err != nil {
				// Just verify error is not nil and contains password-related message
				errStr := err.Error()
				if !contains(errStr, "password") {
					t.Errorf("Expected password-related error, got %v", err)
				}
			}
		})
	}
}

// TestSystemUserUUID_FixedValue tests that system user UUID is the expected constant
func TestSystemUserUUID_FixedValue(t *testing.T) {
	expected := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	if SystemUserUUID != expected {
		t.Errorf("SystemUserUUID = %v, want %v", SystemUserUUID, expected)
	}
}

// TestEnsureSystemUser_ConcurrentCalls tests concurrent calls to EnsureSystemUser
func TestEnsureSystemUser_ConcurrentCalls(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	// Run 5 concurrent calls (reduced to minimize SQLite lock contention)
	const numGoroutines = 5
	var wg sync.WaitGroup
	results := make([]*ent.User, numGoroutines)
	errs := make([]error, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			user, err := service.EnsureSystemUser(ctx)
			results[idx] = user
			errs[idx] = err
		}(i)
	}

	wg.Wait()

	// Count successes (at least one should succeed)
	successCount := 0
	var successUser *ent.User
	for i, err := range errs {
		if err == nil {
			successCount++
			if successUser == nil {
				successUser = results[i]
			}
		}
	}

	if successCount == 0 {
		t.Fatal("All goroutines failed, expected at least one success")
	}

	// Verify all successful calls returned the same user
	for i, user := range results {
		if errs[i] == nil && user != nil {
			if user.ID != successUser.ID {
				t.Errorf("Goroutine %d returned different user ID: %d vs %d", i, user.ID, successUser.ID)
			}
			if user.UUID != successUser.UUID {
				t.Errorf("Goroutine %d returned different UUID: %v vs %v", i, user.UUID, successUser.UUID)
			}
		}
	}

	// Verify only one user exists in database
	count, err := client.User.Query().Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 user in database, got %d", count)
	}
}

// TestEnsureAdminUser_ConcurrentCalls tests concurrent calls to EnsureAdminUser
func TestEnsureAdminUser_ConcurrentCalls(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	email := "admin@example.com"
	password := "AdminPass123"

	// Run 5 concurrent calls (reduced to minimize SQLite lock contention)
	const numGoroutines = 5
	var wg sync.WaitGroup
	results := make([]*ent.User, numGoroutines)
	errs := make([]error, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			user, err := service.EnsureAdminUser(ctx, email, password)
			results[idx] = user
			errs[idx] = err
		}(i)
	}

	wg.Wait()

	// Count successes (at least one should succeed)
	successCount := 0
	var successUser *ent.User
	for i, err := range errs {
		if err == nil {
			successCount++
			if successUser == nil {
				successUser = results[i]
			}
		}
	}

	if successCount == 0 {
		t.Fatal("All goroutines failed, expected at least one success")
	}

	// Verify all successful calls returned the same user
	for i, user := range results {
		if errs[i] == nil && user != nil {
			if user.ID != successUser.ID {
				t.Errorf("Goroutine %d returned different user ID: %d vs %d", i, user.ID, successUser.ID)
			}
			if user.UUID != successUser.UUID {
				t.Errorf("Goroutine %d returned different UUID: %v vs %v", i, user.UUID, successUser.UUID)
			}
		}
	}

	// Verify only one user exists in database
	count, err := client.User.Query().Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 user in database, got %d", count)
	}
}
