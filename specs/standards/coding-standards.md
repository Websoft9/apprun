# Coding Decisions
# apprun BaaS Platform

**Created**: 2025-12-25  
**Status**: Active  
**Authority**: Architecture Team

---

## Language & Tools

### Required
- Go 1.25+
- `gofmt`, `goimports` for formatting
- `golangci-lint` for static analysis
- English for all code and comments

### Standard Packages
- **Errors**: `apprun/pkg/errors`
- **Response**: `apprun/pkg/response`
- **Logger**: `apprun/pkg/logger`
- Other packages in `pkg` directory

---

## Project Structure

### Architecture Pattern
**Modular Monolith** - Vertical slice by business capability

### Directory Structure
```
core/
├── main.go                 # Application entry
├── cmd/                    # Cobra commands (flat)
├── modules/                # Business modules (vertical)
│   └── auth/
│       ├── handler.go
│       ├── service.go
│       ├── repository.go
│       ├── types.go
│       └── config.go
├── internal/               # Infrastructure
│   ├── bootstrap/         # App startup
│   ├── middleware/
│   └── jwt/
├── pkg/                    # Reusable utilities
└── ent/                    # Ent ORM schemas
```

### Key Decisions
- **main.go location**: `core/` root (avoid package conflicts)
- **cmd/ structure**: CLI subdirectory
- **Module organization**: Vertical slice per business domain
- **internal/**: Inner Infrastructure only, no business logic
- **pkg/**: Reusable libraries, zero business dependencies

---

## Constants Organization

**Decision Date**: 2026-01-12  
**Principle**: Business cohesion over file type separation

### Rule
- **All module constants in `config.go`**, not separate `constants.go`
- Exception: Only if module has 50+ constants or requires complex initialization

### Naming Patterns
| Type | Pattern | Example |
|------|---------|---------|
| Min/Max | `Min<Name>` / `Max<Name>` | `MinPasswordLength` |
| Default | `Default<Name>` | `DefaultTimeout` |
| Status | `Status<Name>` | `StatusActive` |
| Error | `Err<Name>` | `ErrInvalidEmail` |

### Applied Modules
- ✅ `pkg/jwt/config.go`
- ✅ `modules/auth/config.go`

---

## Module Layer Structure

### Standard Pattern
```
modules/<name>/
├── handler.go      # HTTP layer
├── service.go      # Business logic
├── repository.go   # Data access
├── types.go        # Domain models
└── config.go       # Module config + constants
```

### Responsibilities
- **handler**: HTTP request/response only, call service
- **service**: Business rules, orchestration
- **repository**: Database operations (Ent client)
- **types**: DTOs and domain models

---

## Error Handling

### Required Package
**Must use**: `apprun/pkg/errors`

### Key Methods
- `errors.New(code, msg)` - Create business error
- `errors.Wrap(err, code, msg)` - Wrap with context
- `errors.WithContext(key, value)` - Add metadata
- `httpmap.ToHTTPStatus(err)` - Map to HTTP status

### Forbidden
- ❌ `fmt.Errorf()` for business errors
- ❌ stdlib `errors.New()` for business errors

---

## Configuration Management

### Architecture
**Config Center Registry Pattern**

### Module Config Structure
Each module must have `config.go` with:
- `Config` struct (YAML-compatible types)
- `DefaultConfig()` function
- `ToRuntimeConfig()` method (parse/transform)
- `NewXXXFromConfig()` factory

### Required Tags
- `mapstructure:"field_name"` - mapstructure key (snake_case)
- `default:"value"` - Documentation
- `db:"true|false"` - DB storage permission
- `validate:"rules"` - Validation (optional)

### Priority (Low → High)
1. Struct tag defaults
2. `default.yaml`
3. Specialized files (`database.yaml`)
4. `conf_d/` directory
5. Database (if `db:"true"`)
6. Environment variables

### Environment Variable Mapping
No prefix: `section.field` → `SECTION_FIELD`

Example:
```
database.host     → DATABASE_HOST
server.http_port  → SERVER_HTTP_PORT
```

### DB Tag Rules
- **Infrastructure config**: `db:"false"` (server, database, logs)
- **Business config**: `db:"true"` (feature flags, API keys)
- **Environment-only**: `envonly:"true"` (AUTO_INIT, DEBUG_MODE)

---

## Testing

See [testing-decisions.md](testing-decisions.md) for complete testing strategy and standards.

---

## Ent ORM

### Field Definition
All fields must have explicit JSON tags (snake_case):

```go
field.String("user_name").
    StorageKey("user_name").
    StructTag(`json:"user_name"`)
```

### Sensitive Fields
Use `json:"-"` for passwords, tokens, API keys

---

## Docker

### Compose Commands
**V2 syntax only**: `docker compose up` (not `docker-compose`)

### Compose Files
- **No version field** (deprecated in V2)
- Start directly with `services:`

### File Naming
- `docker-compose.yml` - Production
- `docker-compose.dev.yml` - Development
- `docker-compose.local.yml` - Local testing

---

## Build System

### Makefile Location
**Must be** in project root (only one per project)

### Rationale
- User expectation
- CI/CD default path
- Centralized build commands

---

**Last Updated**: 2026-01-20  
**Approved By**: Websoft9
