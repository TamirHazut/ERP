# infra Service Makefile

.PHONY: build test test-coverage lint clean help mocks mocks-clean

# Service configuration
SERVICE_NAME := infra
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

mocks: mocks-clean
	@echo "Generating infra mocks..."
	@go generate ./...
	@echo "✅ Infra mocks generated successfully"

mocks-clean:
	@echo "Cleaning generated infra mocks..."
	@find . -path "*/mocks/mock_*.go" -type f -delete
	@echo "✅ Infra mocks cleaned"


# ============================================================================
# CERTIFICATE GENERATION (mTLS)
# ============================================================================

# Certificate configuration
CA_DIR := resources/certs/ca
CA_KEY := $(CA_DIR)/ca-key.pem
CA_CERT := $(CA_DIR)/ca-cert.pem
CA_DAYS := 3650
CERT_DAYS := 365

certs: certs-clean ## Create CA and certificates for all services
	@echo "Creating Certificate Authority (CA)..."
	@mkdir -p $(CA_DIR)
	@if [ ! -f $(CA_CERT) ]; then \
		echo "Creating root CA certificate..."; \
		openssl genrsa -out $(CA_KEY) 4096; \
		openssl req -new -x509 -days $(CA_DAYS) -key $(CA_KEY) -out $(CA_CERT) \
			-subj "/C=US/ST=State/L=City/O=ERP System/OU=Certificate Authority/CN=ERP Root CA"; \
		echo "✓ CA certificate created: $(CA_CERT)"; \
		echo "✓ CA private key created: $(CA_KEY)"; \
		echo ""; \
		echo "⚠️  IMPORTANT: Keep $(CA_KEY) secure and never commit to version control!"; \
		echo ""; \
	else \
		echo "✓ CA certificate already exists: $(CA_CERT)"; \
	fi

certs-clean: ## Remove all certificates (CA and service certificates)
	@echo "Removing all certificates..."
	@rm -rf resources/certs
	@for service in $(SERVICES); do \
		$(MAKE) -C internal/$$service certs-clean; \
	done
	@echo "✓ All certificates removed"