package response

// Error code constants following API design standards
// Format: <MODULE>_<ERROR_TYPE>_<NUMBER>

const (
	// Validation Errors (VAL_*)
	ErrCodeInvalidParam = "VAL_INVALID_PARAM_001"
	ErrCodeMissingField = "VAL_MISSING_FIELD_002"
	ErrCodeFormatError  = "VAL_FORMAT_ERROR_003"

	// Resource Errors (RES_*)
	ErrCodeNotFound      = "RES_NOT_FOUND_001"
	ErrCodeAlreadyExists = "RES_ALREADY_EXISTS_002"
	ErrCodeConflict      = "RES_CONFLICT_003"

	// Authentication Errors (AUTH_*)
	ErrCodeInvalidToken = "AUTH_INVALID_TOKEN_001"
	ErrCodeTokenExpired = "AUTH_TOKEN_EXPIRED_002"
	ErrCodeUnauthorized = "AUTH_UNAUTHORIZED_003"

	// Permission Errors (PERM_*)
	ErrCodeForbidden        = "PERM_FORBIDDEN_001"
	ErrCodeInsufficientRole = "PERM_INSUFFICIENT_ROLE_002"

	// Business Errors (BIZ_*)
	ErrCodeQuotaExceeded   = "BIZ_QUOTA_EXCEEDED_001"
	ErrCodeOperationFailed = "BIZ_OPERATION_FAILED_002"

	// System Errors (SYS_*)
	ErrCodeInternalError      = "SYS_INTERNAL_ERROR_001"
	ErrCodeServiceUnavailable = "SYS_SERVICE_UNAVAILABLE_002"
	ErrCodeDatabaseError      = "SYS_DATABASE_ERROR_003"
)
