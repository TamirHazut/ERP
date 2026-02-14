# Auth Service Makefile

.PHONY: help \
		build run lint test test-coverage clean \
		certs certs-clean \
		mocks mocks-clean

# Service configuration
SERVICE_NAME := auth
SERVICE_PORT := 5000
BIN_DIR := ./bin
CMD_DIR := ./cmd
INTERNAL_DIR := .

help: ## Show this help message
	@echo "$(SERVICE_NAME) Service - Available targets:"
	@echo ""
	@echo "  make build          		- Build $(SERVICE_NAME) service"
	@echo "  make run            		- Run $(SERVICE_NAME) service (port $(SERVICE_PORT))"
	@echo "  make lint           		- Run linter on $(SERVICE_NAME) service"
	@echo "  make test           		- Run $(SERVICE_NAME) service tests"
	@echo "  make test-coverage  		- Run tests with coverage"
	@echo "  make clean            		- Clean artifacts"
	@echo "  make certs          		- Generate mTLS certificates for $(SERVICE_NAME) service"
	@echo "  make certs-clean    		- Remove $(SERVICE_NAME) service certificates"
	@echo "	 make help			 		- Display this menu"
	@echo ""

build: ## Build auth service
	@echo "Building $(SERVICE_NAME) service..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(SERVICE_NAME) $(CMD_DIR)
	@echo "✓ $(SERVICE_NAME) service built: $(BIN_DIR)/$(SERVICE_NAME)"

run: ## Run auth service
	@echo "Starting $(SERVICE_NAME) service on port $(SERVICE_PORT)..."
	@go run $(CMD_DIR)

test: mocks ## Run auth service tests
	@echo "Running $(SERVICE_NAME) service tests..."
	@go test -v -count=1 ./...
	@echo "✓ $(SERVICE_NAME) tests complete"

test-coverage: mocks ## Run tests with coverage
	@echo "Running $(SERVICE_NAME) service tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"

lint: ## Run linter on auth service
	@echo "Running linter on $(SERVICE_NAME) service..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

clean: ## Clean build artifacts
	@echo "Cleaning $(SERVICE_NAME) service artifacts..."
	@rm -f $(INTERNAL_DIR)/coverage.out $(INTERNAL_DIR)/coverage.html $(BIN_DIR)/*
	@echo "✓ Clean complete"


# ============================================================================
# MOCKS
# ============================================================================

mocks: mocks-clean
	@echo "Generating $(SERVICE_NAME) mocks..."
	@go generate ./...
	@echo "✅ $(SERVICE_NAME) mocks generated successfully"

mocks-clean:
	@echo "Cleaning generated infra mocks..."
	@find . -name "mock_*.go" -type f -delete
	@echo "✅ Infra mocks cleaned"

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

certs: ## Generate mTLS certificates for auth service
	@echo "Creating certificates for $(SERVICE_NAME) service..."
	@if [ ! -f $(CA_CERT) ]; then \
		echo "❌ CA certificate not found. Run 'make certs-ca' from root directory first."; \
		exit 1; \
	fi
	@mkdir -p $(CERTS_DIR)
	@openssl genrsa -out $(SERVICE_KEY) 2048
	@openssl req -new -key $(SERVICE_KEY) -out $(SERVICE_CSR) \
		-subj "/C=US/ST=State/L=City/O=ERP System/OU=$(SERVICE_NAME) Service/CN=$(SERVICE_NAME).erp.localhost"
	@openssl x509 -req -in $(SERVICE_CSR) -CA $(CA_CERT) -CAkey $(CA_KEY) \
		-CAcreateserial -out $(SERVICE_CERT) -days $(CERT_DAYS) \
		-extfile <(printf "subjectAltName=DNS:$(SERVICE_NAME).erp.localhost,DNS:localhost,IP:127.0.0.1")
	@cp $(CA_CERT) $(SERVICE_CA_COPY)
	@rm -f $(SERVICE_CSR)
	@echo "✓ Service certificate created: $(SERVICE_CERT)"
	@echo "✓ Service private key created: $(SERVICE_KEY)"
	@echo "✓ CA certificate copied: $(SERVICE_CA_COPY)"

certs-clean: ## Remove auth service certificates
	@echo "Removing $(SERVICE_NAME) service certificates..."
	@rm -rf $(CERTS_DIR)
	@echo "✓ $(SERVICE_NAME) certificates removed"