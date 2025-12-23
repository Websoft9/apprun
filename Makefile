# apprun Makefile

.PHONY: help build test test-all test-unit test-integration test-e2e clean docker-build docker-up docker-down

# 默认目标
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  test-all       - Run all tests"
	@echo "  test-unit      - Run unit tests"
	@echo "  test-unit-html - Run unit tests with HTML coverage report"
	@echo "  test-unit-setup- Setup unit test environment"
	@echo "  test-unit-run  - Run unit tests via script"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e       - Run end-to-end tests"
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-up      - Start Docker services"
	@echo "  docker-down    - Stop Docker services"
	@echo "  clean          - Clean build artifacts"

# 构建
build:
	cd core && go build -o bin/server ./cmd/server

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
	cd docker && docker compose up -d

docker-down:
	cd docker && docker compose down

# 清理
clean:
	cd core && rm -rf bin/ coverage.out coverage.html
	find . -name "*.log" -delete

# 开发环境
dev: docker-up
	@echo "Development environment started"
	@echo "App: http://localhost:8080"
	@echo "Config API: http://localhost:8080/config"

# 快速测试
test-config: test-unit
	@echo "Running config module tests..."
	@tests/scripts/setup-test-db.sh
	@tests/integration/config/test-api.sh
	@tests/scripts/cleanup.sh