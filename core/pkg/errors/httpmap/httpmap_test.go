package httpmap

import (
	"fmt"
	"net/http"
	"testing"

	"apprun/pkg/errors"
)

func TestToHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: http.StatusOK,
		},
		{
			name:     "validation error",
			err:      errors.New("CORE_VAL_INVALID_PARAM_001", "Invalid parameter"),
			expected: http.StatusBadRequest,
		},
		{
			name:     "resource not found error",
			err:      errors.New("CORE_RES_NOT_FOUND_001", "Resource not found"),
			expected: http.StatusNotFound,
		},
		{
			name:     "resource error without NOT_FOUND",
			err:      errors.New("CORE_RES_INVALID_001", "Invalid resource"),
			expected: http.StatusBadRequest,
		},
		{
			name:     "auth error",
			err:      errors.New("CORE_AUTH_UNAUTHORIZED_001", "Unauthorized"),
			expected: http.StatusUnauthorized,
		},
		{
			name:     "permission error",
			err:      errors.New("CORE_PERM_FORBIDDEN_001", "Forbidden"),
			expected: http.StatusForbidden,
		},
		{
			name:     "business error",
			err:      errors.New("CORE_BIZ_INVALID_STATE_001", "Invalid state"),
			expected: http.StatusUnprocessableEntity,
		},
		{
			name:     "system error",
			err:      errors.New("CORE_SYS_INTERNAL_ERROR_001", "Internal error"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "unknown category",
			err:      errors.New("UNKNOWN_ERROR", "Unknown error"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "standard error",
			err:      fmt.Errorf("standard error"),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := ToHTTPStatus(tt.err)
			if status != tt.expected {
				t.Errorf("Expected HTTP status %d, got %d", tt.expected, status)
			}
		})
	}
}

func TestToHTTPStatus_WrappedError(t *testing.T) {
	baseErr := fmt.Errorf("database connection failed")
	wrappedErr := errors.Wrap(baseErr, "CORE_SYS_INTERNAL_ERROR_001", "System error")

	status := ToHTTPStatus(wrappedErr)
	if status != http.StatusInternalServerError {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusInternalServerError, status)
	}
}

func TestToHTTPStatus_WithContext(t *testing.T) {
	err := errors.New("CORE_VAL_INVALID_PARAM_001", "Invalid parameter").
		WithContext(errors.ContextKeyUserID, "user123")

	status := ToHTTPStatus(err)
	if status != http.StatusBadRequest {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusBadRequest, status)
	}
}
