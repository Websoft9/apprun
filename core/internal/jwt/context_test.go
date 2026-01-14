package jwt

import (
	"context"
	"testing"
)

func TestContextSettersAndGetters(t *testing.T) {
	ctx := context.Background()

	// Test SetUserID and GetUserID
	userID := int64(12345)
	ctx = SetUserID(ctx, userID)
	if got := GetUserID(ctx); got != userID {
		t.Errorf("GetUserID() = %d, want %d", got, userID)
	}

	// Test SetUsername and GetUsername
	username := "testuser"
	ctx = SetUsername(ctx, username)
	if got := GetUsername(ctx); got != username {
		t.Errorf("GetUsername() = %q, want %q", got, username)
	}

	// Test SetEmail and GetEmail
	email := "test@example.com"
	ctx = SetEmail(ctx, email)
	if got := GetEmail(ctx); got != email {
		t.Errorf("GetEmail() = %q, want %q", got, email)
	}
}

func TestContextGetters_EmptyContext(t *testing.T) {
	ctx := context.Background()

	// Test GetUserID on empty context
	if got := GetUserID(ctx); got != 0 {
		t.Errorf("GetUserID(empty) = %d, want 0", got)
	}

	// Test GetUsername on empty context
	if got := GetUsername(ctx); got != "" {
		t.Errorf("GetUsername(empty) = %q, want empty string", got)
	}

	// Test GetEmail on empty context
	if got := GetEmail(ctx); got != "" {
		t.Errorf("GetEmail(empty) = %q, want empty string", got)
	}
}

func TestContextChaining(t *testing.T) {
	ctx := context.Background()

	// Test chaining all setters
	ctx = SetUserID(ctx, 999)
	ctx = SetUsername(ctx, "chainuser")
	ctx = SetEmail(ctx, "chain@example.com")

	// Verify all values are preserved
	if got := GetUserID(ctx); got != 999 {
		t.Errorf("GetUserID() after chaining = %d, want 999", got)
	}
	if got := GetUsername(ctx); got != "chainuser" {
		t.Errorf("GetUsername() after chaining = %q, want 'chainuser'", got)
	}
	if got := GetEmail(ctx); got != "chain@example.com" {
		t.Errorf("GetEmail() after chaining = %q, want 'chain@example.com'", got)
	}
}
