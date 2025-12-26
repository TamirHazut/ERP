# ERP System Development Roadmap

## Overview
This roadmap outlines the development order for building the multi-tenant ERP system. Services are organized by priority and dependencies to ensure efficient development.

## Pre-Phase: Infrastructure Setup 🏗️

Before starting service development, we need to set up foundational infrastructure that all services will depend on.

**Status:** ✅ Completed (gRPC ✅, JWT ✅, Error Handling ✅, Service Structure ⏭️, Build Tooling ✅, Models ✅)

### 1. gRPC Infrastructure (Critical) 📡
**Status:** ✅ Completed

**Why First:** All inter-service communication uses gRPC. Must be set up before any service development.

**What to Build:**
- [x] Create proto files directory structure (service-specific proto dirs + `proto/common/`)
  - [x] `proto/common/` - Shared types
  - [x] `internal/auth/proto/` - Auth service proto files
  - [x] `internal/config/proto/` - Config service proto files
  - [x] `internal/core/proto/` - Core service proto files
  - [x] `internal/gateway/proto/` - Gateway service proto files (if needed)
  - [x] `internal/events/proto/` - Events service proto files (if needed)
- [x] Add gRPC Go dependencies to `go.mod`
  - [x] `google.golang.org/grpc`
  - [x] `google.golang.org/protobuf`
  - [x] `google.golang.org/protobuf/cmd/protoc-gen-go`
  - [x] `google.golang.org/grpc/cmd/protoc-gen-go-grpc`
- [x] Set up proto code generation (Makefile or script)
  - [x] Makefile for Linux/Mac
  - [x] PowerShell script for Windows (`scripts/generate-proto.ps1`)
  - [x] Bash script for Linux/Mac (`scripts/generate-proto.sh`)
- [x] Create proto file template/structure for services
  - [x] Common proto file (`proto/common/common.proto`)
  - [x] Template documentation in `docs/proto/README.md`
- [x] Document proto generation workflow

**Note:** Proto definitions for each service will be created as part of that service's development.

**Directory Structure:**
```
proto/
├── common/              # Shared types (errors, base messages)

internal/
├── auth/proto/          # Auth service proto files
├── config/proto/        # Config service proto files
├── core/proto/          # Core service proto files
├── gateway/proto/       # Gateway service proto files
└── events/proto/        # Events service proto files
```

---

### 2. JWT Library (Critical for Auth) 🔑
**Status:** ✅ Completed

**Why Second:** Required for Auth Service. Should be added early.

**What to Build:**
- [x] Add JWT library to `go.mod`
  - [x] `github.com/golang-jwt/jwt/v5`
- [x] Create JWT utility package/helpers
  - [x] TokenManager struct (`internal/auth/token_manager.go`) - Unified JWT and Redis token management
  - [x] GenerateAccessToken method (with userID, tenantID, role, permissions)
  - [x] VerifyAccessToken method
  - [x] GenerateRefreshToken method
  - [x] VerifyRefreshToken method
  - [x] Token storage in Redis (AccessTokenKeyHandler, RefreshTokenKeyHandler)
  - [x] Token revocation and management (Revoke, RevokeAll)

---

### 3. Error Handling Patterns (Important) ⚠️
**Status:** ✅ Completed

**Why Third:** Standardized error handling ensures consistency across services.

**What to Build:**
- [x] Define gRPC error codes mapping
  - [x] `internal/errors/grpc.go` - ToGRPCError/FromGRPCError functions
  - [x] Category to gRPC code mapping (AUTH → Unauthenticated, VALIDATION → InvalidArgument, etc.)
- [x] Create error handling utilities
  - [x] `internal/errors/errors.go` - AppError type with constructors
  - [x] Helper functions: New(), Wrap(), Auth(), Validation(), NotFound(), Conflict(), Business(), Internal()
- [x] Document error response format
  - [x] Updated `proto/common/common.proto` with ErrorCategory enum and enhanced Error message
- [x] Create common error types
  - [x] `internal/errors/codes.go` - Categorized error codes (AUTH, VALIDATION, NOT_FOUND, CONFLICT, BUSINESS, INTERNAL)

**Files Created:**
- `internal/errors/errors.go` - Core error types and constructors
- `internal/errors/codes.go` - Error code definitions by category
- `internal/errors/grpc.go` - gRPC status code mapping
- `internal/errors/errors_test.go` - Unit tests for error handling

---

### 4. Service Structure & Lifecycle (Helpful) 🔄
**Status:** ⏭️ Skipped (will evolve as we build services)

**Why Fourth:** Establishes patterns for how services are organized and started.

**What to Build:**
- [ ] Define service initialization pattern
- [ ] Create service lifecycle management (start/stop)
- [ ] Add health check endpoint pattern
- [ ] Create service main entry point template

**Note:** Skipped for now. Will evolve organically as we build the Auth service.

---

### 5. Build Tooling (Helpful) 🛠️
**Status:** ✅ Completed

**Why Last:** Helpful for development workflow, but not blocking.

**What to Build:**
- [x] Create root Makefile with common tasks:
  - [x] `make proto` - Generate proto code (delegates to service Makefiles)
  - [x] `make build` / `make build-auth` / `make build-config` / `make build-core` - Build services (delegates to service Makefiles)
  - [x] `make run-auth` / `make run-config` / `make run-core` - Run services (delegates to service Makefiles)
  - [x] `make test` / `make test-coverage` - Run tests (delegates to service Makefiles)
  - [x] `make lint` - Run linter (delegates to service Makefiles)
  - [x] `make clean` - Clean artifacts (delegates to service Makefiles)
  - [x] `make tidy` - Update dependencies
  - [x] `make proto-common` - Generate common proto files
- [x] Create service-specific Makefiles (independent, can be run standalone):
  - [x] `internal/auth/Makefile` - Auth service targets (proto, build, run, test, lint, clean)
  - [x] `internal/config/Makefile` - Config service targets (proto, build, run, test, lint, clean)
  - [x] `internal/core/Makefile` - Core service targets (proto, build, run, test, lint, clean)
  - [ ] `internal/gateway/Makefile` - Gateway service targets (to be created when Gateway service is developed)
  - [ ] `internal/events/Makefile` - Events service targets (to be created when Events service is developed)
  - [ ] `internal/webui/Makefile` - WebUI service targets (to be created when WebUI service is developed)
- [x] Create PowerShell scripts for Windows:
  - [x] `scripts/build.ps1` - Build script
  - [x] `scripts/run.ps1` - Run services script
  - [x] `scripts/generate-proto.ps1` - Proto generation script

**Files:**
- `Makefile` - Root Makefile (delegates to service Makefiles, handles shared targets like docker, tidy, proto-common)
- `internal/{service}/Makefile` - Service-specific Makefiles (independent, can be run from service directory)
- `docker-compose.yml` - MongoDB and Redis containers
- `scripts/build.ps1` - Windows build script
- `scripts/run.ps1` - Windows service runner
- `scripts/generate-proto.ps1` - Windows proto generation
- `scripts/generate-proto.sh` - Linux/Mac proto generation

**Makefile Structure:**
- **Root Makefile**: Delegates service-specific targets to service Makefiles using `make -C internal/{service} {target}`
- **Service Makefiles**: Independent Makefiles in each service directory with targets: `proto`, `build`, `run`, `test`, `test-coverage`, `lint`, `clean`, `help`
- **Usage**: 
  - From root: `make build-auth` (delegates to `internal/auth/Makefile`)
  - From service: `cd internal/auth && make build` (runs independently)

**Docker Commands:**
- `make docker-up` or `.\scripts\build.ps1 -Target docker-up` - Start containers
- `make docker-down` or `.\scripts\build.ps1 -Target docker-down` - Stop containers
- `make docker-logs` - View logs
- `make docker-ps` - List containers

**Note:** MongoDB and Redis connection URIs are currently hardcoded. Will be moved to environment configuration later.

---

### 6. Model Organization (Completed) 📦
**Status:** ✅ Completed

**What was Built:**
- [x] Organized models by service for future microservice separation
- [x] `internal/auth/models/models.go` - Auth models (Tenant, User, Role, Permission, UserGroup, AuditLog)
- [x] `internal/core/models/models.go` - Core models (Product, Order, Vendor, Customer, Inventory, Warehouse, Category)
- [x] `internal/config/models/models.go` - Config models (ServiceConfig, FeatureFlag)
- [x] Updated Redis cache models to reference new locations
- [x] Validation methods on all models (`Validate(createOperation bool)`)
- [x] Removed deprecated `internal/db/models.go` and `internal/db/mongo/models/`

**Directory Structure:**
```
internal/
├── auth/
│   ├── models/
│   │   └── models.go      # Tenant, User, Role, Permission
│   └── repository/
│       ├── users_repo.go
│       ├── roles_repo.go
│       ├── permissions_repo.go
│       └── tenants_repo.go
├── core/
│   └── models/
│       └── models.go      # Product, Order, Vendor, Customer, etc.
├── config/
│   └── models/
│       └── models.go      # ServiceConfig, FeatureFlag
└── db/
    └── redis/
        └── models/
            └── models.go  # Session, TokenMetadata, caches
```

---

## Development Phases

### Phase 1: Foundation ⚙️

#### 1. Auth Service (Priority 1) 🔐
**Status:** 🟡 In Progress (Repositories ✅, gRPC Server ⬜)

**Why First:** Required by all other services for authentication/authorization. Foundation for the entire system.

**Prerequisites:**
- ✅ Pre-Phase infrastructure setup must be completed first (gRPC infrastructure, JWT library)

**Dependencies:** 
- Uses existing `db` package
- MongoDB (`auth_db` collection)
- Redis (sessions/tokens)
- gRPC infrastructure (from Pre-Phase)
- JWT library (from Pre-Phase)

**What to Build:**
- [ ] gRPC server implementation
- [ ] Auth service proto definitions (`.proto` files)
- [x] User repository using generic Repository pattern (MongoDB: `auth_db.users`)
  - [x] `internal/auth/repository/users_repo.go`
  - [x] CRUD operations with tenant isolation
  - [x] `GetUsersByTenantID`, `GetUsersByRoleID` methods
  - [x] Model validation tests (`internal/auth/models/models_test.go`)
  - [x] Unit tests (`internal/auth/repository/users_repo_test.go`)
- [x] JWT generation/validation library integration
  - [x] TokenManager implementation (`internal/auth/token_manager.go`) - Unified JWT and Redis token management
  - [x] GenerateAccessToken with tenantID support
  - [x] VerifyAccessToken implementation
  - [x] GenerateRefreshToken implementation
  - [x] VerifyRefreshToken implementation
  - [x] Unit tests (`internal/auth/token_manager_test.go`)
- [x] JWT claims structure (include tenant ID)
  - [x] Claims include `sub` (userID), `tenant_id`, `username`, `role`, `permissions`, and `exp`
- [x] Password hashing (bcrypt)
  - [x] `internal/auth/hash.go` - HashPassword, VerifyPassword functions
  - [x] Password strength validation
  - [x] Unit tests (`internal/auth/hash_test.go`)
- [x] Token management infrastructure (Redis: `tokens:{tenant_id}:{token_id}`, `refresh_tokens:{tenant_id}:{user_id}:{token_id}`)
  - [x] AccessTokenKeyHandler (`internal/auth/keys_handlers/access_token.go`)
    - [x] Store, Get, Validate, Revoke, RevokeAll, Delete methods
    - [x] Unit tests (`internal/auth/keys_handlers/access_token_test.go`)
  - [x] RefreshTokenKeyHandler (`internal/auth/keys_handlers/refresh_token.go`)
    - [x] Store, Get, Validate, Revoke, RevokeAll, UpdateLastUsed, Delete methods
    - [x] Unit tests (`internal/auth/keys_handlers/refresh_token_test.go`)
  - [x] TokenIndex (`internal/auth/keys_handlers/token_index.go`)
    - [x] Redis Sets for efficient RevokeAll operations
    - [x] Indexes access and refresh tokens per tenant+user
    - [x] Unit tests (`internal/auth/keys_handlers/token_index_test.go`)
  - [x] TokenManager (`internal/auth/token_manager.go`)
    - [x] Unified interface for JWT operations and Redis storage
    - [x] StoreTokens, ValidateAccessTokenFromRedis, ValidateRefreshTokenFromRedis
    - [x] RefreshTokens (with token rotation), RevokeAllTokens
    - [x] Unit tests (`internal/auth/token_manager_test.go`)
  - [x] Documentation (`docs/auth/TOKEN_INFRASTRUCTURE.md`)
- [ ] Login endpoint (`Authenticate()` gRPC method)
- [ ] Session management (Redis: `sessions:{session_id}`)
- [ ] Logout endpoint
- [ ] Token refresh endpoint
- [ ] RBAC permission checking logic
- [ ] Permission checking endpoint (`CheckPermission()` gRPC method)
- [x] Role repository (MongoDB: `auth_db.roles`)
  - [x] `internal/auth/repository/roles_repo.go`
  - [x] CRUD operations with tenant isolation
  - [x] `GetRolesByTenantID`, `GetRolesByPermissionsIDs` methods
  - [x] Unit tests (`internal/auth/repository/roles_repo_test.go`)
- [x] Permission repository (MongoDB: `auth_db.permissions`)
  - [x] `internal/auth/repository/permissions_repo.go`
  - [x] CRUD operations with tenant isolation
  - [x] `GetPermissionsByTenantID`, `GetPermissionsByResource`, `GetPermissionsByAction` methods
  - [x] Unit tests (`internal/auth/repository/permissions_repo_test.go`)
- [x] Tenant repository (MongoDB: `auth_db.tenants`)
  - [x] `internal/auth/repository/tenants_repo.go`
  - [x] CRUD operations
  - [x] Unit tests (`internal/auth/repository/tenants_repo_test.go`)

**Key Endpoints:**
- `POST /auth/login` → gRPC `Authenticate()`
- `POST /auth/logout` → gRPC `Logout()`
- `POST /auth/refresh` → gRPC `RefreshToken()`
- `GET /auth/verify` → gRPC `VerifyToken()`
- `POST /rbac/check-permission` → gRPC `CheckPermission()`

**Port:** 5000

---

#### 2. Config Service (Priority 2) ⚙️
**Status:** ⬜ Not Started

**Why Second:** Simple service, needed for feature flags and dynamic configuration. Low complexity, high value.

**Prerequisites:**
- ✅ Pre-Phase infrastructure setup must be completed first (gRPC infrastructure)

**Dependencies:**
- Uses existing `db` package
- MongoDB (`config_db` collection)
- Redis (caching)
- gRPC infrastructure (from Pre-Phase)

**What to Build:**
- [ ] gRPC server implementation
- [ ] Config service proto definitions
- [ ] Config repository (MongoDB: `config_db.configurations`)
- [ ] Environment settings repository (MongoDB: `config_db.environment_settings`)
- [ ] Feature flags repository (MongoDB: `config_db.feature_flags`)
- [ ] Redis caching layer
- [ ] Config validation logic
- [ ] Config versioning
- [ ] GetConfig gRPC method
- [ ] UpdateConfig gRPC method
- [ ] Cache invalidation on updates
- [ ] Broadcast config updates to services

**Port:** 5002

---

### Phase 2: Core Business Logic 💼

#### 3. Core Service (Priority 3) 🏢
**Status:** ⬜ Not Started

**Why Third:** Contains main business logic. Depends on Auth for RBAC checks and Config for feature flags.

**Prerequisites:**
- ✅ Pre-Phase infrastructure setup (gRPC infrastructure)
- ✅ Auth Service (for RBAC permission checks)
- ✅ Config Service (for feature flags)

**Dependencies:**
- Auth Service (for RBAC permission checks)
- Config Service (for feature flags)
- MongoDB (`core_db` collection)
- Kafka (event publishing)
- gRPC infrastructure (from Pre-Phase)

**What to Build:**
- [ ] gRPC server implementation
- [ ] Core service proto definitions
- [ ] Products repository (MongoDB: `core_db.products`)
- [ ] Orders repository (MongoDB: `core_db.orders`)
- [ ] Vendors repository (MongoDB: `core_db.vendors`)
- [ ] Inventory repository (MongoDB: `core_db.inventory`)
- [ ] Business rules and validation
- [ ] Transaction management
- [ ] Kafka event publisher integration
- [ ] CreateOrder gRPC method
- [ ] UpdateOrder gRPC method
- [ ] Product CRUD operations
- [ ] Vendor CRUD operations
- [ ] Inventory management operations
- [ ] Event publishing for: `order.created`, `order.updated`, `product.updated`, `vendor.approved`
- [ ] Multi-tenancy filtering (tenant_id in all queries)

**Modules:**
- [ ] Users module
- [ ] Vendors module
- [ ] Products module
- [ ] Orders module
- [ ] Inventory module

**Port:** 5001

---

### Phase 3: Integration Layer 🔗

#### 4. Gateway (Priority 4) 🌐
**Status:** ⬜ Not Started

**Why Fourth:** Single entry point for WebUI. Depends on Auth and Core services being ready.

**Prerequisites:**
- ✅ Auth Service (JWT validation)
- ✅ Core Service (business operations)

**Dependencies:**
- Auth Service (JWT validation)
- Core Service (business operations)
- Config Service (optional, for config queries)
- Redis (caching, rate limiting)

**What to Build:**
- [ ] Create service Makefile (`internal/gateway/Makefile`) - Independent Makefile with proto, build, run, test, lint, clean targets
- [ ] GraphQL server setup (gqlgen)
- [ ] GraphQL schema definitions
- [ ] Auth middleware (JWT validation via Auth service gRPC)
- [ ] Request routing to backend services
- [ ] Query/Mutation resolvers
- [ ] Rate limiting & throttling (Redis: `rate_limit:{user_id}`)
- [ ] Response caching (Redis: `query_cache:{query_hash}`)
- [ ] Request aggregation
- [ ] Error handling and formatting
- [ ] Login mutation (calls Auth service)
- [ ] CreateOrder mutation (calls Core service)
- [ ] Query resolvers for products, orders, vendors, etc.

**Port:** 4000

---

#### 5. Events Service (Priority 5) 📡
**Status:** ⬜ Not Started

**Why Fifth:** Consumes events from Kafka. Can be built in parallel with Gateway.

**Dependencies:**
- Core Service (consumes its events)
- Kafka (consumer)

**What to Build:**
- [ ] Create service Makefile (`internal/events/Makefile`) - Independent Makefile with proto, build, run, test, lint, clean targets
- [ ] Kafka consumer setup (sarama/confluent-kafka-go)
- [ ] Event handlers for different event types
- [ ] Notification system (Email, SMS, Push)
- [ ] Audit logging
- [ ] Alerting & monitoring
- [ ] Observability metrics (OpenTelemetry, Prometheus)
- [ ] Event routing logic
- [ ] Handler for `user.created` events
- [ ] Handler for `order.placed` events
- [ ] Handler for `product.updated` events
- [ ] Handler for `vendor.approved` events
- [ ] Handler for `system.alert` events

**Port:** 5003

---

### Phase 4: Frontend 🎨

#### 6. WebUI (Priority 6) 💻
**Status:** ⬜ Not Started

**Why Last:** Depends on Gateway being ready to provide GraphQL API.

**Dependencies:**
- Gateway (GraphQL API)

**What to Build:**
- [ ] Create service Makefile (`internal/webui/Makefile`) - Independent Makefile with build, run, test, lint, clean targets (may not need proto if using GraphQL)
- [ ] React 18+ project setup
- [ ] TypeScript configuration
- [ ] Apollo Client setup (GraphQL client)
- [ ] State management (Redux/Zustand)
- [ ] UI framework setup (TailwindCSS/Material-UI)
- [ ] Login page with tenant selection
- [ ] Dashboard
- [ ] Form handling & validation
- [ ] Order management UI
- [ ] Product management UI
- [ ] Vendor management UI
- [ ] Inventory management UI
- [ ] User management UI (for admins)
- [ ] Configuration UI (for admins)

**Port:** 443 (HTTPS)

---

## Key Flows Implementation Status

### 1. User Login Flow
**Status:** ⬜ Not Started

1. [ ] WebUI → Gateway: mutation login(email, password, tenant_id)
2. [ ] Gateway → Auth: gRPC Authenticate()
3. [ ] Auth → Redis: Validate & create session
4. [ ] Auth → Gateway: Return JWT + refresh token (with tenant_id in claims)
5. [ ] Gateway → WebUI: Return tokens + user info
6. [ ] WebUI: Store tokens, redirect to dashboard

### 2. Create Order Flow
**Status:** ⬜ Not Started

1. [ ] WebUI → Gateway: mutation createOrder(input)
2. [ ] Gateway → Auth: Verify JWT token (extract tenant_id)
3. [ ] Gateway → Core: gRPC CreateOrder()
4. [ ] Core → MongoDB: Insert order document (with tenant_id)
5. [ ] Core → Kafka: Publish "order.created" event
6. [ ] Core → Gateway: Return order data
7. [ ] Events Service: Consume event → Send notification
8. [ ] Gateway → WebUI: Return created order

### 3. Configuration Update Flow
**Status:** ⬜ Not Started

1. [ ] Admin changes feature flag in WebUI
2. [ ] Gateway → Config: UpdateConfig()
3. [ ] Config → MongoDB: Update config document
4. [ ] Config → Redis: Invalidate cache
5. [ ] Config: Broadcast update to all services
6. [ ] Services: Reload configuration

---

## Technical Decisions

### Multi-tenancy
- ✅ Tenant ID captured from login form
- ✅ Tenant ID stored in JWT claims
- ✅ All queries filtered by tenant_id

### Inter-service Communication
- ✅ All inter-service communication via gRPC

### Authentication & Authorization
- ✅ User credentials stored in MongoDB (`auth_db.users`)
- ✅ Credentials cached in Redis
- ✅ Auth service enforces RBAC based on operation and role permissions

### Database Access
- ✅ Each component creates a repository service for its db+collection
- ✅ Uses generic Repository pattern from `internal/db/repository.go`

### Code Organization
- ✅ Starting as monorepo with multiple packages
- ✅ Will break down to microservices and shared Go modules later
- ✅ Models organized by service for easy future separation:
  - `internal/auth/models/` - Auth models (Tenant, User, Role, Permission, UserGroup, AuditLog)
  - `internal/core/models/` - Core models (Product, Order, Vendor, Customer, Inventory, etc.)
  - `internal/config/models/` - Config models (ServiceConfig, FeatureFlag)
  - `internal/db/redis/models/` - Redis cache models (Session, TokenMetadata, caches)

### Infrastructure Notes
- ⚠️ MongoDB and Redis connection URIs are currently hardcoded in `internal/db/mongo/mongo.go` and `internal/db/redis/redis.go`
- ⚠️ Will be moved to environment configuration later (not blocking for initial development)

---

---

## Future Features (To Be Planned)

### Data Import from Files 📁
**Status:** 📝 Planned (not yet designed)

Import data from external files (CSV, JSON, Excel, etc.) into the ERP system.

**Potential scope:**
- Products import
- Vendors import
- Customers import
- Inventory import
- Orders import (historical)

*Details to be planned when we reach this phase.*

---

## Development Standards

### Unit Testing Requirements 🧪
**Every feature/component must include unit tests with:**
- ✅ Positive test cases (expected successful behavior)
- ✅ Negative test cases (error handling, edge cases, invalid inputs)
- ✅ Table-driven tests where applicable
- ✅ Use `testify` for assertions (`assert`, `require`)

**Test file naming:** `<filename>_test.go` in the same package

**Example structure:**
```go
func TestFunctionName(t *testing.T) {
    testCases := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {name: "valid input", input: ..., want: ..., wantErr: false},
        {name: "invalid input", input: ..., want: ..., wantErr: true},
    }
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

---

## Notes
- Update status checkboxes (⬜ → ✅) as items are completed
- Add notes or blockers in the relevant sections
- Update this roadmap as architecture evolves
- Infrastructure setup (Pre-Phase) should be completed before starting Phase 1 services
- **All new code must include unit tests before marking as complete**

