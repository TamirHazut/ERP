# Auth Service Makefile

.PHONY: proto build run test test-coverage lint clean help

# Service configuration
SERVICE_NAME := auth
SERVICE_PORT := 5000
BIN_DIR := ../../bin
CMD_DIR := ../../cmd/$(SERVICE_NAME)
INTERNAL_DIR := .

# Proto generation
PROTO_OUT := ../..
PROTO_DIR := $(INTERNAL_DIR)/proto
PROTO_COMMON := ../../proto/common
MODULE := erp.localhost

help: ## Show this help message
	@echo "Auth Service - Available targets:"
	@echo ""
	@echo "  make proto          - Generate auth service proto files"
	@echo "  make build          - Build auth service"
	@echo "  make run            - Run auth service (port $(SERVICE_PORT))"
	@echo "  make test           - Run auth service tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter on auth service"
	@echo "  make clean          - Clean build artifacts"
	@echo ""

proto: ## Generate auth service proto files
	@echo "Generating auth service proto files..."
	@if [ -f "$(PROTO_DIR)/auth.proto" ]; then \
		protoc --go_out=$(PROTO_OUT) \
			--go_opt=module=$(MODULE) \
			--go-grpc_out=$(PROTO_OUT) \
			--go-grpc_opt=module=$(MODULE) \
			-I=../../proto \
			-I=$(PROTO_DIR) \
			$(PROTO_DIR)/*.proto; \
		echo "✓ Auth proto files generated"; \
	else \
		echo "⚠ No auth.proto file found, skipping..."; \
	fi

build: ## Build auth service
	@echo "Building auth service..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(SERVICE_NAME) $(CMD_DIR)
	@echo "✓ Auth service built: $(BIN_DIR)/$(SERVICE_NAME)"

run: ## Run auth service
	@echo "Starting auth service on port $(SERVICE_PORT)..."
	@go run $(CMD_DIR)

test: ## Run auth service tests
	@echo "Running auth service tests..."
	@go test -v -count=1 ./...
	@echo "✓ Auth tests complete"

test-coverage: ## Run tests with coverage
	@echo "Running auth service tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"

lint: ## Run linter on auth service
	@echo "Running linter on auth service..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

clean: ## Clean build artifacts
	@echo "Cleaning auth service artifacts..."
	@rm -f $(INTERNAL_DIR)/coverage.out $(INTERNAL_DIR)/coverage.html
	@echo "✓ Clean complete"

