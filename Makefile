# ERP System Makefile

.PHONY: proto proto-auth proto-config proto-core proto-common help

# Proto generation output directories
PROTO_OUT := internal
PROTO_COMMON := proto/common
PROTO_AUTH := internal/auth/proto
PROTO_CONFIG := internal/config/proto
PROTO_CORE := internal/core/proto

# Go module path
MODULE := erp.localhost

help: ## Show this help message
	@echo "Available targets:"
	@echo "  make proto          - Generate all proto files"
	@echo "  make proto-common   - Generate common proto files"
	@echo "  make proto-auth     - Generate auth service proto files"
	@echo "  make proto-config   - Generate config service proto files"
	@echo "  make proto-core     - Generate core service proto files"
	@echo "  make tidy           - Run go mod tidy"
	@echo "  make build          - Build all services"

proto: proto-common proto-auth proto-config proto-core ## Generate all proto files

proto-common: ## Generate common proto files
	@echo "Generating common proto files..."
	@if [ -f "$(PROTO_COMMON)/common.proto" ]; then \
		protoc --go_out=$(PROTO_OUT) \
			--go_opt=module=$(MODULE) \
			--go-grpc_out=$(PROTO_OUT) \
			--go-grpc_opt=module=$(MODULE) \
			-I=proto \
			$(PROTO_COMMON)/common.proto; \
		echo "✓ Common proto files generated"; \
	else \
		echo "⚠ No common.proto file found, skipping..."; \
	fi

proto-auth: ## Generate auth service proto files
	@echo "Generating auth service proto files..."
	@if [ -f "$(PROTO_AUTH)/auth.proto" ]; then \
		protoc --go_out=$(PROTO_OUT) \
			--go_opt=module=$(MODULE) \
			--go-grpc_out=$(PROTO_OUT) \
			--go-grpc_opt=module=$(MODULE) \
			-I=proto \
			-I=$(PROTO_AUTH) \
			$(PROTO_AUTH)/*.proto; \
		echo "✓ Auth proto files generated"; \
	else \
		echo "⚠ No auth.proto file found, skipping..."; \
	fi

proto-config: ## Generate config service proto files
	@echo "Generating config service proto files..."
	@if [ -f "$(PROTO_CONFIG)/config.proto" ]; then \
		protoc --go_out=$(PROTO_OUT) \
			--go_opt=module=$(MODULE) \
			--go-grpc_out=$(PROTO_OUT) \
			--go-grpc_opt=module=$(MODULE) \
			-I=proto \
			-I=$(PROTO_CONFIG) \
			$(PROTO_CONFIG)/*.proto; \
		echo "✓ Config proto files generated"; \
	else \
		echo "⚠ No config.proto file found, skipping..."; \
	fi

proto-core: ## Generate core service proto files
	@echo "Generating core service proto files..."
	@if [ -f "$(PROTO_CORE)/core.proto" ]; then \
		protoc --go_out=$(PROTO_OUT) \
			--go_opt=module=$(MODULE) \
			--go-grpc_out=$(PROTO_OUT) \
			--go-grpc_opt=module=$(MODULE) \
			-I=proto \
			-I=$(PROTO_CORE) \
			$(PROTO_CORE)/*.proto; \
		echo "✓ Core proto files generated"; \
	else \
		echo "⚠ No core.proto file found, skipping..."; \
	fi

tidy: ## Run go mod tidy
	@echo "Running go mod tidy..."
	@go mod tidy
	@echo "✓ Dependencies updated"

build: ## Build all services
	@echo "Building services..."
	@go build ./...
	@echo "✓ Build complete"

