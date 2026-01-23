# Story 5.9.1: Audit Integration (审计集成)
# Sprint 2: 认证与授权 (Auth Epic)

**Priority**: P1 (High Priority)  
**Effort**: 2 days  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 5.9 (审计日志中间件) - ✅ Completed (Audit module infrastructure exists)
- Story 5.5.2 (RBAC API Endpoints) - ✅ Completed
- Story 5.6 (User Self-Service) - ✅ Completed
- Story 5.7 (Platform User Management) - ✅ Completed

**Status**: not-started  
**Module**: Audit / Integration  
**Epic**: [auth-epic](../../epics/5-auth-epic.md)  
**Issue**: #TBD  

---

## User Story

As a **platform security auditor and developer**, I want all critical modules to be integrated with the audit logging system, so that:
- All security-sensitive operations are automatically captured without manual intervention
- Business-critical operations are logged with rich contextual information
- Audit trails are consistent, searchable, and meet compliance requirements
- Integration patterns are standardized and easy to implement

**Background**: Story 5.9 has implemented the audit logging infrastructure (AuditLog entity, middleware, service layer). This story focuses on systematically integrating audit logging across all existing modules to ensure comprehensive coverage.

---

## Acceptance Criteria

### Functional Requirements
- [ ] Auth module integration (Login, Register, Password Reset, Logout)
- [ ] RBAC module integration (Member add/remove, Role changes, Permission grants)
- [ ] Platform Admin integration (User status changes, User deletion, Role assignments)
- [ ] User Self-Service integration (Profile updates, Password changes) - Story 5.6
- [ ] All audit logs include: timestamp, operator_id, action, target_type, target_id, changes, IP, user_agent
- [ ] Failed operations are also logged with error details

### Module Coverage Requirements
- [ ] **Auth Handler**: Login success/failure, register, logout, password reset, token refresh
- [ ] **RBAC Handler**: Member add/remove, role update, permission check failures (403)
- [ ] **Platform Admin Handler**: User status update, user deletion, admin role assignment
- [ ] **User Self-Service Handler** (Story 5.6): Profile update, password change

### Non-Functional Requirements
- [ ] No performance degradation (< 5ms overhead per request)
- [ ] Backward compatible (no breaking changes to existing APIs)
- [ ] Consistent audit log format across all modules
- [ ] Clear documentation for future module integrations

### Testing Requirements
- [ ] Integration tests verify audit logs are created for all critical operations
- [ ] Test both successful and failed operation logging
- [ ] Verify audit log content includes all required fields
- [ ] Performance tests ensure no significant latency increase

---

## Technical Design

### Integration Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Audit Infrastructure (Story 5.9)          │
│  - AuditLog Ent Entity                                       │
│  - AuditService (Log, LogAction, Query)                      │
│  - AuditMiddleware (HTTP request auto-capture)              │
│  - DatabaseStorage + FileStorage                             │
└─────────────────────────────────────────────────────────────┘
                            ▲
                            │ Integration
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
        ▼                                       ▼
┌─────────────────┐                  ┌─────────────────┐
│  Auto-Capture   │                  │ Manual Logging  │
│  (Middleware)   │                  │ (Service Call)  │
└─────────────────┘                  └─────────────────┘
        │                                       │
        ▼                                       ▼
┌─────────────────┐                  ┌─────────────────┐
│ Generic HTTP    │                  │ Business-Rich   │
│ Logs (Basic)    │                  │ Logs (Detailed) │
└─────────────────┘                  └─────────────────┘
```

### Module Integration Matrix

| Module | Handler/Endpoint | Integration Method | Audit Action | Priority |
|--------|-----------------|-------------------|--------------|----------|
| **Auth** | `POST /api/auth/register` | Manual | `auth.register` | P0 |
| Auth | `POST /api/auth/login` | Manual | `auth.login.success` / `auth.login.failed` | P0 |
| Auth | `POST /api/auth/logout` | Manual | `auth.logout` | P1 |
| Auth | `POST /api/auth/refresh` | Manual | `auth.token.refresh` | P1 |
| Auth | `POST /api/auth/password-reset` | Manual | `auth.password.reset` | P1 |
| **RBAC** | `POST /api/projects/{id}/members` | Manual | `rbac.member.add` | P0 |
| RBAC | `DELETE /api/projects/{id}/members/{user_id}` | Manual | `rbac.member.remove` | P0 |
| RBAC | `PUT /api/projects/{id}/members/{user_id}/role` | Manual | `rbac.member.role_update` | P0 |
| RBAC | Permission check failures (403) | Middleware | `rbac.permission.denied` | P1 |
| **Platform Admin** | `POST /api/admin/users` | Manual | `admin.user.create` | P0 |
| Platform Admin | `PUT /api/admin/users/:id/status` | Manual | `admin.user.status_update` | P0 |
| Platform Admin | `PUT /api/admin/users/:id/role` | Manual | `admin.user.role_update` | P0 |
| Platform Admin | `DELETE /api/admin/users/:id` | Manual | `admin.user.delete` | P0 |
| **User Self-Service** | `PUT /api/profile` | Manual | `user.profile.update` | P1 |
| User Self-Service | `PUT /api/profile/password` | Manual | `user.password.change` | P1 |

---

## Integration Patterns

### Pattern 1: Manual Logging (Recommended for Business Operations)

**Use Case**: Operations with rich business context that need detailed tracking

**Implementation Example** (Auth Handler - Login):

```go
// Before Integration (Story 5.2 - existing code)
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // ... password verification logic ...
    
    if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(req.Password)); err != nil {
        logger.Warn("Login failed: invalid password",
            logger.Field{Key: "email", Value: req.Email},
            logger.Field{Key: "ip", Value: r.RemoteAddr},
        )
        response.AppError(w, errors.Unauthorized("invalid_credentials", "Invalid email or password"))
        return
    }
    
    // Generate JWT tokens...
    logger.Info("User logged in successfully",
        logger.Field{Key: "user_id", Value: user.ID},
        logger.Field{Key: "email", Value: user.Email},
    )
    
    response.Success(w, loginResponse)
}
```

**After Integration** (with Audit Service):

```go
import (
    "apprun/modules/audit/service"
    "apprun/modules/audit/storage"
)

type AuthHandler struct {
    authService  *service.AuthService
    auditService *service.AuditService  // ✨ NEW
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // ... password verification logic ...
    
    // ❌ Failed login
    if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(req.Password)); err != nil {
        // Log failed login attempt
        h.auditService.LogAction(ctx, service.LogActionParams{
            Action:     "auth.login.failed",
            TargetID:   user.ID.String(),  // Still log user ID even for failed attempts
            TargetType: "user",
            Changes: map[string]interface{}{
                "email":  req.Email,
                "reason": "invalid_password",
            },
        })
        
        logger.Warn("Login failed: invalid password",
            logger.Field{Key: "email", Value: req.Email},
        )
        response.AppError(w, errors.Unauthorized("invalid_credentials", "Invalid email or password"))
        return
    }
    
    // ✅ Successful login
    // Generate JWT tokens...
    
    // Log successful login
    h.auditService.LogAction(ctx, service.LogActionParams{
        Action:     "auth.login.success",
        TargetID:   user.ID.String(),
        TargetType: "user",
        Changes: map[string]interface{}{
            "email":      user.Email,
            "login_time": time.Now(),
        },
    })
    
    logger.Info("User logged in successfully",
        logger.Field{Key: "user_id", Value: user.ID},
    )
    
    response.Success(w, loginResponse)
}
```

**Benefits**:
- ✅ Rich business context (email, reason for failure)
- ✅ Distinguishes success vs failure
- ✅ Captures intent (login attempt) even when failed
- ✅ Easy to search and analyze

---

### Pattern 2: Middleware Auto-Capture (For Generic HTTP Logging)

**Use Case**: Automatic capture of all HTTP operations without code changes

**Configuration** (Already implemented in Story 5.9):

```go
// routes/routes.go (Apply middleware to protected routes)
router.Group(func(r chi.Router) {
    r.Use(middleware.JWTAuth)                    // Story 5.3
    r.Use(middleware.AuditMiddleware(auditSvc))  // Story 5.9
    
    // All routes here are auto-audited
    r.Post("/projects/{id}/members", handler.AddMember)
    r.Delete("/projects/{id}/members/{user_id}", handler.RemoveMember)
})
```

**What Gets Logged** (Automatically):
```json
{
  "timestamp": "2026-01-23T10:30:45Z",
  "operator_id": "user-uuid",
  "action": "POST /api/projects/123/members",
  "target_id": "123",
  "target_type": "http_request",
  "status_code": 201,
  "response_time_ms": 45,
  "method": "POST",
  "path": "/api/projects/123/members",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0..."
}
```

**Benefits**:
- ✅ Zero code changes required
- ✅ Comprehensive coverage
- ✅ Performance metrics included

**Limitations**:
- ❌ Generic context only (no business details)
- ❌ Cannot capture operation intent

---

### Pattern 3: Hybrid (Manual + Middleware) - Best Practice

**Use Case**: Combine automatic HTTP logging with manual business-level logging

**Example** (RBAC - Add Member):

```go
func (h *ProjectMemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    projectID := chi.URLParam(r, "id")
    
    // Business logic...
    member, err := h.memberService.AddMember(ctx, projectID, req.UserID, req.Role)
    if err != nil {
        // Manual audit for business-level failure
        h.auditService.LogAction(ctx, service.LogActionParams{
            Action:     "rbac.member.add.failed",
            TargetID:   projectID,
            TargetType: "project",
            Changes: map[string]interface{}{
                "target_user_id": req.UserID,
                "target_role":    req.Role,
                "error":          err.Error(),
            },
        })
        response.AppError(w, err)
        return
    }
    
    // Manual audit for business-level success
    h.auditService.LogAction(ctx, service.LogActionParams{
        Action:     "rbac.member.add.success",
        TargetID:   projectID,
        TargetType: "project",
        Changes: map[string]interface{}{
            "target_user_id": req.UserID,
            "target_role":    req.Role,
            "member_id":      member.ID.String(),
            "operator_name":  GetOperatorName(ctx), // Optional: enrich context
        },
    })
    
    response.Created(w, member)
}
// Middleware will also log generic HTTP info (method, path, status_code, duration)
```

**Result**: Two audit logs created
1. **HTTP-level log** (middleware): Generic request info
2. **Business-level log** (manual): Rich business context

**Benefits**:
- ✅ Complete audit trail (HTTP + business)
- ✅ Correlation via `request_id`
- ✅ Best of both worlds

---

## Implementation Tasks

**Implementation Order** (按优先级):
1. **Task 1** (Preparation) → **Task 2** (Auth) → **Task 3** (RBAC) - Core security modules (P0)
2. **Task 4** (Platform Admin) - Administrative operations (P0)
3. **Task 5** (User Self-Service) - User operations (P1)
4. **Task 6** (Documentation) - Final polish

---

### Task 1: Prepare Audit Service Dependency Injection
**Effort**: 0.5 days  
**Files**: 
- `core/cmd/serve.go` - Initialize AuditService and inject into handlers
- `core/modules/auth/handler/*.go` - Update constructors to accept AuditService
- `core/modules/admin/handler/*.go` - Update constructors to accept AuditService

**Subtasks**:
- [ ] Initialize AuditService in `serve.go` (after DB connection)
- [ ] Pass AuditService to all handler constructors
- [ ] Update handler struct definitions to include `auditService *service.AuditService`
- [ ] Verify no compilation errors

**Acceptance Criteria**:
- All handlers have access to AuditService
- No breaking changes to existing APIs

---

### Task 2: Integrate Auth Module
**Effort**: 0.5 days  
**Files**: 
- `core/modules/auth/handler/register.go`
- `core/modules/auth/handler/login.go`
- `core/modules/auth/handler/logout.go` (if exists)
- `core/modules/auth/handler/refresh.go` (if exists)

**Subtasks**:
- [ ] Login success: Log `auth.login.success` with user_id, email, timestamp
- [ ] Login failure: Log `auth.login.failed` with email, reason, IP
- [ ] Registration: Log `auth.register` with user_id, email
- [ ] Token refresh: Log `auth.token.refresh` with user_id
- [ ] Password reset: Log `auth.password.reset` with user_id
- [ ] Replace existing `logger.Info()` audit logs with `auditService.LogAction()`
- [ ] Add integration tests for audit log creation

**Audit Actions**:
- `auth.register` - User registration
- `auth.login.success` - Successful login
- `auth.login.failed` - Failed login attempt
- `auth.logout` - User logout
- `auth.token.refresh` - Token refresh
- `auth.password.reset` - Password reset

**Acceptance Criteria**:
- All auth operations create audit logs
- Failed operations are also logged
- Audit logs include email (but not password)

---

### Task 3: Integrate RBAC Module
**Effort**: 0.5 days  
**Files**: 
- `core/modules/auth/handler/project_member.go`

**Subtasks**:
- [ ] AddMember: Log `rbac.member.add` with project_id, target_user_id, role
- [ ] RemoveMember: Log `rbac.member.remove` with project_id, target_user_id
- [ ] UpdateRole: Log `rbac.member.role_update` with project_id, target_user_id, old_role, new_role
- [ ] Permission denied (403): Middleware already logs generic HTTP, but add business context if needed
- [ ] Replace existing manual audit logs (currently using `logger.Info()`)
- [ ] Add integration tests

**Audit Actions**:
- `rbac.member.add` - Member added to project
- `rbac.member.remove` - Member removed from project
- `rbac.member.role_update` - Member role changed
- `rbac.permission.denied` - Permission check failed (403)

**Acceptance Criteria**:
- All member management operations create audit logs
- Audit logs include operator_id, target_user_id, role changes

---

### Task 4: Integrate Platform Admin Module
**Effort**: 0.5 days  
**Files**: 
- `core/modules/admin/handler/user_management.go` (or similar)

**Subtasks**:
- [ ] Create User: Log `admin.user.create` with target_user_id, email, role
- [ ] Update User Status: Log `admin.user.status_update` with target_user_id, old_status, new_status
- [ ] Update User Role: Log `admin.user.role_update` with target_user_id, old_role, new_role
- [ ] Delete User: Log `admin.user.delete` with target_user_id, email
- [ ] Add integration tests

**Audit Actions**:
- `admin.user.create` - Admin created new user
- `admin.user.status_update` - Admin changed user status
- `admin.user.role_update` - Admin changed user role
- `admin.user.delete` - Admin deleted user

**Acceptance Criteria**:
- All admin operations create audit logs
- Audit logs include both operator (admin) and target (affected user)

---

### Task 5: Integrate User Self-Service Module
**Effort**: 0.25 days  
**Files**: 
- `core/modules/auth/handler/profile.go` (Story 5.6)
- `core/modules/auth/handler/password.go` (Story 5.6, if separate file)

**Subtasks**:
- [ ] Profile Update: Log `user.profile.update` with user_id, changed_fields
- [ ] Password Change: Log `user.password.change` with user_id, timestamp (DO NOT log old/new passwords)
- [ ] Add integration tests

**Audit Actions**:
- `user.profile.update` - User updated their profile
- `user.password.change` - User changed their password

**Acceptance Criteria**:
- Self-service operations create audit logs
- Passwords are NEVER logged

---

### Task 6: Documentation & Testing
**Effort**: 0.25 days  
**Files**: 
- `docs/guides/audit-integration.md` - Integration guide
- `core/modules/audit/README.md` - Module documentation
- `core/modules/auth/handler/*_test.go` - Integration tests

**Subtasks**:
- [ ] Document all audit action constants
- [ ] Document integration patterns (manual vs middleware vs hybrid)
- [ ] Document required fields for each action type
- [ ] Create integration test examples
- [ ] Update module READMEs with audit integration notes

**Acceptance Criteria**:
- Clear documentation for future module integrations
- All new audit integrations have test coverage

---

## Audit Action Constants

### Auth Module Actions
```go
const (
    ActionAuthRegister       = "auth.register"
    ActionAuthLoginSuccess   = "auth.login.success"
    ActionAuthLoginFailed    = "auth.login.failed"
    ActionAuthLogout         = "auth.logout"
    ActionAuthTokenRefresh   = "auth.token.refresh"
    ActionAuthPasswordReset  = "auth.password.reset"
)
```

### RBAC Module Actions
```go
const (
    ActionRBACMemberAdd          = "rbac.member.add"
    ActionRBACMemberRemove       = "rbac.member.remove"
    ActionRBACMemberRoleUpdate   = "rbac.member.role_update"
    ActionRBACPermissionDenied   = "rbac.permission.denied"
)
```

### Platform Admin Module Actions
```go
const (
    ActionAdminUserCreate       = "admin.user.create"
    ActionAdminUserStatusUpdate = "admin.user.status_update"
    ActionAdminUserRoleUpdate   = "admin.user.role_update"
    ActionAdminUserDelete       = "admin.user.delete"
)
```

### User Self-Service Module Actions
```go
const (
    ActionUserProfileUpdate  = "user.profile.update"
    ActionUserPasswordChange = "user.password.change"
)
```

---

## Testing Strategy

### Unit Tests
- Verify AuditService is called with correct parameters
- Mock AuditService to test handler logic independently
- Test both success and failure scenarios

### Integration Tests
- Verify audit logs are created in database
- Verify audit log content includes all required fields
- Verify correlation with HTTP logs (request_id)
- Test failed operations also create audit logs

### Example Integration Test
```go
func TestLogin_CreatesAuditLog(t *testing.T) {
    // Setup
    entClient := setupTestDB(t)
    auditSvc := service.NewAuditService(storage.NewDatabaseStorage(entClient), service.Config{Enabled: true})
    authHandler := handler.NewAuthHandler(authSvc, auditSvc)
    
    // Execute login
    req := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginPayload)
    rec := httptest.NewRecorder()
    authHandler.Login(rec, req)
    
    // Verify audit log created
    auditLogs, err := entClient.AuditLog.Query().
        Where(auditlog.Action("auth.login.success")).
        All(context.Background())
    
    require.NoError(t, err)
    assert.Len(t, auditLogs, 1)
    assert.Equal(t, user.ID, auditLogs[0].OperatorID)
    assert.Equal(t, "user", auditLogs[0].TargetType)
}
```

---

## Migration Plan

### Phase 1: Preparation (0.5 days)
- Initialize AuditService in `serve.go`
- Inject AuditService into all handlers
- Verify no breaking changes

### Phase 2: Module Integration (1 day)
- Integrate Auth module (0.5 days)
- Integrate RBAC module (0.25 days)
- Integrate Platform Admin module (0.25 days)

### Phase 3: Testing & Documentation (0.5 days)
- Write integration tests for all modules
- Document audit action constants
- Create integration guide

---

## Rollback Plan

If issues are discovered:
1. Audit logging is non-blocking (async), so failures don't impact business logic
2. Remove `auditService.LogAction()` calls without breaking APIs
3. Middleware can be disabled via configuration flag

**Risk Mitigation**:
- All audit logging is wrapped in error handling (failures logged but don't crash)
- Feature flag for gradual rollout (`audit.enabled` config)
- Comprehensive integration tests before deployment

---

## Success Metrics

### Technical Metrics
- ✅ 100% of critical operations have audit logs
- ✅ Audit logging overhead < 5ms (p95)
- ✅ Zero failed operations due to audit logging issues
- ✅ Test coverage > 85% for audit integration

### Business Metrics
- ✅ All security events are traceable
- ✅ Audit logs meet compliance requirements (SOC2, ISO27001)
- ✅ < 1 hour to investigate security incidents
- ✅ Complete audit trail for all admin operations

---

## References

### Internal Documentation
- [Story 5.9: 审计日志中间件](./5-9-audit-logging.md) - Audit infrastructure
- [Story 5.5.2: RBAC API Endpoints](../sprint-1/5-5-2-rbac-api-endpoints.md) - RBAC integration points
- [Story 5.6: User Self-Service](./5-6-user-self-service.md) - User profile integration points
- [Story 5.7: Platform User Management](./5-7-platform-user-management.md) - Admin integration points
- [Architecture Standards](../../standards/architecture-standards.md) - Design principles

### External Resources
- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)
- [Audit Log Best Practices](https://www.imperva.com/learn/data-security/audit-log/)

---

## Dev Notes

### Design Decisions

**Q: Why manual logging instead of just middleware?**  
A: Middleware provides generic HTTP logs (method, path, status_code) but lacks business context. Manual logging allows rich contextual information (e.g., "promoted user X to admin", "failed login due to invalid password").

**Q: Why not use Ent hooks for automatic audit logging?**  
A: Ent hooks are data-layer focused and don't have access to HTTP context (user_id, IP, user_agent). Handler-level integration captures both HTTP context and business intent.

**Q: How to handle high-volume operations?**  
A: Audit logging is asynchronous (Story 5.9 uses worker pool), so it doesn't block business logic. For extremely high-volume endpoints, use middleware auto-capture only (skip manual logging).

**Q: What if AuditService fails?**  
A: Audit logging failures are logged but don't crash the application. Business operations continue normally even if audit logging temporarily fails.

### Project Structure Alignment
- Follows existing handler patterns (dependency injection via constructor)
- Reuses Story 5.9 infrastructure (no new components)
- Consistent with error handling patterns (Story 1.3)
- Integrates with existing JWT context (Story 5.3)

### Testing Notes
- Use mock AuditService for unit tests (test handler logic independently)
- Use real AuditService + in-memory DB for integration tests
- Test both success and failure scenarios
- Verify audit log content matches expected format

---

## Completion Checklist

### Implementation Checklist
- [ ] All handlers have AuditService dependency
- [ ] Auth module integration completed and tested
- [ ] RBAC module integration completed and tested
- [ ] Platform Admin module integration completed and tested
- [ ] User Self-Service module integration completed and tested

### Testing Checklist
- [ ] Integration tests passing for all modules
- [ ] Unit tests with mocked AuditService passing
- [ ] Performance benchmarks meet targets (< 5ms overhead)
- [ ] Failed operations logging verified

### Documentation Checklist
- [ ] Documentation updated (action constants, integration guide)
- [ ] Code review approved
- [ ] Story marked as complete in sprint-status.yaml

### Verification Checklist
**Run these commands to verify integration**:
```bash
# Check audit logs for auth operations
make test FILTER=TestLogin_CreatesAuditLog

# Check audit logs for RBAC operations
make test FILTER=TestAddMember_CreatesAuditLog

# Check audit logs for admin operations
make test FILTER=TestUpdateUserStatus_CreatesAuditLog

# Performance benchmark
go test -bench=BenchmarkAuditIntegration -benchmem ./...
```

---

_Story created following BMM standards | Last updated: 2026-01-23_
