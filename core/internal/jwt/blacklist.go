package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"apprun/pkg/cache"
	pkgconfig "apprun/pkg/config"
	"apprun/pkg/logger"
)

const (
	// blacklistOperationTimeout is the maximum time allowed for blacklist operations
	blacklistOperationTimeout = 2 * time.Second
)

var (
	cacheClient      cache.Client
	blacklistEnabled bool
	configProvider   pkgconfig.Provider
)

// InitBlacklist initializes the cache client for token blacklisting.
// Handles nil client gracefully - blacklist will be disabled with warnings.
func InitBlacklist(client cache.Client, cfg pkgconfig.Provider) {
	cacheClient = client
	configProvider = cfg

	if configProvider != nil {
		blacklistEnabled = configProvider.GetBool("jwt.blacklist_enabled")
	}

	if blacklistEnabled && cacheClient == nil {
		logger.Warn("JWT blacklist enabled but cache client not provided - blacklist will be disabled")
		blacklistEnabled = false
	}

	if blacklistEnabled {
		logger.Info("JWT blacklist initialized with cache backend")
	} else {
		logger.Info("JWT blacklist disabled - tokens will not be tracked")
	}
}

// AddToBlacklist adds a token ID to the blacklist with TTL.
// Fails gracefully if cache unavailable - logs warning and continues.
func AddToBlacklist(tokenID string, ttl time.Duration) error {
	if !blacklistEnabled {
		return nil // Silently succeed if blacklist disabled
	}

	if cacheClient == nil {
		logger.Warn("Cache client unavailable, cannot blacklist token",
			logger.Field{Key: "token_id", Value: tokenID})
		return nil // Non-blocking: allow operation to continue
	}

	ctx, cancel := context.WithTimeout(context.Background(), blacklistOperationTimeout)
	defer cancel()

	key := fmt.Sprintf("jwt:blacklist:%s", tokenID)
	err := cacheClient.Set(ctx, key, "1", ttl)
	if err != nil {
		logger.Error("Failed to add token to blacklist",
			logger.Field{Key: "token_id", Value: tokenID},
			logger.Field{Key: "error", Value: err.Error()})
		return err // Return error but don't fail the operation
	}

	logger.Debug("Token added to blacklist",
		logger.Field{Key: "token_id", Value: tokenID},
		logger.Field{Key: "ttl", Value: ttl.String()})
	return nil
}

// IsBlacklisted checks if a token ID is in the blacklist.
// Returns false if blacklist disabled or cache unavailable (fail-open design).
func IsBlacklisted(tokenID string) (bool, error) {
	if !blacklistEnabled {
		return false, nil
	}

	if cacheClient == nil {
		logger.Warn("Cache client unavailable, cannot check blacklist",
			logger.Field{Key: "token_id", Value: tokenID})
		return false, nil // Fail open: allow if cache unavailable
	}

	ctx, cancel := context.WithTimeout(context.Background(), blacklistOperationTimeout)
	defer cancel()

	key := fmt.Sprintf("jwt:blacklist:%s", tokenID)
	val, err := cacheClient.Get(ctx, key)

	if errors.Is(err, cache.ErrKeyNotFound) {
		return false, nil // Not in blacklist
	}
	if err != nil {
		logger.Error("Failed to check blacklist",
			logger.Field{Key: "token_id", Value: tokenID},
			logger.Field{Key: "error", Value: err.Error()})
		return false, err // Fail open on error
	}

	return val == "1", nil
}

// IsBlacklistEnabled returns whether the blacklist feature is enabled.
func IsBlacklistEnabled() bool {
	return blacklistEnabled
}
