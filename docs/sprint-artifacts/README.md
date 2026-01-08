# Sprint Artifacts - BMad Method Workflow

**Last Updated**: 2026-01-08 | **Maintainer**: Architect Agent

> **📌 Single Source of Truth**: [`sprint-status.yaml`](./sprint-status.yaml) - Central tracking for all Epic-Story relationships

---

## Overview

BMad Method workflow management:
- **Epics**: Business-level features (what to build)
- **Stories**: Executable tasks (how to build)
- **sprint-status.yaml**: Central tracking file

**Hierarchy**: `Epic → Stories`

---

## Quick Commands

```bash
# View status
make sprint-status                              # Summary
make sprint-status-summary                      # Detailed with epic breakdown

# List stories
./scripts/manage-sprint-status.py list-stories --epic epic-auth
./scripts/manage-sprint-status.py list-stories --status planning

# Update story
./scripts/manage-sprint-status.py update-story story-18 done

# Add new story
./scripts/manage-sprint-status.py add-story story-21 "Password Reset" epic-auth

# Update statistics
make sprint-status-update
```

---

## Agent Workflows

### Architect: Create Story
1. Create file: `sprint-X/story-XX-title.md`
2. Add to YAML: `./scripts/manage-sprint-status.py add-story story-21 "Title" epic-auth`
3. Validate: `make validate-stories`
4. Commit (git hooks auto-update)

### Dev: Update Status
1. Update story file status field
2. Update YAML: `./scripts/manage-sprint-status.py update-story story-18 done`
3. Commit

### PM: Sprint Planning
1. View status: `make sprint-status-summary`
2. Filter stories: `./scripts/manage-sprint-status.py list-stories --status planning`
3. Organize stories into epics in sprint-status.yaml

---

## sprint-status.yaml Structure

```yaml
epics:
  - id: "epic-infrastructure"
    name: "Infrastructure & Foundation"
    status: "in-progress"              # planning | in-progress | done | paused
    stories: ["story-01", "story-02"]

stories:
  - id: "story-01"
    title: "Docker Development"
    status: "done"                      # planning | in-progress | done | blocked
    epic: "epic-infrastructure"
```

---

## Python API

```python
from scripts.manage_sprint_status import SprintStatusManager

mgr = SprintStatusManager()

# Read
epic = mgr.get_epic("epic-auth")
story = mgr.get_story("story-18")
stories = mgr.list_stories_by_status("planning")
stories = mgr.list_stories_by_epic("epic-auth")

# Write
mgr.update_story_status("story-18", "done")
mgr.add_story({'id': 'story-21', 'title': 'Title', 'status': 'planning', 'epic': 'epic-auth'})
mgr.update_statistics()
```

---

## Story Standards

### Required Metadata
```markdown
**Priority**: P0 (Critical) | P1 (High) | P2 (Medium)
**Effort**: 2 days
**Owner**: Backend Dev
**Dependencies**: story-18 or -
**Status**: Planning | In Progress | Done | Blocked
**Module**: Infrastructure | Auth | Storage | Functions | Management
**Issue**: #123 or #TBD
```

### Module Categories
- **Infrastructure**: Docker, CI/CD, Config, Testing, Logging
- **Auth**: Login, Registration, JWT, RBAC
- **Storage**: File upload, Object storage
- **Functions**: Serverless execution, Scheduling
- **Management**: User admin, Monitoring

---

## Available Epics

| Epic ID | Name | Status |
|---------|------|--------|
| epic-infrastructure | Infrastructure & Foundation | in-progress |
| epic-i18n | Internationalization | planning |
| epic-config | Configuration Management | in-progress |
| epic-docs | API Documentation | done |
| epic-auth | Authentication & Authorization | in-progress |
| epic-storage | File Storage Service | planning |
| epic-functions | Serverless Functions | planning |

---

## Automation

### Git Hooks
Pre-commit hook auto-updates sprint-status.yaml statistics

### Manual Commands
```bash
make sprint-status-update    # Update statistics
make validate-stories        # Validate story format
make sync-index             # Update legacy README table
```

---

## Validation

```bash
# Validate stories
make validate-stories

# Check YAML syntax
python3 -c "import yaml; yaml.safe_load(open('docs/sprint-artifacts/sprint-status.yaml'))"
```

---

## Best Practices

**All Agents:**
- ✅ Read from sprint-status.yaml (single source of truth)
- ✅ Update YAML when creating/modifying stories
- ✅ Run validation before commit
- ✅ Use Python API for automation

**Architect:** Clear acceptance criteria, validate format  
**Dev:** Update status as work progresses, add issue numbers  
**PM:** Use YAML for planning, track via epic status

---

## Troubleshooting

**YAML Syntax Error:**
```bash
python3 -c "import yaml; yaml.safe_load(open('docs/sprint-artifacts/sprint-status.yaml'))"
```

**Statistics Out of Sync:**
```bash
make sprint-status-update
```

**Story Not Found:**
- Check `id` matches pattern (e.g., `story-01`)
- Verify story in epic's story list
- Ensure file exists in sprint directory

---

**Details**: See story files in `sprint-0/` and `sprint-1/` directories.
