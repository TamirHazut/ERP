# Config Service Makefile

.PHONY: proto build run test test-coverage lint clean help

# Service configuration
SERVICE_NAME := config
SERVICE_PORT := 5002
BIN_DIR := ../../bin
CMD_DIR := ../../cmd/$(SERVICE_NAME)
INTERNAL_DIR := .

# Proto generation
PROTO_OUT := ../..
PROTO_DIR := $(INTERNAL_DIR)/proto
PROTO_COMMON := ../../proto/common
MODULE := erp.localhost

help: ## Show this help message
	@echo "Config Service - Available targets:"
	@echo ""
	@echo "  make proto          - Generate config service proto files"
	@echo "  make build          - Build config service"
	@echo "  make run            - Run config service (port $(SERVICE_PORT))"
	@echo "  make test           - Run config service tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter on config service"
	@echo "  make clean          - Clean build artifacts"
	@echo ""

proto: ## Generate config service proto files
	@echo "Generating config service proto files..."
	@if [ -f "$(PROTO_DIR)/config.proto" ]; then \
		protoc --go_out=$(PROTO_OUT) \
			--go_opt=module=$(MODULE) \
			--go-grpc_out=$(PROTO_OUT) \
			--go-grpc_opt=module=$(MODULE) \
			-I=../../proto \
			-I=$(PROTO_DIR) \
			$(PROTO_DIR)/*.proto; \
		echo "✓ Config proto files generated"; \
	else \
		echo "⚠ No config.proto file found, skipping..."; \
	fi

build: ## Build config service
	@echo "Building config service..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(SERVICE_NAME) $(CMD_DIR)
	@echo "✓ Config service built: $(BIN_DIR)/$(SERVICE_NAME)"

run: ## Run config service
	@echo "Starting config service on port $(SERVICE_PORT)..."
	@go run $(CMD_DIR)

test: ## Run config service tests
	@echo "Running config service tests..."
	@go test -v -count=1 ./...
	@echo "✓ Config tests complete"

test-coverage: ## Run tests with coverage
	@echo "Running config service tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"

lint: ## Run linter on config service
	@echo "Running linter on config service..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

clean: ## Clean build artifacts
	@echo "Cleaning config service artifacts..."
	@rm -f $(INTERNAL_DIR)/coverage.out $(INTERNAL_DIR)/coverage.html
	@echo "✓ Clean complete"

