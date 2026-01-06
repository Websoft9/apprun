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
apprun-base:latest (599MB)
├── golang:1.24-alpine
├── Build tools (git, make)
├── All Go dependencies (cached)
├── Atlas CLI (from official image)
└── Multi-arch support (amd64/arm64)
      ↓
apprun:latest (178MB)
└── Application binary + configs
```

## Multi-Architecture Support

The base image supports both AMD64 and ARM64 architectures:

- **AMD64**: Native x86_64 support
- **ARM64**: Via QEMU emulation in CI/CD
- **Atlas CLI**: Downloaded directly from official releases

### Atlas CLI Installation Method

**Previous approach** (deprecated):
```dockerfile
COPY --from=arigaio/atlas:latest /atlas /go/bin/atlas
```
❌ Problem: Atlas Docker image contains `unknown/unknown` platform layers

**Current approach** (recommended):
```dockerfile
RUN curl -sSL -o /go/bin/atlas \
    "https://release.ariga.io/atlas/atlas-linux-${ATLAS_ARCH}-latest"
```
✅ Benefits:
- Clean manifest (only `linux/amd64` and `linux/arm64`)
- No `unknown/unknown` pollution
- Faster downloads
- Direct architecture control

### Testing Multi-Arch Builds

```bash
# Test locally (requires Docker Desktop with Buildx)
./scripts/test-multi-arch.sh

# Manual build for specific architecture
docker buildx build \
  --platform linux/amd64 \
  --file docker/Dockerfile.base \
  --tag apprun-base:amd64 \
  --load \
  .
```

## GitHub Actions Workflows

### 1. Build Base Image (`.github/workflows/build-base.yml`)

**Purpose**: Build and publish base image to GHCR

**Triggers**:
- **Weekly**: Every Sunday 2 AM UTC
- **On change**: When `go.mod`/`go.sum`/`Dockerfile.base` modified
- **Manual**: Workflow dispatch with force rebuild option

**Output**: `ghcr.io/websoft9/apprun-base:latest`

**Usage**:
```bash
# Manual trigger via GitHub UI:
Actions > Build Base Image > Run workflow > Force rebuild ✓
```

### 2. Docker Build and Publish (`.github/workflows/docker-build.yml`)

**Purpose**: Build and publish application image

**Triggers**:
- Push to `main`/`develop` branches
- Tags matching `v*`
- Pull requests (build only, no push)

**Dependencies**: Requires base image from GHCR

**Behavior**:
- ✅ Auto-pulls base image from `ghcr.io/websoft9/apprun-base:latest`
- ⚠️ Shows warning if base image not found
- 🔄 Falls back to full build if necessary

**Output**: `ghcr.io/websoft9/apprun:latest`

## First-Time Setup

If you're setting up CI for the first time:

1. **Build base image first**:
   ```bash
   # Go to GitHub Actions
   Actions > Build Base Image > Run workflow > Force rebuild ✓
   ```

2. **Wait for completion** (~5-10 minutes)

3. **Then build application**:
   - Push to main branch
   - Or trigger docker-build.yml manually

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
