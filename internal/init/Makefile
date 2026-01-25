# init Service Makefile

.PHONY: build test test-coverage run lint clean help mocks mocks-clean certs certs-clean

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

test: mocks ## Run init service tests
	@echo "Running init service tests..."
	@go test -v -count=1 ./...
	@echo "✓ Init tests complete"

test-coverage: mocks ## Run tests with coverage
	@echo "Running init service tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"

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

mocks: mocks-clean
	@echo "Generating infra mocks..."
	@go generate ./...
	@echo "✅ Init mocks generated successfully"

mocks-clean:
	@echo "Cleaning generated infra mocks..."
	@find . -path "*/mocks/mock_*.go" -type f -delete
	@echo "✅ Init mocks cleaned"

certs: certs-clean ## Create CA and certificates for all services
	@echo "No certs in init module"

certs-clean: ## Remove all certificates (CA and service certificates)
	@echo "No certs in init module"