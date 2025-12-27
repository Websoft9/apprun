# Story Metadata Specification
**Version**: 2.0  
**Last Updated**: 2025-12-27  
**Purpose**: Agent-driven Story management

---

## 📋 Required Metadata Fields

| Field | Required | Format | Example |
|-------|----------|--------|---------|
| **Priority** | ✅ | P0/P1/P2 | P0 |
| **Effort** | ✅ | X days | 2 days |
| **Owner** | ✅ | Role | Backend Dev |
| **Dependencies** | ✅ | Story IDs or - | Story 1, 2 or - |
| **Status** | ✅ | Planning/In Progress/Done/Blocked | Planning |
| **Module** | ✅ | Module name | Infrastructure |
| **Issue** | ✅ | #TBD or #123 | #TBD |

---

## 📝 Document Structure

```markdown
# Story {N}: {Title}
# Sprint {N}: {Sprint Title}

**Priority**: P0/P1/P2  
**Effort**: X days  
**Owner**: {Role}  
**Dependencies**: Story X, Y or -  
**Status**: Planning/In Progress/Done/Blocked  
**Module**: {Module name}  
**Issue**: #TBD or #123  
**Related**: [Doc name](link)

---

## User Story
As a {role}, I want {feature}, so that {value}.

---

## Acceptance Criteria
- [ ] {Verifiable criterion 1}
- [ ] {Verifiable criterion 2}

---

## Implementation Tasks
- [ ] {Task 1}
- [ ] {Task 2}

---

## Technical Details
{Code examples, architecture diagrams, configurations}

---

## Test Cases
- [ ] {Test scenario 1}
- [ ] {Test scenario 2}

---

## Related Docs
- [Doc title](link)

---

**Created**: YYYY-MM-DD  
**Updated**: YYYY-MM-DD  
**Maintainer**: {Agent Name}
```

---

## 🎯 Module Categories

- **Infrastructure**: Docker, CI/CD, Config, Testing
- **Auth**: Login, Permissions, Token, RBAC
- **Storage**: File upload, Object storage
- **Functions**: Execution, Scheduling
- **Management**: User admin, Monitoring

---

## ✅ Validation

```bash
# Validate all Stories (run after Agent generates)
make validate-stories
```

**Validation checks**:
- ✅ All required fields present
- ✅ Priority format (P0/P1/P2)
- ✅ Status format (4 valid values)
- ✅ Required sections exist
