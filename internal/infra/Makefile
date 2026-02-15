# infra Module Makefile

.PHONY: help \
		lint test test-coverage clean \
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
	@echo "  make clean     			- Clean tests artifacts"
	@echo "	 make certs		 			- Generate all required certificates"
	@echo "	 make certs-clean		 	- Clean all certificates"
	@echo "	 make generate		 		- Generate all required files"
	@echo "	 make generate-clean		- Clean all generated files"
	@echo "	 make help			 		- Display this menu"
	@echo ""

test: generate ## Run infra module tests
	@echo "Running $(MODULE_NAME) module tests..."
	@go test -v -count=1 ./go/...
	@echo "✓ $(MODULE_NAME) tests complete"

test-coverage: generate ## Run tests with coverage
	@echo "Running $(MODULE_NAME) module tests with coverage..."
	@go test -v -coverprofile=$(INTERNAL_DIR)/coverage.out ./go/...
	@go tool cover -html=$(INTERNAL_DIR)/coverage.out -o $(INTERNAL_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(INTERNAL_DIR)/coverage.html"

lint: ## Run linter on infra module
	@echo "Running linter on $(MODULE_NAME) module..."
	@golangci-lint run ./go/...
	@echo "✓ Linting complete"

clean: ## Clean artifacts
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

generate-mocks:
	@echo "Generating $(MODULE_NAME) mocks..."
	@go generate ./go/...
	@echo "✅ $(MODULE_NAME) mocks generated successfully"

generate-mocks-clean:
	@echo "Cleaning generated $(MODULE_NAME) mocks..."
	@find . -name "mock_*.go" -type f -delete
	@echo "✅ $(MODULE_NAME) mocks cleaned"

# ============================================================================
# Kafka
# ============================================================================

generate-topics:
	@echo "Generating Kafka topics"
	@python ./resources/scripts/topics_generator.py \
		--input ./resources/topics.yaml \
		--go ./go/model/event/topics.go \
		--py ./python/lib/model/event/topics.py \
		--js ./js/model/event/topics.js
	@echo "✅ Kafka topics generated successfully"

generate-topics-clean:
	@echo "Cleaning generated Kafka topics..."
	@rm -f ./go/model/event/topics.go ./python/lib/model/event/topics.py ./js/model/event/topics.js
	@echo "✅ Kafka topics cleaned"

# ============================================================================
# PROTO
# ============================================================================

MODULE := erp.localhost/infra
INFRA_BASE := .
PROTO_IN := $(INFRA_BASE)/proto
THIRD_PARTY := $(PROTO_IN)/third_party

# --- 1. DEFINE OUTPUT DIRECTORIES FIRST ---
# (Must come before flags that use them)
PYTHON_LIB_DIR 	 := python/lib
PYTHON_MODEL_DIR 	 := $(PYTHON_LIB_DIR)/model
GENERATED_GO_OUT     := $(INFRA_BASE)/go
GENERATED_PYTHON_OUT := $(INFRA_BASE)/$(PYTHON_MODEL_DIR)
GENERATED_JS_OUT     := $(INFRA_BASE)/js/model


# --- NOTE FOR WINDOWS USERS ---
# JavaScript gRPC code generation requires grpc_tools_node_protoc, which has
# a broken native plugin on Windows. This Makefile automatically uses WSL
# (Windows Subsystem for Linux) when running on Windows.
#
# WSL Setup (one-time):
#   1. Install WSL: wsl --install
#   2. Install Node.js in WSL: (in WSL) sudo apt update && sudo apt install nodejs npm
#   3. Install grpc-tools in WSL: (in WSL) npm install -g grpc-tools
#   4. Verify: wsl grpc_tools_node_protoc --version
#
# After setup, 'make generate-proto' will work automatically on Windows.
# Linux/macOS users don't need WSL - grpc_tools_node_protoc works natively.
# Detect operating system for gRPC plugin selection
# Windows (Cygwin/MSYS) requires WSL for working grpc_tools_node_protoc
UNAME_S := $(shell uname -s)
ifeq ($(findstring CYGWIN,$(UNAME_S)),CYGWIN)
    # Windows (Cygwin) - use WSL
    GRPC_NODE_PROTOC := wsl grpc_tools_node_protoc
    JS_OUT_PATH := $(shell wslpath -u '$(shell cygpath -w $(GENERATED_JS_OUT))')
else ifeq ($(findstring MINGW,$(UNAME_S)),MINGW)
    # Windows (MinGW) - use WSL
    GRPC_NODE_PROTOC := wsl grpc_tools_node_protoc
    JS_OUT_PATH := $(GENERATED_JS_OUT)
else ifeq ($(findstring MSYS,$(UNAME_S)),MSYS)
    # Windows (MSYS) - use WSL
    GRPC_NODE_PROTOC := wsl grpc_tools_node_protoc
    JS_OUT_PATH := $(GENERATED_JS_OUT)
else
    # Linux/macOS - use native grpc_tools_node_protoc
    GRPC_NODE_PROTOC := grpc_tools_node_protoc
    JS_OUT_PATH := $(GENERATED_JS_OUT)
endif

PROTOC_COMMON_FLAGS := -I=$(PROTO_IN) -I=$(THIRD_PARTY)

# --- 2. DEFINE FLAGS ---
# Step 1: Generate Base Go Files
GO_GEN_FLAGS := --go_out=$(GENERATED_GO_OUT) --go_opt=module=$(MODULE) \
                --go-grpc_out=$(GENERATED_GO_OUT) --go-grpc_opt=module=$(MODULE)

# Step 2: Inject Tags 
GO_TAG_FLAGS := --gotag_out=module=$(MODULE):$(INFRA_BASE)

PYTHON_GEN_FLAGS := --python_out=$(GENERATED_PYTHON_OUT) --grpc_python_out=$(GENERATED_PYTHON_OUT)

JS_GEN_FLAGS := --js_out=import_style=commonjs,binary:$(JS_OUT_PATH) --grpc_out=grpc_js:$(JS_OUT_PATH)

# --- 3. THE MACRO ---
define generate_proto
    WORD=$(1); \
    SERVICE_DIR="$(PROTO_IN)/$$WORD"; \
    if [ -d "$$SERVICE_DIR" ]; then \
        PROTO_FILES=$$(find $$SERVICE_DIR -maxdepth 3 -name "*.proto" -type f); \
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
			mkdir -p $(GENERATED_PYTHON_OUT)/$(1); \
            python -m grpc_tools.protoc $(PROTOC_COMMON_FLAGS) $(PYTHON_GEN_FLAGS) $$PROTO_FILES; \
			\
            echo "   > JS..."; \
			mkdir -p $(GENERATED_JS_OUT)/$(1); \
            $(GRPC_NODE_PROTOC) $(PROTOC_COMMON_FLAGS) $(JS_GEN_FLAGS) $$PROTO_FILES; \
        fi; \
    else \
        echo "Skipping $$WORD: directory $$SERVICE_DIR not found"; \
    fi
endef

generate-proto:
	@echo "Generating proto files"
	@MODULES=$$(find $(PROTO_IN) -name "*.proto" -type f -exec dirname {} \; | cut -d'/' -f3 | sort -u | grep -v third_party); \
	for module in $$MODULES; do \
		$(call generate_proto,$$module); \
	done;
	@echo "Generating third_party proto (tagger)..."
	@echo "   > Python..."
	@cd $(THIRD_PARTY) && python -m grpc_tools.protoc -I=. --python_out=../../$(PYTHON_MODEL_DIR) --grpc_python_out=../../$(PYTHON_MODEL_DIR) tagger/tagger.proto
	@echo "   > JS..."
	@cd $(THIRD_PARTY) && $(GRPC_NODE_PROTOC) -I=. --js_out=import_style=commonjs,binary:../../$(GENERATED_JS_OUT) --grpc_out=grpc_js:../../$(GENERATED_JS_OUT) tagger/tagger.proto
	@echo "   > Go (skipped - using vendor package)..."
	@find $(GENERATED_PYTHON_OUT) -type d -exec touch '{}/__init__.py' \;
	@python $(INFRA_BASE)/resources/scripts/fix_proto_imports.py $(GENERATED_PYTHON_OUT)
	@echo "✅ Proto files generated successfully"

generate-proto-clean: ## Surgically remove ONLY generated proto files
	@echo "Cleaning generated proto files..."
	@echo "Remove Go generated proto files"
	@if [ -d $$GENERATED_GO_OUT/model ]; then find $$GENERATED_GO_OUT/model -type f \( -name "*.pb.go" -o -name "*_grpc.pb.go" \) -delete; fi
	@echo "Remove Python generated proto files"
	@if [ -d $$GENERATED_PYTHON_OUT ]; then find $$GENERATED_PYTHON_OUT -type f -name "*_pb2.py" -delete; fi
	@echo "Remove JS generated proto files"
	@if [ -d $$GENERATED_JS_OUT ]; then find $$GENERATED_JS_OUT -type f \( -name "*_pb.js" -o -name "*_grpc_pb.js" \) -delete; fi
	@echo "✅ Generated proto cleaned"

# ============================================================================
# GENERATED
# ============================================================================

generate: generate-clean generate-proto generate-topics generate-mocks

generate-clean: generate-proto-clean generate-topics-clean generate-mocks-clean 

# ============================================================================
# FUNCTIONAL TESTS
# ============================================================================
test-functional-setup: 
	@echo "Setting up Python functional test environment..."
	@cd $(PYTHON_LIB_DIR)/functional && python -m pip install -r requirements.txt
	@echo "✓ Python dependencies installed"