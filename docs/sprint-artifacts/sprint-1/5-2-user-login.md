# Story 5.2: User Login

Status: drafted

## Story

As a registered apprun platform user,
I want to log in using my email/username and password and receive a JWT token,
so that I can securely access protected APIs and retrieve my profile information.

## Acceptance Criteria

1. **POST /api/auth/login endpoint implemented**
   - Accept `identifier` (username or email) and `password` fields
   - Return JWT token + user info on successful authentication
   - Return 401 for invalid credentials (generic error message)
   - Return 403 for disabled accounts

2. **GET /api/auth/me endpoint implemented**
   - Require valid JWT token in Authorization header
   - Return complete user profile (excluding password hash)
   - Return 401 for missing/invalid token

3. **Password verification using bcrypt**
   - Use `core/internal/password` package from Story 5.1
   - Constant-time comparison to prevent timing attacks

4. **JWT token generation**
   - Create `core/internal/jwt` package
   - Generate tokens with user claims (user_id, username, email)
   - Set expiration time (24h default, configurable)
   - Sign with HMAC-SHA256

5. **Login history tracking**
   - Update `last_login_at` timestamp on successful login
   - Record `last_login_ip` from request

6. **User status validation**
   - Check user `status = 1` (active) before issuing token
   - Reject login for disabled accounts (status = 0)

7. **Security requirements**
   - Generic error messages (don't reveal if user exists)
   - No passwords in logs or responses
   - Structured logging with trace IDs
   - Input validation on all fields

8. **Performance and quality**
   - Login response time P95 < 500ms
   - /auth/me response time P95 < 100ms
   - Unit test coverage ≥ 80%
   - Integration tests for all scenarios

## Tasks / Subtasks

- [ ] **Task 1: Create JWT infrastructure** (AC: #4)
  - [ ] Add dependency: `github.com/golang-jwt/jwt/v5 v5.3.0` to go.mod
  - [ ] Create `core/internal/jwt/token.go` with `GenerateToken(userID int64, claims map[string]interface{}) (string, time.Time, error)`
  - [ ] Create `core/internal/jwt/claims.go` with CustomClaims struct (see Dev Notes for structure)
  - [ ] Implement `ValidateToken(tokenString string) (*CustomClaims, error)` for token verification
  - [ ] Add JWT configuration to `core/config/default.yaml` with secret, expiration, issuer, audience
  - [ ] Write unit tests: token generation, validation success, expired token, invalid signature (4 tests minimum)

- [ ] **Task 2: Implement login handler** (AC: #1, #3, #5, #6, #7)
  - [ ] Create `modules/auth/handlers/login.go`
  - [ ] Parse and validate login request (identifier + password required, return 400 if missing)
  - [ ] Query user by username OR email using Ent: `client.User.Query().Where(user.Or(user.UsernameEQ(...), user.EmailEQ(...)))` (see Dev Notes)
  - [ ] Verify password: call `password.VerifyPassword(req.Password, user.PasswordHash)` from Story 5.1
  - [ ] Check user status: reject if `user.Status != 1`, return 403 with error code `AUTH_ACCOUNT_DISABLED`
  - [ ] Generate JWT token: call `jwt.GenerateToken(user.ID, map[string]interface{}{"username": user.Username, "email": user.Email})`
  - [ ] Update login history: `user.Update().SetLastLoginAt(time.Now()).SetLastLoginIP(getClientIP(r)).SaveX(ctx)` (non-blocking, log errors but don't fail login)
  - [ ] Return response: use `response.Success(w, LoginResponse{User: sanitizeUser(user), Token: token, ExpiresAt: expiresAt})`
  - [ ] Add structured logging: 5 log points (request start, user query, password verify, token generate, response) with trace ID

- [ ] **Task 3: Implement /me handler** (AC: #2)
  - [ ] Create `modules/auth/handlers/me.go`
  - [ ] Extract user ID from JWT claims: `userID, ok := r.Context().Value("user_id").(int64)` (see Dev Notes for pattern)
  - [ ] Query user by ID: `client.User.Get(ctx, userID)`
  - [ ] Return sanitized user profile: exclude `password_hash`, include all other fields
  - [ ] Handle user-not-found: return 404 with error code `USER_NOT_FOUND` (user may be deleted after token issued)

- [ ] **Task 4: Register routes** (AC: #1, #2)
  - [ ] Add `POST /api/auth/login` route in `core/routes/router.go` without JWT middleware
  - [ ] Add `GET /api/auth/me` route with JWT middleware: `r.With(middleware.RequireAuth()).Get("/me", handlers.Me)`
  - [ ] Apply middleware stack order: RequestID → Logger → CORS → JWT (for /me only)

- [ ] **Task 5: Write comprehensive tests** (AC: #8)
  - [ ] JWT package tests: generation, validation, expired, invalid signature (4+ tests)
  - [ ] Login handler tests: valid username, valid email, user not found, wrong password, disabled account, missing fields (6+ tests)
  - [ ] Me handler tests: valid token, missing token, invalid token, user deleted (4+ tests)
  - [ ] Integration tests: full login flow with database, /me with real JWT
  - [ ] Performance benchmarks: use `go test -bench` to verify login < 500ms, /me < 100ms at P95

- [ ] **Task 6: Documentation** (AC: #8)
  - [ ] Update `docs/swagger.yaml` with `/auth/login` and `/auth/me` specs
  - [ ] Add inline code comments explaining JWT claims structure and error handling
  - [ ] Update `docs/api.md` with login flow sequence diagram

## Dev Notes

### Quick Start Implementation Guide

#### 1. JWT Package Setup (core/internal/jwt/)

**claims.go** - Custom JWT claims structure:
```go
package jwt

import "github.com/golang-jwt/jwt/v5"

type CustomClaims struct {
    UserID   int64  `json:"user_id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    jwt.RegisteredClaims
}
```

**token.go** - Token generation and validation:
```go
package jwt

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/spf13/viper"
)

func GenerateToken(userID int64, userClaims map[string]interface{}) (string, time.Time, error) {
    secret := viper.GetString("jwt.secret")
    expiresIn := viper.GetDuration("jwt.access_token_expiration") // 24h
    expiresAt := time.Now().Add(expiresIn)
    
    claims := CustomClaims{
        UserID:   userID,
        Username: userClaims["username"].(string),
        Email:    userClaims["email"].(string),
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expiresAt),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    viper.GetString("jwt.issuer"), // "apprun-platform"
            Audience:  []string{viper.GetString("jwt.audience")}, // "apprun-api"
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(secret))
    return tokenString, expiresAt, err
}

func ValidateToken(tokenString string) (*CustomClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
        return []byte(viper.GetString("jwt.secret")), nil
    })
    if err != nil || !token.Valid {
        return nil, err
    }
    return token.Claims.(*CustomClaims), nil
}
```

**Configuration** (core/config/default.yaml):
```yaml
jwt:
  secret: "${JWT_SECRET}"  # MUST be set via environment variable in production
  access_token_expiration: 24h
  refresh_token_expiration: 168h  # 7 days (for future refresh token story)
  issuer: "apprun-platform"
  audience: "apprun-api"
```

#### 2. Login Handler (modules/auth/handlers/login.go)

**Ent ORM Query Pattern** - Query user by username OR email:
```go
user, err := client.User.Query().
    Where(
        user.Or(
            user.UsernameEQ(req.Identifier),
            user.EmailEQ(req.Identifier),
        ),
    ).
    Only(ctx)
if ent.IsNotFound(err) {
    return response.Error(w, errors.Unauthorized("AUTH_INVALID_CREDENTIALS", "Invalid username or password"))
}
```

**Password Verification** - Use Story 5.1 password package:
```go
import "apprun/core/internal/password"

if err := password.VerifyPassword(req.Password, user.PasswordHash); err != nil {
    logger.Warn("Login failed: invalid password", "user_id", user.ID, "trace_id", traceID)
    return response.Error(w, errors.Unauthorized("AUTH_INVALID_CREDENTIALS", "Invalid username or password"))
}
```

**Login History Update** - Non-blocking, don't fail login on error:
```go
go func() {
    updateErr := client.User.UpdateOneID(user.ID).
        SetLastLoginAt(time.Now()).
        SetLastLoginIP(getClientIP(r)).
        Exec(context.Background())
    if updateErr != nil {
        logger.Error("Failed to update login history", "user_id", user.ID, "error", updateErr)
    }
}()
```

#### 3. /me Handler (modules/auth/handlers/me.go)

**Extract User ID from JWT Context** - Pattern for getting authenticated user:
```go
func Me(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value("user_id").(int64)
    if !ok {
        return response.Error(w, errors.Unauthorized("AUTH_TOKEN_INVALID", "Invalid token claims"))
    }
    
    user, err := client.User.Get(r.Context(), userID)
    if ent.IsNotFound(err) {
        return response.Error(w, errors.NotFound("USER_NOT_FOUND", "User account not found"))
    }
    
    response.Success(w, sanitizeUser(user))
}
```

#### 4. Error Code Definitions

**Add to pkg/errors or modules/auth/errors.go**:
```go
const (
    AUTH_INVALID_CREDENTIALS = "AUTH_INVALID_CREDENTIALS"  // Generic login failure
    AUTH_ACCOUNT_DISABLED    = "AUTH_ACCOUNT_DISABLED"     // User status != 1
    AUTH_TOKEN_MISSING       = "AUTH_TOKEN_MISSING"        // No Authorization header
    AUTH_TOKEN_INVALID       = "AUTH_TOKEN_INVALID"        // Invalid/expired token
    USER_NOT_FOUND           = "USER_NOT_FOUND"            // User deleted after login
)
```

### Learning from Story 5.1

**Reuse these exact patterns from Story 5.1** (refer to [5-1-user-registration.md](./5-1-user-registration.md)):

1. **Password Package** (`core/internal/password/hash.go`):
   - `func VerifyPassword(plainPassword, hashedPassword string) error`
   - Already implements constant-time comparison
   - Already has bcrypt cost=12

2. **User Entity Fields** (from Ent schema in 5.1):
   - `id` (int64, primary key)
   - `uuid` (string, for external API reference) - **Added in Story 5.1**
   - `username` (string, optional, unique)
   - `email` (string, required, unique)
   - `password_hash` (string, sensitive field)
   - `status` (int8, 1=active, 0=disabled)
   - `last_login_at` (time, nullable)
   - `last_login_ip` (string, max 45 chars for IPv6)

3. **Error Handling Pattern** (from Story 5.1):
   - Use `pkg/errors` package with error codes and context
   - Return generic messages for security-sensitive errors
   - Log detailed errors server-side with trace ID

4. **Logging Pattern** (from Story 5.1):
   - Use `pkg/logger` with structured fields
   - Filter sensitive fields (passwords never logged)
   - Include trace ID in all log entries
   - Log at appropriate levels: Info (success), Warn (auth failure), Error (system error)

5. **Response Format** (from Story 5.1):
   - Use `pkg/response.Success()` and `pkg/response.Error()`
   - Always include timestamp
   - Sanitize user objects (exclude password_hash)

### Architecture & Security Constraints

**From [Epic 5: Authentication & Authorization](../../epics/5-auth-epic.md)**:

- **Technology Stack**:
  - bcrypt: golang.org/x/crypto/bcrypt (from Story 5.1)
  - JWT: github.com/golang-jwt/jwt/v5 v5.2.0 (NEW)
  - Chi Router: HTTP routing with middleware chain
  - Ent ORM: User queries and updates

- **Security Requirements**:
  - Password validation failures MUST NOT reveal if user exists (prevents user enumeration attacks)
  - Login attempts logged without exposing passwords or hashes
  - JWT signed with HMAC-SHA256 (HS256)
  - Token expiration strictly enforced (validate `exp` claim)
  - HTTPS only in production
  - Never store sensitive data in JWT payload (Base64 encoded, not encrypted)

- **Performance Notes**:
  - ⚡ bcrypt.CompareHashAndPassword is CPU-intensive (cost=12)
  - Consider caching failed login attempts in Redis to prevent brute force without repeatedly computing expensive hashes
  - Login history update is non-blocking (goroutine) to avoid impacting login latency
  - Use database connection pooling for concurrent logins

### JWT Security Best Practices

1. **Token Storage**:
   - Client should store token in memory or httpOnly cookie (never localStorage for XSS protection)
   - Tokens are stateless - cannot be revoked without additional infrastructure

2. **Claims Validation**:
   - Always validate `exp` (expiration) claim
   - Validate `iss` (issuer) and `aud` (audience) claims
   - Consider adding `jti` (JWT ID) claim for future token revocation support

3. **Secret Management**:
   - JWT secret MUST be at least 256 bits (32 characters)
   - MUST be set via environment variable `${JWT_SECRET}`
   - Never commit secrets to Git
   - Rotate secrets periodically (requires token invalidation strategy)

4. **Payload Recommendations**:
   - Keep payload minimal (user_id, username, email only)
   - Never include password, password_hash, or sensitive PII
   - JWT payload is Base64 encoded, NOT encrypted (anyone can decode it)

### Source Tree Components

**New files to create**:
```
core/internal/jwt/
├── token.go          # JWT generation and validation (185 lines estimated)
├── claims.go         # CustomClaims struct (25 lines)
└── token_test.go     # Unit tests (150 lines, 4 test functions)

modules/auth/handlers/
├── login.go          # Login HTTP handler (120 lines)
├── login_test.go     # Handler tests (200 lines, 6 test functions)
├── me.go             # Get current user handler (60 lines)
└── me_test.go        # Me handler tests (100 lines, 4 test functions)
```

**Files to modify**:
```
core/routes/router.go      # +5 lines (register 2 routes)
core/config/default.yaml   # +6 lines (JWT config)
go.mod                     # +1 line (jwt/v5 dependency)
docs/swagger.yaml          # +80 lines (2 endpoint specs)
```

**Files to reference from Story 5.1**:
```
core/internal/password/hash.go       # VerifyPassword(plain, hash) function
core/ent/user.go                     # User entity with UUID field
core/ent/user/where.go               # Ent query predicates (Or, UsernameEQ, EmailEQ)
pkg/response/response.go             # Success() and Error() functions
pkg/logger/logger.go                 # Structured logging with sensitive filtering
pkg/errors/errors.go                 # Error types with codes
```

### Testing Standards

**Coverage Target**: ≥ 80% for all new code

**Critical Test Scenarios**:
- JWT: generation, validation, expiration, invalid signature
- Login: valid user/email, not found, wrong password, disabled, missing fields
- /me: valid token, missing, invalid, user deleted
- Integration: full flow with real database and JWT

**Performance Benchmarks**:
```bash
go test -bench=BenchmarkLogin -benchmem    # Target: < 500ms
go test -bench=BenchmarkMe -benchmem       # Target: < 100ms
```

### References

- [Source: docs/epics/5-auth-epic.md#2-技术规范] - Auth architecture, API contracts
- [Source: docs/prd.md#FR-AUTH-001] - Functional requirements
- [Source: docs/sprint-artifacts/sprint-1/5-1-user-registration.md] - User schema, password patterns, error handling
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [OWASP JWT Security](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [jwt/v5 Go Package](https://pkg.go.dev/github.com/golang-jwt/jwt/v5)

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
