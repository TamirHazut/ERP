# infra Module Makefile

.PHONY: help \
		lint test test-coverage test-clean \
		certs certs-clean \
		generate generate-clean

# Module configuration
MODULE_NAME := infra
INTERNAL_DIR := .

help: ## Show this help message
	@echo "$(MODULE_NAME) - Available targets:"
	@echo ""
	@echo "  make lint           		- Run linter on $(MODULE_NAME) module"
	@echo "  make test           		- Run $(MODULE_NAME) module tests"
	@echo "  make test-coverage  		- Run tests with coverage"
	@echo "  make test-clean     		- Clean tests artifacts"
	@echo "	 make certs		 			- Generate all required certificates"
	@echo "	 make certs-clean		 	- Clean all certificates"
	@echo "	 make generate		 		- Generate all required files"
	@echo "	 make generate-clean		- Clean all generated files"
	@echo "	 make help			 		- Display this menu"
	@echo ""

test: generate ## Run infra module tests
	@echo "Running $(MODULE_NAME) module tests..."
	@go test -v -count=1 ./...
	@echo "✓ $(MODULE_NAME) tests complete"

test-coverage: generate ## Run tests with coverage
	@echo "Running $(MODULE_NAME) module tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"

lint: ## Run linter on infra module
	@echo "Running linter on $(MODULE_NAME) module..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

test-clean: ## Clean build artifacts
	@echo "Cleaning $(MODULE_NAME) module artifacts..."
	@rm -f $(INTERNAL_DIR)/coverage.out $(INTERNAL_DIR)/coverage.html
	@echo "✓ Clean complete"


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

certs-clean: ## Remove all certificates (CA and module certificates)
	@echo "Removing all certificates..."
	@rm -rf resources/certs
	@echo "✓ All certificates removed"


# ============================================================================
# MOCKS
# ============================================================================

mocks:
	@echo "Generating $(MODULE_NAME) mocks..."
	@go generate ./...
	@echo "✅ $(MODULE_NAME) mocks generated successfully"

mocks-clean:
	@echo "Cleaning generated $(MODULE_NAME) mocks..."
	@find . -name "mock_*.go" -type f -delete
	@echo "✅ $(MODULE_NAME) mocks cleaned"

# ============================================================================
# Kafka
# ============================================================================

generate-topics:
	python ./resources/scripts/topics_generator.py \
		--input ./resources/topics.yaml \
		--go ./go/model/event/topics.go \
		--py ./python/model/event/topics.py \
		--js ./js/model/event/topics.js

generate-topics-clean:
	rm -rf 

# ============================================================================
# PROTO
# ============================================================================

MODULE := erp.localhost/infra
INFRA_BASE := .
PROTO_IN := $(INFRA_BASE)/proto
THIRD_PARTY := $(PROTO_IN)/third_party

# --- 1. DEFINE OUTPUT DIRECTORIES FIRST ---
# (Must come before flags that use them)
GENERATED_GO_OUT     := $(INFRA_BASE)/go
GENERATED_PYTHON_OUT := $(INFRA_BASE)/python/model
GENERATED_JS_OUT     := $(INFRA_BASE)/js/model

PROTOC_COMMON_FLAGS := -I=$(PROTO_IN) -I=$(THIRD_PARTY)

# --- 2. DEFINE FLAGS ---
# Step 1: Generate Base Go Files
GO_GEN_FLAGS := --go_out=$(GENERATED_GO_OUT) --go_opt=module=$(MODULE) \
                --go-grpc_out=$(GENERATED_GO_OUT) --go-grpc_opt=module=$(MODULE)

# Step 2: Inject Tags 
GO_TAG_FLAGS := --gotag_out=module=$(MODULE):$(INFRA_BASE)

# --- 3. THE MACRO ---
define generate_proto
    WORD=$$(basename $(1)); \
    SERVICE_DIR="$(PROTO_IN)/$$WORD/v1"; \
    if [ -d "$$SERVICE_DIR" ]; then \
        PROTO_FILES=$$(find $$SERVICE_DIR -maxdepth 2 -name "*.proto" -type f); \
        if [ -n "$$PROTO_FILES" ]; then \
            echo "--- Generating [$$WORD] ---"; \
			echo "   > Go..."; \
            protoc $(PROTOC_COMMON_FLAGS) $(GO_GEN_FLAGS) $$PROTO_FILES; \
			\
			echo "   > Go Tags..."; \
			cd $(GENERATED_GO_OUT) && \
            protoc -I=../$(PROTO_IN) -I=../$(THIRD_PARTY) $(GO_TAG_FLAGS) \
            $$(echo $$PROTO_FILES | sed 's|$(PROTO_IN)|../$(PROTO_IN)|g'); \
			\
			echo "   > Python..."; \
			cd $(CURDIR); \
			mkdir -p $(GENERATED_PYTHON_OUT)/$$WORD; \
            protoc $(PROTOC_COMMON_FLAGS) --python_out=$(GENERATED_PYTHON_OUT)/$$WORD $$PROTO_FILES; \
            touch $(GENERATED_PYTHON_OUT)/$$WORD/__init__.py; \
			\
            echo "   > JS..."; \
            mkdir -p $(GENERATED_JS_OUT)/$$WORD; \
            protoc $(PROTOC_COMMON_FLAGS) --js_out=import_style=commonjs,binary:$(GENERATED_JS_OUT)/$$WORD $$PROTO_FILES; \
        fi; \
    else \
        echo "Skipping $$WORD: directory $$SERVICE_DIR not found"; \
    fi
endef

proto:
	@if [ -z "$(MODULES)" ]; then echo "Error: No modules specified"; exit 1; fi
	@$(foreach module,$(MODULES), \
		$(call generate_proto,$(module)); \
	)
	@echo "✓ All proto files generated successfully"

proto-clean: ## Surgically remove ONLY generated proto files
	@echo "Cleaning generated files safely..."
	@# Remove Go generated files
	@find $(GENERATED_GO_OUT)/model -type f \( -name "*.pb.go" -o -name "*_grpc.pb.go" \) -delete
	@# Remove Python generated files
	@find $(GENERATED_PYTHON_OUT) -type f -name "*_pb2.py" -delete
	@# Remove JS generated files
	@find $(GENERATED_JS_OUT) -type f -name "*_pb.js" -delete
	@echo "✓ Generated files removed. Manual files preserved."

# ============================================================================
# GENERATED
# ============================================================================

generate: generate-clean proto generate-topics mocks

generate-clean: proto-clean generate-topics-clean mocks-clean 