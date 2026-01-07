# infra Service Makefile

.PHONY: build test test-coverage lint clean help mocks mocks-clean

# Service configuration
BIN_DIR := ../../bin
CMD_DIR := ../../cmd/$(SERVICE_NAME)
INTERNAL_DIR := .

help: ## Show this help message
	@echo "Infra Service - Available targets:"
	@echo ""
	@echo "  make build          - Build infra service"
	@echo "  make test           - Run infra service tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter on infra service"
	@echo "  make certs          - Generate mTLS certificates for infra service"
	@echo "  make certs-clean    - Remove infra service certificates"
	@echo "  make clean          - Clean build artifacts"
	@echo ""

build: ## Build infra service
	@echo "Building infra service..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(SERVICE_NAME) $(CMD_DIR)
	@echo "✓ Infra service built: $(BIN_DIR)/$(SERVICE_NAME)"

test: mocks ## Run infra service tests
	@echo "Running infra service tests..."
	@go test -v -count=1 ./...
	@echo "✓ Infra tests complete"

test-coverage: mocks ## Run tests with coverage
	@echo "Running infra service tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"

lint: ## Run linter on infra service
	@echo "Running linter on infra service..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

clean: ## Clean build artifacts
	@echo "Cleaning infra service artifacts..."
	@rm -f $(INTERNAL_DIR)/coverage.out $(INTERNAL_DIR)/coverage.html
	@echo "✓ Clean complete"

.PHONY: mocks
mocks: mocks-clean
	@echo "Generating infra mocks..."
	@go generate ./...
	@echo "✅ Infra mocks generated successfully"

.PHONY: mocks-clean
mocks-clean:
	@echo "Cleaning generated infra mocks..."
	@find . -path "*/mocks/mock_*.go" -type f -delete
	@echo "✅ Infra mocks cleaned"