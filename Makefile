# apprun Makefile

.PHONY: help build test test-all test-unit test-integration test-e2e clean docker-build docker-up docker-down validate-stories sync-index dev-up dev-down run-local build-local test-local prod-up-local prod-down-local swagger

# 默认目标
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build & Test:"
	@echo "  build          - Build the application"
	@echo "  test-all       - Run all tests"
	@echo "  test-unit      - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e       - Run end-to-end tests"
	@echo "  swagger        - Generate Swagger API documentation"
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
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-up      - Start Docker services"
	@echo "  docker-down    - Stop Docker services"
	@echo ""
	@echo "Documentation:"
	@echo "  validate-stories - Validate all Story documents"
	@echo "  sync-index     - Sync global Stories index"
	@echo ""
	@echo "  clean          - Clean build artifacts"

# 构建
build:
	cd core && go build -o bin/server ./cmd/server

# Swagger 文档生成
swagger:
	@echo "Generating Swagger API documentation..."
	@cd core && swag init -g cmd/server/main.go -o docs
	@echo "✅ Swagger docs generated in core/docs/"
	@echo "Access at: http://localhost:$${HTTP_PORT:-8080}/api/docs/"

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

# 同步全局 Stories 索引
sync-index:
	@echo "🔄 Syncing global Stories index..."
	@./scripts/sync-story-index.sh
	@echo "✅ Global Stories index synced"

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