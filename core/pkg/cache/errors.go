package cache

import (
	"apprun/pkg/errors"
)

// Cache error codes following pkg/errors framework
const (
	CodeConnectFailed = "CACHE_CONNECT_001"
	CodeOperationFail = "CACHE_OP_002"
	CodeInvalidConfig = "CACHE_CONFIG_003"
	CodeKeyNotFound   = "CACHE_NOT_FOUND_004"
	CodeTimeout       = "CACHE_TIMEOUT_005"
	CodePubSubFailed  = "CACHE_PUBSUB_006"
	CodeSerializeFail = "CACHE_SERIALIZE_007"
)

// Predefined cache errors
var (
	// ErrConnectFailed is returned when connection to Redis fails
	ErrConnectFailed = errors.New(CodeConnectFailed, "failed to connect to cache")

	// ErrOperationFailed is returned when a cache operation fails
	ErrOperationFailed = errors.New(CodeOperationFail, "cache operation failed")

	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New(CodeInvalidConfig, "invalid cache configuration")

	// ErrKeyNotFound is returned when a key doesn't exist
	ErrKeyNotFound = errors.New(CodeKeyNotFound, "key not found in cache")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New(CodeTimeout, "cache operation timeout")

	// ErrPubSubFailed is returned when pub/sub operation fails
	ErrPubSubFailed = errors.New(CodePubSubFailed, "pub/sub operation failed")

	// ErrSerializationFailed is returned when value serialization/deserialization fails
	ErrSerializationFailed = errors.New(CodeSerializeFail, "value serialization failed")
)
