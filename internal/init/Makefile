# init Module Makefile

.PHONY: help \
		build run lint test test-coverage clean \

# Module configuration
MODULE_NAME := init
BIN_DIR := ./bin
CMD_DIR := ./cmd
INTERNAL_DIR := .

help: ## Show this help message
	@echo "$(MODULE_NAME) Module - Available targets:"
	@echo ""
	@echo "  make build          		- Build $(MODULE_NAME) module"
	@echo "  make run          	 		- Run $(MODULE_NAME) module"
	@echo "  make lint           		- Run linter on $(MODULE_NAME) module"
	@echo "  make test-coverage  		- Run tests with coverage"
	@echo "  make clean            		- Clean artifacts"
	@echo "	 make help			 		- Display this menu"
	@echo ""

build: ## Build init module
	@echo "Building $(MODULE_NAME) module..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(MODULE_NAME) $(CMD_DIR)
	@chmod +x $(BIN_DIR)/$(MODULE_NAME)
	@echo "✓ Init module built: $(BIN_DIR)/$(MODULE_NAME)"

run: ## Run init module
	@echo "Starting $(MODULE_NAME) module..."
	@go run $(CMD_DIR)

test: mocks ## Run init module tests
	@echo "Running $(MODULE_NAME) module tests..."
	@go test -v -count=1 ./...
	@echo "✓ Init tests complete"

test-coverage: mocks ## Run tests with coverage
	@echo "Running $(MODULE_NAME) module tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"


lint: ## Run linter on init module
	@echo "Running linter on $(MODULE_NAME) module..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

clean: ## Clean build artifacts
	@echo "Cleaning $(MODULE_NAME) module artifacts..."
	@rm -f $(INTERNAL_DIR)/coverage.out $(INTERNAL_DIR)/coverage.html $(BIN_DIR)/*
	@echo "✓ Clean complete"