# init Service Makefile

.PHONY: build run lint clean help

# Service configuration
SERVICE_NAME := init
BIN_DIR := ../../bin
CMD_DIR := ../../cmd/$(SERVICE_NAME)
INTERNAL_DIR := .

help: ## Show this help message
	@echo "Init Service - Available targets:"
	@echo ""
	@echo "  make build          - Build init service"
	@echo "  make run          	 - Run init service"
	@echo "  make lint           - Run linter on init service"
	@echo "  make clean          - Clean build artifacts"
	@echo ""

build: ## Build init service
	@echo "Building init service..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(SERVICE_NAME) $(CMD_DIR)
	@echo "✓ Init service built: $(BIN_DIR)/$(SERVICE_NAME)"

run: ## Run init service
	@echo "Starting init service..."
	@go run $(CMD_DIR)

lint: ## Run linter on init service
	@echo "Running linter on init service..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

clean: ## Clean build artifacts
	@echo "Cleaning init service artifacts..."
	@rm -f $(INTERNAL_DIR)/coverage.out $(INTERNAL_DIR)/coverage.html
	@echo "✓ Clean complete"