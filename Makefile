# ERP System Makefile

.PHONY: proto proto-auth proto-config proto-core proto-common \
        build build-auth build-config build-core \
        run-auth run-config run-core \
        test test-coverage lint clean tidy help \
        docker-up docker-down docker-logs docker-ps

# Proto generation output directories
PROTO_OUT := internal
PROTO_COMMON := proto/common
PROTO_AUTH := internal/auth/proto
PROTO_CONFIG := internal/config/proto
PROTO_CORE := internal/core/proto

# Go module path
MODULE := erp.localhost

# Service ports
AUTH_PORT := 5000
CONFIG_PORT := 5002
CORE_PORT := 5001

# Binary output directory
BIN_DIR := bin

help: ## Show this help message
	@echo "ERP System - Available targets:"
	@echo ""
	@echo "Proto Generation:"
	@echo "  make proto          - Generate all proto files"
	@echo "  make proto-common   - Generate common proto files"
	@echo "  make proto-auth     - Generate auth service proto files"
	@echo "  make proto-config   - Generate config service proto files"
	@echo "  make proto-core     - Generate core service proto files"
	@echo ""
	@echo "Build:"
	@echo "  make build          - Build all services"
	@echo "  make build-auth     - Build auth service"
	@echo "  make build-config   - Build config service"
	@echo "  make build-core     - Build core service"
	@echo ""
	@echo "Run:"
	@echo "  make run-auth       - Run auth service (port $(AUTH_PORT))"
	@echo "  make run-config     - Run config service (port $(CONFIG_PORT))"
	@echo "  make run-core       - Run core service (port $(CORE_PORT))"
	@echo ""
	@echo "Test & Quality:"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up      - Start MongoDB and Redis containers"
	@echo "  make docker-down    - Stop and remove containers"
	@echo "  make docker-logs    - View container logs"
	@echo "  make docker-ps      - List running containers"
	@echo ""
	@echo "Utilities:"
	@echo "  make tidy           - Run go mod tidy"
	@echo "  make clean          - Clean build artifacts"

proto: proto-common proto-auth proto-config proto-core ## Generate all proto files

proto-common: ## Generate common proto files
	@echo "Generating common proto files..."
	@if [ -f "$(PROTO_COMMON)/common.proto" ]; then \
		protoc --go_out=$(PROTO_OUT) \
			--go_opt=module=$(MODULE) \
			--go-grpc_out=$(PROTO_OUT) \
			--go-grpc_opt=module=$(MODULE) \
			-I=proto \
			$(PROTO_COMMON)/common.proto; \
		echo "✓ Common proto files generated"; \
	else \
		echo "⚠ No common.proto file found, skipping..."; \
	fi

proto-auth: ## Generate auth service proto files
	@echo "Generating auth service proto files..."
	@if [ -f "$(PROTO_AUTH)/auth.proto" ]; then \
		protoc --go_out=$(PROTO_OUT) \
			--go_opt=module=$(MODULE) \
			--go-grpc_out=$(PROTO_OUT) \
			--go-grpc_opt=module=$(MODULE) \
			-I=proto \
			-I=$(PROTO_AUTH) \
			$(PROTO_AUTH)/*.proto; \
		echo "✓ Auth proto files generated"; \
	else \
		echo "⚠ No auth.proto file found, skipping..."; \
	fi

proto-config: ## Generate config service proto files
	@echo "Generating config service proto files..."
	@if [ -f "$(PROTO_CONFIG)/config.proto" ]; then \
		protoc --go_out=$(PROTO_OUT) \
			--go_opt=module=$(MODULE) \
			--go-grpc_out=$(PROTO_OUT) \
			--go-grpc_opt=module=$(MODULE) \
			-I=proto \
			-I=$(PROTO_CONFIG) \
			$(PROTO_CONFIG)/*.proto; \
		echo "✓ Config proto files generated"; \
	else \
		echo "⚠ No config.proto file found, skipping..."; \
	fi

proto-core: ## Generate core service proto files
	@echo "Generating core service proto files..."
	@if [ -f "$(PROTO_CORE)/core.proto" ]; then \
		protoc --go_out=$(PROTO_OUT) \
			--go_opt=module=$(MODULE) \
			--go-grpc_out=$(PROTO_OUT) \
			--go-grpc_opt=module=$(MODULE) \
			-I=proto \
			-I=$(PROTO_CORE) \
			$(PROTO_CORE)/*.proto; \
		echo "✓ Core proto files generated"; \
	else \
		echo "⚠ No core.proto file found, skipping..."; \
	fi

tidy: ## Run go mod tidy
	@echo "Running go mod tidy..."
	@go mod tidy
	@echo "✓ Dependencies updated"

# ============================================================================
# BUILD TARGETS
# ============================================================================

build: ## Build all services
	@echo "Building all services..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/ ./...
	@echo "✓ Build complete"

build-auth: ## Build auth service
	@echo "Building auth service..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/auth ./cmd/auth
	@echo "✓ Auth service built"

build-config: ## Build config service
	@echo "Building config service..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/config ./cmd/config
	@echo "✓ Config service built"

build-core: ## Build core service
	@echo "Building core service..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/core ./cmd/core
	@echo "✓ Core service built"

# ============================================================================
# RUN TARGETS
# ============================================================================

run-auth: ## Run auth service
	@echo "Starting auth service on port $(AUTH_PORT)..."
	@go run ./cmd/auth

run-config: ## Run config service
	@echo "Starting config service on port $(CONFIG_PORT)..."
	@go run ./cmd/config

run-core: ## Run core service
	@echo "Starting core service on port $(CORE_PORT)..."
	@go run ./cmd/core

# ============================================================================
# TEST TARGETS
# ============================================================================

test: ## Run all tests from all services
	@echo "Running all tests..."
	@go test -v -count=1 ./...
	@echo "✓ All tests complete"

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# ============================================================================
# QUALITY TARGETS
# ============================================================================

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

# ============================================================================
# UTILITY TARGETS
# ============================================================================

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

# ============================================================================
# DOCKER TARGETS
# ============================================================================

docker-up: ## Start MongoDB and Redis containers
	@echo "Starting Docker containers..."
	@docker compose up -d
	@echo "✓ Containers started"
	@echo ""
	@echo "Services:"
	@echo "  MongoDB: mongodb://root:secret@localhost:27017"
	@echo "  Redis:   redis://:supersecretredis@localhost:6379"

docker-down: ## Stop and remove containers
	@echo "Stopping Docker containers..."
	@docker compose down
	@echo "✓ Containers stopped"

docker-logs: ## View container logs
	@docker compose logs -f

docker-ps: ## List running containers
	@docker compose ps

