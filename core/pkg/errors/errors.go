// Package errors provides business error wrapping with error codes, context, and HTTP mapping.
//
// Example usage:
//
//	// Create a new error
//	err := errors.New(errors.ErrCodeInvalidParam, "Invalid username")
//
//	// Create with formatting
//	err := errors.Newf(errors.ErrCodeInvalidParam, "Invalid parameter: %s", paramName)
//
//	// Wrap an existing error
//	if err := db.Query(...); err != nil {
//	    return errors.Wrap(err, errors.ErrCodeInternalError, "Database query failed")
//	}
//
//	// Add context
//	err := errors.New(errors.ErrCodeNotFound, "User not found").
//	    WithContext(errors.ContextKeyUserID, userID).
//	    WithContext(errors.ContextKeyRequestID, requestID)
//
//	// Check error type
//	if errors.IsValidation(err) {
//	    // Handle validation error
//	}
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// AppError represents a business error with code, message, underlying error, and context
type AppError struct {
	Code    string                 // Error code in format: <MODULE>_<CATEGORY>_<NAME>_<NNN>
	Message string                 // User-friendly error message
	Err     error                  // Underlying error (for error chain)
	Context map[string]interface{} // Additional context information
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for error chain support
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithContext adds a key-value pair to the error context (chainable).
// It accepts both string keys and ContextKey typed keys for type safety.
func (e *AppError) WithContext(key interface{}, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}

	// Convert key to string
	var keyStr string
	switch k := key.(type) {
	case string:
		keyStr = k
	case ContextKey:
		keyStr = k.String()
	default:
		keyStr = fmt.Sprintf("%v", k)
	}

	e.Context[keyStr] = value
	return e
}

// Category extracts and returns the category from the error code
// Format: <MODULE>_<CATEGORY>_<NAME>_<NNN>
// Returns: CATEGORY (e.g., "VAL", "RES", "AUTH")
func (e *AppError) Category() string {
	parts := strings.Split(e.Code, "_")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "UNKNOWN"
}

// New creates a new AppError with the given code and message
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Newf creates a new AppError with formatted message
func Newf(code, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with error code and message
func Wrap(err error, code, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrapf wraps an existing error with error code and formatted message
func Wrapf(err error, code, format string, args ...interface{}) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Category() == CategoryValidation
	}
	return false
}

// IsNotFound checks if the error is a resource not found error
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Category() == CategoryResource && strings.Contains(appErr.Code, "NOT_FOUND")
	}
	return false
}

// IsAuth checks if the error is an authentication error
func IsAuth(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Category() == CategoryAuth
	}
	return false
}

// IsSystem checks if the error is a system error
func IsSystem(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Category() == CategorySystem
	}
	return false
}
