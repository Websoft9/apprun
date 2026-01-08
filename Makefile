# apprun Makefile

.PHONY: help build test test-all test-unit test-integration test-e2e clean docker-build docker-up docker-down validate-stories sync-index sprint-status sprint-status-update sprint-status-summary dev-up dev-down run-local build-local build-base pull-base test-local prod-up-local prod-down-local swagger i18n i18n-extract i18n-merge lint lint-fix check-go-version

# Go 版本检查（防止工具链不匹配）
check-go-version:
	@echo "🔍 Checking Go toolchain version..."
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

# 默认目标
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build & Test:"
	@echo "  build          - Generate code and build the application (includes i18n, swagger)"
	@echo "  generate       - Generate Ent ORM code only"
	@echo "  i18n           - Extract and merge translation keys"
	@echo "  i18n-extract   - Extract translation keys from code"
	@echo "  i18n-merge     - Merge extracted keys to translation files"
	@echo "  config-example - Generate config.example from module registry (Story 10a)"
	@echo "  lint           - Run golangci-lint (same as CI)"
	@echo "  lint-fix       - Run golangci-lint with auto-fix"
	@echo "  test-all       - Run all tests"
	@echo "  test-unit      - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e       - Run end-to-end tests"
	@echo "  swagger        - Generate Swagger API documentation"
	@echo ""
	@echo "Database Migration (Using Docker):"
	@echo "  migrate-diff   - Generate migration from schema changes (requires NAME=xxx)"
	@echo "  migrate-apply  - Apply pending migrations to dev database"
	@echo "  migrate-status - Show migration status"
	@echo "  migrate-baseline - Set baseline version for existing database"
	@echo "  migrate-validate - Validate migration files"
	@echo "  migrate-lint   - Lint migrations for common issues"
	@echo "  migrate-hash   - Generate migration checksums (atlas.sum)"
	@echo ""
	@echo "Development Environment (Story 1):"
	@echo "  dev-up         - Start dev dependencies (postgres + redis)"
	@echo "  dev-down       - Stop dev dependencies"
	@echo "  run-local      - Run app locally with go run"
	@echo "  build-local    - Build Docker image locally"
	@echo "  test-local     - Run integration tests with local image"
	@echo "  prod-up-local  - Start production-like environment locally"
	@echo "  prod-down-local- Stop local production environment"
	@echo ""
	@echo "Docker:"
	@echo "  build-base     - Build base image with Go dependencies (for faster builds)"
	@echo "  pull-base      - Pull pre-built base image from registry"
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-up      - Start Docker services"
	@echo "  docker-down    - Stop Docker services"
	@echo ""
	@echo "Documentation & Sprint Management:"
	@echo "  validate-stories - Validate all Story documents"
	@echo "  sync-index     - Sync global Stories index (legacy table in README)"
	@echo "  sprint-status  - Show sprint status summary from sprint-status.yaml"
	@echo "  sprint-status-update - Update sprint-status.yaml statistics"
	@echo "  sprint-status-summary - Show detailed sprint status report"
	@echo ""
	@echo "Utilities:"
	@echo "  check-go-version - Check and fix Go toolchain version"
	@echo ""
	@echo "  clean          - Clean build artifacts"

# 构建（正确顺序：生成代码 -> 提取翻译 -> 生成文档 -> 编译）
build: generate i18n swagger
	cd core && go build -o bin/server ./cmd/server

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
# Database Migration Commands (Using Docker)
# ============================================

# Atlas Docker image
ATLAS_IMAGE := arigaio/atlas:latest
POSTGRES_URL := postgres://apprun:dev_password_123@host.docker.internal:5432/apprun_dev?sslmode=disable

# Generate migration from schema changes
# Usage: make migrate-diff NAME=add_project_table
migrate-diff:
ifndef NAME
	$(error NAME is required. Usage: make migrate-diff NAME=add_project_table)
endif
	@echo "📝 Generating migration: $(NAME)..."
	@docker run --rm \
		-v $(PWD)/core:/app \
		-w /app \
		$(ATLAS_IMAGE) \
		migrate diff $(NAME) \
		--dir "file://migrations" \
		--to "ent://ent/schema" \
		--dev-url "docker://postgres/15/dev?search_path=public"
	@echo "✅ Migration generated! Please review:"
	@ls -la core/migrations/*.sql | tail -1
	@echo ""
	@echo "⚠️  IMPORTANT: Review the generated SQL before committing!"

# Apply pending migrations to dev database
migrate-apply:
	@echo "🚀 Applying migrations to dev database..."
	@docker run --rm \
		-v $(PWD)/core:/app \
		-w /app \
		--add-host=host.docker.internal:host-gateway \
		$(ATLAS_IMAGE) \
		migrate apply \
		--dir "file://migrations" \
		--url "$(POSTGRES_URL)"
	@echo "✅ Migrations applied!"

# Show migration status
migrate-status:
	@echo "📊 Migration status:"
	@docker run --rm \
		-v $(PWD)/core:/app \
		-w /app \
		--add-host=host.docker.internal:host-gateway \
		$(ATLAS_IMAGE) \
		migrate status \
		--dir "file://migrations" \
		--url "$(POSTGRES_URL)"

# Validate migrations
migrate-validate:
	@echo "✅ Validating migrations..."
	@docker run --rm \
		-v $(PWD)/core:/app \
		-w /app \
		$(ATLAS_IMAGE) \
		migrate validate \
		--dir "file://migrations" \
		--dev-url "docker://postgres/15/test?search_path=public"

# Lint migrations for common issues
migrate-lint:
	@echo "🔍 Linting migrations..."
	@docker run --rm \
		-v $(PWD)/core:/app \
		-w /app \
		$(ATLAS_IMAGE) \
		migrate lint \
		--dir "file://migrations" \
		--dev-url "docker://postgres/15/test?search_path=public" \
		--latest 1
	@echo "✅ Migration lint completed"

# Generate migration checksum (atlas.sum)
migrate-hash:
	@echo "🔐 Generating migration checksums..."
	@docker run --rm \
		-v $(PWD)/core:/app \
		-w /app \
		$(ATLAS_IMAGE) \
		migrate hash \
		--dir "file://migrations"
	@echo "✅ Checksums generated in core/migrations/atlas.sum"

# Set baseline version for existing database
migrate-baseline:
	@echo "📍 Setting migration baseline..."
	@docker run --rm \
		-v $(PWD)/core:/app \
		-w /app \
		--add-host=host.docker.internal:host-gateway \
		$(ATLAS_IMAGE) \
		migrate set 002 \
		--dir "file://migrations" \
		--url "$(POSTGRES_URL)"
	@echo "✅ Baseline set to version 002"
	@echo "💡 Now you can run 'make migrate-apply' to apply new migrations"

# Swagger 文档生成
swagger:
	@echo "Generating Swagger API documentation..."
	@cd core && swag init -g cmd/server/main.go -o docs
	@echo "✅ Swagger docs generated in core/docs/"
	@echo "Access at: http://localhost:$${HTTP_PORT:-8080}/api/docs/"

# ============================================
# Code Quality (Story 5 - CI/CD)
# ============================================

# Run linter (same configuration as CI)
lint:
	@echo "🔍 Running golangci-lint..."
	@which golangci-lint > /dev/null 2>&1 || { \
		echo "❌ golangci-lint not installed"; \
		echo "📥 Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
		echo "💡 Then add $$(go env GOPATH)/bin to your PATH"; \
		exit 1; \
	}
	@echo "🔧 Verifying golangci-lint configuration..."
	@cd core && ( \
		for i in 1 2 3; do \
			output=$$(golangci-lint config verify --config=.golangci.yml 2>&1); \
			exit_code=$$?; \
			if [ $$exit_code -eq 0 ]; then \
				echo "✅ Configuration verified"; \
				break; \
			else \
				if echo "$$output" | grep -q -i "timeout\|deadline\|network"; then \
					if [ $$i -lt 3 ]; then \
						echo "⚠️  Network timeout on attempt $$i/3, retrying in 2 seconds..."; \
						sleep 2; \
					else \
						echo "❌ Config verification failed after 3 attempts due to network timeout"; \
						echo "💡 Please check your internet connection or try again later"; \
						exit 1; \
					fi \
				else \
					echo "❌ Config verification failed:"; \
					echo "$$output"; \
					exit 1; \
				fi \
			fi \
		done \
	)
	@echo ""
	@echo "🔍 Running lint checks..."
	@cd core && golangci-lint run --timeout=5m --config=.golangci.yml
	@echo "✅ Linting completed"
	@echo ""
	@echo "🔒 Running govulncheck (dependency vulnerability scan)..."
	@which govulncheck > /dev/null 2>&1 || { \
		echo "📥 Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	}
	@cd core && govulncheck ./...
	@echo "✅ Vulnerability check completed"

# Run linter with auto-fix
lint-fix:
	@echo "🔧 Running golangci-lint with auto-fix..."
	@which golangci-lint > /dev/null 2>&1 || { \
		echo "❌ golangci-lint not installed"; \
		echo "📥 Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
		echo "💡 Then add $$(go env GOPATH)/bin to your PATH"; \
		exit 1; \
	}
	@cd core && golangci-lint run --timeout=5m --config=.golangci.yml --fix
	@echo "✅ Linting with fixes completed"

# ============================================
# Testing
# ============================================

# 测试
test-all: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	cd core && go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage summary:"
	@cd core && go tool cover -func=coverage.out

test-unit-html: test-unit
	@echo ""
	@echo "Generating HTML coverage report..."
	cd core && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: core/coverage.html"

test-unit-setup:
	@tests/scripts/unit-test-setup.sh

test-unit-run:
	@tests/scripts/run-unit-tests.sh

test-integration:
	@echo "Running integration tests..."
	@tests/scripts/setup-test-db.sh
	@tests/integration/config/test-api.sh
	@tests/integration/config/test-priority.sh
	@tests/scripts/cleanup.sh

test-e2e:
	@echo "Running E2E tests..."
	@echo "E2E tests not implemented yet"

# Docker
docker-build:
	cd docker && docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# ============================================
# Docker Base Image Commands (Story 1 Enhancement)
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

# 清理
clean:
	cd core && rm -rf bin/ coverage.out coverage.html
	find . -name "*.log" -delete

# 开发环境
dev: docker-up
	@echo "Development environment started"
	@echo "App: http://localhost:$${HTTP_PORT:-8080}"
	@echo "Config API: http://localhost:$${HTTP_PORT:-8080}/config"

# 快速测试
test-config: test-unit
	@echo "Running config module tests..."
	@tests/scripts/setup-test-db.sh
	@tests/integration/config/test-api.sh

# 验证 Story 文档
validate-stories:
	@echo "🔍 Validating Story documents..."
	@for file in docs/sprint-artifacts/sprint-*/story-*.md; do \
		if [ -f "$$file" ]; then \
			./scripts/validate-story.sh "$$file" || exit 1; \
		fi \
	done
	@echo ""
	@echo "✅ All Story documents validated successfully"
	@tests/scripts/cleanup.sh

# 同步全局 Stories 索引 (Legacy table in README)
sync-index:
	@echo "🔄 Syncing global Stories index..."
	@./scripts/sync-story-index.sh
	@echo "✅ Global Stories index synced"

# Sprint Status Management (sprint-status.yaml)
sprint-status:
	@echo "📊 Sprint Status Summary (from sprint-status.yaml)"
	@./scripts/manage-sprint-status.py summary

sprint-status-update:
	@echo "🔄 Updating statistics in sprint-status.yaml..."
	@./scripts/manage-sprint-status.py update-stats
	@echo "✅ Statistics updated"

sprint-status-summary:
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

# ============================================
# Story 1: Development Environment Commands
# ============================================

# Start development dependencies only (postgres + redis)
dev-up:
	@echo "🚀 Starting development dependencies..."
	@docker compose -f docker-compose.dev.yml up -d
	@echo "✅ Development dependencies ready!"
	@echo ""
	@echo "📊 Services:"
	@echo "  PostgreSQL: localhost:5432 (user: apprun, password: dev_password_123)"
	@echo "  Redis:      localhost:6379"
	@echo ""
	@echo "💡 Next step: Run your app locally"
	@echo "   go run core/cmd/server/main.go"

# Stop development dependencies
dev-down:
	@echo "🛑 Stopping development dependencies..."
	@docker compose -f docker-compose.dev.yml down
	@echo "✅ Development dependencies stopped"

# Run app locally (assumes dev-up is running)
run-local:
	@echo "🏃 Running app locally..."
	@echo "📌 Make sure dependencies are running: make dev-up"
	@echo "📝 Using development database configuration"
	@echo ""
	cd core && \
		DATABASE_USER=apprun \
		DATABASE_PASSWORD=dev_password_123 \
		DATABASE_DB_NAME=apprun_dev \
		go run ./cmd/server/main.go

# Build Docker image locally
build-local:
	@echo "🔨 Building Docker image locally..."
	@$(MAKE) .check-base
	@docker build -t apprun:local -f docker/Dockerfile .
	@echo "✅ Docker image built: apprun:local"
	@echo ""
	@docker images apprun:local

# Run integration tests with local build
test-local: build-local
	@echo "🧪 Running integration tests..."
	@docker compose -f docker-compose.local.yml up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 15
	@echo "🔍 Checking health..."
	@docker exec apprun-app-local wget -q -O- http://localhost:$${HTTP_PORT:-8080}/health || (echo "❌ Health check failed" && docker compose -f docker-compose.local.yml down && exit 1)
	@echo "✅ Integration tests passed!"
	@docker compose -f docker-compose.local.yml down

# Start production-like environment locally
prod-up-local:
	@echo "🚀 Starting production-like environment locally..."
	@docker compose -f docker-compose.local.yml up -d
	@echo "✅ Local production environment started!"
	@echo ""
	@echo "🔗 Access:"
	@echo "   HTTP:  http://localhost:$${HTTP_PORT:-8080}"
	@echo "   HTTPS: https://localhost:$${HTTPS_PORT:-8443}"
	@echo ""
	@echo "📊 View logs:"
	@echo "   docker compose -f docker-compose.local.yml logs -f"

# Stop local production environment
prod-down-local:
	@echo "🛑 Stopping local production environment..."
	@docker compose -f docker-compose.local.yml down
	@echo "✅ Local production environment stopped"

# Clean all Docker resources
clean-docker:
	@echo "🧹 Cleaning Docker resources..."
	@docker compose -f docker-compose.dev.yml down -v
	@docker compose -f docker-compose.local.yml down -v
	@docker rmi apprun:local 2>/dev/null || true
	@echo "✅ Docker resources cleaned"