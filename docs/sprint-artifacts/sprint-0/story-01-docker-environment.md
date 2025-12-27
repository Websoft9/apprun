# Story 1: Docker Development & Deployment Environment
# Sprint 0: Infrastructure

**Priority**: P0  
**Effort**: 2 days  
**Owner**: DevOps/Backend Dev  
**Dependencies**: -  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [Deployment Architecture](../../architecture/deployment-architecture.md)

---

## User Story

As a developer and DevOps engineer, I want a unified Docker environment configuration that can switch between development and production modes via environment variables, enabling "configure once, run anywhere".

---

## Acceptance Criteria

- [ ] Create `docker/Dockerfile` (multi-stage: builder, dev, prod)
- [ ] Create `docker/docker-compose.yml` (environment-driven)
- [ ] Create `.env.dev` and `.env.prod.example`
- [ ] Create `docker/nginx/nginx.conf` (HTTPS support)
- [ ] Create `Makefile` with dev/prod commands
- [ ] Create `scripts/prod-deploy.sh`
- [ ] Update `docs/product/setup/*.md`
- [ ] Dev mode starts in < 5 minutes
- [ ] Prod deployment completes in < 15 minutes

---

## Technical Design

### Dockerfile (Multi-stage)

| Stage | Purpose | Base Image |
|-------|---------|------------|
| builder | Compile Go binary | golang:1.21-alpine |
| dev | Hot reload development | golang:1.21-alpine |
| prod | Minimal runtime (~20MB) | alpine:latest |

**Features**: Static binary, non-root user, health check

### docker-compose.yml

| Service | Dev Mode | Prod Mode |
|---------|----------|-----------|
| app | Run locally | Built from prod stage |
| postgres | Port 5432 exposed | Internal only |
| redis | Port 6379 exposed | Internal only |
| nginx | Not started | HTTPS enabled |

**Environment Variables**:
- `BUILD_TARGET`: dev / prod
- `SERVER_ENV`: development / production
- `APP_PROFILE`: default / production

### Makefile Commands

| Command | Action |
|---------|--------|
| `make dev-up` | Start postgres + redis |
| `make dev-down` | Stop dev services |
| `make prod-deploy` | Full production deployment |
| `make prod-down` | Stop prod services |

---

## Implementation Tasks

## Test Cases

### Development
- [ ] `make dev-up` succeeds
- [ ] Ports 5432/6379 accessible
- [ ] `go run cmd/server/main.go` connects

### Production
- [ ] `make prod-deploy` succeeds
- [ ] Health checks pass
- [ ] HTTPS works
- [ ] Ports not exposed externally

---

## Deliverables

- `docker/Dockerfile`
- `docker/docker-compose.yml`
- `.env.dev`, `.env.prod.example`
- `docker/nginx/nginx.conf`
- `Makefile`
- `scripts/prod-deploy.sh`
- `docs/product/setup/local-deployment.md`
- `docs/product/setup/production-deployment.md`


---

## Related Docs

---

**Created**: 2025-12-27  
**Updated**: 2025-12-27
