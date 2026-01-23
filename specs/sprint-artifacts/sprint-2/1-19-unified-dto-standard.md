# Story 1-19: Unified API Module DTO Design and Implementation
# Sprint 2: Infrastructure Enhancement

**Priority**: P1 (High)  
**Effort**: 3-5 days  
**Owner**: Backend Dev  
**Dependencies**: None  
**Status**: ready-for-dev  
**Module**: Infrastructure  
**Related**: [Coding Standards - Module Layer Structure](../../standards/coding-standards.md#module-layer-structure)

---

## User Story

**As a** Developer,  
**I want** to implement standardized Data Transfer Objects (DTOs) for all API modules,  
**So that** the API layer is decoupled from domain models, providing type safety, explicit validation, and consistent response structures.

---

## Context

Currently, several modules mix domain models (Ent entities) with API contracts or use unstructured `map[string]interface{}` for request/response handling. This creates:
- **Coupling**: Direct exposure of database schema through API
- **Validation gaps**: No explicit validation rules at API boundary
- **Maintenance issues**: Schema changes break API contracts unexpectedly

This story enforces the DTO pattern decision recorded in `docs/standards/coding-standards.md` to establish a clean separation between API and domain layers.

---

## Target Modules

### Refactor Required
1. **`core/modules/auth`** - Currently uses Ent User entity directly in responses
2. **`core/modules/config`** - Mixed usage of domain models and maps

### New Implementation
3. **`core/modules/admin`** - No DTO structure exists
4. **`core/modules/audit`** - No DTO structure exists

---

## Acceptance Criteria

- [ ] All 4 target modules have `types.go` file with complete DTO definitions
- [ ] Request DTOs include `validate` tags (e.g., `validate:"required,email"`)
- [ ] Response DTOs include `json` tags with snake_case naming
- [ ] No Ent entities used directly in HTTP request body binding
- [ ] No Ent entities serialized directly in HTTP responses (must map to DTOs)
- [ ] All handlers refactored to use Request/Response DTOs
- [ ] Validation middleware correctly processes `validate` tags
- [ ] Unit tests updated to verify DTO behavior
- [ ] Integration tests validate end-to-end DTO flow
- [ ] Swagger/OpenAPI specs reflect DTO schemas accurately

---

## Implementation Tasks

### Phase 1: Auth Module Refactoring (Day 1-2)
- [ ] Create `core/modules/auth/types.go` if not exists
- [ ] Define DTOs:
  - `RegisterRequest` (email, password, name + validation tags)
  - `LoginRequest` (email, password + validation tags)
  - `UserResponse` (id, email, name, created_at - exclude password)
  - `TokenResponse` (access_token, refresh_token, expires_in)
- [ ] Refactor handlers:
  - `RegisterHandler`: Bind to `RegisterRequest`, return `UserResponse`
  - `LoginHandler`: Bind to `LoginRequest`, return `TokenResponse`
- [ ] Update tests

### Phase 2: Config Module Refactoring (Day 2-3)
- [ ] Create `core/modules/config/types.go` if not exists
- [ ] Define DTOs:
  - `CreateConfigRequest` (key, value, namespace + validation)
  - `UpdateConfigRequest` (value + validation)
  - `ConfigResponse` (id, key, value, namespace, updated_at)
  - `ConfigListResponse` (items []ConfigResponse, total int)
- [ ] Refactor handlers
- [ ] Update tests

### Phase 3: Admin Module Implementation (Day 3-4)
- [ ] Create `core/modules/admin/types.go`
- [ ] Define DTOs based on admin operations (user management, system settings)
- [ ] Implement handlers using DTOs
- [ ] Write tests

### Phase 4: Audit Module Implementation (Day 4-5)
- [ ] Create `core/modules/audit/types.go`
- [ ] Define DTOs:
  - `AuditLogResponse` (id, user_id, action, resource, timestamp, details)
  - `AuditQueryRequest` (filters, pagination)
- [ ] Implement handlers using DTOs
- [ ] Write tests

### Phase 5: Validation & Documentation (Day 5)
- [ ] Verify validation middleware integration
- [ ] Update Swagger annotations if needed
- [ ] Run integration tests across all modules
- [ ] Document DTO patterns in module READMEs

---

## Technical Details

### DTO Naming Conventions
```go
// Request DTOs
type Create<Resource>Request struct { ... }
type Update<Resource>Request struct { ... }
type <Resource>QueryRequest struct { ... }

// Response DTOs
type <Resource>Response struct { ... }
type <Resource>ListResponse struct {
    Items []<Resource>Response `json:"items"`
    Total int                  `json:"total"`
}
```

### Required Struct Tags
- **Request DTOs**: `json`, `validate`
- **Response DTOs**: `json` (snake_case)
- **Sensitive fields**: `json:"-"` (passwords, tokens)

### Example
```go
// types.go
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required,min=2"`
}

type UserResponse struct {
    ID        int       `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}
```

---

## Testing Strategy

### Unit Tests
- DTO validation behavior (valid/invalid inputs)
- Handler DTO binding logic
- Service-to-DTO mapping functions

### Integration Tests
- End-to-end API flow with DTOs
- Validation error responses (400 Bad Request)
- Response structure compliance

---

## Definition of Done

- [ ] Code merged to main branch
- [ ] All acceptance criteria met
- [ ] Unit test coverage ≥ 80%
- [ ] Integration tests passing
- [ ] Code review approved
- [ ] Documentation updated
- [ ] No golangci-lint warnings

---

## Notes

- **Migration Strategy**: Refactor one module at a time to minimize risk
- **Backward Compatibility**: Consider deprecation period if external clients exist
- **Performance**: DTO mapping adds minimal overhead (struct-to-struct copy)
- **Future Work**: Consider auto-generation tools for DTO boilerplate

---

## References

- [Coding Standards - Module Layer Structure](../../standards/coding-standards.md#module-layer-structure)
- [Coding Standards - Error Handling](../../standards/coding-standards.md#error-handling)
- [Testing Decisions](../../standards/testing-decisions.md)
