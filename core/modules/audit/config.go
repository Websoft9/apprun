// Package audit provides audit logging configuration management.
package audit

import "time"

// ============================================================================
// Module Constants (Action Types)
// ============================================================================

const (
	// Authentication actions
	ActionAuthLogin         = "auth.login"
	ActionAuthLogout        = "auth.logout"
	ActionAuthPasswordReset = "auth.password_reset"
	//nolint:gosec // G101: This is an action type constant, not a credential
	ActionAuthTokenRefresh = "auth.token_refresh"
	ActionAuthFailed       = "auth.failed"

	// Resource CRUD actions
	ActionResourceCreate = "resource.create"
	ActionResourceRead   = "resource.read"
	ActionResourceUpdate = "resource.update"
	ActionResourceDelete = "resource.delete"
	ActionResourceList   = "resource.list"

	// Permission actions
	ActionPermissionCheck  = "permission.check"
	ActionPermissionGrant  = "permission.grant"
	ActionPermissionRevoke = "permission.revoke"

	// Configuration actions
	ActionConfigUpdate = "config.update"
	ActionConfigExport = "config.export"
	ActionConfigImport = "config.import"

	// System actions
	ActionSystemStartup  = "system.startup"
	ActionSystemShutdown = "system.shutdown"
	ActionSystemBackup   = "system.backup"
	ActionSystemRestore  = "system.restore"

	// Data export/import
	ActionDataExport = "data.export"
	ActionDataImport = "data.import"
)

// ============================================================================
// Target Types for Audit Logs
// ============================================================================

const (
	TargetTypeUser         = "user"
	TargetTypeProject      = "project"
	TargetTypeRole         = "role"
	TargetTypePermission   = "permission"
	TargetTypeConfig       = "config"
	TargetTypeSession      = "session"
	TargetTypeAuditLog     = "audit_log"
	TargetTypeBackup       = "backup"
	TargetTypeApplication  = "application"
	TargetTypeEnvironment  = "environment"
	TargetTypeDeployment   = "deployment"
	TargetTypeNotification = "notification"
)

// ============================================================================
// Default Configuration Values
// ============================================================================

const (
	// Service defaults
	DefaultBufferSize       = 1000
	DefaultWorkerCount      = 4
	DefaultFlushInterval    = 5 * time.Second
	DefaultFallbackFilePath = "/var/log/apprun/audit-fallback.log"

	// Middleware defaults
	DefaultMiddlewareEnabled = true

	// Query defaults
	DefaultPageSize    = 20
	MaxPageSize        = 200
	DefaultQueryPeriod = 30 * 24 * time.Hour // 30 days

	// Retention defaults
	DefaultRetentionDays  = 365 // 1 year
	DefaultArchiveEnabled = false
	DefaultArchivePath    = "/var/log/apprun/audit-archive"
)

// ============================================================================
// Sensitive Fields for Redaction
// ============================================================================

var (
	// DefaultSensitiveFields defines fields that should be redacted in audit logs
	DefaultSensitiveFields = []string{
		"password",
		"token",
		"secret",
		"api_key",
		"authorization",
		"access_token",
		"refresh_token",
		"jwt",
		"private_key",
		"certificate",
		"credential",
		"auth_code",
	}

	// DefaultExcludePaths defines paths that should not be audited
	DefaultExcludePaths = []string{
		"/health",
		"/metrics",
		"/api/docs",
		"/swagger",
		"/favicon.ico",
		"/static",
	}
)

// ============================================================================
// Configuration Structures (Config Center Managed)
// ============================================================================

// Config holds all audit-related configuration.
// This struct follows the Configuration Center architecture.
type Config struct {
	Service    ServiceConfig    `mapstructure:"service" json:"service"`       // Audit service configuration
	Middleware MiddlewareConfig `mapstructure:"middleware" json:"middleware"` // Middleware configuration
	Storage    StorageConfig    `mapstructure:"storage" json:"storage"`       // Storage configuration
	Retention  RetentionConfig  `mapstructure:"retention" json:"retention"`   // Log retention policy
}

// ServiceConfig defines the audit service settings.
type ServiceConfig struct {
	// Enabled controls whether audit logging is active
	// db:"true" - Can be toggled at runtime
	Enabled bool `mapstructure:"enabled" json:"enabled" default:"true" db:"true"`

	// BufferSize defines the size of the async processing queue
	// Larger = more memory, better burst handling
	// db:"true" - Can be adjusted for performance tuning
	BufferSize int `mapstructure:"buffer_size" json:"buffer_size" default:"1000" db:"true" validate:"omitempty,min=100,max=10000"`

	// WorkerCount defines the number of concurrent workers processing audit logs
	// Higher = more CPU usage, faster processing
	// db:"true" - Can be adjusted for performance tuning
	WorkerCount int `mapstructure:"worker_count" json:"worker_count" default:"4" db:"true" validate:"omitempty,min=1,max=32"`

	// FlushInterval defines how often to flush buffered logs (if using batch writes)
	// db:"true" - Can be adjusted for performance/latency balance
	FlushInterval time.Duration `mapstructure:"flush_interval" json:"flush_interval" default:"5s" db:"true" validate:"omitempty,min=1s,max=60s"`

	// FallbackFilePath defines where to write logs if database is unavailable
	// db:"true" - Can be updated at runtime
	FallbackFilePath string `mapstructure:"fallback_file_path" json:"fallback_file_path" default:"/var/log/apprun/audit-fallback.log" db:"true"`
}

// MiddlewareConfig defines the HTTP middleware settings for audit logging.
type MiddlewareConfig struct {
	// Enabled controls whether HTTP request auditing is active
	// db:"true" - Can be toggled at runtime
	Enabled bool `mapstructure:"enabled" json:"enabled" default:"true" db:"true"`

	// ExcludePaths defines paths that should not be audited
	// db:"true" - Can be updated at runtime
	ExcludePaths []string `mapstructure:"exclude_paths" json:"exclude_paths" db:"true"`

	// SensitiveFields defines field names that should be redacted in audit logs
	// db:"true" - Can be updated at runtime
	SensitiveFields []string `mapstructure:"sensitive_fields" json:"sensitive_fields" db:"true"`

	// IncludeRequestBody controls whether to log request bodies
	// db:"true" - Can be toggled at runtime
	IncludeRequestBody bool `mapstructure:"include_request_body" json:"include_request_body" default:"false" db:"true"`

	// IncludeResponseBody controls whether to log response bodies
	// db:"true" - Can be toggled at runtime
	IncludeResponseBody bool `mapstructure:"include_response_body" json:"include_response_body" default:"false" db:"true"`

	// MaxBodySize defines maximum size of request/response body to log (in bytes)
	// db:"true" - Can be adjusted at runtime
	MaxBodySize int `mapstructure:"max_body_size" json:"max_body_size" default:"4096" db:"true" validate:"omitempty,min=0,max=65536"`
}

// StorageConfig defines the storage backend settings.
type StorageConfig struct {
	// Type defines the storage backend: "database", "file", or "both"
	// db:"true" - Can be changed at runtime
	Type string `mapstructure:"type" json:"type" default:"database" db:"true" validate:"oneof=database file both"`

	// AsyncWrite enables asynchronous writes to storage
	// db:"true" - Can be toggled at runtime
	AsyncWrite bool `mapstructure:"async_write" json:"async_write" default:"true" db:"true"`

	// FilePath defines the path for file-based storage
	// db:"true" - Can be updated at runtime
	FilePath string `mapstructure:"file_path" json:"file_path" default:"/var/log/apprun/audit.log" db:"true"`

	// FileRotateSize defines max size (MB) before rotating log file
	// db:"true" - Can be adjusted at runtime
	FileRotateSize int `mapstructure:"file_rotate_size" json:"file_rotate_size" default:"100" db:"true" validate:"omitempty,min=1,max=1000"`

	// FileRotateAge defines max age (days) before rotating log file
	// db:"true" - Can be adjusted at runtime
	FileRotateAge int `mapstructure:"file_rotate_age" json:"file_rotate_age" default:"7" db:"true" validate:"omitempty,min=1,max=365"`
}

// RetentionConfig defines log retention policy.
type RetentionConfig struct {
	// Enabled controls whether automatic cleanup is active
	// db:"true" - Can be toggled at runtime
	Enabled bool `mapstructure:"enabled" json:"enabled" default:"false" db:"true"`

	// Days defines how many days to retain audit logs
	// db:"true" - Can be adjusted at runtime
	Days int `mapstructure:"days" json:"days" default:"365" db:"true" validate:"omitempty,min=1,max=3650"`

	// ArchiveEnabled controls whether to archive old logs before deletion
	// db:"true" - Can be toggled at runtime
	ArchiveEnabled bool `mapstructure:"archive_enabled" json:"archive_enabled" default:"false" db:"true"`

	// ArchivePath defines where to store archived logs
	// db:"true" - Can be updated at runtime
	ArchivePath string `mapstructure:"archive_path" json:"archive_path" default:"/var/log/apprun/audit-archive" db:"true"`
}

// DefaultConfig returns a Config with sensible defaults.
// This is used when no configuration file is present.
func DefaultConfig() *Config {
	return &Config{
		Service:    DefaultServiceConfig(),
		Middleware: DefaultMiddlewareConfig(),
		Storage:    DefaultStorageConfig(),
		Retention:  DefaultRetentionConfig(),
	}
}

// DefaultServiceConfig returns default service configuration.
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		Enabled:          true,
		BufferSize:       DefaultBufferSize,
		WorkerCount:      DefaultWorkerCount,
		FlushInterval:    DefaultFlushInterval,
		FallbackFilePath: DefaultFallbackFilePath,
	}
}

// DefaultMiddlewareConfig returns default middleware configuration.
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		Enabled:             DefaultMiddlewareEnabled,
		ExcludePaths:        DefaultExcludePaths,
		SensitiveFields:     DefaultSensitiveFields,
		IncludeRequestBody:  false,
		IncludeResponseBody: false,
		MaxBodySize:         4096,
	}
}

// DefaultStorageConfig returns default storage configuration.
func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Type:           "database",
		AsyncWrite:     true,
		FilePath:       "/var/log/apprun/audit.log",
		FileRotateSize: 100,
		FileRotateAge:  7,
	}
}

// DefaultRetentionConfig returns default retention configuration.
func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		Enabled:        false,
		Days:           DefaultRetentionDays,
		ArchiveEnabled: DefaultArchiveEnabled,
		ArchivePath:    DefaultArchivePath,
	}
}
