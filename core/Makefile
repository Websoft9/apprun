.PHONY: help build run test clean docker-build docker-up docker-down migrate lint fmt

# Variables
APP_NAME=apprun-core
DOCKER_COMPOSE=docker-compose
GO=go

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the Go binary
	@echo "Building $(APP_NAME)..."
	@$(GO) build -o bin/$(APP_NAME) ./cmd/server

run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	@$(GO) run ./cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	@$(GO) test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(GO) test -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out

fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GO) fmt ./...

lint: ## Lint Go code
	@echo "Linting code..."
	@golangci-lint run

mod-download: ## Download Go dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download

mod-tidy: ## Tidy Go dependencies
	@echo "Tidying dependencies..."
	@$(GO) mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):latest .

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	@$(DOCKER_COMPOSE) up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) down

docker-logs: ## View logs from all services
	@$(DOCKER_COMPOSE) logs -f

docker-logs-app: ## View logs from apprun-core only
	@$(DOCKER_COMPOSE) logs -f apprun-core

docker-restart: ## Restart all services
	@echo "Restarting services..."
	@$(DOCKER_COMPOSE) restart

migrate: ## Run database migrations
	@echo "Running migrations..."
	@$(GO) run ./cmd/server/main.go migrate

dev: ## Start development environment
	@echo "Starting development environment..."
	@$(DOCKER_COMPOSE) up -d postgres nats temporal jaeger
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Starting application..."
	@$(GO) run ./cmd/server/main.go

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest

.DEFAULT_GOAL := help
