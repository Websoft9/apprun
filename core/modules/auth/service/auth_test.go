package service

import (
	"context"
	"errors"
	"testing"
)

// TestRegister_InvalidEmail tests email format validation
func TestRegister_InvalidEmail(t *testing.T) {
	// This test doesn't need real database
	// It will fail at email validation before database access

	tests := []struct {
		name  string
		email string
	}{
		{"missing @", "invalid-email"},
		{"no domain", "test@"},
		{"no local part", "@example.com"},
		{"no dot in domain", "test@example"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service with nil repo since we won't reach database
			service := &AuthService{userRepo: nil}
			ctx := context.Background()

			req := &RegisterRequest{
				Email:    tt.email,
				Password: "SecurePass123",
			}

			_, err := service.Register(ctx, req)
			if !errors.Is(err, ErrInvalidEmail) {
				t.Errorf("Register() error = %v, want ErrInvalidEmail", err)
			}
		})
	}
}

// TestRegister_WeakPassword tests password strength validation
func TestRegister_WeakPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"too short", "Short1"},
		{"no uppercase", "lowercase123"},
		{"no lowercase", "UPPERCASE123"},
		{"no digit", "NoDigitPass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &AuthService{userRepo: nil}
			ctx := context.Background()

			req := &RegisterRequest{
				Email:    "test@example.com",
				Password: tt.password,
			}

			_, err := service.Register(ctx, req)
			if err == nil {
				t.Error("Register() expected password validation error, got nil")
			}
		})
	}
}

// TestRegister_InvalidUsername tests username validation
func TestRegister_InvalidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
	}{
		{"too short", "ab"},
		{"special chars", "user@name"},
		{"spaces", "user name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &AuthService{userRepo: nil}
			ctx := context.Background()

			req := &RegisterRequest{
				Email:    "test@example.com",
				Password: "SecurePass123",
				Username: &tt.username,
			}

			_, err := service.Register(ctx, req)
			if err == nil {
				t.Error("Register() expected username validation error, got nil")
			}
		})
	}
}

// TestRegister_ReservedUsername tests that reserved usernames are rejected
func TestRegister_ReservedUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
	}{
		{"admin lowercase", "admin"},
		{"admin uppercase", "ADMIN"},
		{"admin mixed case", "AdMiN"},
		{"administrator", "administrator"},
		{"root", "root"},
		{"ROOT uppercase", "ROOT"},
		{"system", "system"},
		{"superuser", "superuser"},
		{"sysadmin", "sysadmin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &AuthService{userRepo: nil}
			ctx := context.Background()

			req := &RegisterRequest{
				Email:    "test@example.com",
				Password: "SecurePass123",
				Username: &tt.username,
			}

			_, err := service.Register(ctx, req)
			if !errors.Is(err, ErrReservedUsername) {
				t.Errorf("Register() error = %v, want ErrReservedUsername", err)
			}
		})
	}
}

// Note: Full integration tests with database mocks are in integration_test_story18.sh
// These unit tests focus on validation logic that doesn't require database

// Note: Login tests require database/repository layer (FindByIdentifier)
// See tests/integration/ for full login flow testing with real database
