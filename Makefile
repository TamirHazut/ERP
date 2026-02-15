# ERP System Root Makefile
# Delegates to service-specific Makefiles

.PHONY: help \
		generate generate-clean \
        build $(addprefix build-,$(MODULES)) \
		run $(addprefix run-,$(MODULES)) \
        test $(addprefix test-,$(MODULES)) test-coverage \
		lint clean \
		test-functional-setup test-functional-% test-functional-all test-functional-clean \
        docker-up docker-down docker-logs docker-ps \
        certs certs-clean

# Binary output directory
BIN_DIR := ./bin

CMD_DIR := ./cmd

# Define entire system modules including non services (including shared for proto generation)
MODULES := $(shell ls ./internal/ | cut -d'/' -f3 | sort -u)

help: ## Show this help message
	@echo "ERP System - Available targets:"
	@echo ""
	@echo "Generation:"
	@echo "  make generate          	- Generate all required files"
	@echo "  make generate-clean		- Clean all generated files"
	@echo ""
	@echo "Build:"
	@echo "  make build          	- Build all modules (with cmd/ directory)"
	@echo "  make build-<module>    - Build specific module (modules: auth, config, core, init)"
	@echo ""
	@echo "Run:"
	@echo "  make run           	- Run all modules"
	@echo "  make run-<module>      - Run specific module (modules: auth, config, core, init)"
	@echo ""
	@echo "Test & Quality:"
	@echo "  make test           	- Run all tests"
	@echo "  make test-<module>		- Run module tests (modules: auth, config, core, gateway, event, infra, init)"
	@echo "  make test-coverage  	- Run tests with coverage"
	@echo "  make lint           	- Run linter on all modules"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up      	- Start MongoDB and Redis containers"
	@echo "  make docker-down    	- Stop and remove containers"
	@echo "  make docker-logs    	- View container logs"
	@echo "  make docker-ps      	- List running containers"
	@echo ""
	@echo "Utilities:"
	@echo "  make tidy           	- Run go mod tidy"
	@echo "  make clean          	- Clean build artifacts from all modules"
	@echo ""
	@echo "Certificates (mTLS):"
	@echo "  make certs          	- Create CA and all module certificates"
	@echo "  make certs-clean    	- Remove all certificates"
	@echo ""
	@echo "	 make help			 	- Display this menu"
	@echo ""


# ============================================================================
# PROTO GENERATION TARGETS
# ============================================================================

generate: ## Generate all required files
	@$(MAKE) -C internal/infra generate"


generate-clean: ## Remove all generated files
	@$(MAKE) -C internal/infra generate-clean

# ============================================================================
# BUILD TARGETS
# ============================================================================

build: ## Build all modules
	@echo "Building all modules..."
	@for module in $(MODULES); do \
		$(MAKE) build-$$module; \
	done
	@echo "✓ All modules processed"

build-%:
	@if [ -d "internal/$*/cmd" ]; then \
		echo "Building $* module..."; \
		$(MAKE) -C internal/$* build; \
		echo "✓ $* built successfully"; \
	else \
		echo "⚠️  Skipping $* (no cmd/ directory)"; \
	fi

# ============================================================================
# RUN TARGETS
# ============================================================================

run:
	@LOG_FILE_PATH=./logs go run $(CMD_DIR)

run-%:
	@if [ -d "internal/$*/cmd" ]; then \
		$(MAKE) -C internal/$* run; \
	else \
		echo "⚠️  Cannot run $* (no cmd/ directory)"; \
	fi

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

test-functional-setup: 
	$(MAKE) -C internal/infra test-functional-setup

test-functional-%:
	@echo "Running $* service functional tests..."
	@$(MAKE) -C internal/$* test-functional
	@echo "✓ $* functional tests completed"

test-functional:
	@echo "Running all functional tests..."
	@for module in $(MODULES); do \
		if [ -d "internal/$$module/functional" ]; then \
			$(MAKE) test-functional-$$module; \
			$(MAKE) test-functional-clean; \
		fi; \
	done
	@echo "✓ All functional tests completed"

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

docker-down: ## Stop and remove containers
	@echo "Stopping Docker containers..."
	@docker compose down
	@echo "✓ Containers stopped"

docker-logs: ## View container logs
	@docker compose logs -f

docker-ps: ## List running containers
	@docker compose ps

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