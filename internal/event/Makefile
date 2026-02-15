# event Module Makefile
# TODO: This module will use Python + Node.js components

.PHONY: help \
		lint test test-coverage clean \
		certs certs-clean

# Module configuration
MODULE_NAME := event
INTERNAL_DIR := .

help: ## Show this help message
	@echo "$(MODULE_NAME) Module - Available targets:"
	@echo ""
	@echo "  make lint           		- Run linter on $(MODULE_NAME) module (TODO: implement)"
	@echo "  make test           		- Run $(MODULE_NAME) module tests (TODO: implement)"
	@echo "  make test-coverage  		- Run tests with coverage (TODO: implement)"
	@echo "  make clean     			- Clean tests artifacts (TODO: implement)"
	@echo "	 make certs		 			- Generate all required certificates"
	@echo "	 make certs-clean		 	- Clean all certificates"
	@echo "	 make help			 		- Display this menu"
	@echo ""

test: ## Run event module tests (TODO: implement Python/Node.js tests)
	@echo "⚠️  $(MODULE_NAME) module tests not implemented yet"
	@echo "TODO: Implement Python + Node.js component tests"

test-coverage: ## Run tests with coverage (TODO)
	@echo "⚠️  $(MODULE_NAME) module test coverage not implemented yet"
	@echo "TODO: Implement coverage for Python + Node.js components"

lint: ## Run linter on event module (TODO: implement)
	@echo "⚠️  $(MODULE_NAME) module linting not implemented yet"
	@echo "TODO: Run pylint for Python, eslint for Node.js"

clean: ## Clean artifacts (TODO)
	@echo "⚠️  $(MODULE_NAME) module clean not implemented yet"
	@echo "TODO: Clean Python/Node.js artifacts"

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

certs: ## Generate mTLS certificates for event module
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

certs-clean: ## Remove event module certificates
	@echo "Removing $(MODULE_NAME) module certificates..."
	@rm -rf $(CERTS_DIR)
	@echo "✓ $(MODULE_NAME) certificates removed"
