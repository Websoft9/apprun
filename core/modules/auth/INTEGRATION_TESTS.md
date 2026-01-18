# Auth Module Integration Tests

## Overview
Comprehensive integration tests for Story 5.2 (User Login) with real database interactions.

## Test Environment
- **Database**: SQLite in-memory (via ent/enttest)
- **Configuration**: Viper with test settings
- **JWT**: Real token generation and validation
- **Password**: bcrypt hashing (cost 12)

## Test Coverage

### Login Integration Tests (6 tests)

1. **TestLoginIntegration_SuccessWithEmail**
   - Creates user in DB
   - Logs in with email identifier
   - Verifies JWT token generated
   - Validates token claims match user
   - **Result**: ✅ PASS (1.3s)

2. **TestLoginIntegration_SuccessWithUsername**
   - Creates user in DB
   - Logs in with username identifier
   - Verifies successful authentication
   - **Result**: ✅ PASS (0.99s)

3. **TestLoginIntegration_InvalidPassword**
   - Creates user with correct password
   - Attempts login with wrong password
   - Verifies generic error returned (no user existence leak)
   - **Result**: ✅ PASS (0.82s)

4. **TestLoginIntegration_UserNotFound**
   - Attempts login with non-existent user
   - Verifies generic error (don't reveal user doesn't exist)
   - **Result**: ✅ PASS (0.00s)

5. **TestLoginIntegration_DisabledAccount**
   - Creates user with status = 0 (disabled)
   - Attempts login
   - Verifies "account disabled" error
   - **Result**: ✅ PASS (0.78s)

6. **TestLoginIntegration_LoginHistoryTracking**
   - Logs in successfully
   - Waits for async update (100ms)
   - Verifies `last_login_at` and `last_login_ip` updated
   - **Result**: ✅ PASS (0.83s)

### User Profile Tests (2 tests)

7. **TestGetUserProfileIntegration**
   - Creates user
   - Retrieves profile by ID
   - Verifies all fields returned correctly
   - Ensures password_hash NOT included
   - **Result**: ✅ PASS (0.73s)

8. **TestGetUserProfileIntegration_UserNotFound**
   - Attempts to get profile for non-existent ID
   - Verifies "not found" error
   - **Result**: ✅ PASS (0.00s)

### JWT Tests (1 test)

9. **TestJWTTokenValidation**
   - Generates JWT token with custom claims
   - Validates token
   - Verifies claims extracted correctly
   - Checks expiration time (~24 hours)
   - **Result**: ✅ PASS (0.00s)

### Repository Tests (2 tests)

10. **TestRepositoryFindByIdentifier**
    - Creates multiple users
    - Finds by email (test OR condition)
    - Finds by username (test OR condition)
    - Verifies NotFound error for non-existent
    - **Result**: ✅ PASS (1.19s)

11. **TestRepositoryUpdateLoginHistory**
    - Updates user's login timestamp and IP
    - Verifies changes persisted
    - Checks timestamp within 2-second tolerance
    - **Result**: ✅ PASS (0.48s)

### End-to-End Test (1 test)

12. **TestIntegration_FullLoginFlow**
    - Complete workflow:
      1. Create user in DB
      2. Initialize full stack (repo → service → handler)
      3. Login with credentials
      4. Validate JWT token
      5. Get profile using token claims
      6. Verify login history updated
    - **Result**: ✅ PASS (0.93s)

## Test Execution

### Run All Integration Tests
```bash
cd /data/cdl/apprun/core/modules/auth
go test -v -run "Integration|Profile|JWT|Repository" -timeout 60s
```

### Run Only Login Tests
```bash
go test -v -run "TestLoginIntegration"
```

### Skip Integration Tests (Short Mode)
```bash
go test -short ./...
```

## Test Statistics

- **Total Tests**: 12 integration tests
- **Total Duration**: ~8.1 seconds
- **Pass Rate**: 100% ✅
- **Database Operations**: Real SQLite transactions
- **Test Isolation**: Each test uses fresh database

## Key Validations

### Security
- ✅ Generic error messages (don't leak user existence)
- ✅ Password hashing with bcrypt (cost 12)
- ✅ Password hash never returned in responses
- ✅ JWT token validation
- ✅ Account status checking (disabled accounts rejected)

### Functionality
- ✅ Login with email or username
- ✅ JWT token generation with correct claims
- ✅ Login history tracking (async)
- ✅ User profile retrieval
- ✅ Repository OR query (FindByIdentifier)

### Performance
- ✅ Async login history update (non-blocking)
- ✅ Fast tests (avg 0.7s per test)
- ✅ In-memory database for speed

## Dependencies

- `apprun/ent/enttest` - Test database setup
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/stretchr/testify` - Assertions
- `github.com/spf13/viper` - Configuration
- `apprun/internal/jwt` - JWT operations
- `apprun/internal/password` - Password hashing

## CI/CD Integration

These tests are designed to run in CI pipelines:
- No external dependencies (in-memory DB)
- Fast execution (~8 seconds)
- No cleanup required (in-memory)
- Exit code 0 on success

## Story 5.2 Requirements Coverage

| Acceptance Criteria | Test Coverage |
|-------------------|--------------|
| AC #1: POST /api/auth/login | ✅ 6 login tests |
| AC #2: GET /api/auth/me | ✅ 2 profile tests |
| AC #3: bcrypt verification | ✅ Implicit in all tests |
| AC #4: JWT generation | ✅ JWT validation test |
| AC #5: Login history | ✅ History tracking test |
| AC #6: Status validation | ✅ Disabled account test |
| AC #7: Security requirements | ✅ Generic errors tested |
| AC #8: Performance/quality | ✅ 100% pass rate |

## Future Enhancements

- [ ] Performance benchmarks (measure P95 latency)
- [ ] Concurrent login tests (race condition detection)
- [ ] JWT expiration tests (time manipulation)
- [ ] Rate limiting tests (prevent brute force)
- [ ] HTTP handler integration tests (with httptest)

