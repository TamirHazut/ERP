# Config Service Makefile

.PHONY: build run test test-coverage lint clean help certs certs-clean

# Service configuration
SERVICE_NAME := config
SERVICE_PORT := 5002
BIN_DIR := ../../bin
CMD_DIR := ../../cmd/$(SERVICE_NAME)
INTERNAL_DIR := .

help: ## Show this help message
	@echo "Config Service - Available targets:"
	@echo ""
	@echo "  make build          - Build config service"
	@echo "  make run            - Run config service (port $(SERVICE_PORT))"
	@echo "  make test           - Run config service tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter on config service"
	@echo "  make certs          - Generate mTLS certificates for config service"
	@echo "  make certs-clean    - Remove config service certificates"
	@echo "  make clean          - Clean build artifacts"
	@echo ""

build: ## Build config service
	@echo "Building config service..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(SERVICE_NAME) $(CMD_DIR)
	@echo "✓ Config service built: $(BIN_DIR)/$(SERVICE_NAME)"

run: ## Run config service
	@echo "Starting config service on port $(SERVICE_PORT)..."
	@go run $(CMD_DIR)

test: mocks ## Run config service tests
	@echo "Running config service tests..."
	@go test -v -count=1 ./...
	@echo "✓ Config tests complete"

test-coverage: mocks ## Run tests with coverage
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

.PHONY: mocks
mocks: mocks-clean
	@echo "Generating config mocks..."
	@go generate ./...
	@echo "✅ Config mocks generated successfully"

.PHONY: mocks-clean
mocks-clean:
	@echo "Cleaning generated config mocks..."
	@find . -path "*/mocks/mock_*.go" -type f -delete
	@echo "✅ Config mocks cleaned"

# ============================================================================
# CERTIFICATE GENERATION (mTLS)
# ============================================================================

# Certificate paths
CERTS_DIR := resources/certs
CA_DIR := ../../resources/certs/ca
CA_KEY := $(CA_DIR)/ca-key.pem
CA_CERT := $(CA_DIR)/ca-cert.pem
SERVICE_KEY := $(CERTS_DIR)/key.pem
SERVICE_CSR := $(CERTS_DIR)/csr.pem
SERVICE_CERT := $(CERTS_DIR)/cert.pem
SERVICE_CA_COPY := $(CERTS_DIR)/ca-cert.pem
CERT_DAYS := 365

certs: ## Generate mTLS certificates for config service
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

certs-clean: ## Remove config service certificates
	@echo "Removing $(SERVICE_NAME) service certificates..."
	@rm -rf $(CERTS_DIR)
	@echo "✓ $(SERVICE_NAME) certificates removed"