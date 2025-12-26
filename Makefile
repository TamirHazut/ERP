# ERP System Root Makefile
# Delegates to service-specific Makefiles

.PHONY: proto proto-auth proto-config proto-core proto-common \
        build build-auth build-config build-core \
        run-auth run-config run-core \
        test test-coverage lint clean tidy help \
        docker-up docker-down docker-logs docker-ps

# Proto generation output directories
PROTO_OUT := internal
PROTO_COMMON := proto/common

# Go module path
MODULE := erp.localhost

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
	@echo "  make run-auth       - Run auth service"
	@echo "  make run-config     - Run config service"
	@echo "  make run-core       - Run core service"
	@echo ""
	@echo "Test & Quality:"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter on all services"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up      - Start MongoDB and Redis containers"
	@echo "  make docker-down    - Stop and remove containers"
	@echo "  make docker-logs    - View container logs"
	@echo "  make docker-ps      - List running containers"
	@echo ""
	@echo "Utilities:"
	@echo "  make tidy           - Run go mod tidy"
	@echo "  make clean          - Clean build artifacts from all services"
	@echo ""
	@echo "Service-specific help:"
	@echo "  make -C internal/auth help    - Show auth service targets"
	@echo "  make -C internal/config help  - Show config service targets"
	@echo "  make -C internal/core help    - Show core service targets"

# ============================================================================
# PROTO GENERATION TARGETS
# ============================================================================

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
	@$(MAKE) -C internal/auth proto

proto-config: ## Generate config service proto files
	@$(MAKE) -C internal/config proto

proto-core: ## Generate core service proto files
	@$(MAKE) -C internal/core proto

# ============================================================================
# BUILD TARGETS
# ============================================================================

build: build-auth build-config build-core ## Build all services
	@echo "✓ All services built"

build-auth: ## Build auth service
	@$(MAKE) -C internal/auth build

build-config: ## Build config service
	@$(MAKE) -C internal/config build

build-core: ## Build core service
	@$(MAKE) -C internal/core build

# ============================================================================
# RUN TARGETS
# ============================================================================

run-auth: ## Run auth service
	@$(MAKE) -C internal/auth run

run-config: ## Run config service
	@$(MAKE) -C internal/config run

run-core: ## Run core service
	@$(MAKE) -C internal/core run

# ============================================================================
# TEST TARGETS
# ============================================================================

test: ## Run all tests from all services
	@echo "Running all tests..."
	@$(MAKE) -C internal/auth test
	@$(MAKE) -C internal/config test
	@$(MAKE) -C internal/core test
	@echo "✓ All tests complete"

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# ============================================================================
# QUALITY TARGETS
# ============================================================================

lint: ## Run linter on all services
	@echo "Running linter on all services..."
	@$(MAKE) -C internal/auth lint
	@$(MAKE) -C internal/config lint
	@$(MAKE) -C internal/core lint
	@echo "✓ All linting complete"

# ============================================================================
# UTILITY TARGETS
# ============================================================================

tidy: ## Run go mod tidy
	@echo "Running go mod tidy..."
	@go mod tidy
	@echo "✓ Dependencies updated"

clean: ## Clean build artifacts from all services
	@echo "Cleaning all build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@$(MAKE) -C internal/auth clean
	@$(MAKE) -C internal/config clean
	@$(MAKE) -C internal/core clean
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
