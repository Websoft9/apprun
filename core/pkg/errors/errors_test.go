package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("CORE_VAL_INVALID_PARAM_001", "Invalid parameter")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Code != "CORE_VAL_INVALID_PARAM_001" {
		t.Errorf("Expected code CORE_VAL_INVALID_PARAM_001, got %s", err.Code)
	}
	if err.Message != "Invalid parameter" {
		t.Errorf("Expected message 'Invalid parameter', got %s", err.Message)
	}
	if err.Err != nil {
		t.Errorf("Expected nil underlying error, got %v", err.Err)
	}
	if err.Context != nil {
		t.Errorf("Expected nil context, got %v", err.Context)
	}
}

func TestNewf(t *testing.T) {
	err := Newf("CORE_VAL_INVALID_PARAM_001", "Invalid parameter: %s", "username")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Code != "CORE_VAL_INVALID_PARAM_001" {
		t.Errorf("Expected code CORE_VAL_INVALID_PARAM_001, got %s", err.Code)
	}
	expected := "Invalid parameter: username"
	if err.Message != expected {
		t.Errorf("Expected message %q, got %q", expected, err.Message)
	}
}

func TestWrap(t *testing.T) {
	t.Run("wrap non-nil error", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		err := Wrap(originalErr, "CORE_SYS_INTERNAL_ERROR_001", "System error occurred")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Code != "CORE_SYS_INTERNAL_ERROR_001" {
			t.Errorf("Expected code CORE_SYS_INTERNAL_ERROR_001, got %s", err.Code)
		}
		if err.Message != "System error occurred" {
			t.Errorf("Expected message 'System error occurred', got %s", err.Message)
		}
		if !errors.Is(err.Err, originalErr) {
			t.Errorf("Expected underlying error to be original error")
		}
	})

	t.Run("wrap nil error", func(t *testing.T) {
		err := Wrap(nil, "CORE_SYS_INTERNAL_ERROR_001", "System error occurred")
		if err != nil {
			t.Errorf("Expected nil when wrapping nil error, got %v", err)
		}
	})
}

func TestWrapf(t *testing.T) {
	t.Run("wrapf non-nil error", func(t *testing.T) {
		originalErr := fmt.Errorf("database connection failed")
		err := Wrapf(originalErr, "CORE_SYS_INTERNAL_ERROR_001", "System error: %s", "DB unavailable")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		expected := "System error: DB unavailable"
		if err.Message != expected {
			t.Errorf("Expected message %q, got %q", expected, err.Message)
		}
		if !errors.Is(err.Err, originalErr) {
			t.Errorf("Expected underlying error to be original error")
		}
	})

	t.Run("wrapf nil error", func(t *testing.T) {
		err := Wrapf(nil, "CORE_SYS_INTERNAL_ERROR_001", "System error: %s", "test")
		if err != nil {
			t.Errorf("Expected nil when wrapping nil error, got %v", err)
		}
	})
}

func TestAppError_Error(t *testing.T) {
	t.Run("error without underlying error", func(t *testing.T) {
		err := New("CORE_VAL_INVALID_PARAM_001", "Invalid parameter")
		expected := "Invalid parameter"
		if err.Error() != expected {
			t.Errorf("Expected error string %q, got %q", expected, err.Error())
		}
	})

	t.Run("error with underlying error", func(t *testing.T) {
		originalErr := fmt.Errorf("connection timeout")
		err := Wrap(originalErr, "CORE_SYS_INTERNAL_ERROR_001", "System error")
		expected := "System error: connection timeout"
		if err.Error() != expected {
			t.Errorf("Expected error string %q, got %q", expected, err.Error())
		}
	})
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("database error")
	err := Wrap(originalErr, "CORE_SYS_INTERNAL_ERROR_001", "System error")

	unwrapped := err.Unwrap()
	if !errors.Is(unwrapped, originalErr) {
		t.Errorf("Expected unwrapped error to be original error")
	}

	// Test with errors.Is
	if !errors.Is(err, originalErr) {
		t.Error("errors.Is should find the original error in the chain")
	}
}

func TestAppError_WithContext(t *testing.T) {
	err := New("CORE_VAL_INVALID_PARAM_001", "Invalid parameter")

	// Test chaining with type-safe context keys
	_ = err.WithContext(ContextKeyUserID, "user123").
		WithContext(ContextKeyRequestID, "req456")

	if err.Context == nil {
		t.Fatal("Expected context to be initialized")
	}
	if err.Context[ContextKeyUserID.String()] != "user123" {
		t.Errorf("Expected user_id to be 'user123', got %v", err.Context[ContextKeyUserID.String()])
	}
	if err.Context[ContextKeyRequestID.String()] != "req456" {
		t.Errorf("Expected request_id to be 'req456', got %v", err.Context[ContextKeyRequestID.String()])
	}

	// Test with string keys (backward compatibility)
	err2 := New("CORE_VAL_INVALID_PARAM_001", "Invalid parameter")
	_ = err2.WithContext("custom_key", "custom_value")

	if err2.Context["custom_key"] != "custom_value" {
		t.Errorf("Expected custom_key to be 'custom_value', got %v", err2.Context["custom_key"])
	}
}

func TestAppError_Category(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"validation error", "CORE_VAL_INVALID_PARAM_001", CategoryValidation},
		{"resource error", "CORE_RES_NOT_FOUND_001", CategoryResource},
		{"auth error", "CORE_AUTH_UNAUTHORIZED_001", CategoryAuth},
		{"permission error", "CORE_PERM_FORBIDDEN_001", CategoryPermission},
		{"business error", "CORE_BIZ_INVALID_STATE_001", CategoryBusiness},
		{"system error", "CORE_SYS_INTERNAL_ERROR_001", CategorySystem},
		{"invalid code", "INVALID", "UNKNOWN"},
		{"empty code", "", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code, "test message")
			category := err.Category()
			if category != tt.expected {
				t.Errorf("Expected category %q, got %q", tt.expected, category)
			}
		})
	}
}

func TestIsValidation(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		err := New("CORE_VAL_INVALID_PARAM_001", "Invalid parameter")
		if !IsValidation(err) {
			t.Error("Expected IsValidation to return true")
		}
	})

	t.Run("non-validation error", func(t *testing.T) {
		err := New("CORE_RES_NOT_FOUND_001", "Not found")
		if IsValidation(err) {
			t.Error("Expected IsValidation to return false")
		}
	})

	t.Run("standard error", func(t *testing.T) {
		err := fmt.Errorf("standard error")
		if IsValidation(err) {
			t.Error("Expected IsValidation to return false for standard error")
		}
	})
}

func TestIsNotFound(t *testing.T) {
	t.Run("not found error", func(t *testing.T) {
		err := New("CORE_RES_NOT_FOUND_001", "Resource not found")
		if !IsNotFound(err) {
			t.Error("Expected IsNotFound to return true")
		}
	})

	t.Run("resource error without NOT_FOUND", func(t *testing.T) {
		err := New("CORE_RES_INVALID_001", "Invalid resource")
		if IsNotFound(err) {
			t.Error("Expected IsNotFound to return false")
		}
	})

	t.Run("non-resource error", func(t *testing.T) {
		err := New("CORE_VAL_INVALID_PARAM_001", "Invalid parameter")
		if IsNotFound(err) {
			t.Error("Expected IsNotFound to return false")
		}
	})

	t.Run("standard error", func(t *testing.T) {
		err := fmt.Errorf("standard error")
		if IsNotFound(err) {
			t.Error("Expected IsNotFound to return false for standard error")
		}
	})
}

func TestIsAuth(t *testing.T) {
	t.Run("auth error", func(t *testing.T) {
		err := New("CORE_AUTH_UNAUTHORIZED_001", "Unauthorized")
		if !IsAuth(err) {
			t.Error("Expected IsAuth to return true")
		}
	})

	t.Run("non-auth error", func(t *testing.T) {
		err := New("CORE_VAL_INVALID_PARAM_001", "Invalid parameter")
		if IsAuth(err) {
			t.Error("Expected IsAuth to return false")
		}
	})

	t.Run("standard error", func(t *testing.T) {
		err := fmt.Errorf("standard error")
		if IsAuth(err) {
			t.Error("Expected IsAuth to return false for standard error")
		}
	})
}

func TestIsSystem(t *testing.T) {
	t.Run("system error", func(t *testing.T) {
		err := New("CORE_SYS_INTERNAL_ERROR_001", "Internal error")
		if !IsSystem(err) {
			t.Error("Expected IsSystem to return true")
		}
	})

	t.Run("non-system error", func(t *testing.T) {
		err := New("CORE_VAL_INVALID_PARAM_001", "Invalid parameter")
		if IsSystem(err) {
			t.Error("Expected IsSystem to return false")
		}
	})

	t.Run("standard error", func(t *testing.T) {
		err := fmt.Errorf("standard error")
		if IsSystem(err) {
			t.Error("Expected IsSystem to return false for standard error")
		}
	})
}

func TestErrorChain(t *testing.T) {
	// Create a chain of errors
	baseErr := fmt.Errorf("base error")
	wrappedErr := Wrap(baseErr, "CORE_SYS_INTERNAL_ERROR_001", "wrapped error")

	// Test errors.Is
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is should find base error in chain")
	}

	// Test errors.As
	var appErr *AppError
	if !errors.As(wrappedErr, &appErr) {
		t.Error("errors.As should extract AppError from chain")
	}
	if appErr.Code != "CORE_SYS_INTERNAL_ERROR_001" {
		t.Errorf("Expected code CORE_SYS_INTERNAL_ERROR_001, got %s", appErr.Code)
	}
}
