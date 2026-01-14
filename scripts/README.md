# About

These scripts are for developer with System-level Test Orchestration, not for application starting

There any other test codes

- Unit and Integration testing in every module
- Infrastructure Validation Script at directory `core/tests`

---

## Story Index Generation

### generate-story-index.py

Generates story index from `sprint-status.yaml` (single source of truth).

**Usage:**

```bash
# Terminal output
python3 scripts/generate-story-index.py

# Generate Markdown file
python3 scripts/generate-story-index.py --format markdown

# Custom output path
python3 scripts/generate-story-index.py --format markdown -o /path/to/output.md
```

**Features:**
- ✅ Reads from `sprint-status.yaml` only (no scanning story files)
- ✅ Automatic epic and sprint mapping
- ✅ Status icons (✅ done, 🔄 in-progress, 📋 backlog, etc.)
- ✅ Summary with breakdown by status and sprint
- ✅ Terminal and Markdown output formats

### sync-story-index.sh

Wrapper script to regenerate `story-index.md` from `sprint-status.yaml`.

**Usage:**

```bash
bash scripts/sync-story-index.sh
```

**Output:** `docs/sprint-artifacts/story-index.md`

---

## Refactoring History

**2026-01-14**: Story index generation refactored to use `sprint-status.yaml` as single source of truth instead of scanning individual story files.