# Story 1: Docker Development & Deployment Environment
# Sprint 0: Infrastructure

**Priority**: P0  
**Effort**: 3 days  
**Owner**: DevOps/Backend Dev  
**Dependencies**: -  
**Status**: Planning  
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

- [ ] Create `docker/Dockerfile`, 3x `docker-compose.yml`, `.env` files
- [ ] Add TLS support: `SSL_CERT_FILE`, `SSL_KEY_FILE` env vars
- [ ] Create `.github/workflows/docker-build.yml` (GHCR auto-publish)
- [ ] Create `Makefile`: `dev-up`, `run-local`, `build-local`, `test-local`
- [ ] Create reverse proxy examples: Nginx, Caddy, Traefik
- [ ] Update 4x setup docs
- [ ] Dev deps start < 1min, prod deploy < 2min, image < 30MB

---

## Technical Design

### Dockerfile Structure
```
golang:1.21-alpine (builder) → alpine:latest (prod)
- Static binary, non-root user, health check
- Expose: 8080 (HTTP), 8443 (HTTPS)
- Size target: < 30MB
```

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

| Command | Action |
|---------|--------|
| `make dev-up` | Start postgres + redis |
| `make build-local` | Build image locally |
| `make test-local` | Run integration tests |

---

## Test Cases

- [ ] Dev: `make dev-up` starts in < 1min, ports 5432/6379 accessible
- [ ] Local: `make build-local` completes < 3min, image < 30MB
- [ ] Prod: Remote image deploys < 2min, health checks pass
- [ ] TLS: HTTP :8080 and HTTPS :8443 both work

---

## Deliverables

**Docker**: `Dockerfile`, 3x `docker-compose.yml`, 2x `.env`  
**CI/CD**: `.github/workflows/docker-build.yml`  
**Tools**: `Makefile`, `scripts/generate-self-signed-cert.sh`  
**Examples**: `examples/reverse-proxy/` (nginx, caddy, traefik)  
**Docs**: `docs/product/setup/` (4 files)

---

**Created**: 2025-12-27  
**Updated**: 2025-12-28
