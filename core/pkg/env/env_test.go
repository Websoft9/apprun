package env

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	if err := os.Setenv("TEST_VAR", "custom"); err != nil {
		t.Fatalf("Failed to set env: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_VAR"); err != nil {
			t.Logf("Failed to unset env: %v", err)
		}
	}()

	assert.Equal(t, "custom", Get("TEST_VAR", "default"))
	assert.Equal(t, "default", Get("NONEXISTENT", "default"))
}

func TestMustGet(t *testing.T) {
	if err := os.Setenv("MUST_HAVE", "value"); err != nil {
		t.Fatalf("Failed to set env: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("MUST_HAVE"); err != nil {
			t.Logf("Failed to unset env: %v", err)
		}
	}()

	assert.NotPanics(t, func() {
		result := MustGet("MUST_HAVE")
		assert.Equal(t, "value", result)
	})

	assert.Panics(t, func() {
		MustGet("NONEXISTENT_REQUIRED")
	})
}

func TestGetInt(t *testing.T) {
	if err := os.Setenv("INT_VAR", "42"); err != nil {
		t.Fatalf("Failed to set env: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("INT_VAR"); err != nil {
			t.Logf("Failed to unset env: %v", err)
		}
	}()

	assert.Equal(t, 42, GetInt("INT_VAR", 10))
	assert.Equal(t, 100, GetInt("MISSING_INT", 100))

	if err := os.Setenv("INVALID_INT", "notanumber"); err != nil {
		t.Fatalf("Failed to set env: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("INVALID_INT"); err != nil {
			t.Logf("Failed to unset env: %v", err)
		}
	}()
	assert.Equal(t, 50, GetInt("INVALID_INT", 50))
}

func TestGetBool(t *testing.T) {
	os.Setenv("BOOL_TRUE", "true")
	os.Setenv("BOOL_FALSE", "false")
	defer os.Unsetenv("BOOL_TRUE")
	defer os.Unsetenv("BOOL_FALSE")

	assert.True(t, GetBool("BOOL_TRUE", false))
	assert.False(t, GetBool("BOOL_FALSE", true))
	assert.True(t, GetBool("MISSING", true))
}

func TestGetDuration(t *testing.T) {
	os.Setenv("DURATION_VAR", "5s")
	defer os.Unsetenv("DURATION_VAR")

	assert.Equal(t, 5*time.Second, GetDuration("DURATION_VAR", 10*time.Second))
	assert.Equal(t, 15*time.Second, GetDuration("MISSING", 15*time.Second))
}
