# apprun Owner Manual

**For project maintainers and operators** - CI/CD, releases, infrastructure management.

---

## 🎯 CI/CD Operations

### Local Development Linting

**Install golangci-lint**:
```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Or via Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Run linters**:
```bash
# Quick lint (modified files)
make lint

# Full project lint
golangci-lint run --config=.golangci.yml

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### CI Pipeline

**Workflow files**:
- `.github/workflows/ci.yml` - Main CI (lint, test, build)
- `.github/workflows/docker-build.yml` - Multi-arch Docker images
- `.github/workflows/release.yml` - Release automation

**CI stages**:
1. **Lint**: golangci-lint with `.golangci.yml`
2. **Test**: Go unit/integration tests (coverage > 80%)
3. **Security**: gosec + trivy scans
4. **Build**: Multi-arch Docker images (amd64, arm64)
5. **Deploy**: Staging → Production (manual approval)

**Triggered by**:
- Push to `main` → Full CI + Docker build
- Pull requests → Lint + test only
- Tags `v*` → Release workflow

### Troubleshooting CI Failures

| Error | Cause | Fix |
|-------|-------|-----|
| Lint failure | Code style violations | Run `make lint-fix` |
| Test failure | Broken tests | Check `make test` locally |
| Docker build timeout | Large layers | Optimize Dockerfile caching |
| Security scan | Known CVEs | Update dependencies |

**View detailed logs**:
```bash
# Download workflow logs
gh run view <run-id> --log

# Re-run failed jobs
gh run rerun <run-id>
```

---

## 📦 Release Management

### Versioning Strategy

- **Semantic Versioning**: `MAJOR.MINOR.PATCH`
- **Pre-releases**: `v1.2.3-rc.1`, `v1.2.3-beta.2`
- **Release branches**: `release/v1.x` for maintenance

### Creating a Release

```bash
# 1. Update version
vim core/pkg/version/version.go  # Set new version

# 2. Update CHANGELOG.md
# Add release notes, breaking changes, migration guide

# 3. Create release branch (for major/minor)
git checkout -b release/v1.2
git push origin release/v1.2

# 4. Tag release
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0

# 5. GitHub Release
gh release create v1.2.0 \
  --title "apprun v1.2.0" \
  --notes-file CHANGELOG.md \
  --draft  # Remove --draft when ready
```

### Release Checklist

- [ ] All tests passing in CI
- [ ] CHANGELOG.md updated
- [ ] Version bumped in `version.go`
- [ ] Migration guide written (if breaking changes)
- [ ] Docker images built for all architectures
- [ ] Staging deployment tested
- [ ] Release notes reviewed
- [ ] Security scan passed
- [ ] Documentation updated

### Hotfix Process

```bash
# 1. Create hotfix branch from tag
git checkout -b hotfix/v1.2.1 v1.2.0

# 2. Apply fix
git cherry-pick <commit-sha>

# 3. Test thoroughly
make test && make check

# 4. Tag and release
git tag -a v1.2.1 -m "Hotfix: critical bug fix"
git push origin v1.2.1

# 5. Merge back to main
git checkout main
git merge hotfix/v1.2.1
git push origin main
```

---

## 🗄️ Database Operations

### Migration Management

**Prerequisites**:
```bash
# Ensure Docker is running
docker --version

# Start development database
docker-compose -f docker-compose.dev.yml up -d postgres
```

### Common Migration Tasks

**Check status**:
```bash
make migrate-status
```

**Apply migrations**:
```bash
# Development
make migrate-apply

# Production (dry-run first)
docker run --rm \
  -v $(PWD)/core:/app \
  arigaio/atlas:latest \
  migrate apply \
  --dir file:///app/migrations \
  --url "$PROD_POSTGRES_URL" \
  --dry-run

# Execute after reviewing dry-run
make migrate-apply
```

**Create new migration**:
```bash
# 1. Modify Ent schema (core/ent/schema/)
vim core/ent/schema/user.go

# 2. Generate migration
make migrate-diff NAME=add_user_phone

# 3. Review generated SQL
cat core/migrations/00X_add_user_phone.sql

# 4. Validate
make migrate-validate

# 5. Apply
make migrate-apply
```

**Validate migrations**:
```bash
make migrate-validate  # Check syntax
make migrate-lint      # Check best practices
make migrate-hash      # Generate checksum
```

### Production Migration Deployment

**Pre-deployment checklist**:
- [ ] Database backed up
- [ ] Tested in staging environment
- [ ] Migration time estimated (for large tables)
- [ ] Rollback plan prepared
- [ ] Team notified
- [ ] Monitoring dashboards ready

**Deployment steps**:
```bash
# 1. Set production database URL
export POSTGRES_URL="postgres://user:pass@prod-db:5432/apprun"

# 2. Backup database
pg_dump -h prod-db -U user -d apprun > backup_$(date +%Y%m%d).sql

# 3. Dry-run
docker run --rm \
  -v $(PWD)/core:/app \
  arigaio/atlas:latest \
  migrate apply \
  --dir file:///app/migrations \
  --url "$POSTGRES_URL" \
  --dry-run

# 4. Review output and execute
make migrate-apply

# 5. Verify
make migrate-status

# 6. Check application logs
kubectl logs -f deployment/apprun-core
```

**Rollback procedure**:
```bash
# Option 1: Apply rollback migration (recommended)
# Create rollback SQL file: 00X_rollback_description.sql
ALTER TABLE users DROP COLUMN phone;

make migrate-apply

# Option 2: Restore from backup
psql -h prod-db -U user -d apprun < backup_20250108.sql
```

### Troubleshooting Database Issues

| Problem | Cause | Solution |
|---------|-------|----------|
| Checksum mismatch | Migration file modified | Run `make migrate-hash` |
| "Database not clean" | Existing schema without baseline | Run `make migrate-baseline` |
| Connection timeout | Network/firewall issue | Check `host.docker.internal` or use direct IP |
| Migration lock timeout | Previous migration crashed | Manually release lock in DB |

**Troubleshooting**: See Story 05a for detailed procedures

---

## 🚀 Deployment Operations

### Environment Configuration

**Environment files**:
- `config/default.yaml` - Default configuration
- `config/conf_d/*.yaml` - Environment overrides
- `.env` - Secrets (never commit)

**Configuration priority**:
```
Environment variables > conf_d/*.yaml > default.yaml
```

### Docker Deployment

**Build images**:
```bash
# Single architecture
docker build -t apprun-core:latest -f docker/Dockerfile .

# Multi-architecture
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t apprun-core:latest \
  -f docker/Dockerfile \
  --push .
```

**Run locally**:
```bash
# Development
docker-compose -f docker-compose.dev.yml up -d

# Local production simulation
docker-compose -f docker-compose.local.yml up -d

# Production
docker-compose up -d
```

### Kubernetes Deployment

**Deploy to cluster**:
```bash
# Apply manifests
kubectl apply -f k8s/

# Update deployment image
kubectl set image deployment/apprun-core \
  apprun-core=apprun-core:v1.2.0

# Rolling update status
kubectl rollout status deployment/apprun-core

# Rollback if needed
kubectl rollout undo deployment/apprun-core
```

**Check health**:
```bash
# Pod status
kubectl get pods -l app=apprun-core

# Logs
kubectl logs -f deployment/apprun-core

# Health endpoints
curl https://apprun.example.com/health
```

---

## 🔐 Security Operations

### Secret Management

**Local development**:
```bash
# Generate secure password
openssl rand -base64 32

# Create .env file
cat > .env <<EOF
POSTGRES_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)
EOF
```

**Production**:
- Use **Kubernetes Secrets** or **AWS Secrets Manager**
- Rotate secrets quarterly
- Never commit secrets to Git

### TLS Certificate Management

**Generate self-signed cert (dev)**:
```bash
./scripts/generate-self-signed-cert.sh
```

**Production certificates**:
- Use **Let's Encrypt** with cert-manager (K8s)
- Or **AWS Certificate Manager** (ACM)
- Auto-renewal enabled

### Security Scanning

**Scan Docker images**:
```bash
# Trivy scan
trivy image apprun-core:latest

# Critical vulnerabilities only
trivy image --severity CRITICAL apprun-core:latest
```

**Dependency audit**:
```bash
# Go modules
go list -m -json all | nancy sleuth

# Or use govulncheck
govulncheck ./...
```

---

## 📊 Monitoring & Operations

### Health Checks

**Endpoints**:
- `GET /health` - Liveness probe
- `GET /health/ready` - Readiness probe
- `GET /metrics` - Prometheus metrics

**Example response**:
```json
{
  "status": "healthy",
  "version": "1.2.0",
  "database": "connected",
  "uptime": "72h15m"
}
```

### Logging

**View logs**:
```bash
# Docker Compose
docker-compose logs -f apprun-core

# Kubernetes
kubectl logs -f deployment/apprun-core

# Last 100 lines
kubectl logs deployment/apprun-core --tail=100
```

**Log levels**:
- `DEBUG` - Development only
- `INFO` - Default for production
- `WARN` - Unexpected but handled
- `ERROR` - Errors requiring attention

### Performance Tuning

**Database connection pool**:
```yaml
# config/default.yaml
database:
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
```

**Go runtime**:
```bash
# Set GOMAXPROCS
export GOMAXPROCS=$(nproc)

# Enable profiling
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## 🛠️ Maintenance Tasks

### Routine Maintenance

**Weekly**:
- [ ] Review CI/CD failures
- [ ] Check security scan reports
- [ ] Update dependencies (patch versions)
- [ ] Review application logs for errors

**Monthly**:
- [ ] Update minor dependencies
- [ ] Review and archive old branches
- [ ] Check disk usage (database, logs)
- [ ] Review monitoring dashboards

**Quarterly**:
- [ ] Major dependency updates
- [ ] Security audit
- [ ] Rotate secrets/credentials
- [ ] Disaster recovery drill

### Backup & Recovery

**Database backups**:
```bash
# Manual backup
pg_dump -h localhost -U apprun -d apprun_prod > backup.sql

# Automated (cron)
0 2 * * * pg_dump -h prod-db -U apprun -d apprun_prod | gzip > /backups/apprun_$(date +\%Y\%m\%d).sql.gz
```

**Restore procedure**:
```bash
# Stop application
kubectl scale deployment apprun-core --replicas=0

# Restore database
psql -h prod-db -U apprun -d apprun_prod < backup.sql

# Restart application
kubectl scale deployment apprun-core --replicas=3
```

---

## 🆘 Emergency Procedures

### Application Down

1. **Check service status**:
   ```bash
   kubectl get pods
   docker ps
   ```

2. **Review logs**:
   ```bash
   kubectl logs deployment/apprun-core --tail=200
   ```

3. **Common fixes**:
   - Restart: `kubectl rollout restart deployment/apprun-core`
   - Scale up: `kubectl scale deployment apprun-core --replicas=5`
   - Rollback: `kubectl rollout undo deployment/apprun-core`

### Database Connection Issues

1. **Check connectivity**:
   ```bash
   psql -h prod-db -U apprun -d apprun_prod -c "SELECT 1;"
   ```

2. **Check connection pool**:
   ```bash
   kubectl logs deployment/apprun-core | grep "connection pool"
   ```

3. **Emergency fix**:
   - Increase pool size in config
   - Restart application pods

### High Load / Performance Degradation

1. **Scale horizontally**:
   ```bash
   kubectl scale deployment apprun-core --replicas=10
   ```

2. **Check resource usage**:
   ```bash
   kubectl top pods
   kubectl top nodes
   ```

3. **Emergency cache**:
   - Enable Redis caching
   - Add CDN for static assets

---

## 📚 Reference Documentation

| Document | Purpose |
|----------|---------|
| [Story 05a](./docs/sprint-artifacts/sprint-0/story-05a-database-migration.md) | Database migration decisions |
| [devops-process.md](./docs/standards/devops-process.md) | DevOps workflows & standards |
| [coding-standards.md](./docs/standards/coding-standards.md) | Code review guidelines |

---

**For contributors**: See [CONTRIBUTING.md](./CONTRIBUTING.md)

**Maintainer contact**: devops@apprun.com
