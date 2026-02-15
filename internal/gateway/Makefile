# gateway Module Makefile
# TODO: This module will use Node.js + GraphQL (implement npm targets)

.PHONY: help \
		build run lint test test-coverage clean \
		certs certs-clean \
		mocks mocks-clean

# Module configuration
MODULE_NAME := gateway
SERVICE_PORT := 4000
# BIN_DIR and CMD_DIR not applicable for Node.js modules
INTERNAL_DIR := .

help: ## Show this help message
	@echo "$(MODULE_NAME) Module - Available targets:"
	@echo ""
	@echo "  make build          		- Build $(MODULE_NAME) module (TODO: npm build)"
	@echo "  make run            		- Run $(MODULE_NAME) module (port $(SERVICE_PORT), TODO: npm start)"
	@echo "  make lint           		- Run linter on $(MODULE_NAME) module (TODO: npm lint)"
	@echo "  make test           		- Run $(MODULE_NAME) module tests (TODO: npm test)"
	@echo "  make test-coverage  		- Run tests with coverage (TODO)"
	@echo "  make clean            		- Clean artifacts (TODO: npm clean)"
	@echo "  make certs          		- Generate mTLS certificates for $(MODULE_NAME) module"
	@echo "  make certs-clean    		- Remove $(MODULE_NAME) module certificates"
	@echo "	 make help			 		- Display this menu"
	@echo ""

build: ## Build gateway module (TODO: implement npm build)
	@echo "⚠️  $(MODULE_NAME) module build not implemented yet (will use Node.js + GraphQL)"
	@echo "TODO: npm install && npm run build"

run: ## Run gateway module (TODO: implement npm start)
	@echo "⚠️  $(MODULE_NAME) module run not implemented yet (will use Node.js + GraphQL)"
	@echo "TODO: npm start (port $(SERVICE_PORT))"

test: ## Run gateway module tests (TODO: implement npm test)
	@echo "⚠️  $(MODULE_NAME) module tests not implemented yet"
	@echo "TODO: npm test"

test-coverage: ## Run tests with coverage (TODO)
	@echo "⚠️  $(MODULE_NAME) module test coverage not implemented yet"
	@echo "TODO: npm run test:coverage"

lint: ## Run linter on gateway module (TODO: implement npm lint)
	@echo "⚠️  $(MODULE_NAME) module linting not implemented yet"
	@echo "TODO: npm run lint"

clean: ## Clean build artifacts (TODO)
	@echo "⚠️  $(MODULE_NAME) module clean not implemented yet"
	@echo "TODO: rm -rf node_modules dist"

# ============================================================================
# MOCKS (TODO: implement for Node.js/TypeScript)
# ============================================================================

mocks: mocks-clean
	@echo "⚠️  $(MODULE_NAME) mocks not implemented yet"
	@echo "TODO: Implement mock generation for Node.js/TypeScript"

mocks-clean:
	@echo "⚠️  $(MODULE_NAME) mock cleanup not implemented yet"

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

certs: ## Generate mTLS certificates for gateway module
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

certs-clean: ## Remove gateway module certificates
	@echo "Removing $(MODULE_NAME) module certificates..."
	@rm -rf $(CERTS_DIR)
	@echo "✓ $(MODULE_NAME) certificates removed"