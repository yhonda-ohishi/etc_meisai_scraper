# Makefile for ETC Meisai System

# Variables
APP_NAME := etc_meisai
DOCKER_IMAGE := $(APP_NAME):latest

# Test targets
.PHONY: test test-coverage test-coverage-html test-coverage-check test-unit test-integration clean-tests setup-tests generate-mocks ci-test

test:
	go test -v -race -timeout 60s ./src/...

test-coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./src/...
	@go tool cover -func=coverage.out

test-coverage-html:
	@echo "Generating HTML coverage report..."
	@go test -coverprofile=coverage.out ./src/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

test-coverage-check:
	@bash scripts/check-coverage.sh

# T013-B: Optimized test execution
test-fast: ## Run tests optimized for speed (< 30s target)
	@echo "$(GREEN)Running optimized tests...$(NC)"
	@# Run tests with optimal parallelism
	@go test -parallel 8 -count=1 ./... -timeout 30s

test-parallel: ## Run tests in parallel groups
	@echo "$(GREEN)Running parallel test groups...$(NC)"
	@if [ -f scripts/parallel-test.sh ]; then \
		bash scripts/parallel-test.sh; \
	else \
		echo "Generating parallel test script..."; \
		go run scripts/test-optimizer.go . coverage; \
		bash scripts/parallel-test.sh; \
	fi

test-profile: ## Profile test execution time
	@echo "$(YELLOW)Profiling test performance...$(NC)"
	@go run scripts/test-optimizer.go . coverage
	@echo "$(GREEN)Test profile generated in coverage/$(NC)"

# T012-E: Enhanced coverage targets with gates and thresholds
# Coverage thresholds
COVERAGE_THRESHOLD_STATEMENT = 95
COVERAGE_THRESHOLD_BRANCH = 90
COVERAGE_DIR = coverage

test-coverage:
	@mkdir -p $(COVERAGE_DIR)
	go test -coverprofile=$(COVERAGE_DIR)/coverage.raw -covermode=atomic -coverpkg=./src/... ./...
	@# Filter out excluded files (generated code, mocks, etc.)
	@grep -v -E '(pb\.go|pb\.gw\.go|_mock\.go|/mocks/|/vendor/|/migrations/)' \
		$(COVERAGE_DIR)/coverage.raw > $(COVERAGE_DIR)/coverage.filtered || cp $(COVERAGE_DIR)/coverage.raw $(COVERAGE_DIR)/coverage.filtered
	go tool cover -func=$(COVERAGE_DIR)/coverage.filtered

test-coverage-html: test-coverage
	go tool cover -html=$(COVERAGE_DIR)/coverage.filtered -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

test-unit:
	go test -v -race ./src/...

test-integration:
	go test -v ./tests/integration/...

clean-tests:
	find . -name "*_test.go" -delete
	rm -rf tests/
	rm -f coverage.out coverage.html

setup-tests:
	mkdir -p mocks tests/unit tests/integration tests/contract
	go get github.com/stretchr/testify/mock
	go get github.com/stretchr/testify/assert
	go get github.com/stretchr/testify/require

generate-mocks:
	@echo "Generating mocks..."
	@echo "Mocks will be created by individual test tasks"

ci-test:
	go test ./... -coverprofile=coverage.out -covermode=atomic -race
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$(echo "$$coverage < 100" | bc -l)" -eq 1 ]; then \
		echo "Coverage is below 100%: $$coverage%"; \
		exit 1; \
	fi
	@echo "Coverage check passed: $$coverage%"

# T012-E: Coverage gate enforcement
coverage-gate test-coverage-check: test-coverage
	@echo "$(GREEN)Checking coverage gates...$(NC)"
	@# Extract coverage percentage
	@coverage=$$(go tool cover -func=$(COVERAGE_DIR)/coverage.filtered | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $$coverage%"; \
	echo "Required coverage: $(COVERAGE_THRESHOLD_STATEMENT)%"; \
	if [ "$$(echo "$$coverage < $(COVERAGE_THRESHOLD_STATEMENT)" | bc)" -eq 1 ]; then \
		echo "$(RED)❌ Coverage $$coverage% is below threshold $(COVERAGE_THRESHOLD_STATEMENT)%$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ Coverage $$coverage% meets threshold $(COVERAGE_THRESHOLD_STATEMENT)%$(NC)"; \
	fi

# Coverage enforcement - strict checking
coverage-enforce: test-coverage
	@echo "$(YELLOW)Enforcing strict coverage requirements...$(NC)"
	@# Check for completely uncovered files
	@echo "Checking for uncovered files..."
	@uncovered=$$(go tool cover -func=$(COVERAGE_DIR)/coverage.filtered | grep -E "0.0%" | wc -l); \
	if [ $$uncovered -gt 0 ]; then \
		echo "$(RED)Found $$uncovered files with 0% coverage:$(NC)"; \
		go tool cover -func=$(COVERAGE_DIR)/coverage.filtered | grep -E "0.0%"; \
		exit 1; \
	else \
		echo "$(GREEN)No completely uncovered files$(NC)"; \
	fi
	@# Check for low coverage files
	@echo "Checking for files below 80% coverage..."
	@low_cov=$$(go tool cover -func=$(COVERAGE_DIR)/coverage.filtered | grep -E "[0-7][0-9]\.[0-9]%" | wc -l); \
	if [ $$low_cov -gt 5 ]; then \
		echo "$(RED)Error: $$low_cov files have coverage below 80%$(NC)"; \
		go tool cover -func=$(COVERAGE_DIR)/coverage.filtered | grep -E "[0-7][0-9]\.[0-9]%"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ All coverage gates passed$(NC)"

# Show detailed coverage report
coverage-detailed: test-coverage
	@echo "$(YELLOW)Generating detailed coverage analysis...$(NC)"
	@# Run advanced coverage analysis if available
	@if [ -f scripts/coverage-advanced.go ]; then \
		go run scripts/coverage-advanced.go $(COVERAGE_DIR)/coverage.filtered . || true; \
	fi
	@if [ -f scripts/coverage-report.go ]; then \
		go run scripts/coverage-report.go $(COVERAGE_DIR)/coverage.filtered . $(COVERAGE_DIR) || true; \
	fi
	@echo "$(GREEN)Detailed reports generated in $(COVERAGE_DIR)/$(NC)"
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