# Story 3.3: Module Configuration Registry Enhancement

## Sprint 0: Foundation (Configuration Epic)

**Priority**: P1 (High)  
**Effort**: 2 days  
**Owner**: Platform Dev  
**Dependencies**:
- Story 10 (Configuration Center Foundation) - Completed

**Status**: Done  
**Module**: Configuration  
**Epic**: [configuration-epic](../../epics/configuration-epic.md)  
**Issue**: #TBD  

---

## User Story

As a **platform developer**, I want a standardized way to manage module configurations so that all configurable modules follow the same pattern, can be automatically registered with the config center, and have their configuration templates generated automatically for easy setup and deployment.

---

## Acceptance Criteria

### Functional Acceptance
- [x] All modules with configuration must have a dedicated `config.go` file
- [x] Factory functions serve as "connectors" between modules and the config center
- [x] Ability to generate `config/config.example` file containing all registered module configurations
- [x] `config.example` includes default values, validation rules, and environment variable examples
- [x] All configurable modules are listed and properly registered in `main.go`

### Non-Functional Acceptance
- [x] Configuration generation is fast and doesn't impact application startup
- [x] Generated config.example is well-documented with comments
- [ ] CI/CD integration for automatic config validation

### Technical Acceptance
- [x] Follows BMad Config Center Registry Pattern
- [x] All existing modules (logger, i18n, jwt) comply with the standard
- [x] New modules automatically inherit the configuration pattern

---

## Technical Design

### Module Configuration Standard

#### 1. Config.go Structure Requirement

Every configurable module MUST have `config.go` with:
- `Config` struct with proper tags (`yaml`, `default`, `db`, `validate`)
- `DefaultConfig()` function
- `ToRuntimeConfig()` method (if type conversion needed)
- `RuntimeConfig` struct (optional, for parsed values like `time.Duration`)

#### 2. Factory Function as Config Connector

Every module MUST provide a factory function:
- `NewXXXFromConfig(cfg *Config)` - converts Config to service instance
- Handles config-to-runtime conversion
- Returns error for invalid configurations

### Config.example Generation

**Timing**: Development-time via `make config-example`

**Location**: `core/scripts/generate-config-example.go` (Inside core module to access types)

**Process**:
1. **Bootstrap Registry**: Script must explicitly import and register all target modules (simulating `main.go`).
2. **Scan Registry**: Iterate through registered modules in the script's registry instance.
3. **Extract Metadata**: Use reflection to read tags (`yaml`, `default`, `validate`).
4. **Generate YAML**: Output with inline comments in proper order.

**Output Format**:
- Each module as top-level YAML key
- Inline comments: default values, validation rules, env vars
- Placeholders for secrets: `${VAR_NAME}`
- Warning comments for dangerous options (e.g., `auto_migrate: true`)

### Module Configuration Inventory

#### Currently Registered Modules

| Module | Location | Status | Notes |
|--------|----------|--------|-------|
| logger | `core/pkg/logger/logger.go` | ✅ Registered | Config inline in logger.go |
| i18n | `core/pkg/i18n/config.go` | ✅ Registered | Standard pattern |
| jwt | `core/internal/jwt/config.go` | ✅ Registered | Standard pattern |

#### Special Case: Database Module

⚠️ **CRITICAL**: Database module MUST NOT be registered to Config Center.

**Reason**: Bootstrap Circular Dependency
```
Config Center startup → Needs Database → Loads database config from Registry → Registry not initialized yet → DEADLOCK
```

**Current Architecture**:
- Database config loaded BEFORE Config Center initialization (Infrastructure Config)
- Used in `bootstrap.LoadInitialConfig()` → `InitDatabase()` → `CreateService()`

**Action**: Document database as exception, do NOT register it.

#### Modules Requiring Action

| Module | Location | Current State | Action |
|--------|----------|---------------|--------|
| database | `core/pkg/database/config.go` | Has config.go | **DO NOT register** (bootstrap dependency) |
| server | `core/pkg/server/` | No config.go | Create if needed (HTTP/HTTPS settings) |

#### Future Module Considerations

| Module | Config Needs | Priority | Notes |
|--------|--------------|----------|-------|
| user | Password policy, session TTL | Medium | Business module |
| server | HTTP/HTTPS ports, timeouts | Low | Infrastructure |
| temporal | Workflow retry policies | Low | External service |
| kratos | Identity provider settings | High | If integrated |

---

## Implementation Plan

### Phase 1: Foundation (Day 1)

**Audit & Documentation**
- [x] Verify logger, i18n, jwt follow config.go standard
- [x] Document database as bootstrap exception (no registration)
- [x] Update coding-standards.md with module inventory

**Generation Script**
- [x] Create `core/scripts/generate-config-example.go`
- [x] Implement explicit module registration in script
- [x] Implement registry reflection logic
- [x] Generate YAML with inline comments (default, validate, env)
- [x] Handle nested structs and array fields

### Phase 2: Integration (Day 2)

**Makefile & CI/CD**
- [x] Add `config-example` target to Makefile
- [x] Integrate into `make dev-setup`
- [ ] Add pre-commit validation (optional)

**Testing & Documentation**
- [x] Test generation with current modules (logger, i18n, jwt)
- [x] Validate YAML syntax
- [x] Update README with config.example usage

---

## Key Design Decisions

### Why Development-Time Generation?

**✅ Chosen**: `make config-example` (on-demand)
- Fast: No impact on application startup
- Clean: Generated file committed to repo
- Predictable: Developers control when to regenerate

**❌ Rejected**: Runtime generation
- Slow: Adds startup overhead
- Complex: Needs write permissions
- Unpredictable: May change unexpectedly

### Why Database is NOT Registered?

**Bootstrap Order**:
```
1. Load database config from YAML
2. Connect to database  
3. Create Config Center service (needs DB connection)
4. Load business module configs from Registry
```

**If database were registered**:
```
1. Config Center needs database config from Registry
2. But Registry is part of Config Center
3. Circular dependency → DEADLOCK
```

**Solution**: Database remains infrastructure config, loaded separately in bootstrap phase.

---

## API Changes

**New Command**:
```bash
make config-example  # Generate config/config.example
```

**Updated main.go** (no changes needed):
```go
// Database is NOT registered (bootstrap dependency)
registry := config.NewRegistry()
registry.Register("logger", &logger.Config{})
registry.Register("i18n", &i18n.Config{})
registry.Register("jwt", &jwt.Config{})
// DO NOT: registry.Register("database", ...) ❌
```

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Secrets in config.example | Critical | Always use `${VAR_NAME}` placeholders |
| Database circular dependency | Critical | Never register database module |
| Forgotten modules | Medium | CI validates all config.go have registration |
| Complex generation logic | Low | Keep reflection-based, avoid custom parsers |

---

## Definition of Done

- [x] All business modules have config.go (logger, i18n, jwt)
- [x] Database documented as bootstrap exception
- [x] `make config-example` generates valid YAML
- [x] Generated file includes all registered modules
- [x] Comments include defaults, validation, env vars
- [x] No secrets in generated file
- [x] Documentation updated

---

## References

- [Story 3.2: Configuration Center Foundation](./3-2-config-basic.md)
- [Story 1.18: CLI Generate - Unified Code Generation](./1-18-cli-generate.md) (covers config example generation)
- [Coding Standards: Section 8 - Configuration Management](../../standards/coding-standards.md#8-配置管理)

---

**Story Owner**: Platform Dev Team  
**Created**: 2026-01-08  
**Last Updated**: 2026-01-19  
**Sprint**: Sprint 0-1 (Foundation)
