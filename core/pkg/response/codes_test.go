package response

import (
	"testing"
)

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantCode string
	}{
		{"validation invalid param", ErrCodeInvalidParam, "VAL_INVALID_PARAM_001"},
		{"validation missing field", ErrCodeMissingField, "VAL_MISSING_FIELD_002"},
		{"validation format error", ErrCodeFormatError, "VAL_FORMAT_ERROR_003"},
		{"resource not found", ErrCodeNotFound, "RES_NOT_FOUND_001"},
		{"resource already exists", ErrCodeAlreadyExists, "RES_ALREADY_EXISTS_002"},
		{"resource conflict", ErrCodeConflict, "RES_CONFLICT_003"},
		{"auth invalid token", ErrCodeInvalidToken, "AUTH_INVALID_TOKEN_001"},
		{"auth token expired", ErrCodeTokenExpired, "AUTH_TOKEN_EXPIRED_002"},
		{"auth unauthorized", ErrCodeUnauthorized, "AUTH_UNAUTHORIZED_003"},
		{"permission forbidden", ErrCodeForbidden, "PERM_FORBIDDEN_001"},
		{"permission insufficient role", ErrCodeInsufficientRole, "PERM_INSUFFICIENT_ROLE_002"},
		{"business quota exceeded", ErrCodeQuotaExceeded, "BIZ_QUOTA_EXCEEDED_001"},
		{"business operation failed", ErrCodeOperationFailed, "BIZ_OPERATION_FAILED_002"},
		{"system internal error", ErrCodeInternalError, "SYS_INTERNAL_ERROR_001"},
		{"system service unavailable", ErrCodeServiceUnavailable, "SYS_SERVICE_UNAVAILABLE_002"},
		{"system database error", ErrCodeDatabaseError, "SYS_DATABASE_ERROR_003"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.wantCode {
				t.Errorf("Error code = %v, want %v", tt.code, tt.wantCode)
			}
		})
	}
}
