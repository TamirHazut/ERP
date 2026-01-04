# ERP System Root Makefile
# Delegates to service-specific Makefiles

.PHONY: proto proto-auth proto-config proto-core proto-infra proto-clean \
        build build-auth build-config build-core \
        run-auth run-config run-core \
        test test-coverage lint clean tidy help \
        docker-up docker-down docker-logs docker-ps \
        certs certs-clean

# Binary output directory
BIN_DIR := bin

help: ## Show this help message
	@echo "ERP System - Available targets:"
	@echo ""
	@echo "Proto Generation:"
	@echo "  make proto          - Generate all proto files"
	@echo "  make proto-infra    - Generate infra proto files"
	@echo "  make proto-auth     - Generate auth service proto files"
	@echo "  make proto-config   - Generate config service proto files"
	@echo "  make proto-core     - Generate core service proto files"
	@echo "  make proto-clean    - Remove all generated proto files"
	@echo ""
	@echo "Build:"
	@echo "  make build          - Build all services"
	@echo "  make build-auth     - Build auth service"
	@echo "  make build-config   - Build config service"
	@echo "  make build-core     - Build core service"
	@echo ""
	@echo "Run:"
	@echo "  make run-auth       - Run auth service"
	@echo "  make run-config     - Run config service"
	@echo "  make run-core       - Run core service"
	@echo ""
	@echo "Test & Quality:"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter on all services"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up      - Start MongoDB and Redis containers"
	@echo "  make docker-down    - Stop and remove containers"
	@echo "  make docker-logs    - View container logs"
	@echo "  make docker-ps      - List running containers"
	@echo ""
	@echo "Utilities:"
	@echo "  make tidy           - Run go mod tidy"
	@echo "  make clean          - Clean build artifacts from all services"
	@echo ""
	@echo "Certificates (mTLS):"
	@echo "  make certs          - Create CA and all service certificates"
	@echo "  make certs-clean    - Remove all certificates"
	@echo ""
	@echo "Service-specific help:"
	@echo "  make -C internal/auth help    - Show auth service targets"
	@echo "  make -C internal/config help  - Show config service targets"
	@echo "  make -C internal/core help    - Show core service targets"

# ============================================================================
# PROTO GENERATION TARGETS
# ============================================================================
MODULE="erp.localhost"
PROTO_OUT="."
INFRA_BASE_PROTO="internal/infra/proto"
INFRA_PROTO="$(INFRA_BASE_PROTO)/infra/v1"
AUTH_PROTO="$(INFRA_BASE_PROTO)/auth/v1"
CONFIG_PROTO="$(INFRA_BASE_PROTO)/config/v1"
CORE_PROTO="$(INFRA_BASE_PROTO)/core/v1"

proto: proto-infra proto-auth proto-config proto-core ## Generate all proto files
	@echo "All proto files generated successfully"

proto-infra: ## Generate infra proto files
	@echo "Generating infra proto files..."
	@protoc --go_out=$(PROTO_OUT) \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=$(PROTO_OUT) \
		--go-grpc_opt=module=$(MODULE) \
		-I=$(INFRA_PROTO) \
		"$(INFRA_PROTO)/*.proto"
	@echo "Infra proto files generated"

proto-auth: ## Generate auth proto files
	@echo "Generating auth proto files..."
	@protoc --go_out=$(PROTO_OUT) \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=$(PROTO_OUT) \
		--go-grpc_opt=module=$(MODULE) \
		-I=$(INFRA_PROTO) \
		-I=$(AUTH_PROTO) \
		"$(AUTH_PROTO)/*.proto"
	@echo "Auth proto files generated"

proto-config: ## Generate config proto files
	@echo "Generating config proto files..."
	@protoc --go_out=$(PROTO_OUT) \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=$(PROTO_OUT) \
		--go-grpc_opt=module=$(MODULE) \
		-I=$(INFRA_PROTO) \
		-I=$(CONFIG_PROTO) \
		"$(CONFIG_PROTO)/*.proto"
	@echo "Config proto files generated"

proto-core: ## Generate core proto files (user service)
	@echo "Generating core proto files..."
	@protoc --go_out=$(PROTO_OUT) \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=$(PROTO_OUT) \
		--go-grpc_opt=module=$(MODULE) \
		-I=$(INFRA_PROTO) \
		-I=$(CORE_PROTO) \
		"$(CORE_PROTO)/*.proto"
	@echo "Core proto files generated"

proto-clean: ## Remove all generated proto files
	@echo "Cleaning generated proto files..."
	@find internal/infra/proto -name "*.pb.go" -type f -delete 2>/dev/null || true
	@echo "Proto files cleaned"

# ============================================================================
# BUILD TARGETS
# ============================================================================

build: build-auth build-config build-core ## Build all services
	@echo "✓ All services built"

build-auth: ## Build auth service
	@$(MAKE) -C internal/auth build

build-config: ## Build config service
	@$(MAKE) -C internal/config build

build-core: ## Build core service
	@$(MAKE) -C internal/core build

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
	@$(MAKE) -C internal/auth test
	@$(MAKE) -C internal/config test
	@$(MAKE) -C internal/core test
	@$(MAKE) -C internal/db test
	@echo "✓ All tests complete"

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# ============================================================================
# QUALITY TARGETS
# ============================================================================

lint: ## Run linter on all services
	@echo "Running linter on all services..."
	@$(MAKE) -C internal/auth lint
	@$(MAKE) -C internal/config lint
	@$(MAKE) -C internal/core lint
	@$(MAKE) -C internal/db lint
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
	@$(MAKE) -C internal/auth clean
	@$(MAKE) -C internal/config clean
	@$(MAKE) -C internal/core clean
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
	@$(MAKE) -C internal/auth mocks
	@$(MAKE) -C internal/config mocks
	@$(MAKE) -C internal/core mocks
	@$(MAKE) -C internal/db mocks
	@echo "✅ Mocks generated successfully"

.PHONY: mocks-clean
mocks-clean:
	@echo "Cleaning generated mocks..."
	@$(MAKE) -C internal/auth mocks-clean
	@$(MAKE) -C internal/config mocks-clean
	@$(MAKE) -C internal/core mocks-clean
	@$(MAKE) -C internal/db mocks-clean
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

certs: ## Create CA and certificates for all services
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
	@$(MAKE) -C internal/auth certs
	@$(MAKE) -C internal/config certs
	@$(MAKE) -C internal/core certs
	@echo "✓ All certificates created successfully"

certs-clean: ## Remove all certificates (CA and service certificates)
	@echo "Removing all certificates..."
	@rm -rf resources/certs
	@$(MAKE) -C internal/auth certs-clean
	@$(MAKE) -C internal/config certs-clean
	@$(MAKE) -C internal/core certs-clean
	@echo "✓ All certificates removed"