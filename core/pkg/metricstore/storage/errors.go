package storage

import (
	stdErrors "errors"
	"fmt"

	"apprun/pkg/errors"
)

// Error codes for metrics storage operations.
// These will be registered in pkg/errors/codes.go.
const (
	// ErrCodeMetricsStorageNotFound is returned when a metric name doesn't exist
	ErrCodeMetricsStorageNotFound = "METRICS_STORAGE_NOT_FOUND_001"

	// ErrCodeMetricsStorageUnavailable is returned when the backend is unreachable
	ErrCodeMetricsStorageUnavailable = "METRICS_STORAGE_UNAVAILABLE_002"

	// ErrCodeMetricsStorageInvalidConfig is returned for configuration errors
	ErrCodeMetricsStorageInvalidConfig = "METRICS_STORAGE_INVALID_CONFIG_003"

	// ErrCodeMetricsStorageTimeout is returned when an operation times out
	ErrCodeMetricsStorageTimeout = "METRICS_STORAGE_TIMEOUT_004"
)

// ErrNotFound creates an error for when a metric key is not found.
//
// Example:
//
//	return ErrNotFound("http_requests_total")
func ErrNotFound(key string) error {
	return errors.New(
		ErrCodeMetricsStorageNotFound,
		fmt.Sprintf("metric not found: %s", key),
	)
}

// ErrUnavailable creates an error for when the storage backend is unreachable.
//
// Example:
//
//	return ErrUnavailable("badger", dbErr)
func ErrUnavailable(backend string, cause error) error {
	msg := fmt.Sprintf("storage backend '%s' unavailable", backend)
	if cause != nil {
		return errors.Wrap(
			cause,
			ErrCodeMetricsStorageUnavailable,
			msg,
		)
	}
	return errors.New(
		ErrCodeMetricsStorageUnavailable,
		msg,
	)
}

// ErrInvalidConfig creates an error for configuration validation failures.
//
// Example:
//
//	return ErrInvalidConfig("backend", "must be 'badger' or 'prometheus'")
func ErrInvalidConfig(field string, reason string) error {
	return errors.New(
		ErrCodeMetricsStorageInvalidConfig,
		fmt.Sprintf("invalid config field '%s': %s", field, reason),
	)
}

// ErrTimeout creates an error for when an operation exceeds its timeout.
//
// Example:
//
//	return ErrTimeout("Query", 5*time.Second)
func ErrTimeout(operation string, timeout interface{}) error {
	return errors.New(
		ErrCodeMetricsStorageTimeout,
		fmt.Sprintf("operation '%s' timed out after %v", operation, timeout),
	)
}

// IsRetryable determines if an error is temporary and should be retried.
// Retryable errors: ErrUnavailable, ErrTimeout
// Non-retryable errors: ErrNotFound, ErrInvalidConfig, validation errors
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Get the error code from our errors package
	var appErr *errors.AppError
	if !stdErrors.As(err, &appErr) {
		// For non-AppError types, don't retry by default
		return false
	}

	switch appErr.Code {
	case ErrCodeMetricsStorageUnavailable,
		ErrCodeMetricsStorageTimeout:
		// These are temporary errors - should retry
		return true
	case ErrCodeMetricsStorageNotFound,
		ErrCodeMetricsStorageInvalidConfig:
		// These are permanent errors - don't retry
		return false
	default:
		// Unknown errors - don't retry by default
		return false
	}
}
