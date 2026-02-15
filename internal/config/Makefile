# Config Module Makefile

.PHONY: help \
		build run lint test test-coverage clean \
		certs certs-clean \
		mocks mocks-clean \
		test-functional-setup test-functional test-functional-clean

# Module configuration
MODULE_NAME := config
SERVICE_PORT := 5002
BIN_DIR := ./bin
CMD_DIR := ./cmd
INTERNAL_DIR := .

help: ## Show this help message
	@echo "$(MODULE_NAME) Module - Available targets:"
	@echo ""
	@echo "  make build          		- Build $(MODULE_NAME) module"
	@echo "  make run            		- Run $(MODULE_NAME) module (port $(SERVICE_PORT))"
	@echo "  make lint           		- Run linter on $(MODULE_NAME) module"
	@echo "  make test           		- Run $(MODULE_NAME) module tests"
	@echo "  make test-coverage  		- Run tests with coverage"
	@echo "  make clean            		- Clean artifacts"
	@echo "  make certs          		- Generate mTLS certificates for $(MODULE_NAME) module"
	@echo "  make certs-clean    		- Remove $(MODULE_NAME) module certificates"
	@echo "	 make help			 		- Display this menu"
	@echo ""

build: ## Build config module
	@echo "Building $(MODULE_NAME) module..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(MODULE_NAME) $(CMD_DIR)
	@echo "✓ $(MODULE_NAME) module built: $(BIN_DIR)/$(MODULE_NAME)"

run: ## Run config module
	@echo "Starting $(MODULE_NAME) module on port $(SERVICE_PORT)..."
	@go run $(CMD_DIR)

test: mocks ## Run config module tests
	@echo "Running $(MODULE_NAME) module tests..."
	@go test -v -count=1 ./...
	@echo "✓ $(MODULE_NAME) tests complete"

test-coverage: mocks ## Run tests with coverage
	@echo "Running $(MODULE_NAME) module tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"

lint: ## Run linter on config module
	@echo "Running linter on $(MODULE_NAME) module..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

clean: ## Clean build artifacts
	@echo "Cleaning $(MODULE_NAME) module artifacts..."
	@rm -f $(INTERNAL_DIR)/coverage.out $(INTERNAL_DIR)/coverage.html $(BIN_DIR)/*
	@echo "✓ Clean complete"

# ============================================================================
# MOCKS
# ============================================================================

mocks: mocks-clean
	@echo "Generating $(MODULE_NAME) mocks..."
	@go generate ./...
	@echo "✅ $(MODULE_NAME) mocks generated successfully"

mocks-clean:
	@echo "Cleaning generated $(MODULE_NAME) mocks..."
	@find . -name "mock_*.go" -type f -delete
	@echo "✅ $(MODULE_NAME) mocks cleaned"

# ============================================================================
# CERTIFICATE GENERATION (mTLS)
# ============================================================================

# Certificate paths
CERTS_DIR := resources/certs
CA_DIR := ../infra/resources/certs/ca
CA_KEY := $(CA_DIR)/ca-key.pem
CA_CERT := $(CA_DIR)/ca-cert.pem
SERVICE_KEY := $(CERTS_DIR)/key.pem
SERVICE_CSR := $(CERTS_DIR)/csr.pem
SERVICE_CERT := $(CERTS_DIR)/cert.pem
SERVICE_CA_COPY := $(CERTS_DIR)/ca-cert.pem
CERT_DAYS := 365

certs: ## Generate mTLS certificates for config module
	@echo "Creating certificates for $(MODULE_NAME) module..."
	@if [ ! -f $(CA_CERT) ]; then \
		echo "❌ CA certificate not found. Run 'make certs-ca' from root directory first."; \
		exit 1; \
	fi
	@mkdir -p $(CERTS_DIR)
	@openssl genrsa -out $(SERVICE_KEY) 2048
	@openssl req -new -key $(SERVICE_KEY) -out $(SERVICE_CSR) \
		-subj "/C=US/ST=State/L=City/O=ERP System/OU=$(MODULE_NAME) Module/CN=$(MODULE_NAME).erp.localhost"
	@openssl x509 -req -in $(SERVICE_CSR) -CA $(CA_CERT) -CAkey $(CA_KEY) \
		-CAcreateserial -out $(SERVICE_CERT) -days $(CERT_DAYS) \
		-extfile <(printf "subjectAltName=DNS:$(MODULE_NAME).erp.localhost,DNS:localhost,IP:127.0.0.1")
	@cp $(CA_CERT) $(SERVICE_CA_COPY)
	@rm -f $(SERVICE_CSR)
	@echo "✓ Service certificate created: $(SERVICE_CERT)"
	@echo "✓ Service private key created: $(SERVICE_KEY)"
	@echo "✓ CA certificate copied: $(SERVICE_CA_COPY)"

certs-clean: ## Remove config module certificates
	@echo "Removing $(MODULE_NAME) module certificates..."
	@rm -rf $(CERTS_DIR)
	@echo "✓ $(MODULE_NAME) certificates removed"

# ============================================================================
# FUNCTIONAL TESTS
# ============================================================================
test-functional-setup:
	$(MAKE) -C ../infra test-functional-setup

test-functional: test-functional-setup
	@echo "Setting up Python functional test environment..."
	@cd ./functional && LOG_LEVEL=DEBUG python -m pytest -v -s --tb=short
	@echo "✓ Python dependencies installed"

test-functional-clean: ## Clean functional test artifacts
	@echo "Cleaning functional test artifacts..."
	@find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	@find . -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true
	@find . -type f -name "*.pyc" -delete 2>/dev/null || true
	@echo "✓ Functional test artifacts cleaned"