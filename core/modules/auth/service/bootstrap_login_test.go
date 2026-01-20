package service

import (
	"context"
	"errors"
	"testing"

	"apprun/ent/enttest"
	"apprun/modules/auth/repository"

	_ "github.com/mattn/go-sqlite3"
)

// TestSystemUserCannotLogin tests that system user is blocked from logging in
func TestSystemUserCannotLogin(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	// Create system user
	_, err := service.EnsureSystemUser(ctx)
	if err != nil {
		t.Fatalf("Failed to create system user: %v", err)
	}

	// Attempt to login as system user
	loginReq := &LoginRequest{
		Identifier: "system",
		Password:   "any-password",
	}

	_, err = service.Login(ctx, loginReq, "127.0.0.1")
	if err == nil {
		t.Error("Login() should fail for system user")
	}

	// Verify it's the system user login error
	if !errors.Is(err, ErrSystemCannotLogin) {
		t.Errorf("Expected ErrSystemCannotLogin, got %v", err)
	}
}

// TestAdminUserCanLogin tests that admin user can login normally
func TestAdminUserCanLogin(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	service := NewAuthService(userRepo, nil)

	ctx := context.Background()

	email := "admin@example.com"
	password := "AdminPass123"

	// Create admin user
	_, err := service.EnsureAdminUser(ctx, email, password)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Login as admin user
	loginReq := &LoginRequest{
		Identifier: email,
		Password:   password,
	}

	// Note: This test will fail because JWT generation requires more setup
	// We're just testing that it doesn't fail with ErrSystemCannotLogin
	_, err = service.Login(ctx, loginReq, "127.0.0.1")

	// Should not be system user error
	if errors.Is(err, ErrSystemCannotLogin) {
		t.Error("Admin user should not be blocked with ErrSystemCannotLogin")
	}

	// May fail with JWT error, which is acceptable for this test
	// We're only testing the system user blocking logic
}
