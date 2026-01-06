package errors

// Category constants for error classification
const (
	CategoryValidation = "VAL"  // Validation errors
	CategoryResource   = "RES"  // Resource errors
	CategoryAuth       = "AUTH" // Authentication errors
	CategoryPermission = "PERM" // Permission errors
	CategoryBusiness   = "BIZ"  // Business logic errors
	CategorySystem     = "SYS"  // System errors
)

// ContextKey is a type-safe key for error context values
type ContextKey string

// Context key constants for standardized context usage
const (
	ContextKeyUserID    ContextKey = "user_id"
	ContextKeyProjectID ContextKey = "project_id"
	ContextKeyRequestID ContextKey = "request_id"
	ContextKeyTraceID   ContextKey = "trace_id"
	ContextKeySessionID ContextKey = "session_id"
	ContextKeyIPAddress ContextKey = "ip_address"
)

// String returns the string value of the context key
func (k ContextKey) String() string {
	return string(k)
}

// =============================================================================
// CORE Module Error Codes
// =============================================================================

// CORE - Common error codes
const (
	ErrCodeNotFound      = "CORE_RES_NOT_FOUND_001"
	ErrCodeInvalidParam  = "CORE_VAL_INVALID_PARAM_001"
	ErrCodeUnauthorized  = "CORE_AUTH_UNAUTHORIZED_001"
	ErrCodeInternalError = "CORE_SYS_INTERNAL_ERROR_001"
	ErrCodeForbidden     = "CORE_PERM_FORBIDDEN_001"
	ErrCodeConflict      = "CORE_BIZ_CONFLICT_001"
)

// =============================================================================
// CONFIG Module Error Codes (Configuration Management)
// =============================================================================

// CONFIG - Validation errors
const (
	ErrCodeConfigInvalidKey    = "CONFIG_VAL_INVALID_KEY_001"
	ErrCodeConfigInvalidValue  = "CONFIG_VAL_INVALID_VALUE_002"
	ErrCodeConfigInvalidFormat = "CONFIG_VAL_INVALID_FORMAT_003"
	ErrCodeConfigInvalidTag    = "CONFIG_VAL_INVALID_TAG_004"
	ErrCodeConfigMissingParam  = "CONFIG_VAL_MISSING_PARAM_005"
)

// CONFIG - Resource errors
const (
	ErrCodeConfigNotFound     = "CONFIG_RES_NOT_FOUND_001"
	ErrCodeConfigFileNotFound = "CONFIG_RES_FILE_NOT_FOUND_002"
)

// CONFIG - Business errors
const (
	ErrCodeConfigReadOnly     = "CONFIG_BIZ_READ_ONLY_001"
	ErrCodeConfigConflict     = "CONFIG_BIZ_CONFLICT_002"
	ErrCodeConfigNotDynamic   = "CONFIG_BIZ_NOT_DYNAMIC_003"
	ErrCodeConfigNotAllowedDB = "CONFIG_BIZ_NOT_ALLOWED_DB_004"
)

// CONFIG - System errors
const (
	ErrCodeConfigQueryFailed     = "CONFIG_SYS_QUERY_FAILED_001"
	ErrCodeConfigUpdateFailed    = "CONFIG_SYS_UPDATE_FAILED_002"
	ErrCodeConfigCreateFailed    = "CONFIG_SYS_CREATE_FAILED_003"
	ErrCodeConfigDeleteFailed    = "CONFIG_SYS_DELETE_FAILED_004"
	ErrCodeConfigLoadFailed      = "CONFIG_SYS_LOAD_FAILED_005"
	ErrCodeConfigParseFailed     = "CONFIG_SYS_PARSE_FAILED_006"
	ErrCodeConfigUnmarshalFailed = "CONFIG_SYS_UNMARSHAL_FAILED_007"
	ErrCodeConfigMetadataFailed  = "CONFIG_SYS_METADATA_FAILED_008"
)

// =============================================================================
// AUTH Module Error Codes (Authentication & Authorization)
// =============================================================================

// AUTH - Validation errors
const (
	ErrCodeAuthInvalidEmail = "AUTH_VAL_INVALID_EMAIL_001"
	// #nosec G101 -- this is an error code constant, not a credential
	ErrCodeAuthWeakPassword = "AUTH_VAL_WEAK_PASSWORD_002"
	// #nosec G101 -- this is an error code constant, not a credential
	ErrCodeAuthInvalidToken = "AUTH_VAL_INVALID_TOKEN_003"
	ErrCodeAuthInvalidCode  = "AUTH_VAL_INVALID_CODE_004"
	ErrCodeAuthMissingParam = "AUTH_VAL_MISSING_PARAM_005"
)

// AUTH - Resource errors
const (
	ErrCodeAuthUserNotFound    = "AUTH_RES_USER_NOT_FOUND_001"
	ErrCodeAuthRoleNotFound    = "AUTH_RES_ROLE_NOT_FOUND_002"
	ErrCodeAuthProjectNotFound = "AUTH_RES_PROJECT_NOT_FOUND_003"
)

// AUTH - Authentication errors
const (
	// #nosec G101 -- this is an error code constant, not a credential
	ErrCodeAuthInvalidCredentials = "AUTH_AUTH_INVALID_CREDENTIALS_001"
	// #nosec G101 -- this is an error code constant, not a credential
	ErrCodeAuthTokenExpired    = "AUTH_AUTH_TOKEN_EXPIRED_002"
	ErrCodeAuthSessionExpired  = "AUTH_AUTH_SESSION_EXPIRED_003"
	ErrCodeAuthAccountDisabled = "AUTH_AUTH_ACCOUNT_DISABLED_004"
	ErrCodeAuthAccountLocked   = "AUTH_AUTH_ACCOUNT_LOCKED_005"
	ErrCodeAuthTooManyAttempts = "AUTH_AUTH_TOO_MANY_ATTEMPTS_006"
)

// AUTH - Permission errors
const (
	ErrCodeAuthNoPermission    = "AUTH_PERM_NO_PERMISSION_001"
	ErrCodeAuthProjectNoAccess = "AUTH_PERM_PROJECT_NO_ACCESS_002"
	ErrCodeAuthRoleNotAssigned = "AUTH_PERM_ROLE_NOT_ASSIGNED_003"
)

// AUTH - Business logic errors
const (
	ErrCodeAuthEmailExists    = "AUTH_BIZ_EMAIL_EXISTS_001"
	ErrCodeAuthUsernameExists = "AUTH_BIZ_USERNAME_EXISTS_002"
	// #nosec G101 -- this is an error code constant, not a credential
	ErrCodeAuthTokenUsed = "AUTH_BIZ_TOKEN_USED_003"
	// #nosec G101 -- this is an error code constant, not a credential
	ErrCodeAuthInvalidOldPass = "AUTH_BIZ_INVALID_OLD_PASS_004"
)

// =============================================================================
// DATA Module Error Codes (Data Modeling & CRUD)
// =============================================================================

// DATA - Validation errors
const (
	ErrCodeDataInvalidSchema    = "DATA_VAL_INVALID_SCHEMA_001"
	ErrCodeDataInvalidField     = "DATA_VAL_INVALID_FIELD_002"
	ErrCodeDataInvalidRelation  = "DATA_VAL_INVALID_RELATION_003"
	ErrCodeDataInvalidQuery     = "DATA_VAL_INVALID_QUERY_004"
	ErrCodeDataConstraintFailed = "DATA_VAL_CONSTRAINT_FAILED_005"
)

// DATA - Resource errors
const (
	ErrCodeDataModelNotFound  = "DATA_RES_MODEL_NOT_FOUND_001"
	ErrCodeDataRecordNotFound = "DATA_RES_RECORD_NOT_FOUND_002"
	ErrCodeDataFieldNotFound  = "DATA_RES_FIELD_NOT_FOUND_003"
)

// DATA - Business logic errors
const (
	ErrCodeDataModelExists      = "DATA_BIZ_MODEL_EXISTS_001"
	ErrCodeDataDuplicateKey     = "DATA_BIZ_DUPLICATE_KEY_002"
	ErrCodeDataForeignKeyFailed = "DATA_BIZ_FOREIGN_KEY_FAILED_003"
	ErrCodeDataMigrationFailed  = "DATA_BIZ_MIGRATION_FAILED_004"
)

// DATA - System errors
const (
	ErrCodeDataDatabaseError = "DATA_SYS_DATABASE_ERROR_001"
	ErrCodeDataQueryTimeout  = "DATA_SYS_QUERY_TIMEOUT_002"
)

// =============================================================================
// FUNC Module Error Codes (Function Service)
// =============================================================================

// FUNC - Validation errors
const (
	ErrCodeFuncInvalidCode    = "FUNC_VAL_INVALID_CODE_001"
	ErrCodeFuncInvalidTrigger = "FUNC_VAL_INVALID_TRIGGER_002"
	ErrCodeFuncInvalidRuntime = "FUNC_VAL_INVALID_RUNTIME_003"
)

// FUNC - Resource errors
const (
	ErrCodeFuncNotFound = "FUNC_RES_NOT_FOUND_001"
)

// FUNC - Business logic errors
const (
	ErrCodeFuncDeployFailed    = "FUNC_BIZ_DEPLOY_FAILED_001"
	ErrCodeFuncExecutionFailed = "FUNC_BIZ_EXECUTION_FAILED_002"
	ErrCodeFuncTimeout         = "FUNC_BIZ_TIMEOUT_003"
	ErrCodeFuncResourceLimit   = "FUNC_BIZ_RESOURCE_LIMIT_004"
)

// =============================================================================
// PLUGIN Module Error Codes (Plugin Extension)
// =============================================================================

// PLUGIN - Validation errors
const (
	ErrCodePluginInvalidManifest = "PLUGIN_VAL_INVALID_MANIFEST_001"
	ErrCodePluginInvalidVersion  = "PLUGIN_VAL_INVALID_VERSION_002"
)

// PLUGIN - Resource errors
const (
	ErrCodePluginNotFound = "PLUGIN_RES_NOT_FOUND_001"
)

// PLUGIN - Business logic errors
const (
	ErrCodePluginLoadFailed    = "PLUGIN_BIZ_LOAD_FAILED_001"
	ErrCodePluginAlreadyLoaded = "PLUGIN_BIZ_ALREADY_LOADED_002"
	ErrCodePluginIncompatible  = "PLUGIN_BIZ_INCOMPATIBLE_003"
)

// =============================================================================
// STORAGE Module Error Codes (File Storage)
// =============================================================================

// STORAGE - Validation errors
const (
	ErrCodeStorageInvalidPath     = "STORAGE_VAL_INVALID_PATH_001"
	ErrCodeStorageInvalidFilename = "STORAGE_VAL_INVALID_FILENAME_002"
	ErrCodeStorageFileTooLarge    = "STORAGE_VAL_FILE_TOO_LARGE_003"
	ErrCodeStorageInvalidMimeType = "STORAGE_VAL_INVALID_MIME_TYPE_004"
)

// STORAGE - Resource errors
const (
	ErrCodeStorageFileNotFound   = "STORAGE_RES_FILE_NOT_FOUND_001"
	ErrCodeStorageFolderNotFound = "STORAGE_RES_FOLDER_NOT_FOUND_002"
)

// STORAGE - Business logic errors
const (
	ErrCodeStorageFileExists    = "STORAGE_BIZ_FILE_EXISTS_001"
	ErrCodeStorageFolderExists  = "STORAGE_BIZ_FOLDER_EXISTS_002"
	ErrCodeStorageQuotaExceeded = "STORAGE_BIZ_QUOTA_EXCEEDED_003"
	ErrCodeStorageUploadFailed  = "STORAGE_BIZ_UPLOAD_FAILED_004"
)

// STORAGE - System errors
const (
	ErrCodeStorageS3Error = "STORAGE_SYS_S3_ERROR_001"
)

// =============================================================================
// WORKFLOW Module Error Codes (Workflow Engine)
// =============================================================================

// WORKFLOW - Validation errors
const (
	ErrCodeWorkflowInvalidDefinition = "WORKFLOW_VAL_INVALID_DEFINITION_001"
	ErrCodeWorkflowInvalidNode       = "WORKFLOW_VAL_INVALID_NODE_002"
	ErrCodeWorkflowInvalidTrigger    = "WORKFLOW_VAL_INVALID_TRIGGER_003"
)

// WORKFLOW - Resource errors
const (
	ErrCodeWorkflowNotFound          = "WORKFLOW_RES_NOT_FOUND_001"
	ErrCodeWorkflowExecutionNotFound = "WORKFLOW_RES_EXECUTION_NOT_FOUND_002"
)

// WORKFLOW - Business logic errors
const (
	ErrCodeWorkflowExecutionFailed = "WORKFLOW_BIZ_EXECUTION_FAILED_001"
	ErrCodeWorkflowNodeFailed      = "WORKFLOW_BIZ_NODE_FAILED_002"
	ErrCodeWorkflowTimeout         = "WORKFLOW_BIZ_TIMEOUT_003"
	ErrCodeWorkflowRetryExhausted  = "WORKFLOW_BIZ_RETRY_EXHAUSTED_004"
)

// =============================================================================
// EVENT Module Error Codes (Event Bus)
// =============================================================================

// EVENT - Validation errors
const (
	ErrCodeEventInvalidTopic   = "EVENT_VAL_INVALID_TOPIC_001"
	ErrCodeEventInvalidPayload = "EVENT_VAL_INVALID_PAYLOAD_002"
)

// EVENT - Resource errors
const (
	ErrCodeEventTopicNotFound        = "EVENT_RES_TOPIC_NOT_FOUND_001"
	ErrCodeEventSubscriptionNotFound = "EVENT_RES_SUBSCRIPTION_NOT_FOUND_002"
)

// EVENT - Business logic errors
const (
	ErrCodeEventPublishFailed   = "EVENT_BIZ_PUBLISH_FAILED_001"
	ErrCodeEventSubscribeFailed = "EVENT_BIZ_SUBSCRIBE_FAILED_002"
)

// =============================================================================
// REALTIME Module Error Codes (Real-time Push)
// =============================================================================

// REALTIME - Validation errors
const (
	ErrCodeRealtimeInvalidChannel = "REALTIME_VAL_INVALID_CHANNEL_001"
	ErrCodeRealtimeInvalidMessage = "REALTIME_VAL_INVALID_MESSAGE_002"
)

// REALTIME - Business logic errors
const (
	ErrCodeRealtimeConnectFailed    = "REALTIME_BIZ_CONNECT_FAILED_001"
	ErrCodeRealtimePushFailed       = "REALTIME_BIZ_PUSH_FAILED_002"
	ErrCodeRealtimeConnectionClosed = "REALTIME_BIZ_CONNECTION_CLOSED_003"
)

// =============================================================================
// I18N Module Error Codes (Internationalization)
// =============================================================================

// I18N - Validation errors
const (
	ErrCodeI18nInvalidLocale = "I18N_VAL_INVALID_LOCALE_001"
	ErrCodeI18nInvalidKey    = "I18N_VAL_INVALID_KEY_002"
)

// I18N - Resource errors
const (
	ErrCodeI18nTranslationNotFound = "I18N_RES_TRANSLATION_NOT_FOUND_001"
)

// I18N - Business logic errors
const (
	ErrCodeI18nImportFailed = "I18N_BIZ_IMPORT_FAILED_001"
)

// =============================================================================
// LOG Module Error Codes (Logging & Monitoring)
// =============================================================================

// LOG - Validation errors
const (
	ErrCodeLogInvalidQuery     = "LOG_VAL_INVALID_QUERY_001"
	ErrCodeLogInvalidTimeRange = "LOG_VAL_INVALID_TIME_RANGE_002"
)

// LOG - Resource errors
const (
	ErrCodeLogNotFound = "LOG_RES_NOT_FOUND_001"
)

// LOG - System errors
const (
	ErrCodeLogStorageError     = "LOG_SYS_STORAGE_ERROR_001"
	ErrCodeLogQueryTimeout     = "LOG_SYS_QUERY_TIMEOUT_002"
	ErrCodeLogInvalidConfig    = "LOG_SYS_INVALID_CONFIG_003"
	ErrCodeLogOutputFailed     = "LOG_SYS_OUTPUT_FAILED_004"
	ErrCodeLogFileOpenFailed   = "LOG_SYS_FILE_OPEN_FAILED_005"
	ErrCodeLogLevelParseFailed = "LOG_SYS_LEVEL_PARSE_FAILED_006"
)

// =============================================================================
// SERVER Module Error Codes (HTTP/HTTPS Server)
// =============================================================================

// SERVER - System errors
const (
	ErrCodeServerStartFailed    = "SERVER_SYS_START_FAILED_001"
	ErrCodeServerShutdownFailed = "SERVER_SYS_SHUTDOWN_FAILED_002"
	ErrCodeServerTLSLoadFailed  = "SERVER_SYS_TLS_LOAD_FAILED_003"
)

// =============================================================================
// DATABASE Module Error Codes (Database Connection & Migration)
// =============================================================================

// DATABASE - System errors
const (
	ErrCodeDatabaseConnectFailed = "DATABASE_SYS_CONNECT_FAILED_001"
	ErrCodeDatabaseMigrateFailed = "DATABASE_SYS_MIGRATE_FAILED_002"
	ErrCodeDatabaseQueryFailed   = "DATABASE_SYS_QUERY_FAILED_003"
	ErrCodeDatabaseCloseFailed   = "DATABASE_SYS_CLOSE_FAILED_004"
)

// =============================================================================
// GATEWAY Module Error Codes (API Gateway)
// =============================================================================

// GATEWAY - Validation errors
const (
	ErrCodeGatewayInvalidRoute  = "GATEWAY_VAL_INVALID_ROUTE_001"
	ErrCodeGatewayInvalidTarget = "GATEWAY_VAL_INVALID_TARGET_002"
)

// GATEWAY - Resource errors
const (
	ErrCodeGatewayRouteNotFound   = "GATEWAY_RES_ROUTE_NOT_FOUND_001"
	ErrCodeGatewayServiceNotFound = "GATEWAY_RES_SERVICE_NOT_FOUND_002"
)

// GATEWAY - Business logic errors
const (
	ErrCodeGatewayProxyFailed = "GATEWAY_BIZ_PROXY_FAILED_001"
	ErrCodeGatewayTimeout     = "GATEWAY_BIZ_TIMEOUT_002"
	ErrCodeGatewayServiceDown = "GATEWAY_BIZ_SERVICE_DOWN_003"
)

// =============================================================================
// LICENSE Module Error Codes (License Management)
// =============================================================================

// LICENSE - Validation errors
const (
	ErrCodeLicenseInvalidFormat = "LICENSE_VAL_INVALID_FORMAT_001"
	ErrCodeLicenseExpired       = "LICENSE_VAL_EXPIRED_002"
)

// LICENSE - Business logic errors
const (
	ErrCodeLicenseVerifyFailed    = "LICENSE_BIZ_VERIFY_FAILED_001"
	ErrCodeLicenseFeatureDisabled = "LICENSE_BIZ_FEATURE_DISABLED_002"
	ErrCodeLicenseQuotaExceeded   = "LICENSE_BIZ_QUOTA_EXCEEDED_003"
)
