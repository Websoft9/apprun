# Contributing to apprun

**BMad Method**: AI-assisted development with specialized agents.

---

## 🚀 Quick Start

### Prerequisites
- Go 1.25.5+
- golangci-lint v2.7.2+
- PostgreSQL 14+
- Redis 7+ (optional, for caching)
- Govulncheck 
- [Atlas 1.0+](https://atlasgo.io/getting-started)
- Docker & Docker Compose
- GitHub account
- AI Coding Agent (GitHub Copilot, Cursor, or similar)

### Setup
```bash
# Clone repository
git clone https://github.com/Websoft9/apprun.git
cd apprun

# Start development environment
make dev-start

# Run tests
make test
```

---

## 📋 Development Workflow

### 1. Find a Task
- Check [`docs/sprint-artifacts/sprint-status.yaml`](docs/sprint-artifacts/sprint-status.yaml)
- Look for issues tagged `good-first-issue` or `help-wanted`
- Generate story index: `make story-index`

### 2. Create Branch
```bash
git checkout -b feature/story-XX-description
```

### 3. AI-Assisted Development

**Load context into your AI agent**:
```bash
@workspace /docs/standards/coding-standards.md
@workspace /docs/standards/api-design.md
```

**Ask AI**:
- "Implement X following coding-standards.md"
- "Does this code follow api-design.md Section 3?"
- "Generate tests with >80% coverage"

### 4. Development Commands
```bash
make build-fast    # Quick build
make test          # Run tests
make lint          # Run linters
make check         # Lint + test + security
```

### 5. Commit
```bash
git commit -m "feat(module): brief description

- Detail 1
- Detail 2

Ref: Sprint-X Story-XX"
```

**Commit format**: See [devops-process.md](./docs/standards/devops-process.md#22-commit-message-规范)

### 6. Submit Pull Request
- Title: `[Story-XX] Brief description`
- Reference story in description
- Link to relevant documentation
- Pass CI checks

---

## 🤖 BMad Method Essentials

### BMad Agents
- **architect**: System design decisions
- **dev**: Code implementation
- **tea**: Testing & automation
- **pm**: Requirements & planning

**Activate**: In GitHub Copilot Chat, select agent from mode dropdown.

### Key Principles
- **Documentation First**: Always reference [`docs/standards`](docs/standards ) before coding
- **AI-Assisted**: Load project standards into AI agent
- **Test-Driven**: Coverage > 80%
- **Review Checklist**: Use [Code Review Checklist](./docs/standards/devops-process.md#33-code-review-清单)

---

## 📚 Essential Documentation

| Document | Purpose |
|----------|---------|
| [coding-standards.md](./docs/standards/coding-standards.md) | Code style & patterns |
| [api-design.md](./docs/standards/api-design.md) | API decisions & formats |
| [devops-process.md](./docs/standards/devops-process.md) | Commit format & reviews |
| [Story 05a](./docs/sprint-artifacts/sprint-0/story-05a-database-migration.md) | Database migration details |
| [sprint-artifacts/](./docs/sprint-artifacts/) | Current sprint stories |

---

## 🛠️ Common Tasks

### Add New API Endpoint
1. Define in Ent schema (`core/ent/schema/`)
2. Generate migration: `make migrate-diff NAME=add_xxx`
3. Apply migration: `make migrate-apply`
4. Implement handler with Swagger annotations
5. Run `make swagger` to update docs
6. Test and commit

### Database Schema Changes

**Prerequisites**:
```bash
# Ensure Docker and database are running
docker ps | grep postgres
```

**Workflow**:
```bash
# 1. Modify Ent schema
vim core/ent/schema/user.go
# Example: field.String("phone").Optional()

# 2. Generate migration
make migrate-diff NAME=add_user_phone

# 3. Review generated SQL
cat core/migrations/00X_add_user_phone.sql

# 4. Apply migration
make migrate-apply

# 5. Verify status
make migrate-status
```

**Common scenarios**:
- Add field: Modify schema → `migrate-diff` → `migrate-apply`
- Create table: New schema file → `migrate-diff` → `migrate-apply`
- Rollback: Check Story 05a for rollback procedures

---

## 🙋 Help & Support

- **Issues**: Tag `help-wanted` or `good-first-issue`
- **Discussions**: [GitHub Discussions](https://github.com/Websoft9/apprun/discussions)
- **Documentation**: Start with [`docs/standards/README.md`](docs/standards/README.md)

---

**For project owners/maintainers**: See [OWNER.md](./OWNER.md)

**Thank you for contributing!** 🎉
