# Docker Image Maintenance Guide

## Overview

The `apprun-base` image is a pre-built cache layer containing all Go dependencies. It dramatically speeds up Docker builds by eliminating repeated dependency downloads.


## Quick Start

### For Developers

```bash
# Pull pre-built base image (recommended)
make pull-base

# Or build locally (one-time, 5-8 min)
make build-base

# Then build app (fast, 1-2 min)
make build-local
```

### For CI/CD

Base image is automatically rebuilt:
- **Weekly**: Every Sunday 2 AM UTC
- **On change**: When `go.mod`/`go.sum` modified
- **Manual**: Via GitHub Actions UI

## Architecture

```
apprun-base:latest (500MB)
├── golang:1.24-alpine
├── Build tools (git, make)
├── All Go dependencies (cached)
└── Pre-compiled stdlib
      ↓
apprun:latest (144MB)
└── Application binary only
```

## When to Rebuild

### Automatic (GitHub Actions)
- Weekly schedule
- `go.mod`/`go.sum` changes
- Manual workflow dispatch

### Manual (Local)
- Go version upgrade
- Team onboarding
- Major dependency changes

## Troubleshooting

### "Base image not found"
```bash
make pull-base  # Pull from registry
# or
make build-base  # Build locally
```

### "Build still slow"
```bash
# Verify base is used
docker build -f docker/Dockerfile . 2>&1 | grep "apprun-base"

# Rebuild if go.mod changed
make build-base
```

## Related Documentation

- [Story 01: Docker Environment](../../sprint-artifacts/sprint-0/story-01-docker-environment.md)
- [Dockerfile.base](../../../docker/Dockerfile.base)
- [Build Base Workflow](../../../.github/workflows/build-base.yml)

---

**Last Updated**: 2026-01-06  
**Maintainer**: DevOps Team
