# Story 1: Docker Development & Deployment Environment
# Sprint 0: Infrastructure

**Priority**: P0  
**Effort**: 3 days  
**Owner**: DevOps/Backend Dev  
**Dependencies**: -  
**Status**: Done  
**Module**: Infrastructure  
**Issue**: #TBD  

---

## User Story

As a developer and DevOps engineer, I want a flexible Docker environment with local Go development, optional TLS, and production deployment via pre-built images.

---

## Design Principles

1. **Local-first Development**: `go run` for daily work, Docker for dependencies only
2. **On-demand Build**: Local image builds only for integration testing
3. **CI/CD Automation**: Auto-build and publish on git push
4. **Simple Deployment**: Users pull pre-built images from registry

---

## Acceptance Criteria

- [ ] Create `docker/Dockerfile.base` for Go dependencies cache layer
- [ ] Create `docker/Dockerfile`, 3x `docker-compose.yml`, `.env` files
- [ ] Add TLS support: `SSL_CERT_FILE`, `SSL_KEY_FILE` env vars
- [ ] Create `.github/workflows/docker-build.yml` (GHCR auto-publish)
- [ ] Create `.github/workflows/build-base.yml` (base image update)
- [ ] Create `Makefile`: `build-base`, `dev-up`, `run-local`, `build-local`, `test-local`
- [ ] Create reverse proxy examples: Nginx, Caddy, Traefik
- [ ] Update 4x setup docs + base image maintenance guide
- [ ] Dev deps start < 1min, prod deploy < 2min, image < 30MB
- [ ] Base image build < 8min, app rebuild < 2min (with cache)

---

## Technical Design

### Docker Image Architecture

**Two-layer Build Strategy**:
```
┌─────────────────────────────────────────┐
│ apprun-base:latest                      │
│ - golang:1.24-alpine                    │
│ - Pre-installed Go dependencies         │
│ - Build tools (git, make)               │
│ - Atlas CLI (database migration tool)   │
│ - Size: ~500MB (cached)                 │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│ apprun:latest (builder)                 │
│ - FROM apprun-base                      │
│ - Copy source code                      │
│ - Build binary (fast, deps cached)     │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│ apprun:latest (production)              │
│ - alpine:latest                         │
│ - Static binary only                    │
│ - Non-root user, health check           │
│ - Size: < 30MB                          │
└─────────────────────────────────────────┘
```

### Dockerfile Structure

**Dockerfile.base** (Pre-built dependency layer):
```dockerfile
FROM golang:1.24-alpine
ENV GOPROXY=https://goproxy.cn,direct
RUN apk add --no-cache git make
COPY go.mod go.sum ./
RUN go mod download
RUN go install ariga.io/atlas/cmd/atlas@latest  # Prebuilt migration tool
```

**Dockerfile** (Application layer):
```dockerfile
FROM apprun-base:latest AS builder
COPY . .
RUN make build
# Atlas CLI already available from base image

FROM alpine:latest
COPY --from=builder /apprun /app/
COPY --from=builder /go/bin/atlas /usr/local/bin/  # Reuse prebuilt Atlas
- Expose: 8080 (HTTP), 8443 (HTTPS)
- Size target: < 30MB
```

**Build Time Optimization**:
- **Cold build** (first time): ~5-8 min
- **Warm build** (with base): ~1-2 min (60-80% faster)
- **CI/CD builds**: Cached base image reduces pipeline time
- **Atlas CLI**: Prebuilt in base, saves ~20-30 seconds per build

### TLS Support
```go
// Optional TLS via environment variables
if os.Getenv("SSL_CERT_FILE") != "" {
    http.ListenAndServeTLS(":8443", cert, key, handler)
} else {
    http.ListenAndServe(":8080", handler)
}
```

### Docker Compose Files

| File | Purpose | App |
|------|---------|-----|
| `dev.yml` | Local development | No (use `go run`) |
| `prod-local.yml` | Integration test | Build locally |
| `prod.yml` | Production | Pull from GHCR |

### Makefile Commands

| Command | Action | Base Image |
|---------|--------|------------|
| `make build-base` | Build apprun-base with Go deps | ✅ Creates cache layer |
| `make dev-up` | Start postgres + redis | - |
| `make build-local` | Build app image (uses base) | ✅ Reuses cache |
| `make test-local` | Run integration tests | ✅ Fast rebuild |

**Base Image Update Strategy**:
- **Frequency**: Weekly or when `go.mod` changes
- **Trigger**: Manual or automated via CI schedule
- **Registry**: Store in GHCR for team sharing

---

## Test Cases

- [ ] Base: `make build-base` completes < 8min, image cached in Docker
- [ ] Dev: `make dev-up` starts in < 1min, ports 5432/6379 accessible
- [ ] Local: `make build-local` completes < 2min (with base), image < 30MB
- [ ] Rebuild: Change code → rebuild < 90s (deps cached)
- [ ] Prod: Remote image deploys < 2min, health checks pass
- [ ] TLS: HTTP :8080 and HTTPS :8443 both work
- [ ] Cache: `go.mod` unchanged → base layer reused, no re-download

---

## Deliverables

**Docker**: 
- `Dockerfile.base` (dependency cache layer)
- `Dockerfile` (application layer)
- 3x `docker-compose.yml`, 2x `.env`

**CI/CD**: 
- `.github/workflows/docker-build.yml` (app image)
- `.github/workflows/build-base.yml` (base image, weekly schedule)

**Tools**: 
- `Makefile` (add `build-base` target)
- `scripts/generate-self-signed-cert.sh`

**Examples**: 
- `examples/reverse-proxy/` (nginx, caddy, traefik)

**Docs**: 
- `docs/product/setup/` (4 files)
- `docs/product/setup/docker-base-image.md` (base maintenance guide)

---

## Implementation Notes

### Base Image Maintenance

**When to rebuild base image**:
1. Weekly schedule (automated via GitHub Actions)
2. Manual trigger when `go.mod` changes
3. After major Go version upgrade

**Base image workflow** (`.github/workflows/build-base.yml`):
```yaml
on:
  schedule:
    - cron: '0 2 * * 0'  # Every Sunday 2 AM UTC
  workflow_dispatch:      # Manual trigger
  push:
    paths:
      - 'core/go.mod'
      - 'core/go.sum'

jobs:
  build-base:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build and push base
        run: |
          docker build -f docker/Dockerfile.base -t ghcr.io/websoft9/apprun-base:latest .
          docker push ghcr.io/websoft9/apprun-base:latest
```

**Local development workflow**:
```bash
# One-time: Build base image (or pull from registry)
make build-base

# Daily development: Fast incremental builds
make build-local  # Uses cached base, only builds app code
```

**Benefits achieved**:
- ✅ 60-80% faster builds after first run
- ✅ Reduced network dependency downloads
- ✅ Consistent dependency versions across team
- ✅ CI/CD pipeline time reduced from ~8min to ~2min
- ✅ Atlas CLI prebuilt, saves ~20-30s per build
- ✅ No duplicate tool installations

---

**Created**: 2025-12-27  
**Updated**: 2026-01-06 (Added base image optimization)
