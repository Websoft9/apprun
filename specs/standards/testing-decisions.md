# Testing Decisions
# apprun BaaS Platform

**Created**: 2026-01-20  
**Status**: Active  
**Authority**: Architecture Team

---

## Test Directory Structure

### Unit Tests
- **Location**: Source directory (`*_test.go`)
- **Example**: `pkg/auth/auth_test.go`
- **Rationale**: Follow Go conventions, enable `go test ./...`

### Integration Tests
- **White-box**: Module directory (`package same_name`)
  - **Rationale**: Access private code, test internal logic
- **Black-box**: `tests/integration/` (`package *_test`)
  - **Structure**: Organized by component (`api/`, `db/`)
  - **Rationale**: Centralized management, CI/CD friendly

### E2E Tests
- **Location**: `tests/e2e/scenarios/`
- **Rationale**: Separate long-running tests from fast tests

### Performance Tests
- **Location**: `tests/performance/`
- **Tool**: k6 scripts (`.js`)
- **Rationale**: Non-blocking, scheduled execution

### Test Utilities
- **Location**: `tests/testutils/`, `tests/fixtures/`
- **Rationale**: Shared across all test types

---

## Test Execution Strategy

### CI/CD Pipeline
- **PR checks**: Unit + Integration (< 10 min)
- **Nightly**: E2E + Performance (< 60 min)
- **On-demand**: Security scans (gosec, govulncheck)

### Local Development
- **Default**: `make test` (unit + integration)
- **Fast**: `go test -short ./...` (unit only)
- **Full**: `make test-e2e-auto` (all tests)

---

## Test Coverage Targets

### By Module
- **Auth**: Unit ≥90%, Integration ≥80%
- **RBAC**: Unit ≥90%, Integration ≥90%
- **Data**: Unit ≥80%, Integration ≥70%
- **Config**: Unit ≥85%, Integration ≥60%
- **Overall**: Unit ≥80%, Integration ≥70%

### Coverage Enforcement
- **Gate**: PR blocked if coverage drops >2%
- **Reporting**: Codecov integration mandatory

---

## Test Data Management

### Database
- **Isolation**: Each test uses separate transaction or fresh DB
- **Cleanup**: Automatic via `t.Cleanup()` or factory methods
- **Seed Data**: Use factories (`fixtures/user_factory.go`)

### Test Database
- **Name**: `apprun_test`
- **Connection**: `TEST_DATABASE_URL` environment variable
- **Reset**: Truncate tables, not DROP/CREATE

---

## Testing Tools

### Required
- **Assertions**: `github.com/stretchr/testify`
- **HTTP Testing**: `net/http/httptest`
- **Database**: PostgreSQL 14+ (real, not mock)

### Optional
- **Containers**: Testcontainers (for isolated DB)
- **Performance**: k6
- **Security**: gosec, govulncheck

### Forbidden
- **In-memory DB**: SQLite not allowed (PostgreSQL-specific features)
- **Global state**: No shared mutable state between tests

---

## Naming Conventions

### Test Functions
- Pattern: `Test<Function><Scenario>`
- Example: `TestUserRegistration_DuplicateEmail_Returns409`

### Test Files
- Unit: `<package>_test.go`
- Integration: `<component>_integration_test.go`
- E2E: `<scenario>_e2e_test.go`

---

## Quality Gates

### Pass Criteria
- ✅ P0 tests: 100% pass
- ✅ P1 tests: ≥95% pass
- ✅ Unit coverage: ≥80%
- ✅ No high/critical security issues

### Fail Criteria
- ❌ Any P0 test failure
- ❌ Coverage < 75%
- ❌ Critical security vulnerability
- ❌ Performance regression > 20%

---

**Last Updated**: 2026-01-20  
**Approved By**: Websoft9
