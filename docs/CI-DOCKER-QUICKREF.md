# Docker CI/CD Quick Reference

## 🚨 Problem Fixed

### Error
```
ERROR: failed to solve: apprun-base:latest: pull access denied
```

### Root Cause
- Dockerfile ARG BASE_IMAGE was not properly scoped
- GitHub Actions couldn't pass GHCR base image path
- Login step had unnecessary conditional

### Solution
1. ✅ Moved `ARG BASE_IMAGE` before `FROM` (global scope)
2. ✅ Re-declared `ARG BASE_IMAGE` after `FROM` (stage scope)
3. ✅ Removed login conditional (needed for pulling base)
4. ✅ Added base image check step in workflow

## 📦 Image Locations

| Environment | Base Image | App Image |
|------------|------------|-----------|
| **GHCR** | `ghcr.io/websoft9/apprun-base:latest` | `ghcr.io/websoft9/apprun:latest` |
| **Local** | `apprun-base:latest` | `apprun:local` |

## 🔄 Workflows

### Build Base Image (`build-base.yml`)
**Triggers:**
- Weekly (Sunday 2 AM UTC)
- go.mod/go.sum changes
- Manual dispatch

**Output:** Base image to GHCR

### Docker Build (`docker-build.yml`)
**Triggers:**
- Push to main/develop
- Tags (v*)
- Pull requests

**Dependencies:** Base image from GHCR

## 🎯 First-Time Setup

1. **Build base image first:**
   ```bash
   # Via GitHub Actions UI:
   Actions > Build Base Image > Run workflow > Force rebuild ✓
   ```

2. **Wait for completion** (5-10 min)

3. **Test locally (optional):**
   ```bash
   ./scripts/test-ci-build.sh
   ```

4. **Push code** - docker-build.yml will run automatically

## 🧪 Local Testing

### Simulate CI Build
```bash
./scripts/test-ci-build.sh
```

### Manual Build with GHCR Base
```bash
# Tag local base as GHCR (for testing)
docker tag apprun-base:latest ghcr.io/websoft9/apprun-base:latest

# Build with explicit args
docker build \
  --build-arg BASE_IMAGE=ghcr.io/websoft9/apprun-base:latest \
  --file docker/Dockerfile \
  --tag apprun:test \
  .
```

### Regular Local Build
```bash
make build-base   # First time only
make build-local  # Uses local base
```

## 🔧 Troubleshooting

### "Base image not found" in CI
**Solution:** Manually trigger `Build Base Image` workflow first

### Local build fails
```bash
# Check if base image exists
docker images apprun-base

# Rebuild if missing
make build-base
```

### CI still fails after fixes
1. Check GitHub Actions logs
2. Verify GITHUB_TOKEN has `packages: write` permission
3. Ensure base image workflow completed successfully
4. Re-run failed job (might be transient issue)

## 📚 Related Files

- **Dockerfiles:**
  - `docker/Dockerfile` - Main application image
  - `docker/Dockerfile.base` - Base image with dependencies

- **Workflows:**
  - `.github/workflows/build-base.yml` - Base image builder
  - `.github/workflows/docker-build.yml` - App image builder

- **Scripts:**
  - `scripts/test-ci-build.sh` - CI simulation tester
  - `Makefile` - Build automation

- **Docs:**
  - `docker/README.md` - Detailed Docker guide
  - This file - Quick reference
