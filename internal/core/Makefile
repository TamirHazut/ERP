# Core Service Makefile

.PHONY: proto build run test test-coverage lint clean help

# Service configuration
SERVICE_NAME := core
SERVICE_PORT := 5001
BIN_DIR := ../../bin
CMD_DIR := ../../cmd/$(SERVICE_NAME)
INTERNAL_DIR := .

help: ## Show this help message
	@echo "Core Service - Available targets:"
	@echo ""
	@echo "  make proto          - Generate core service proto files"
	@echo "  make build          - Build core service"
	@echo "  make run            - Run core service (port $(SERVICE_PORT))"
	@echo "  make test           - Run core service tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter on core service"
	@echo "  make clean          - Clean build artifacts"
	@echo ""

proto: ## Generate core service proto files
	@bash ../../scripts/generate-proto.sh core

build: ## Build core service
	@echo "Building core service..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(SERVICE_NAME) $(CMD_DIR)
	@echo "✓ Core service built: $(BIN_DIR)/$(SERVICE_NAME)"

run: ## Run core service
	@echo "Starting core service on port $(SERVICE_PORT)..."
	@go run $(CMD_DIR)

test: ## Run core service tests
	@echo "Running core service tests..."
	@go test -v -count=1 ./...
	@echo "✓ Core tests complete"

test-coverage: ## Run tests with coverage
	@echo "Running core service tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"

lint: ## Run linter on core service
	@echo "Running linter on core service..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

clean: ## Clean build artifacts
	@echo "Cleaning core service artifacts..."
	@rm -f $(INTERNAL_DIR)/coverage.out $(INTERNAL_DIR)/coverage.html
	@echo "✓ Clean complete"

