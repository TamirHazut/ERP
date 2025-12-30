# DB Makefile

.PHONY: test test-coverage lint clean help mocks mocks-clean

help: ## Show this help message
	@echo "DB - Available targets:"
	@echo ""
	@echo "  make test           - Run db tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter on db"
	@echo "  make mocks          - Generate mocks"
	@echo "  make mocks-clean    - Clean generated mocks"
	@echo ""


test: mocks ## Run db tests
	@echo "Running db tests..."
	@go test -v -count=1 ./...
	@echo "✓ Db tests complete"

test-coverage: mocks ## Run tests with coverage
	@echo "Running db tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"

lint: ## Run linter on db
	@echo "Running linter on db..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"


.PHONY: mocks
mocks: mocks-clean
	@echo "Generating db mocks..."
	@go generate ./...
	@echo "✅ Db mocks generated successfully"

.PHONY: mocks-clean
mocks-clean:
	@echo "Cleaning generated db mocks..."
	@find . -path "*/mocks/mock_*.go" -type f -delete
	@echo "✅ Db mocks cleaned"
