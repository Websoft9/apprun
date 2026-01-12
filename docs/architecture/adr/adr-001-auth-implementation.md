# ADR-001: Authentication Implementation Strategy

**Status**: Accepted  
**Date**: 2026-01-08  
**Decision Makers**: Architect Agent, BMad Master  
**Stakeholders**: Development Team, Product Team

---

## Context

The apprun BaaS platform requires a secure authentication system to support:
- User registration and login
- JWT-based API authentication
- Project-based multi-tenancy
- RBAC authorization (via Casbin)

Two primary approaches were considered:
1. **Ory Kratos**: External identity management service
2. **Go Native**: Built-in authentication using Go standard libraries

---

## Decision

**For MVP Phase (Next 3-6 months)**: Use **Go Native Authentication**

**For Production Phase (Post-MVP)**: Evaluate migration to **Enterprise Auth** (Ory Kratos or similar)

---

## Rationale

### Why Go Native for MVP?

#### 1. **Speed to Market**
- **Implementation time**: 7-8 days vs 3-4 weeks
- **Reduced complexity**: No external service dependencies
- **Faster iteration**: Direct code control enables rapid feature changes

#### 2. **Sufficient for MVP Requirements**
The PRD (FR-AUTH-001) only requires:
- ✅ User registration/login (Email + Password)
- ✅ JWT Token authentication
- ✅ Project-based RBAC
- ❌ MFA (not required for MVP)
- ❌ SSO/SAML (not required for MVP)
- ❌ Account recovery flows (nice-to-have)

#### 3. **Operational Simplicity**
- Single binary deployment
- No additional service orchestration
- Simpler POC/testing environment
- Reduced Docker Compose complexity

#### 4. **Technical Feasibility**
Go ecosystem provides battle-tested libraries:
- `bcrypt`: Industry-standard password hashing (OWASP recommended)
- `jwt/v5`: Mature JWT implementation
- `gorilla/sessions`: Secure session management
- `x/time/rate`: Rate limiting for brute-force protection

#### 5. **Team Capability**
- Team has Go expertise
- Security best practices well-documented
- Reference implementations available (GitHub, enterprise projects)

### Why NOT Ory Kratos for MVP?

#### 1. **Over-engineering for Current Needs**
- Kratos provides MFA, email verification, account recovery
- These features are not MVP requirements
- Adds 2-3 weeks of integration work

#### 2. **POC Complexity**
Current POC shows:
- Shared database schema management issues
- Docker service coordination overhead
- Dual authentication flow (Session → JWT) complexity

#### 3. **External Dependency Risk**
- Additional point of failure
- Upgrade/maintenance burden
- Learning curve for team (Identity Schemas, Kratos flows)

---

## Consequences

### Positive

✅ **Fast MVP delivery**: Ship auth in 7-8 days vs 3-4 weeks  
✅ **Simple architecture**: Single service, single codebase  
✅ **Full control**: Easy to customize for BaaS-specific needs  
✅ **Lower ops burden**: No external service to monitor  
✅ **Cost effective**: No Kratos infrastructure costs

### Negative

❌ **Missing advanced features**: No MFA, SSO, account recovery (acceptable for MVP)  
❌ **Security responsibility**: Team must maintain security patches  
❌ **Compliance burden**: GDPR/CCPA compliance is team responsibility  
❌ **Feature gap**: May need to rebuild features Kratos provides  
❌ **Migration cost**: Will require effort to migrate to enterprise auth later

### Mitigations

1. **Security Best Practices**
   - Use bcrypt with cost factor 12
   - Implement rate limiting (5 failed attempts → 15min lockout)
   - Enforce strong passwords (8+ chars, uppercase, lowercase, numbers)
   - Use HTTPS in production
   - Rotate JWT secrets regularly

2. **Code Quality**
   - Unit test coverage > 80%
   - Security audit of auth code
   - Follow OWASP guidelines
   - Document security assumptions

3. **Future Migration Path**
   - Design interface abstraction for AuthService
   - Keep JWT format standard-compliant
   - Document migration plan to Kratos
   - Budget 2 weeks for future migration

---

## Migration Plan (Post-MVP)

### When to Migrate?

Migrate to enterprise auth (Kratos or equivalent) when:
- ✅ Product achieves product-market fit
- ✅ Enterprise customers require MFA/SSO
- ✅ Team size > 10 developers (can support multi-service architecture)
- ✅ Revenue supports dedicated DevOps for service orchestration
- ✅ Compliance requirements (SOC2, ISO 27001) needed

### Migration Strategy

```go
// Phase 1: Interface Abstraction (MVP)
type AuthService interface {
    Register(ctx context.Context, email, password string) (*User, error)
    Login(ctx context.Context, email, password string) (string, error)
    ValidateToken(token string) (*Claims, error)
}

// Phase 2: Parallel Implementation
type NativeAuthService struct { ... }   // MVP implementation
type KratosAuthService struct { ... }   // Enterprise implementation

// Phase 3: Blue-Green Migration
// - Run both services in parallel
// - Migrate users batch-by-batch
// - Use feature flags for gradual rollout
// - Monitor error rates and rollback if needed
```

**Estimated Migration Effort**: 2 weeks
- 3 days: Kratos integration
- 2 days: User data migration
- 2 days: Parallel testing
- 3 days: Gradual rollout + monitoring

---

## Technical Specification

### Go Packages Selected

| Package | Version | Purpose | Rationale |
|---------|---------|---------|-----------|
| `golang.org/x/crypto/bcrypt` | latest | Password hashing | OWASP recommended, adjustable cost |
| `github.com/golang-jwt/jwt/v5` | v5 | JWT tokens | Most popular Go JWT library |
| `github.com/gorilla/sessions` | latest | Session management | Industry standard for Go web apps |
| `golang.org/x/time/rate` | latest | Rate limiting | Official Go rate limiter |

### Security Configuration

```yaml
auth:
  password:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_number: true
    bcrypt_cost: 12
  
  jwt:
    algorithm: HS256
    access_token_ttl: 3600      # 1 hour
    refresh_token_ttl: 604800   # 7 days
  
  rate_limit:
    login_attempts: 5
    lockout_duration: 900       # 15 minutes
```

---

## Alternatives Considered

### Alternative 1: Keep Ory Kratos (Rejected)

**Pros**: Enterprise-ready, MFA, SSO, compliance  
**Cons**: Over-engineered for MVP, 3-4 week implementation, operational complexity

**Verdict**: Right technology, wrong timing

### Alternative 2: Firebase Auth (Rejected)

**Pros**: Managed service, zero ops  
**Cons**: Vendor lock-in, cost scaling, no self-hosted option for BaaS platform

**Verdict**: Conflicts with BaaS self-hosted positioning

### Alternative 3: Auth0 SDK (Rejected)

**Pros**: Comprehensive features  
**Cons**: Same issues as Firebase (cost, vendor lock-in)

**Verdict**: Not suitable for open-source BaaS

### Alternative 4: Keycloak (Rejected)

**Pros**: Open-source, feature-complete  
**Cons**: Java-based (team is Go), heavy infrastructure, steep learning curve

**Verdict**: Too complex for MVP

---

## Success Criteria

### MVP Success (3 months)
- [ ] Auth implementation completed in < 8 days
- [ ] Unit test coverage > 80%
- [ ] Zero critical security vulnerabilities (gosec scan)
- [ ] API auth latency P95 < 10ms
- [ ] Successfully onboard first 100 users

### Post-MVP Review (6 months)
- [ ] No security incidents related to auth
- [ ] User feedback on auth UX is positive
- [ ] Decision point: Migrate to enterprise auth or continue native?

---

## References

- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
- [Go Crypto bcrypt Documentation](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [Ory Kratos Documentation](https://www.ory.sh/kratos/docs/) (for future reference)

---

## Related Documents

- [5-auth-epic.md](../epics/5-auth-epic.md) - Implementation epic
- [tech-architecture.md](./tech-architecture.md) - Technical architecture
- [prd.md](../prd.md#21-认证与权限) - Product requirements

---

**Document Owner**: Winston (Architect Agent)  
**Last Updated**: 2026-01-08  
**Next Review**: 2026-04-08 (3 months)
