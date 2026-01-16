# apprun Makefile
# Organized according to Story 05c standards

.PHONY: help \
	app-start app-stop app-clean \
	deps-start deps-stop deps-clean \
	dev-start dev-stop dev-clean \
	build build-fast generate swagger i18n i18n-extract i18n-merge config-example \
	test test-unit test-integration test-e2e test-cover \
	lint lint-fix security check \
	db-diff db-migrate db-rollback db-status db-reset db-validate db-lint db-hash db-baseline \
	docker-build docker-up docker-down docker-logs docker-clean docker-build-base docker-pull-base \
	docs-api docs-validate story-validate story-sync story-index sprint-status sprint-summary \
	clean clean-all install check-deps version

# ============================================
# Help - Quick Reference
# ============================================

help:
	@echo "╔══════════════════════════════════════════════════════════════╗"
	@echo "║  AppRun BaaS Platform - Development Commands                ║"
	@echo "╚══════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make dev-start     - Start dev environment (deps + app)"
	@echo "  make test          - Run all tests"
	@echo "  make check         - Run quality checks (lint + test + security)"
	@echo ""
	@echo "🔨 Core Development:"
	@echo "  make app-start     - Start application (bin/server)"
	@echo "  make app-stop      - Stop application process"
	@echo "  make app-clean     - Clean app artifacts"
	@echo "  make deps-start    - Start dependencies (postgres, redis)"
	@echo "  make deps-stop     - Stop dependencies"
	@echo "  make deps-clean    - Clean dependency data (volumes)"
	@echo "  make dev-start     - Start all (deps + app)"
	@echo "  make dev-stop      - Stop all (app + deps)"
	@echo "  make dev-clean     - Clean all (app + deps)"
	@echo ""
	@echo "📦 Build & Generate:"
	@echo "  make build         - Full build (generate + i18n + swagger + compile)"
	@echo "  make build-fast    - Quick build (skip docs)"
	@echo "  make generate      - Generate Ent ORM code"
	@echo "  make swagger       - Generate API documentation"
	@echo "  make i18n          - Process translations"
	@echo "  make config-example - Generate config.example"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test          - Run all tests"
	@echo "  make test-unit     - Unit tests only"
	@echo "  make test-integration - Integration tests"
	@echo "  make test-cover    - Generate coverage report"
	@echo ""
	@echo "✨ Code Quality:"
	@echo "  make lint          - Run linter"
	@echo "  make lint-fix      - Auto-fix lint issues"
	@echo "  make security      - Run security scan (govulncheck)"
	@echo "  make check         - Full quality check (lint + test + security)"
	@echo ""
	@echo "🗄️  Database:"
	@echo "  make db-sync       - Auto-sync schema (dev mode)"
	@echo "  make db-migrate    - Apply migrations"
	@echo "  make db-diff       - Generate migration (NAME=xxx)"
	@echo "  make db-status     - Show migration status"
	@echo "  make db-rollback   - Rollback last migration"
	@echo "  make db-reset      - Reset database"
	@echo "  make db-inspect    - Check for schema drift"
	@echo "  make db-repair     - Preview repair SQL (dry-run)"
	@echo "  make db-repair-execute - Apply repair (dev only)"
	@echo "  ⚠️  Note: Requires Atlas CLI - install: curl -sSf https://atlasgo.sh | sh"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  make docker-build  - Build Docker images"
	@echo "  make docker-up     - Start all services"
	@echo "  make docker-down   - Stop all services"
	@echo "  make docker-logs   - View logs"
	@echo "  make docker-clean  - Clean Docker resources"
	@echo ""
	@echo "📚 Documentation:"
	@echo "  make docs-api      - Generate API docs (Swagger)"
	@echo "  make story-validate - Validate Story documents"
	@echo "  make story-index   - Generate story status table"
	@echo "  make sprint-status - View sprint status"
	@echo ""
	@echo "🛠️  Utilities:"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make install       - Install dev tools"
	@echo "  make check-deps    - Check dependencies"
	@echo "  make help          - Show this help"
	@echo ""

# ============================================
# 1. Core Development (核心开发)
# ============================================

# Start application using new CLI (apprun serve)
app-start:
	@echo "🚀 Starting application..."
	@if [ ! -f core/bin/apprun ]; then \
		echo "⚠️  bin/apprun not found, building..." && $(MAKE) build-fast; \
	fi
	@cd core && ./bin/apprun serve

# Stop application process
app-stop:
	@echo "🛑 Stopping application..."
	@pkill -f "bin/apprun" || pkill -f "bin/server" || echo "⚠️  No running app process found"
	@echo "✅ Application stopped"

# Clean application artifacts
app-clean:
	@echo "🧹 Cleaning application artifacts..."
	@rm -rf core/bin core/tmp core/*.log
	@echo "✅ Application artifacts cleaned"

# Start dependencies (postgres, redis via docker-compose)
deps-start:
	@echo "🚀 Starting dependencies..."
	@which docker >/dev/null 2>&1 || (echo "❌ docker not found. Install Docker first." && exit 1)
	@docker compose -f docker-compose.dev.yml up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 3
	@echo "✅ Dependencies started"
	@echo "  PostgreSQL: localhost:5432 (user: apprun, db: apprun_dev)"
	@echo "  Redis: localhost:6379"

# Stop dependencies
deps-stop:
	@echo "🛑 Stopping dependencies..."
	@docker compose -f docker-compose.dev.yml down
	@echo "✅ Dependencies stopped"

# Clean dependency data (volumes)
deps-clean:
	@echo "🧹 Cleaning dependency data..."
	@docker compose -f docker-compose.dev.yml down -v
	@echo "✅ Dependency data cleaned"

# Start development environment (deps + app)
dev-start: deps-start
	@echo "🚀 Starting development environment..."
	@echo "💡 Run 'make app-start' in another terminal to start the app"
	@echo "   Or use 'make dev-all' to start everything in background"

# Start everything (background mode)
dev-all: deps-start
	@echo "🚀 Starting all services..."
	@$(MAKE) app-start &
	@echo "✅ Development environment running"
	@echo "💡 Use 'make dev-stop' to stop everything"

# Stop development environment
dev-stop: app-stop deps-stop
	@echo "✅ Development environment stopped"

# Clean development environment
dev-clean: app-clean deps-clean
	@echo "✅ Development environment cleaned"

# ============================================
# 2. Build & Generate (构建与生成)
# ============================================

# 构建（正确顺序：生成代码 -> 提取翻译 -> 生成文档 -> 编译）
build: generate i18n swagger
	@echo "🔨 Building application with version info..."
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev"); \
	GIT_COMMIT=$$(git rev-parse HEAD 2>/dev/null || echo "unknown"); \
	BUILD_TIME=$$(date -u +"%Y-%m-%dT%H:%M:%SZ"); \
	cd core && go build -ldflags="-X apprun/pkg/version.Version=$$VERSION \
		-X apprun/pkg/version.GitCommit=$$GIT_COMMIT \
		-X apprun/pkg/version.BuildTime=$$BUILD_TIME" \
		-o bin/apprun .
	@echo "✅ Build complete: core/bin/apprun"
	@# Create backward compatibility symlink
	@cd core/bin && rm -f server && ln -sf apprun server
	@echo "✅ Backward compatibility: bin/server -> bin/apprun"

# 快速构建（跳过文档生成）
build-fast: generate
	@echo "⚡ Quick build (skip docs)..."
	@cd core && go build -o bin/apprun .
	@echo "✅ Quick build complete: core/bin/apprun"
	@# Create backward compatibility symlink
	@cd core/bin && rm -f server && ln -sf apprun server
	@echo "✅ Backward compatibility: bin/server -> bin/apprun"

# 代码生成 (Ent ORM)
generate:
	@echo "🔄 Generating Ent code..."
	@cd core && go generate ./ent
	@echo "✅ Ent code generated"

# ============================================
# i18n Commands (Story 8)
# ============================================

# Extract and merge translation keys
i18n: i18n-extract i18n-merge

# Extract translation keys from code
i18n-extract:
	@echo "🔍 Extracting translation keys from code..."
	@go run scripts/i18n-extract.go -source=./core -output=./core/locales/template.toml
	@echo "✅ Translation keys extracted"

# Merge extracted keys to translation files
i18n-merge:
	@echo "🔄 Merging translations to en-US..."
	@cd core && go run ../scripts/i18n-merge.go -template=./locales/template.toml -target=./locales/active.en-US.toml
	@echo "🔄 Merging translations to zh-CN..."
	@cd core && go run ../scripts/i18n-merge.go -template=./locales/template.toml -target=./locales/active.zh-CN.toml
	@echo "✅ Translations merged"
	@echo "⚠️  Please review and translate new keys in core/locales/"

# ============================================
# Configuration Management (Story 10a)
# ============================================

# Generate config.example from registered modules
config-example:
	@echo "🔧 Generating config.example from module registry..."
	@cd core && go run ./scripts/generate-config-example.go
	@echo "✅ config/config.example generated"
	@echo "💡 Review and customize for your environment"

# ============================================
# 5. Database (数据库迁移)
# ============================================

# Auto-sync schema changes (development mode)
db-sync:
	@cd core && ./bin/apprun migrate sync

# Generate migration from schema changes
# Usage: make db-diff NAME=add_project_table
db-diff:
ifndef NAME
	$(error NAME is required. Usage: make db-diff NAME=add_project_table)
endif
	@echo "📝 Generating migration: $(NAME)..."
	@cd core && ./bin/apprun migrate diff $(NAME)
	@echo ""
	@echo "⚠️  IMPORTANT: Review the generated SQL before committing!"

# Apply pending migrations
db-migrate:
	@cd core && ./bin/apprun migrate apply

# Show migration status
db-status:
	@cd core && ./bin/apprun migrate status

# Validate migrations
db-validate:
	@cd core && ./bin/apprun migrate validate

# Rollback last migration
db-rollback:
	@cd core && ./bin/apprun migrate rollback

# Reset database (DANGER!)
db-reset:
	@cd core && ./bin/apprun migrate reset

# Inspect database schema (check for drift)
db-inspect:
	@cd core && ./bin/apprun migrate inspect

# Repair database schema (declarative migration)
db-repair:
	@cd core && ./bin/apprun migrate repair

# Repair database schema (execute)
db-repair-execute:
	@cd core && ./bin/apprun migrate repair --execute

# ============================================
# 3. Testing (测试)
# ============================================

# Run all tests (alias for test-all)
test: test-unit test-integration
	@echo "✅ All tests passed"

# Run unit tests
test-unit:
	@echo "🧪 Running unit tests..."
	cd core && go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "📊 Coverage summary:"
	@cd core && go tool cover -func=coverage.out

# Generate coverage report (HTML)
test-cover: test-unit
	@echo ""
	@echo "📊 Generating HTML coverage report..."
	cd core && go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: core/coverage.html"

# Run integration tests
test-integration:
	@echo "🧪 Running integration tests..."
	@./tests/scripts/setup-test-db.sh
	@./tests/integration/config/test-api.sh
	@./tests/integration/config/test-priority.sh
	@./tests/scripts/cleanup.sh
	@echo "✅ Integration tests passed"

# Run end-to-end tests
test-e2e:
	@echo "🧪 Running E2E tests..."
	@echo "⚠️  E2E tests not implemented yet"

# Backward compatibility
test-all: test
test-unit-html: test-cover

# ============================================
# 4. Code Quality (代码质量)
# ============================================

# Run linter
lint:
	@echo "🔍 Running golangci-lint..."
	@which golangci-lint >/dev/null 2>&1 || (echo "❌ golangci-lint not installed" && echo "📥 Install: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin" && exit 1)
	@cd core && golangci-lint run --timeout=5m --config=.golangci.yml
	@echo "✅ Linting completed"

# Run linter with auto-fix
lint-fix:
	@echo "🔧 Running golangci-lint with auto-fix..."
	@which golangci-lint >/dev/null 2>&1 || (echo "❌ golangci-lint not installed" && exit 1)
	@cd core && golangci-lint run --timeout=5m --config=.golangci.yml --fix
	@echo "✅ Auto-fix completed"

# Run security scan (govulncheck)
security:
	@echo "🔒 Running security scan (govulncheck)..."
	@which govulncheck >/dev/null 2>&1 || (echo "📥 Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	@cd core && govulncheck ./...
	@echo "✅ Security scan completed"

# Full quality check (lint + test + security)
check: lint test security
	@echo "✅ All quality checks passed"

# ============================================
# Swagger Documentation
# ============================================

# Generate Swagger API documentation
swagger:
	@echo "📚 Generating Swagger API documentation..."
	@cd core && swag init -g internal/bootstrap/server.go -o docs
	@echo "✅ Swagger docs generated in core/docs/"
	@echo "💡 Access at: http://localhost:$${HTTP_PORT:-8080}/api/docs/"

# Alias for swagger
docs-api: swagger

# ============================================
# 5. Database (数据库迁移)
# ============================================

# Build base image locally (one-time setup or after go.mod changes)
build-base:
	@echo "🔨 Building apprun-base image with Go dependencies..."
	@echo "⏱️  This may take 5-8 minutes on first run"
	@docker build \
		-f docker/Dockerfile.base \
		-t ghcr.io/websoft9/apprun-base:latest \
		-t apprun-base:latest \
		--build-arg BUILD_DATE=$$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
		.
	@echo "✅ Base image built successfully"
	@echo "💡 App builds will now be 60-80% faster"
	@docker images apprun-base:latest

# Pull pre-built base image from registry (recommended for team)
pull-base:
	@echo "📥 Pulling pre-built base image from registry..."
	@docker pull ghcr.io/websoft9/apprun-base:latest || \
		(echo "⚠️  Failed to pull, will build locally" && $(MAKE) build-base)
	@echo "✅ Base image ready"

# Verify base image exists (helper target)
.check-base:
	@if ! docker image inspect apprun-base:latest >/dev/null 2>&1 && \
	   ! docker image inspect ghcr.io/websoft9/apprun-base:latest >/dev/null 2>&1; then \
		echo "⚠️  Base image not found, pulling from registry..."; \
		$(MAKE) pull-base; \
	fi

# ============================================
# 6. Docker (容器化)
# ============================================

# Build Docker images
docker-build:
	@echo "🔨 Building Docker images..."
	cd docker && docker compose build
	@echo "✅ Docker images built"

# Start Docker services
docker-up:
	@echo "🚀 Starting Docker services..."
	docker compose up -d
	@echo "✅ Docker services started"

# Stop Docker services
docker-down:
	@echo "🛑 Stopping Docker services..."
	docker compose down
	@echo "✅ Docker services stopped"

# View Docker logs
docker-logs:
	@echo "📋 Docker logs (Ctrl+C to exit)..."
	docker compose logs -f

# Clean Docker resources
docker-clean:
	@echo "🧹 Cleaning Docker resources..."
	@docker compose -f docker-compose.dev.yml down -v
	@docker compose -f docker-compose.yml down -v
	@echo "✅ Docker resources cleaned"

# Build base image with Go dependencies
docker-build-base:
	@echo "🔨 Building apprun-base image with Go dependencies..."
	@echo "⏱️  This may take 5-8 minutes on first run"
	@docker build \
		-f docker/Dockerfile.base \
		-t ghcr.io/websoft9/apprun-base:latest \
		-t apprun-base:latest \
		--build-arg BUILD_DATE=$$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
		.
	@echo "✅ Base image built successfully"
	@echo "� App builds will now be 60-80% faster"

# Pull pre-built base image from registry
docker-pull-base:
	@echo "📥 Pulling pre-built base image from registry..."
	@docker pull ghcr.io/websoft9/apprun-base:latest || (echo "⚠️  Failed to pull, will build locally" && $(MAKE) docker-build-base)
	@echo "✅ Base image ready"

# Backward compatibility
build-base: docker-build-base
pull-base: docker-pull-base
build-local: docker-build
clean-docker: docker-clean

# ============================================
# 7. Documentation (文档)
# ============================================

# Validate all Story documents
docs-validate:
	@echo "�🔍 Validating Story documents..."
	@for file in docs/sprint-artifacts/sprint-*/story-*.md; do \
		if [ -f "$$file" ]; then \
			./scripts/validate-story.sh "$$file" || exit 1; \
		fi \
	done
	@echo ""
	@echo "✅ All Story documents validated successfully"

# Sync global Stories index (legacy table in README)
story-sync:
	@echo "🔄 Syncing global Stories index..."
	@./scripts/sync-story-index.sh
	@echo "✅ Global Stories index synced"

# Generate story status index table (saves to story-index.md)
story-index:
	@./scripts/generate-story-index.py --format markdown
	@echo "📖 View: docs/sprint-artifacts/story-index.md"

# Show sprint status summary
sprint-status:
	@echo "📊 Sprint Status Summary (from sprint-status.yaml)"
	@./scripts/manage-sprint-status.py summary

# Update sprint status statistics
sprint-update:
	@echo "🔄 Updating statistics in sprint-status.yaml..."
	@./scripts/manage-sprint-status.py update-stats
	@echo "✅ Statistics updated"

# Show detailed sprint status report
sprint-summary:
	@echo "📊 Detailed Sprint Status Report"
	@echo ""
	@./scripts/manage-sprint-status.py summary
	@echo ""
	@echo "📋 Stories by Epic:"
	@for epic in epic-infrastructure epic-i18n epic-config epic-docs epic-auth epic-storage epic-functions; do \
		echo ""; \
		echo "🏗️  $$epic:"; \
		./scripts/manage-sprint-status.py list-stories --epic $$epic 2>/dev/null || true; \
	done

# Backward compatibility aliases
validate-stories: docs-validate
sync-index: story-sync
sprint-status-update: sprint-update
sprint-status-summary: sprint-summary
story-validate: docs-validate

# ============================================
# 8. Utilities (工具)
# ============================================

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf core/bin core/coverage.out core/coverage.html core/tmp core/*.log
	@find . -name "*.log" -type f -delete
	@echo "✅ Build artifacts cleaned"

# Deep clean (all artifacts + dependencies)
clean-all: clean app-clean deps-clean docker-clean
	@echo "✅ Complete cleanup done"

# Check Go toolchain version and required dependencies
check-deps:
	@echo "🔍 Checking dependencies..."
	@echo ""
	@echo "1️⃣ Checking Go toolchain..."
	@required_version=$$(grep "^go " core/go.mod | awk '{print $$2}'); \
	current_toolchain=$$(go env GOTOOLCHAIN); \
	if [ "$$current_toolchain" = "auto" ] || [ "$$current_toolchain" = "local" ]; then \
		echo "⚠️  GOTOOLCHAIN=$$current_toolchain may cause version mismatch"; \
		echo "💡 Setting GOTOOLCHAIN=go$$required_version..."; \
		go env -w GOTOOLCHAIN=go$$required_version; \
		echo "✅ GOTOOLCHAIN fixed to go$$required_version"; \
	else \
		echo "✅ GOTOOLCHAIN=$$current_toolchain (fixed version)"; \
	fi
	@echo ""
	@echo "2️⃣ Checking Atlas CLI (required for database migrations)..."
	@if command -v atlas >/dev/null 2>&1; then \
		echo "✅ Atlas CLI installed: $$(atlas version | head -1)"; \
	else \
		echo "❌ Atlas CLI not found!"; \
		echo ""; \
		echo "Atlas CLI is required for database migrations."; \
		echo "Install with:"; \
		echo "  curl -sSf https://atlasgo.sh | sh"; \
		echo ""; \
		echo "Or on macOS:"; \
		echo "  brew install ariga/tap/atlas"; \
		echo ""; \
		exit 1; \
	fi
	@echo ""
	@echo "3️⃣ Checking Docker (optional, but recommended)..."
	@if command -v docker >/dev/null 2>&1; then \
		echo "✅ Docker installed: $$(docker --version)"; \
	else \
		echo "⚠️  Docker not found (optional for dev environment)"; \
	fi
	@echo ""
	@echo "✅ All required dependencies are available"

# Install development tools
install:
	@echo "� Installing development tools..."
	@echo "Installing golangci-lint..."
	@which golangci-lint >/dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@echo "Installing govulncheck..."
	@which govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "Installing swag..."
	@which swag >/dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ Development tools installed"

# Show version information
version:
	@echo "AppRun BaaS Platform"
	@echo "Go version: $$(go version)"
	@echo "Docker version: $$(docker --version 2>/dev/null || echo 'not installed')"
	@echo "golangci-lint: $$(golangci-lint --version 2>/dev/null || echo 'not installed')"

# ============================================
# Backward Compatibility Aliases
# ============================================

# Legacy command aliases (deprecated, will be removed)
dev-up:
	@echo "⚠️  'make dev-up' is deprecated, use 'make deps-start' instead"
	@$(MAKE) deps-start

dev-down:
	@echo "⚠️  'make dev-down' is deprecated, use 'make deps-stop' instead"
	@$(MAKE) deps-stop

run-local:
	@echo "⚠️  'make run-local' is deprecated, use 'make app-start' instead"
	@$(MAKE) app-start

check-go-version:
	@echo "⚠️  'make check-go-version' is deprecated, use 'make check-deps' instead"
	@$(MAKE) check-deps

# Legacy Docker targets (keeping for now)
prod-up-local:
	@echo "🚀 Starting production-like environment locally..."
	@docker compose -f docker-compose.yml up -d
	@echo "✅ Local production environment started!"

prod-down-local:
	@echo "🛑 Stopping local production environment..."
	@docker compose -f docker-compose.yml down
	@echo "✅ Local production environment stopped"

test-local: docker-build
	@echo "� Running integration tests..."
	@docker compose -f docker-compose.yml up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 15
	@echo "🔍 Checking health..."
	@docker exec apprun-app wget -q -O- http://localhost:$${HTTP_PORT:-8080}/health || (echo "❌ Health check failed" && docker compose down && exit 1)
	@echo "✅ Integration tests passed!"
	@docker compose down

test-unit-setup:
	@./tests/scripts/unit-test-setup.sh

test-unit-run:
	@./tests/scripts/run-unit-tests.sh
