# Makefile for ETC Meisai System

# Variables
APP_NAME := etc_meisai
DOCKER_IMAGE := $(APP_NAME):latest

# Test targets
.PHONY: test test-coverage test-coverage-html test-coverage-check

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out -coverpkg=./... ./...
	go tool cover -func=coverage.out

test-coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-check: test-coverage
	@echo "Checking coverage threshold (100%)..."
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$(echo "$$coverage < 100" | bc)" -eq 1 ]; then \
		echo "❌ Coverage is below 100%: $$coverage%"; \
		exit 1; \
	else \
		echo "✅ Coverage meets threshold: $$coverage%"; \
	fi
GO := go
GOFLAGS := -v
CGO_ENABLED := 1

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: help
help: ## Display this help message
	@echo "$(GREEN)Available commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

.PHONY: all
all: clean test build ## Clean, test, and build the application

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod verify

.PHONY: tidy
tidy: ## Clean up go.mod and go.sum
	@echo "$(GREEN)Tidying modules...$(NC)"
	$(GO) mod tidy

.PHONY: build
build: ## Build the server binary
	@echo "$(GREEN)Building server...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -o bin/server ./cmd/server

.PHONY: build-migrate
build-migrate: ## Build the migration tool
	@echo "$(GREEN)Building migration tool...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -o bin/migrate ./cmd/migrate

.PHONY: build-all
build-all: build build-migrate ## Build all binaries

.PHONY: run
run: ## Run the server locally
	@echo "$(GREEN)Starting server...$(NC)"
	$(GO) run ./cmd/server

.PHONY: migrate
migrate: ## Run database migrations
	@echo "$(GREEN)Running migrations...$(NC)"
	$(GO) run ./cmd/migrate/main.go migrate

.PHONY: migrate-down
migrate-down: ## Reset database
	@echo "$(RED)Resetting database...$(NC)"
	$(GO) run ./cmd/migrate/main.go reset

.PHONY: migrate-status
migrate-status: ## Show migration status
	@echo "$(GREEN)Migration status:$(NC)"
	$(GO) run ./cmd/migrate/main.go status

.PHONY: test
test: ## Run all tests
	@echo "$(GREEN)Running tests...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(GOFLAGS) ./...

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "$(GREEN)Running unit tests...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(GOFLAGS) ./src/...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(GOFLAGS) ./tests/integration/

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: lint
lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Installing...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GO) vet ./...

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -rf coverage.out coverage.html
	rm -rf temp/ downloads/
	$(GO) clean -cache

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run: ## Run application in Docker
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run -d \
		--name $(APP_NAME) \
		-p 8080:8080 \
		-v $(PWD)/data:/root/data \
		-v $(PWD)/logs:/root/logs \
		$(DOCKER_IMAGE)

.PHONY: docker-stop
docker-stop: ## Stop Docker container
	@echo "$(YELLOW)Stopping Docker container...$(NC)"
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

.PHONY: docker-compose-up
docker-compose-up: ## Start services with docker-compose
	@echo "$(GREEN)Starting services with docker-compose...$(NC)"
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop services with docker-compose
	@echo "$(YELLOW)Stopping services...$(NC)"
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## Show docker-compose logs
	docker-compose logs -f

.PHONY: dev
dev: deps migrate run ## Setup and run for development

.PHONY: prod
prod: clean test build docker-build ## Build for production

.PHONY: install
install: build ## Install the binary to GOPATH/bin
	@echo "$(GREEN)Installing binary...$(NC)"
	$(GO) install ./cmd/server

.PHONY: proto
proto: ## Generate protobuf files (if proto files exist)
	@echo "$(GREEN)Generating protobuf files...$(NC)"
	@if [ -d "proto" ]; then \
		protoc --go_out=. --go-grpc_out=. proto/*.proto; \
	else \
		echo "$(YELLOW)No proto directory found$(NC)"; \
	fi

.PHONY: seed
seed: ## Seed test data
	@echo "$(GREEN)Seeding test data...$(NC)"
	$(GO) run ./cmd/migrate/main.go seed

.PHONY: health
health: ## Check server health
	@echo "$(GREEN)Checking server health...$(NC)"
	@curl -s http://localhost:8080/health | jq '.' || echo "$(RED)Server not running$(NC)"

.PHONY: api-test
api-test: ## Test API endpoints
	@echo "$(GREEN)Testing API endpoints...$(NC)"
	@curl -s http://localhost:8080/ping && echo " - Ping: OK" || echo " - Ping: FAILED"
	@curl -s http://localhost:8080/health | grep -q "healthy" && echo " - Health: OK" || echo " - Health: FAILED"

# Default target
.DEFAULT_GOAL := help