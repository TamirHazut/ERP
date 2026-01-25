# ERP System Root Makefile
# Delegates to service-specific Makefiles

.PHONY: proto $(addprefix proto-,$(MODULES)) proto-clean \
		run $(addprefix run-,$(SERVICES)) \
        build $(addprefix build-,$(SERVICES)) \
        test $(addprefix test-,$(MODULES)) test-coverage \
		lint clean tidy help \
        docker-up docker-down docker-logs docker-ps \
        certs certs-clean

# Binary output directory
BIN_DIR := bin

# Define services
SERVICES := auth config core gateway event

# Define entire system modules including non services (including shared for proto generation)
MODULES := infra $(SERVICES) init #shared

help: ## Show this help message
	@echo "ERP System - Available targets:"
	@echo ""
	@echo "Proto Generation:"
	@echo "  make proto          	- Generate all proto files"
	@echo "  make proto-infra     	- Generate infra service proto files"
	@echo "  make proto-<module>	- Generate module proto files (modules: infra, auth, config, core, gateway, event)"
	@echo ""
	@echo "Build:"
	@echo "  make build          	- Build all services"
	@echo "  make build-<service>   - Build service (services: auth, config, core, gateway, event)"
	@echo ""	
	@echo "Run:"	
	@echo "  make run           	- Run all services"
	@echo "  make run-<service>     - Run service (services: auth, config, core, gateway, event)"
	@echo ""	
	@echo "Test & Quality:"	
	@echo "  make test           	- Run all tests"
	@echo "  make test-<module>		- Run module tests (modules: infra, auth, config, core, gateway, event)"
	@echo "  make test-coverage  	- Run tests with coverage"
	@echo "  make lint           	- Run linter on all services"
	@echo ""	
	@echo "Docker:"	
	@echo "  make docker-up      	- Start MongoDB and Redis containers"
	@echo "  make docker-down    	- Stop and remove containers"
	@echo "  make docker-logs    	- View container logs"
	@echo "  make docker-ps      	- List running containers"
	@echo ""	
	@echo "Utilities:"	
	@echo "  make tidy           	- Run go mod tidy"
	@echo "  make clean          	- Clean build artifacts from all services"
	@echo ""	
	@echo "Certificates (mTLS):"	
	@echo "  make certs          	- Create CA and all service certificates"
	@echo "  make certs-clean    	- Remove all certificates"


# ============================================================================
# PROTO GENERATION TARGETS
# ============================================================================
MODULE := erp.localhost
PROTO_OUT := .
INFRA_BASE := internal/infra
PROTO_IN := $(INFRA_BASE)/proto
THIRD_PARTY := $(PROTO_IN)/third_party
GENERATED_OUT = $(INFRA_BASE)/model
PROTOC_COMMON_FLAGS := -I=$(PROTO_IN) -I=$(THIRD_PARTY)
GO_GEN_FLAGS := --go_out=$(PROTO_OUT) --go_opt=module=$(MODULE) \
                --go-grpc_out=$(PROTO_OUT) --go-grpc_opt=module=$(MODULE)
GO_TAG_FLAGS := --gotag_out=module=$(MODULE):$(PROTO_OUT)

define generate_proto
	@SERVICE_DIR="$(PROTO_IN)/$(1)/v1"; \
	if [ ! -d "$$SERVICE_DIR" ]; then \
		echo "Warning: Proto directory $$SERVICE_DIR not found"; \
		exit 0; \
	fi; \
	echo "Generating $(1) proto files..."; \
	for dir in $$SERVICE_DIR $$SERVICE_DIR/cache; do \
		if [ -d "$$dir" ] && [ "$$(ls $$dir/*.proto 2>/dev/null)" ]; then \
			mkdir -p $(GENERATED_OUT)/$${dir#$(PROTO_IN)/}; \
			protoc $(PROTOC_COMMON_FLAGS) $(GO_GEN_FLAGS) \
				$$dir/*.proto || exit 1; \
			protoc $(PROTOC_COMMON_FLAGS) $(GO_TAG_FLAGS) \
				$$dir/*.proto || exit 1; \
		fi; \
	done; \
	echo "✓ $(1) all files generated and tagged"
endef

proto: ## Generate all proto files
	@for module in $(MODULES); do \
		$(MAKE) proto-$$module; \
	done
	@echo "✓ All proto files generated successfully"

proto-%:
	$(call generate_proto,$*)


proto-clean: ## Remove all generated proto files
	@echo "Cleaning generated proto files..."
	@find $(GENERATED_OUT) -name "*.pb.go" -type f -delete 2>/dev/null || true
	@echo "Proto files cleaned"

# Python proto generation directory
PYTHON_PROTO_OUT := internal/infra/functional/proto

proto-python: ## Generate Python gRPC stubs from proto files
	@echo "Generating Python proto files..."
	@mkdir -p $(PYTHON_PROTO_OUT)
	@# Generate for each service
	@for service in infra auth config core gateway event; do \
		SERVICE_DIR="$(PROTO_IN)/$$service/v1"; \
		if [ -d "$$SERVICE_DIR" ]; then \
			echo "Generating Python stubs for $$service..."; \
			python -m grpc_tools.protoc \
				-I=$(PROTO_IN) \
				-I=$(THIRD_PARTY) \
				--python_out=$(PYTHON_PROTO_OUT) \
				--grpc_python_out=$(PYTHON_PROTO_OUT) \
				$$SERVICE_DIR/*.proto; \
			if [ -d "$$SERVICE_DIR/cache" ]; then \
				python -m grpc_tools.protoc \
					-I=$(PROTO_IN) \
					-I=$(THIRD_PARTY) \
					--python_out=$(PYTHON_PROTO_OUT) \
					--grpc_python_out=$(PYTHON_PROTO_OUT) \
					$$SERVICE_DIR/cache/*.proto; \
			fi; \
		fi; \
	done
	@echo "✓ Python proto files generated in $(PYTHON_PROTO_OUT)"

proto-python-clean: ## Clean generated Python proto files
	@echo "Cleaning Python proto files..."
	@rm -rf $(PYTHON_PROTO_OUT)
	@echo "✓ Python proto files cleaned"

# ============================================================================
# BUILD TARGETS
# ============================================================================

define build_service
	@echo "Building $(1) ..."
	@$(MAKE) -C internal/$$service build;
	@echo "✓ $(1) build successfully"
endef

build: ## Build all services
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		$(MAKE) build-$$service; \
	done
	@echo "✓ All services built"

build-%:
	$(call run build_service,$*)

# ============================================================================
# RUN TARGETS
# ============================================================================

# Function to generate proto for any service
define run_service
	@$(MAKE) -C internal/$(1) run
endef

run: 
	@go run ./cmd/

run-%:
	$(call run_service,$*)

# ============================================================================
# TEST TARGETS
# ============================================================================

define test_module
	@echo "Running $(1) tests..."
	@$(MAKE) -C internal/$(1) test;
	@echo "✓ $(1) tests passed"
endef


test: ## mocks ## Run all tests from all services
	@echo "Running all tests..."
	@for module in $(MODULES); do \
		$(MAKE) test-$$module; \
	done
	@echo "✓ All tests complete"

test-%:
	$(call test_module,$*)


test-coverage: ## Run tests with coverage for all services
	@echo "Running tests with coverage for all modules..."
	@for module in $(MODULES); do \
		echo "=== $$module ===" && \
		$(MAKE) -C internal/$$module test-coverage; \
	done
	@echo "✓ All modules coverage reports generated"

# ============================================================================
# FUNCTIONAL TEST TARGETS
# ============================================================================

.PHONY: test-functional-setup test-functional-% test-functional-all test-functional-clean

test-functional-setup: proto-python ## Setup Python test environment
	@echo "Setting up Python functional test environment..."
	@cd internal/infra/functional && python -m pip install -r requirements.txt
	@echo "✓ Python dependencies installed"

test-functional-%: test-functional-setup ## Run functional tests for a specific service
	@echo "Running $* service functional tests..."
	@cd internal/$*/functional && python -m pytest -v --tb=short
	@echo "✓ $* functional tests completed"

test-functional-all: test-functional-setup ## Run all functional tests
	@echo "Running all functional tests..."
	@for service in $(SERVICES); do \
		if [ -d "internal/$$service/functional" ]; then \
			$(MAKE) test-functional-$$service; \
		fi; \
	done
	@echo "✓ All functional tests completed"

test-functional-clean: proto-python-clean ## Clean functional test artifacts
	@echo "Cleaning functional test artifacts..."
	@find internal -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	@find internal -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true
	@find internal -type f -name "*.pyc" -delete 2>/dev/null || true
	@echo "✓ Functional test artifacts cleaned"

# ============================================================================
# QUALITY TARGETS
# ============================================================================

lint: ## Run linter on all services
	@echo "Running linter on all services..."
	@for module in $(MODULES); do \
		$(MAKE) -C internal/$$module lint; \
	done
	@echo "✓ All linting complete"

# ============================================================================
# UTILITY TARGETS
# ============================================================================

tidy: ## Run go mod tidy
	@echo "Running go mod tidy..."
	@go mod tidy
	@echo "✓ Dependencies updated"

clean: ## Clean build artifacts from all services
	@echo "Cleaning all build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@for module in $(MODULES); do \
		$(MAKE) -C internal/$$module clean; \
	done
	@echo "✓ Clean complete"

# ============================================================================
# DOCKER TARGETS
# ============================================================================

docker-up: ## Start MongoDB and Redis containers
	@echo "Starting Docker containers..."
	@docker compose up -d
	@echo "✓ Containers started"
	@echo ""
	@echo "Services:"
	@echo "  MongoDB: mongodb://root:secret@localhost:27017"
	@echo "  Redis:   redis://:supersecretredis@localhost:6379"

docker-down: ## Stop and remove containers
	@echo "Stopping Docker containers..."
	@docker compose down
	@echo "✓ Containers stopped"

docker-logs: ## View container logs
	@docker compose logs -f

docker-ps: ## List running containers
	@docker compose ps

# ============================================================================
# Mock Generation
# ============================================================================

.PHONY: mocks
mocks: mocks-clean
	@echo "Generating mocks..."
	@for module in $(MODULES); do \
		$(MAKE) -C internal/$$module mocks; \
	done
	@echo "✅ Mocks generated successfully"

.PHONY: mocks-clean
mocks-clean:
	@echo "Cleaning generated mocks..."
	@for module in $(MODULES); do \
		echo "Cleaning $$module mocks..." && \
		$(MAKE) -C internal/$$module mocks-clean; \
	done
	@echo "✅ Mocks cleaned"

.PHONY: mocks-verify
mocks-verify: mocks
	@echo "Verifying mocks are up to date..."
	@if [ -n "$$(git status --porcelain | grep 'mock_')" ]; then \
		echo "❌ Generated mocks are out of date. Run 'make mocks' and commit changes."; \
		git status --porcelain | grep 'mock_'; \
		exit 1; \
	else \
		echo "✅ Mocks are up to date"; \
	fi

# ============================================================================
# CERTIFICATE GENERATION (mTLS)
# ============================================================================

certs: certs-clean ## Create CA and certificates for all services
	@echo "Creating certificates..."
	@for module in $(MODULES); do \
		$(MAKE) -C internal/$$module certs; \
	done
	@echo "✓ All certificates created successfully"

certs-clean: ## Remove all certificates (CA and service certificates)
	@echo "Removing all certificates..."
	@rm -rf internal/infra/resources/certs
	@for module in $(MODULES); do \
		$(MAKE) -C internal/$$module certs-clean; \
	done
	@echo "✓ All certificates removed"