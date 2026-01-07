# ERP System Root Makefile
# Delegates to service-specific Makefiles

.PHONY: proto $(addprefix proto-,$(MODULES)) proto-clean \
        build \
        run-auth run-config run-core \
        test test-coverage lint clean tidy help \
        docker-up docker-down docker-logs docker-ps \
        certs certs-clean

# Binary output directory
BIN_DIR := bin

# Define services
SERVICES := auth config core #gateway event

# Define entire system modules including non services
MODULES := infra $(SERVICES)

help: ## Show this help message
	@echo "ERP System - Available targets:"
	@echo ""
	@echo "Proto Generation:"
	@echo "  make proto          	- Generate all proto files"
	@echo "  make proto-infra     	- Generate infra service proto files"
	@echo "  make proto-<service>	- Generate service proto files (services: auth, config, core, gateway, event)"
	@echo ""
	@echo "Build:"
	@echo "  make build          	- Build all services"
	@echo ""	
	@echo "Run:"	
	@echo "  make run-auth       	- Run auth service"
	@echo "  make run-config     	- Run config service"
	@echo "  make run-core       	- Run core service"
	@echo ""	
	@echo "Test & Quality:"	
	@echo "  make test           	- Run all tests"
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
INFRA_BASE_PROTO := internal/infra/proto

# Function to generate proto for any service
define generate_proto
	@echo "Generating $(1) proto files..."
	@if [ ! -d "$(INFRA_BASE_PROTO)/$(1)/v1" ]; then \
		echo "Warning: Proto directory $(INFRA_BASE_PROTO)/$(1)/v1 not found"; \
		exit 0; \
	fi
	@protoc --go_out=$(PROTO_OUT) \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=$(PROTO_OUT) \
		--go-grpc_opt=module=$(MODULE) \
		-I=$(INFRA_BASE_PROTO) \
		"$(INFRA_BASE_PROTO)/$(1)/v1/*.proto"
	@echo "✓ $(1) proto files generated"
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
	@find internal/infra/proto -name "*.pb.go" -type f -delete 2>/dev/null || true
	@echo "Proto files cleaned"

# ============================================================================
# BUILD TARGETS
# ============================================================================

build: ## Build all services
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		$(MAKE) -C internal/$$service build; \
	done
	@echo "✓ All services built"

# ============================================================================
# RUN TARGETS
# ============================================================================

run-auth: ## Run auth service
	@$(MAKE) -C internal/auth run

run-config: ## Run config service
	@$(MAKE) -C internal/config run

run-core: ## Run core service
	@$(MAKE) -C internal/core run

# ============================================================================
# TEST TARGETS
# ============================================================================

test: mocks ## Run all tests from all services
	@echo "Running all tests..."
	@for module in $(MODULES); do \
		$(MAKE) -C internal/$$module test; \
	done
	@echo "✓ All tests complete"

test-coverage: ## Run tests with coverage for all services
	@echo "Running tests with coverage for all modules..."
	@for module in $(MODULES); do \
		echo "=== $$module ===" && \
		$(MAKE) -C internal/$$module test-coverage; \
	done
	@echo "✓ All modules coverage reports generated"

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

# Certificate configuration
CA_DIR := resources/certs/ca
CA_KEY := $(CA_DIR)/ca-key.pem
CA_CERT := $(CA_DIR)/ca-cert.pem
CA_DAYS := 3650
CERT_DAYS := 365

certs: certs-clean ## Create CA and certificates for all services
	@echo "Creating Certificate Authority (CA) and all service certificates..."
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
	@echo "Creating service certificates..."
	@for service in $(SERVICES); do \
		$(MAKE) -C internal/$$service certs; \
	done
	@echo "✓ All certificates created successfully"

certs-clean: ## Remove all certificates (CA and service certificates)
	@echo "Removing all certificates..."
	@rm -rf resources/certs
	@for service in $(SERVICES); do \
		$(MAKE) -C internal/$$service certs-clean; \
	done
	@echo "✓ All certificates removed"