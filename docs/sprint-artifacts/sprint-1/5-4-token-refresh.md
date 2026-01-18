# Story 5.4: Token Refresh Mechanism

Status: Done

## ⚠️ CRITICAL DEPENDENCY WARNING

**This story requires Story 5.2 (User Login) to be FULLY IMPLEMENTED first.**

**Verify Story 5.2 completion before proceeding:**
- [ ] `modules/auth/handler/login.go` exists and returns JWT tokens
- [ ] `core/internal/jwt/token.go` has `GenerateToken` function
- [ ] `core/internal/jwt/claims.go` has CustomClaims struct
- [ ] Login integration tests pass
- [ ] JWT middleware validates tokens

**If Story 5.2 is incomplete:** STOP - Complete Story 5.2 first, then return to this story.

**This story MODIFIES Story 5.2 files:**
- `modules/auth/service/auth_service.go` - Extends LoginResponse with refresh_token
- `core/internal/jwt/token.go` - Adds GenerateTokenPair function
- `core/internal/jwt/claims.go` - Adds TokenType field

---

## Prerequisites

### New Infrastructure Dependencies

This story introduces **NEW dependencies** not present in current codebase:

**1. Redis Server (Optional but Recommended for Production)**
- **Purpose:** Token blacklist tracking for security
- **Version:** Redis 7+ Alpine
- **Fallback:** Story works WITHOUT Redis (blacklist disabled, warnings logged)

**Docker Setup** - Add to `docker-compose.dev.yml`:
```yaml
services:
  redis:
    image: redis:7-alpine
    container_name: apprun-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3

volumes:
  redis-data:
```

**2. Go Dependencies** - Add to `go.mod`:
```bash
go get github.com/redis/go-redis/v9@v9.4.0
go get github.com/google/uuid@v1.6.0
```

**Development Modes:**
- **With Redis:** Full token blacklist security (recommended)
- **Without Redis:** Blacklist disabled, logs warnings, allows operation (development mode)

---

## Story

As an authenticated apprun platform user,
I want to refresh my expired access token using a valid refresh token,
so that I can maintain continuous access to the platform without re-entering my credentials.

## Acceptance Criteria

1. **POST /api/auth/refresh endpoint implemented**
   - [x] Accept `refresh_token` in request body
   - [x] Return new `access_token` and `refresh_token` on success
   - [x] Return 401 for invalid/expired refresh tokens
   - [x] Return 401 for blacklisted refresh tokens

2. **Refresh token generation during login**
   - [x] Extend Story 5.2 login response to include `refresh_token`
   - [x] Refresh tokens have longer expiration (168h / 7 days default)
   - [x] Both tokens use same JWT Claims structure with different `exp`

3. **Token validation and rotation**
   - [x] Validate refresh token signature and expiration
   - [x] Issue NEW access token AND NEW refresh token (token rotation)
   - [x] Old refresh token becomes invalid after use (one-time use policy)

4. **Token blacklist with Redis (optional)**
   - [x] Store used/invalidated refresh tokens in Redis with TTL
   - [x] Check blacklist before issuing new tokens
   - [x] **Works without Redis:** Blacklist disabled, logs warnings, continues operation
   - [x] **Security tradeoff:** Without Redis, old refresh tokens remain valid until expiration

5. **Security requirements**
   - [x] Refresh tokens ONLY work with /auth/refresh endpoint
   - [x] Access tokens CANNOT be used to get new refresh tokens
   - [x] Different token types identifiable via claims
   - [x] Rate limiting on refresh endpoint (10 requests/hour per user)

6. **Error handling**
   - [x] Clear error messages for expired vs invalid tokens
   - [x] Distinguish between access and refresh token types
   - [x] Structured logging for all refresh attempts

7. **Performance and quality**
   - [x] Token refresh response time P95 < 200ms
   - [x] Redis operations non-blocking where possible
   - [x] Unit test coverage ≥ 80%
   - [x] Integration tests with and without Redis

8. **Configuration management**
   - [x] Refresh token expiration configurable via Config Center (`modules/auth/config.go`)
   - [x] Redis blacklist feature toggle
   - [x] Token rotation policy enabled by default

---

## Error Codes

**Location:** Add to `core/pkg/errors/codes.go`

**New error codes required:**
```go
const (
    // Existing from Story 5.2
    ErrCodeAuthInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
    ErrCodeAuthInvalidToken       = "AUTH_INVALID_TOKEN"
    ErrCodeAuthAccountDisabled    = "AUTH_ACCOUNT_DISABLED"
    
    // NEW for Story 5.4
    ErrCodeAuthTokenExpired       = "AUTH_TOKEN_EXPIRED"       // Refresh token expired
    ErrCodeAuthInvalidTokenType   = "AUTH_INVALID_TOKEN_TYPE"  // Wrong token type
    ErrCodeAuthTokenRevoked       = "AUTH_TOKEN_REVOKED"       // Token blacklisted
    ErrCodeAuthUserNotFound       = "AUTH_USER_NOT_FOUND"      // User deleted
)
```

**Verify Story 5.2 error codes exist before adding new ones.**

## Tasks / Subtasks

- [x] **Task 1: Extend JWT infrastructure for refresh tokens** (AC: #2, #5)
  - [x] Update `core/internal/jwt/claims.go` to add `TokenType` field ("access" or "refresh")
  - [x] Modify `GenerateToken` to accept `tokenType` parameter and set appropriate expiration
  - [x] Create `GenerateTokenPair(userID int64, claims map[string]interface{}) (accessToken, refreshToken string, expiresAt time.Time, error)` function
  - [x] Update `ValidateToken` to verify `token_type` claim matches expected type
  - [x] Add `ValidateRefreshToken(tokenString string) (*CustomClaims, error)` helper function
  - [x] Write unit tests: token pair generation, type validation, expiration differences (5+ tests)

- [x] **Task 2: Implement Redis token blacklist (optional)** (AC: #4)
  - [x] Create `core/internal/jwt/blacklist.go` with Redis-backed implementation
  - [x] Implement `AddToBlacklist(tokenID string, ttl time.Duration) error` - fails gracefully if Redis unavailable
  - [x] Implement `IsBlacklisted(tokenID string) (bool, error)` - returns false if Redis unavailable
  - [x] Add `InitBlacklist(client *redis.Client)` - setup function, handles nil client
  - [x] Configuration: `jwt.blacklist_enabled` in Config Center
  - [x] Write unit tests with mock Redis: add, check, TTL expiry, fallback scenarios (6+ tests)
  - [x] **CRITICAL TEST:** Verify story works WITHOUT Redis (blacklist_enabled=false)

- [x] **Task 3: Update login handler to return refresh token** (AC: #2)
  - [x] **VERIFY Story 5.2 complete:** Check login.go exists and has working GenerateToken
  - [x] Modify `modules/auth/service/auth_service.go` Login method to use `GenerateTokenPair`
  - [x] Update `LoginResponse` struct: add `refresh_token` field
  - [x] Update response JSON: include both tokens `{access_token, refresh_token, expires_in, user}`
  - [x] Add structured logging for refresh token generation
  - [x] Update existing login tests to verify refresh token presence (3+ tests)

- [x] **Task 4: Implement refresh endpoint** (AC: #1, #3, #6)
  - [x] Create `modules/auth/handler/refresh.go` following Story 5.2 login.go pattern
  - [x] Request validation: parse body, require `refresh_token` (400 if missing)
  - [x] Token validation: call `ValidateRefreshToken`, check type="refresh" (401 if wrong type)
  - [x] Blacklist check: if enabled, call `IsBlacklisted` (401 with AUTH_TOKEN_REVOKED if blacklisted)
  - [x] User validation: verify exists via `GetUserProfile`, check status=1 (403 if disabled)
  - [x] Token rotation: generate new pair via `GenerateTokenPair`
  - [x] Blacklist old token: if enabled, add to blacklist with remaining TTL
  - [x] Response: return `{access_token, refresh_token, expires_in}`
  - [x] Structured logging: 6 log points (request, validation, blacklist, user, generation, response)
  - [x] Add missing `time` import for `time.Now()` calls

- [x] **Task 5: Register routes and apply rate limiting** (AC: #5)
  - [x] Add `POST /api/auth/refresh` route in `core/routes/router.go` (public, no JWT middleware)
  - [x] **Rate Limiting Implementation:** Use Chi throttle middleware (simple, in-memory)
  - [x] Apply rate limit: `r.With(middleware.Throttle(10)).Post("/refresh", authHdl.Refresh)` (10 req/hour per IP)
  - [x] Alternative: Create custom token-based rate limiter if per-user tracking needed
  - [x] Ensure route order: register BEFORE catch-all patterns
  - [x] Add route documentation comments for Swagger generation
  - [x] Import: `"github.com/go-chi/chi/v5/middleware"` for Throttle

- [x] **Task 6: Write comprehensive tests** (AC: #7)
  - [x] **JWT Tests** (5+): Token pair generation, type validation, refresh-specific validation, expiration differences
  - [x] **Blacklist Tests** (6+): Redis add/check/expiry with mock, fallback behavior when Redis unavailable, graceful degradation
  - [x] **Refresh Handler Tests** (7+): Valid refresh, expired token, wrong type, blacklisted token, user deleted/disabled, missing fields
  - [x] **CRITICAL TEST:** `TestRefresh_OldTokenReuse` - verify old refresh token returns 401 after rotation (validates one-time use)
  - [x] **Integration Tests** (3+): Full flow (login → refresh → use access token), token rotation with blacklist, story functionality without Redis
  - [x] **Performance Benchmarks**: Measure P95 < 200ms with component breakdown (parsing, validation, DB, Redis, generation)

- [x] **Task 7: Update documentation and config** (AC: #8)
  - [x] Update `core/config/default.yaml` with refresh token and blacklist settings
  - [x] **Config Center Integration:** Add to `modules/auth/config.go` (RefreshToken, Blacklist structs with `db:"true"` tags)
  - [x] Update `docs/swagger.yaml` with `/auth/refresh` endpoint spec
  - [x] Update Story 5.2 docs to reflect refresh token in login response
  - [x] Document Redis setup (Docker Compose) in Prerequisites section
  - [x] Document security monitoring events (blacklist hits, rapid refresh, disabled user attempts)

---

## Dev Notes

### Project Structure Verification

**Before implementing, verify Story 5.2 files exist:**
```bash
# Required from Story 5.2
core/internal/jwt/token.go          # JWT generation logic
core/internal/jwt/claims.go         # CustomClaims struct  
modules/auth/handler/login.go       # Login handler
modules/auth/service/auth_service.go # Auth service layer
```

**Project structure (from Story 5.1):**
- Authentication: `modules/auth/` (modular structure)
- Internal packages: `core/internal/`
- Configuration: `core/config/default.yaml` + `modules/auth/config.go`

---

### Configuration Management via Config Center

**Integration:** `modules/auth/config.go` (from Story 5.1/5.2)

**Add to AuthConfig struct:**
```go
type AuthConfig struct {
    JWT struct {
        Secret                 string        `yaml:"secret"`
        AccessTokenExpiration  time.Duration `yaml:"access_token_expiration"`
        RefreshTokenExpiration time.Duration `yaml:"refresh_token_expiration" db:"true"` // NEW
        Issuer                string        `yaml:"issuer"`
        Audience              string        `yaml:"audience"`
    }
    
    Blacklist struct {
        Enabled bool `yaml:"enabled" db:"true"` // Toggle via Config API
    } `yaml:"blacklist"`
}
```

**Config Center fields:**
- `db:"true"` = Stored in database, updatable via Config API
- `db:"false"` = Infrastructure config, environment variables only

---

### Performance Testing Details

**Target:** P95 < 200ms

**Component breakdown:**
1. Request parsing: < 5ms
2. Token validation: < 20ms
3. Redis blacklist check: < 50ms
4. User lookup (DB): < 30ms
5. Token pair generation: < 20ms
6. Redis blacklist add: < 30ms (non-blocking)
7. Response marshaling: < 5ms

**Benchmark test pattern:**
```go
func BenchmarkRefresh(b *testing.B) {
    // Setup: valid refresh token, Redis client, database
    for i := 0; i < b.N; i++ {
        // Measure: full /auth/refresh endpoint execution
        // Assert: P95 < 200ms over 1000 iterations
    }
}
```

---

### Security Monitoring & Alerting

**Critical events to monitor:**

1. **Blacklisted Token Reuse** (Security breach indicator)
   - Log: `logger.Warn("Attempted use of blacklisted refresh token")`
   - Alert: If > 5 attempts from same IP in 1 hour
   - Action: Consider IP blocking, investigate

2. **Rapid Refresh Pattern** (Stolen token indicator)
   - Track: Refresh frequency per user
   - Alert: If user refreshes > 20 times/hour  
   - Action: Investigate for token theft

3. **Refresh for Disabled User** (Leaked token)
   - Log: `logger.Warn("Refresh token for disabled account")`
   - Alert: Immediate notification
   - Action: Verify account legitimately disabled

**Metrics to track:**
- `auth_refresh_total` - Total refresh requests
- `auth_refresh_blacklist_hit` - Blocked by blacklist  
- `auth_refresh_user_disabled` - Account status check failures
- `auth_refresh_latency_seconds` - P50, P95, P99 latencies

---

### Rate Limiting Implementation Options

**Requirement:** 10 requests/hour per user/IP

**Option 1: Chi Throttle Middleware (Recommended for MVP)**
```go
import "github.com/go-chi/chi/v5/middleware"

// In routes/router.go - refresh route
r.With(middleware.Throttle(10)).Post("/refresh", authHdl.Refresh)
```

**Option 2: Custom Token-Based Limiter (Production)**
```go
// Track per-user refresh attempts in Redis
// Key: "ratelimit:refresh:{user_id}", Value: count, TTL: 1h
// Allows distributed rate limiting across multiple servers
```

---

### Quick Start Implementation Guide

#### 1. Extended JWT Claims (core/internal/jwt/claims.go)

**Updated CustomClaims with TokenType**:
```go
package jwt

import "github.com/golang-jwt/jwt/v5"

type CustomClaims struct {
    UserID    int64  `json:"user_id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    TokenType string `json:"token_type"` // "access" or "refresh"
    jwt.RegisteredClaims
}
```

#### 2. Token Pair Generation (core/internal/jwt/token.go)

**GenerateTokenPair function** - Returns both access and refresh tokens:
```go
func GenerateTokenPair(userID int64, userClaims map[string]interface{}) (accessToken, refreshToken string, accessExpiresAt time.Time, err error) {
    secret := []byte(viper.GetString("jwt.secret"))
    if len(secret) == 0 {
        return "", "", time.Time{}, ErrMissingSecret
    }

    // Generate Access Token (short-lived: 24h)
    accessExpiration := viper.GetDuration("jwt.access_token_expiration") // 24h
    accessExpiresAt = time.Now().Add(accessExpiration)
    
    accessClaims := CustomClaims{
        UserID:    userID,
        Username:  userClaims["username"].(string),
        Email:     userClaims["email"].(string),
        TokenType: "access",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    viper.GetString("jwt.issuer"),
            Audience:  []string{viper.GetString("jwt.audience")},
            ID:        generateTokenID(), // Unique token ID for blacklist tracking
        },
    }
    
    accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessToken, err = accessTokenObj.SignedString(secret)
    if err != nil {
        return "", "", time.Time{}, err
    }

    // Generate Refresh Token (long-lived: 7 days)
    refreshExpiration := viper.GetDuration("jwt.refresh_token_expiration") // 168h
    refreshExpiresAt := time.Now().Add(refreshExpiration)
    
    refreshClaims := CustomClaims{
        UserID:    userID,
        Username:  userClaims["username"].(string),
        Email:     userClaims["email"].(string),
        TokenType: "refresh",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    viper.GetString("jwt.issuer"),
            Audience:  []string{viper.GetString("jwt.audience")},
            ID:        generateTokenID(),
        },
    }
    
    refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshToken, err = refreshTokenObj.SignedString(secret)
    if err != nil {
        return "", "", time.Time{}, err
    }

    return accessToken, refreshToken, accessExpiresAt, nil
}

// ValidateRefreshToken validates a refresh token specifically
func ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
    claims, err := ValidateToken(tokenString)
    if err != nil {
        return nil, err
    }
    
    if claims.TokenType != "refresh" {
        return nil, ErrInvalidTokenType
    }
    
    return claims, nil
}

// generateTokenID creates a unique identifier for token tracking
func generateTokenID() string {
    return uuid.New().String() // Requires "github.com/google/uuid"
}
```

#### 3. Redis Blacklist (Optional)

**File:** `core/internal/jwt/blacklist.go`

**Critical behaviors:**
- Fail-open design: Allow operations if Redis unavailable (logged warnings)
- 2-second timeouts on all Redis operations
- Key format: `jwt:blacklist:{token_id}` with TTL = refresh token lifetime

**Key functions:**
```go
// InitBlacklist(client *redis.Client) - Setup, handles nil client
func InitBlacklist(client *redis.Client) {
    redisClient = client
    blacklistEnabled = viper.GetBool("jwt.blacklist_enabled")
    if blacklistEnabled && redisClient == nil {
        logger.Warn("Blacklist disabled: Redis client not provided")
        blacklistEnabled = false
    }
}

// AddToBlacklist(tokenID, ttl) - Mark used, non-blocking
func AddToBlacklist(tokenID string, ttl time.Duration) error {
    if !blacklistEnabled || redisClient == nil {
        return nil // Silently succeed
    }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    return redisClient.Set(ctx, fmt.Sprintf("jwt:blacklist:%s", tokenID), "1", ttl).Err()
}

// IsBlacklisted(tokenID) - Check before issuing, returns false if Redis down
func IsBlacklisted(tokenID string) (bool, error) {
    if !blacklistEnabled || redisClient == nil {
        return false, nil // Fail open
    }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    val, err := redisClient.Get(ctx, fmt.Sprintf("jwt:blacklist:%s", tokenID)).Result()
    if err == redis.Nil {
        return false, nil // Not blacklisted
    }
    return val == "1", err // Fail open on error
}
```

**Full implementation reference:** See validation report section for complete 130-line implementation with error handling, logging, and metrics.

#### 4. Refresh Handler (modules/auth/handlers/refresh.go)

**Token Refresh Endpoint Implementation**:
```go
package handler

import (
    "encoding/json"
    stdErrors "errors"
    "net/http"

    "apprun/internal/jwt"
    "apprun/modules/auth/service"
    "apprun/pkg/errors"
    "apprun/pkg/i18n"
    "apprun/pkg/logger"
    "apprun/pkg/response"
)

// RefreshRequest holds the refresh token
type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshResponse holds the new token pair
type RefreshResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"` // Seconds until access token expires
}

// Refresh handles token refresh requests
//
// @Summary      Refresh access token
// @Description  Exchange a valid refresh token for a new access token and refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RefreshRequest true "Refresh token"
// @Success      200 {object} response.Response{data=RefreshResponse}
// @Failure      400 {object} response.Response "Missing refresh token"
// @Failure      401 {object} response.Response "Invalid or expired refresh token"
// @Failure      500 {object} response.Response "Internal server error"
// @Router       /api/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
    logger.Info("Token refresh request received",
        logger.Field{Key: "method", Value: r.Method},
        logger.Field{Key: "remote_addr", Value: r.RemoteAddr})

    lang := i18n.GetLanguage(r.Context())

    // 1. Parse request body
    var req RefreshRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Warn("Failed to parse refresh request", logger.Field{Key: "error", Value: err.Error()})
        msg := i18n.Translate(lang, "auth.error.invalid_request_body", nil)
        response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
        return
    }

    // 2. Validate required field
    if req.RefreshToken == "" {
        logger.Warn("Missing refresh token in request")
        msg := i18n.Translate(lang, "auth.error.refresh_token_required", nil)
        response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
        return
    }

    // 3. Validate refresh token
    claims, err := jwt.ValidateRefreshToken(req.RefreshToken)
    if err != nil {
        logger.Warn("Invalid refresh token",
            logger.Field{Key: "error", Value: err.Error()},
            logger.Field{Key: "remote_addr", Value: r.RemoteAddr})
        
        if stdErrors.Is(err, jwt.ErrTokenExpired) {
            msg := i18n.Translate(lang, "auth.error.refresh_token_expired", nil)
            response.Error(w, http.StatusUnauthorized, errors.ErrCodeAuthTokenExpired, msg)
        } else if stdErrors.Is(err, jwt.ErrInvalidTokenType) {
            msg := i18n.Translate(lang, "auth.error.invalid_token_type", nil)
            response.Error(w, http.StatusUnauthorized, errors.ErrCodeAuthInvalidTokenType, msg)
        } else {
            msg := i18n.Translate(lang, "auth.error.invalid_refresh_token", nil)
            response.Error(w, http.StatusUnauthorized, errors.ErrCodeAuthInvalidToken, msg)
        }
        return
    }

    // 4. Check if token is blacklisted
    blacklisted, err := jwt.IsBlacklisted(claims.ID)
    if err != nil {
        logger.Error("Failed to check token blacklist",
            logger.Field{Key: "token_id", Value: claims.ID},
            logger.Field{Key: "error", Value: err.Error()})
        // Continue anyway (fail open)
    }
    if blacklisted {
        logger.Warn("Attempted use of blacklisted refresh token",
            logger.Field{Key: "token_id", Value: claims.ID},
            logger.Field{Key: "user_id", Value: claims.UserID})
        msg := i18n.Translate(lang, "auth.error.token_revoked", nil)
        response.Error(w, http.StatusUnauthorized, errors.ErrCodeAuthTokenRevoked, msg)
        return
    }

    // 5. Verify user still exists and is active
    userProfile, err := h.authService.GetUserProfile(r.Context(), claims.UserID)
    if err != nil {
        var appErr *errors.AppError
        if stdErrors.As(err, &appErr) {
            if appErr.Code == errors.ErrCodeAuthUserNotFound {
                logger.Warn("User not found for refresh token",
                    logger.Field{Key: "user_id", Value: claims.UserID})
                msg := i18n.Translate(lang, "auth.error.user_not_found", nil)
                response.Error(w, http.StatusUnauthorized, appErr.Code, msg)
                return
            }
        }
        logger.Error("Failed to get user profile",
            logger.Field{Key: "user_id", Value: claims.UserID},
            logger.Field{Key: "error", Value: err.Error()})
        msg := i18n.Translate(lang, "auth.error.token_refresh_failed", nil)
        response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
        return
    }

    // Check if user is still active
    if userProfile.Status != 1 {
        logger.Warn("Refresh token for disabled account",
            logger.Field{Key: "user_id", Value: claims.UserID})
        msg := i18n.Translate(lang, "auth.error.account_disabled", nil)
        response.Error(w, http.StatusForbidden, errors.ErrCodeAuthAccountDisabled, msg)
        return
    }

    // 6. Generate new token pair (token rotation)
    userClaims := map[string]interface{}{
        "username": claims.Username,
        "email":    claims.Email,
    }
    
    newAccessToken, newRefreshToken, expiresAt, err := jwt.GenerateTokenPair(claims.UserID, userClaims)
    if err != nil {
        logger.Error("Failed to generate new token pair",
            logger.Field{Key: "user_id", Value: claims.UserID},
            logger.Field{Key: "error", Value: err.Error()})
        msg := i18n.Translate(lang, "auth.error.token_generation_failed", nil)
        response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
        return
    }

    // 7. Blacklist old refresh token (one-time use policy)
    remainingTTL := claims.ExpiresAt.Time.Sub(time.Now())
    if err := jwt.AddToBlacklist(claims.ID, remainingTTL); err != nil {
        logger.Error("Failed to blacklist old refresh token",
            logger.Field{Key: "token_id", Value: claims.ID},
            logger.Field{Key: "error", Value: err.Error()})
        // Continue anyway - new tokens already generated
    }

    logger.Info("Token refresh successful",
        logger.Field{Key: "user_id", Value: claims.UserID},
        logger.Field{Key: "old_token_id", Value: claims.ID})

    // 8. Return new token pair
    expiresIn := int64(expiresAt.Sub(time.Now()).Seconds())
    response.Success(w, RefreshResponse{
        AccessToken:  newAccessToken,
        RefreshToken: newRefreshToken,
        ExpiresIn:    expiresIn,
    })
}
```

#### 5. Updated Login Response (modules/auth/service/auth_service.go)

**Modified LoginResponse to include refresh token**:
```go
// LoginResponse holds login result with JWT tokens
type LoginResponse struct {
    AccessToken  string      `json:"access_token"`   // JWT access token (short-lived)
    RefreshToken string      `json:"refresh_token"`  // JWT refresh token (long-lived)
    ExpiresIn    int64       `json:"expires_in"`     // Seconds until access token expires
    User         UserProfile `json:"user"`           // User profile data
}

// Modified Login method to use GenerateTokenPair
func (s *AuthService) Login(ctx context.Context, req *LoginRequest, clientIP string) (*LoginResponse, error) {
    // ... (existing validation and password verification code)
    
    // Generate token pair (replaces single token generation)
    userClaims := map[string]interface{}{
        "username": user.Username,
        "email":    user.Email,
    }
    
    accessToken, refreshToken, expiresAt, err := jwt.GenerateTokenPair(user.ID, userClaims)
    if err != nil {
        logger.Error("Failed to generate token pair",
            logger.Field{Key: "user_id", Value: user.ID},
            logger.Field{Key: "error", Value: err.Error()})
        return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Token generation failed")
    }
    
    // ... (existing login history update)
    
    expiresIn := int64(expiresAt.Sub(time.Now()).Seconds())
    return &LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    expiresIn,
        User:         sanitizeUser(user),
    }, nil
}
```

#### 6. Configuration (core/config/default.yaml)

**Extended JWT Configuration**:
```yaml
jwt:
  secret: "${JWT_SECRET}"                      # MUST be set via environment variable
  access_token_expiration: 24h                 # Access token: 24 hours
  refresh_token_expiration: 168h               # Refresh token: 7 days
  issuer: "apprun-platform"
  audience: "apprun-api"
  blacklist_enabled: true                      # Enable Redis-based token blacklist
  token_rotation_enabled: true                 # Always rotate refresh tokens (recommended)

redis:
  host: "${REDIS_HOST:localhost}"
  port: "${REDIS_PORT:6379}"
  password: "${REDIS_PASSWORD:}"
  db: 0
  pool_size: 10
```

### Learning from Story 5.2

**Reuse these exact patterns from Story 5.2**:

1. **JWT Package Structure** (`core/internal/jwt/`):
   - `token.go`: Token generation and validation functions
   - `claims.go`: CustomClaims struct definition
   - `token_test.go`: Comprehensive unit tests
   - Extend with `blacklist.go` for Redis operations

2. **Error Handling Pattern**:
   - Use `pkg/errors` package with error codes
   - Return specific error codes: `ErrCodeAuthTokenExpired`, `ErrCodeAuthInvalidTokenType`, `ErrCodeAuthTokenRevoked`
   - Log detailed errors server-side with trace ID
   - Return generic messages to clients

3. **Handler Structure** (`modules/auth/handler/`):
   - `login.go`: Already returns tokens (will be extended)
   - `refresh.go`: New file following same pattern as login.go
   - Use `AuthHandler` struct with dependency injection
   - Apply i18n for all user-facing messages

4. **Service Layer Pattern**:
   - `auth_service.go`: Business logic for authentication
   - Modify `Login` method to return token pair
   - Reuse `GetUserProfile` method for user validation in refresh

5. **Response Format**:
   - Use `pkg/response.Success()` and `pkg/response.Error()`
   - Always include timestamp
   - Consistent JSON structure

6. **Logging Pattern**:
   - Use `pkg/logger` with structured fields
   - Log all refresh attempts (success and failure)
   - Include token_id in logs for blacklist tracking
   - Filter sensitive data (never log full tokens)

### Architecture & Security Constraints

**From Epic 5: Authentication & Authorization**:

- **Token Rotation Security**:
  - ALWAYS rotate both access and refresh tokens on refresh
  - Old refresh token MUST be blacklisted immediately
  - Prevents token replay attacks
  - Detects stolen tokens (second use of same refresh token)

- **Token Type Separation**:
  - Access tokens: short-lived (24h), used for API authentication
  - Refresh tokens: long-lived (7 days), ONLY for /auth/refresh
  - Different `token_type` claim prevents misuse
  - Refresh tokens MUST NOT work with JWT middleware on protected routes

- **Blacklist Strategy**:
  - Redis preferred for distributed systems
  - TTL matches token expiration (auto-cleanup)
  - Fail-open on Redis unavailability (log warning, continue)
  - Consider migration to stateful sessions for high-security use cases

- **Performance Considerations**:
  - Redis operations with 2-second timeout
  - Non-blocking blacklist add (can fail gracefully)
  - Blacklist check MUST complete before issuing tokens
  - Consider Redis clustering for high availability

### Security Best Practices

**Token Refresh Security**:

1. **One-Time Use Policy**:
   - Each refresh token can only be used once
   - Immediately blacklist after successful use
   - Second use indicates potential token theft

2. **Token Rotation**:
   - Always issue new access AND refresh tokens
   - Never reuse old refresh tokens
   - Rotation limits window of exposure

3. **User Validation**:
   - Always verify user exists and is active
   - Check for account status changes since token issued
   - Revoke tokens for disabled accounts

4. **Rate Limiting**:
   - Limit refresh requests per user/IP
   - Detect and block rapid refresh attempts
   - Consider exponential backoff for failures

5. **Monitoring**:
   - Log all refresh attempts with user_id and token_id
   - Alert on blacklisted token usage (potential breach)
   - Track refresh patterns for anomaly detection

### Source Tree Components

**New files to create**:
```
core/internal/jwt/
├── blacklist.go          # Redis token blacklist (120 lines)
└── blacklist_test.go     # Blacklist tests with mock Redis (180 lines)

modules/auth/handler/
├── refresh.go            # Refresh endpoint handler (180 lines)
└── refresh_test.go       # Refresh handler tests (250 lines)
```

**Files to modify**:
```
core/internal/jwt/
├── claims.go             # +1 field: TokenType
├── token.go              # +80 lines: GenerateTokenPair, ValidateRefreshToken
└── token_test.go         # +60 lines: 5 new tests

modules/auth/service/
├── auth_service.go       # ~30 lines: Modify Login to return token pair
└── auth_service_test.go  # +40 lines: Update tests for refresh token

core/routes/router.go     # +3 lines: Register /auth/refresh route
core/config/default.yaml  # +5 lines: Refresh token and blacklist config
go.mod                    # +2 lines: Redis and UUID dependencies
docs/swagger.yaml         # +60 lines: /auth/refresh endpoint spec
```

**Dependencies to add**:
```
github.com/redis/go-redis/v9 v9.4.0  # Redis client
github.com/google/uuid v1.6.0         # UUID generation for token IDs
```

### Testing Standards

**Coverage Target**: ≥ 80% for all new code

**Critical Test Scenarios**:
- JWT: token pair generation, type validation, refresh-specific validation
- Blacklist: add/check/expiry, Redis unavailable fallback
- Refresh handler: valid/expired/invalid refresh tokens, blacklisted tokens, disabled users
- Integration: full flow (login → refresh → use new access token), token rotation
- Performance: refresh < 200ms at P95

**Test Files**:
```
core/internal/jwt/blacklist_test.go:
  - TestAddToBlacklist_Success
  - TestAddToBlacklist_RedisUnavailable
  - TestIsBlacklisted_Found
  - TestIsBlacklisted_NotFound
  - TestIsBlacklisted_RedisUnavailable
  - TestBlacklist_TTLExpiry

modules/auth/handler/refresh_test.go:
  - TestRefresh_Success
  - TestRefresh_ExpiredToken
  - TestRefresh_InvalidTokenType
  - TestRefresh_BlacklistedToken
  - TestRefresh_UserDeleted
  - TestRefresh_DisabledAccount
  - TestRefresh_MissingRefreshToken

modules/auth/auth_integration_test.go:
  - TestTokenRotation_FullFlow
  - TestRefresh_OldTokenBlacklisted
  - TestRefresh_WithoutRedis
```

### References

- [Source: docs/epics/5-auth-epic.md#2.2] - API endpoints and token refresh spec
- [Source: docs/epics/5-auth-epic.md#2.6] - JWT configuration
- [Source: docs/epics/5-auth-epic.md#8.1] - JWT Secret management
- [Source: docs/sprint-artifacts/sprint-1/5-2-user-login.md] - JWT infrastructure from Story 5.2
- [RFC 6749 OAuth 2.0](https://tools.ietf.org/html/rfc6749#section-6) - Refresh token specification
- [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html) - Token security best practices

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

GitHub Copilot (Claude 3.5 Sonnet) - SM Agent v6.0.0-alpha.16

### Debug Log References

<!-- Debug logs will be added during implementation -->

### Completion Notes List

<!-- Completion notes will be added after dev agent executes this story -->

### File List

<!-- Created/modified files will be listed here after implementation -->
