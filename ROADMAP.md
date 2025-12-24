# ERP System Development Roadmap

## Overview
This roadmap outlines the development order for building the multi-tenant ERP system. Services are organized by priority and dependencies to ensure efficient development.

## Pre-Phase: Infrastructure Setup 🏗️

Before starting service development, we need to set up foundational infrastructure that all services will depend on.

**Status:** 🟡 In Progress (gRPC ✅, JWT ✅, Error Handling ✅, Service Structure ⬜, Build Tooling ⬜)

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
  - [x] JWTManager struct (`internal/auth/jwt.go`)
  - [x] GenerateToken method (with userID and tenantID)
  - [x] VerifyToken method
  - [x] RefreshToken method
  - [x] RevokeToken method

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

---

### 4. Service Structure & Lifecycle (Helpful) 🔄
**Status:** ⬜ Not Started

**Why Fourth:** Establishes patterns for how services are organized and started.

**What to Build:**
- [ ] Define service initialization pattern
- [ ] Create service lifecycle management (start/stop)
- [ ] Add health check endpoint pattern
- [ ] Create service main entry point template

**Note:** Can evolve as we build services, but good to have a starting pattern.

---

### 5. Build Tooling (Helpful) 🛠️
**Status:** ⬜ Not Started

**Why Last:** Helpful for development workflow, but not blocking.

**What to Build:**
- [ ] Create Makefile with common tasks:
  - [ ] `make proto` - Generate proto code
  - [ ] `make build` - Build services
  - [ ] `make run-auth` - Run auth service
  - [ ] `make test` - Run tests
- [ ] Or create build scripts (`.sh` or `.bat`)

**Note:** MongoDB and Redis connection URIs are currently hardcoded. Will be moved to environment configuration later.

---

## Development Phases

### Phase 1: Foundation ⚙️

#### 1. Auth Service (Priority 1) 🔐
**Status:** ⬜ Not Started

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
- [ ] User repository using generic Repository pattern (MongoDB: `auth_db.users`)
- [x] JWT generation/validation library integration
  - [x] JWTManager implementation (`internal/auth/jwt.go`)
  - [x] GenerateToken with tenantID support
  - [x] VerifyToken implementation
  - [x] RefreshToken implementation
- [x] JWT claims structure (include tenant ID)
  - [x] Claims include `sub` (userID), `tenant_id`, and `exp`
- [ ] Password hashing (bcrypt)
- [ ] Login endpoint (`Authenticate()` gRPC method)
- [ ] Session management (Redis: `sessions:{session_id}`)
- [ ] Token management (Redis: `tokens:{token_id}`, `refresh_tokens:{user_id}`)
- [ ] Logout endpoint
- [ ] Token refresh endpoint
- [ ] RBAC permission checking logic
- [ ] Permission checking endpoint (`CheckPermission()` gRPC method)
- [ ] Role repository (MongoDB: `auth_db.roles`)
- [ ] Permission repository (MongoDB: `auth_db.permissions`)
- [ ] Tenant repository (MongoDB: `auth_db.tenants`)

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

### Infrastructure Notes
- ⚠️ MongoDB and Redis connection URIs are currently hardcoded in `internal/db/mongo/mongo.go` and `internal/db/redis/redis.go`
- ⚠️ Will be moved to environment configuration later (not blocking for initial development)

---

## Notes
- Update status checkboxes (⬜ → ✅) as items are completed
- Add notes or blockers in the relevant sections
- Update this roadmap as architecture evolves
- Infrastructure setup (Pre-Phase) should be completed before starting Phase 1 services

