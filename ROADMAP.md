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
- [x] Create proto files directory structure (service-specific proto dirs + `internal/infra/proto/`)
  - [x] `internal/infra/proto/` - Shared types
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
  - [x] Common proto file (`internal/infra/proto/shared.proto`)
  - [x] Template documentation in `docs/proto/README.md`
- [x] Document proto generation workflow

**Note:** Proto definitions for each service will be created as part of that service's development.

**Directory Structure:**
```
internal/
├── infra/proto/         # Shared types (errors, base messages)
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
  - [x] `internal/infra/errors/grpc.go` - ToGRPCError/FromGRPCError functions
  - [x] Category to gRPC code mapping (AUTH → Unauthenticated, VALIDATION → InvalidArgument, etc.)
- [x] Create error handling utilities
  - [x] `internal/infra/errors/errors.go` - AppError type with constructors
  - [x] Helper functions: New(), Wrap(), Auth(), Validation(), NotFound(), Conflict(), Business(), Internal()
- [x] Document error response format
  - [x] Updated `internal/infra/proto/shared.proto` with ErrorCategory enum and enhanced Error message
- [x] Create common error types
  - [x] `internal/infra/errors/codes.go` - Categorized error codes (AUTH, VALIDATION, NOT_FOUND, CONFLICT, BUSINESS, INTERNAL)

**Files Created:**
- `internal/infra/errors/errors.go` - Core error types and constructors
- `internal/infra/errors/codes.go` - Error code definitions by category
- `internal/infra/errors/grpc.go` - gRPC status code mapping
- `internal/infra/errors/errors_test.go` - Unit tests for error handling

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

### 6. Model Organization & Architecture Refactoring (Completed) 📦
**Status:** ✅ Completed

**What was Built:**
- [x] **Major Architecture Refactoring**:
  - [x] `common/` directory renamed to `infra/` and moved to `internal/infra/`
  - [x] All shared infrastructure consolidated in `internal/infra/`
  - [x] User collection ownership moved from Auth → Core module
  - [x] RBAC manager refactored to only check permissions (no user CRUD)
  - [x] Audit log ownership moved from Auth → Events module
- [x] Organized infrastructure by purpose for easier management
- [x] `internal/infra/models/auth/` - Auth models (Tenant, Role, Permission, Token, Session, etc.)
- [x] `internal/infra/models/core/` - Core models (User, Product, Order, Vendor, Customer, Inventory, etc.)
- [x] `internal/infra/models/config/` - Config models (ServiceConfig, FeatureFlag)
- [x] `internal/infra/models/gateway/` - Gateway models (RateLimitInfo, QueryCache, etc.)
- [x] `internal/infra/models/events/` - Events models (AuditLog, etc.)
- [x] Validation methods on all models (`Validate(createOperation bool)`)

**Directory Structure:**
```
internal/
├── infra/                          # All shared infrastructure
│   ├── models/                     # Domain models by module
│   │   ├── auth/                   # Tenant, Role, Permission, Token, Session, etc.
│   │   ├── core/                   # User, Product, Order, Vendor, Inventory, etc.
│   │   ├── config/                 # ServiceConfig, FeatureFlag
│   │   ├── gateway/                # RateLimitInfo, QueryCache
│   │   ├── events/                 # AuditLog, etc.
│   │   └── shared/                 # All Shared models
│   ├── grpc/                       # Generic gRPC infrastructure (to be built)
│   ├── proto/                      # Shared proto definitions
│   ├── errors/                     # Error handling utilities
│   ├── db/                         # Database interfaces (MongoDB, Redis)
│   ├── logging/                    # Logging utilities
│   └── events/                     # Event publishing/consuming
│
├── auth/                           # Auth service
│   ├── proto/                      # Auth-specific proto
│   ├── repository/                 # Roles, Permissions, Tenants repos
│   ├── token/                      # Token management
│   └── rbac/                       # RBAC manager
│
├── core/                           # Core service
│   ├── proto/                      # Core-specific proto (user.proto, etc.)
│   └── repository/                 # Users repo
│
├── config/                         # Config service
├── gateway/                        # Gateway service
└── events/                         # Events service
```

**Impact on Services:**
- ⚠️ Auth Service: Has TODOs for Core User service integration
- ⚠️ Unit Tests: Need updates for new import paths (`internal/infra/...`)

---

## Code Quality Initiative: Infrastructure Consolidation 📦

**Status:** ✅ Complete - All shared code centralized to `internal/infra/`

**Why Important:** Centralized infrastructure management in `internal/infra/` improves code organization, eliminates circular dependencies, and provides clear separation between services and shared utilities.

**What Was Done:**

### ✅ Major Architecture Refactoring (100% Complete)
- [x] **Directory Restructure**: `common/` → `internal/infra/`
- [x] **Infrastructure Consolidation**: All shared code moved to `internal/infra/`
  - [x] `internal/infra/models/` - Domain models by module
    - [x] `internal/infra/models/auth/` - Auth models (Tenant, Role, Permission, Token, Session, etc.)
    - [x] `internal/infra/models/core/` - Core models (User, Product, Order, Vendor, Customer, Inventory, etc.)
    - [x] `internal/infra/models/config/` - Config models (ServiceConfig, FeatureFlag, etc.)
    - [x] `internal/infra/models/gateway/` - Gateway models (RateLimitInfo, QueryCache, etc.)
    - [x] `internal/infra/models/events/` - Events models (AuditLog, etc.)
  - [x] `internal/infra/proto/` - Shared proto definitions
  - [x] `internal/infra/errors/` - Error handling (moved from `internal/errors/`)
  - [x] `internal/infra/db/` - Database interfaces (moved from `internal/db/`)
  - [x] `internal/infra/logging/` - Logging utilities (moved from `internal/logging/`)
  - [x] `internal/infra/events/` - Event infrastructure (moved from `internal/events/`)
- [x] **Service Ownership Changes**:
  - [x] User collection: Auth → Core module
  - [x] Audit log: Auth → Events module
  - [x] RBAC manager: Refactored to only check permissions (no user CRUD)
- [x] **TODO Comments**: Added throughout for cross-service dependencies

**Benefits Achieved:**
- ✅ Single source of truth - all infrastructure in `internal/infra/`
- ✅ Clear separation - services vs infrastructure
- ✅ Easier cross-service sharing - no circular dependencies
- ✅ Simplified imports - `internal/infra/models/auth`, `internal/infra/db`, etc.
- ✅ Future-proof - easy to extract as separate module

**Impact:**
- ⚠️ Import paths need updates: `internal/{module}/models` → `internal/infra/models/{module}`
- ⚠️ Infrastructure imports: `internal/errors` → `internal/infra/errors`, etc.
- ⚠️ Auth service has TODOs for Core User service integration
- ⚠️ Unit tests need fixes for new architecture

**Previous Breakdown Work (Preserved):**
The detailed model breakdown work from the previous organization (breaking monolithic `models.go` into focused files) has been preserved during the move to `internal/infra/`. Models are still organized by entity with validation methods and comprehensive tests.

---

## Code Quality Initiative: Test Refactoring (gomock) 🧪

**Status:** ✅ Complete (All Tests Refactored and Stable)

**Why Important:** Using `gomock.Any()` in tests makes them too generic and doesn't properly validate that correct parameters are being passed to mocked methods. Specific test values improve test quality and catch more bugs.

**Refactoring Rules Applied:**
1. ✅ NEVER use `gomock.Any()` under any circumstances
2. ✅ Create custom matchers ONLY for objects/structs with dynamically-set timestamps (CreatedAt, UpdatedAt, Timestamp)
3. ✅ Matchers skip validating ONLY timestamp fields
4. ✅ Pass specific values directly (no `gomock.Eq()` wrappers)
5. ✅ Use specific expected values in test cases (expectedFilter, expectedKey, etc.)
6. ✅ Use specific names like "users", "roles", "tenants", "permissions", "audit_logs"

### ✅ Completed: Collection Tests (internal/auth/collections/)

**Files Refactored:**
- [x] `permissions_test.go` - Created `permissionMatcher` to skip CreatedAt/UpdatedAt validation
- [x] `audit_logs_test.go` - Created `auditLogMatcher` to skip Timestamp validation
- [x] `roles_test.go` - Created `roleMatcher` to skip CreatedAt/UpdatedAt validation
- [x] `tenants_test.go` - Created `tenantMatcher` to skip CreatedAt/UpdatedAt validation
- [x] `users_test.go` - Created `userMatcher` to skip CreatedAt/UpdatedAt validation

### ✅ Completed: RBAC Manager Tests (internal/auth/rbac/)

**Files Created:**
- [x] `rbac_manager_test.go` - Comprehensive unit tests using MockCollectionHandler[T]
  - [x] TestRBACManager_GetUserPermissions (5 test cases)
  - [x] TestRBACManager_GetUserRoles (3 test cases)
  - [x] TestRBACManager_GetRolePermissions (3 test cases)
  - [x] TestRBACManager_CheckUserPermissions (3 test cases)
  - [x] TestRBACManager_VerifyUserRole (3 test cases)
  - [x] TestRBACManager_VerifyRolePermissions (2 test cases)

**Test Strategy:**
- Test helpers create collections with mocked CollectionHandler[T]
- No logic code modified - leveraged existing generic mocks
- All tests use specific expected values (no gomock.Any())

### ✅ Completed: Redis Handler Tests (internal/infra/db/redis/handlers/)

**Files Created:**
- [x] `set_handler_test.go` - Comprehensive tests for BaseSetHandler
  - [x] TestNewBaseSetHandler (constructor tests)
  - [x] TestBaseSetHandler_Add (with and without TTL)
  - [x] TestBaseSetHandler_Remove
  - [x] TestBaseSetHandler_Members (multiple scenarios)
  - [x] TestBaseSetHandler_Clear

### ✅ Completed: Token Index Tests (internal/auth/token/)

**Files Created:**
- [x] `token_index_test.go` - Complete test coverage from scratch (11 test functions, 21 test cases)
  - [x] Constructor tests (with mocks and nil handlers)
  - [x] Access token operations (Add, Remove, Get, Clear)
  - [x] Refresh token operations (Add, Remove, Get, Clear)
  - [x] Integration test (multiple operations workflow)

**Pattern Established:**
```go
// Custom matcher for objects with dynamic timestamps
type userMatcher struct {
    expected models.User
}

func (m userMatcher) Matches(x interface{}) bool {
    user, ok := x.(models.User)
    if !ok {
        return false
    }
    // Match all fields EXCEPT CreatedAt/UpdatedAt
    return user.TenantID == m.expected.TenantID &&
        user.Email == m.expected.Email &&
        user.Username == m.expected.Username &&
        // ... other fields
}

func (m userMatcher) String() string {
    return "matches user fields except CreatedAt and UpdatedAt"
}

// Usage in tests
mockHandler.EXPECT().
    Create("users", userMatcher{expected: tc.user}).
    Return(tc.returnID, tc.returnError).
    Times(tc.expectedCallTimes)

mockHandler.EXPECT().
    FindOne("users", tc.expectedFilter).
    Return(tc.returnData, tc.returnError)
```

**Benefits Achieved:**
- ✅ More robust tests that validate exact parameters
- ✅ Better error detection (tests fail if wrong parameters are used)
- ✅ Improved test readability (explicit values instead of wildcards)
- ✅ Verified: NO `gomock.Any()` usage in any tests
- ✅ All tests passing and stable
- ✅ 100+ comprehensive test cases across all modules

---

## Development Phases

### Phase 1: Foundation ⚙️

#### 1. Auth Service (Priority 1) 🔐
**Status:** ✅ Complete (pagination support pending)

**Why First:** Required by all other services for authentication/authorization. Foundation for the entire system.

**Note:** Auth service manages users directly (UserService, UserAPI, UserHandler). RBAC verification is handled by VerificationManager with wildcard permission support (`*:*`, `resource:*`, `*:action`) and admin short-circuits. Audit logs deferred to Events module. Pagination for ListUsers/ListRoles/ListTenants is the only remaining feature gap.

**Prerequisites:**
- ✅ Pre-Phase infrastructure setup must be completed first (gRPC infrastructure, JWT library)

**Dependencies:**
- Uses existing `db` package (✅ Enhanced with opts parameter for future TTL support)
- MongoDB (`auth_db` collection) - ✅ Auto-creates collections via CreateCollectionInDBIfNotExists
- Redis (sessions/tokens)
- gRPC infrastructure (from Pre-Phase)
- JWT library (from Pre-Phase)

**What to Build:**
- [x] gRPC server implementation (structure complete, mTLS disabled for local testing, needs main.go entry point)
- [x] Auth service proto definitions (`.proto` files)
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
- [x] Login endpoint (`Authenticate()` gRPC method)
- [x] Logout endpoint (`Logout()` gRPC method) - ✅ Implemented with token revocation and audit logging (audit logs commented out)
- [x] Token verification endpoint (`VerifyToken()` gRPC method)
- [x] Token refresh endpoint (`RefreshToken()` gRPC method) - ✅ Implemented with token rotation
- [x] Token revocation endpoint (`RevokeToken()` gRPC method)
- [x] RBAC permission checking endpoint (`CheckPermissions()` gRPC method)
- [x] RBAC manager implementation (`internal/auth/rbac/rbac_manager.go`)
  - [x] CRUD resource operations (Create, Update, Delete, Get, GetAll) with permission checks
  - [x] Permission management (GetUserPermissions, GetUserRoles, GetRolePermissions)
  - [x] Permission verification (CheckUserPermissions, VerifyUserRole, VerifyRolePermissions)
  - [x] Supports User, Role, and Permission resource types
  - [x] Handles role-based permissions, additional permissions, and revoked permissions
  - [x] Unit tests (`internal/auth/rbac/rbac_manager_test.go`) - Comprehensive table-driven tests
- [ ] Session management (Redis: `sessions:{session_id}`) - Deferred to later phase
- [ ] Audit logs collection (`internal/auth/collections/audit_logs.go`)
  - [ ] CRUD operations with tenant isolation
  - [ ] Enhanced audit models with detailed change tracking
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
- `POST /auth/login` → gRPC `Login()` ✅
- `POST /auth/logout` → gRPC `Logout()` ✅
- `POST /auth/refresh` → gRPC `RefreshToken()` ✅
- `POST /auth/verify` → gRPC `VerifyToken()` ✅
- `POST /auth/revoke` → gRPC `RevokeToken()` ✅
- `POST /rbac/check-permissions` → gRPC `CheckPermissions()` ✅

**Infrastructure Improvements (Added During Auth Service Development):**
- [x] Enhanced DBHandler interface with opts parameter
  - [x] `Create(db string, data any, opts ...map[string]any)` - Support for future TTL configuration
  - [x] `Update(db string, filter map[string]any, data any, opts ...map[string]any)` - Support for future options
  - [x] `Close()` method added for proper cleanup
  - [x] MongoDB implementation updated
  - [x] Redis implementation updated
  - [x] MockDBHandler updated for testing
- [x] Auto-create MongoDB collections
  - [x] `CreateCollectionInDBIfNotExists()` in MongoDBManager
  - [x] Called automatically in `NewCollectionHandler`
  - [x] Gracefully handles mocks (returns nil for non-MongoDB handlers)
- [x] Helper methods in AuthService
  - [x] `generateAccessToken()` - Extract access token generation logic
  - [x] `generateRefreshToken()` - Extract refresh token generation logic
  - [x] `generateAndStoreTokens()` - Unified token generation and storage
  - [x] `revokeTokens()` - Unified token revocation logic

**Test Status:**
- ✅ **Functional tests passing** - 108 tests total, 105 running, 3 skipped (all pagination)
  - Happy-path: auth (6), user (5), role (5), permission (3), tenant (6), rbac (5)
  - Negative: auth (13), user (20), role (14), permission (5), tenant (12), rbac (14)
  - Skipped: `test_list_users_invalid_pagination`, `test_list_roles_invalid_pagination`, `test_list_tenants_invalid_pagination`
- ⚠️ **Unit tests** - Need review after recent error handling and permission registry refactoring
- ✅ Test quality standards preserved:
  - NO gomock.Any() usage
  - Custom matchers for dynamic timestamps
  - Specific expected values in all tests

**Completed Infrastructure:**
- [x] Token infrastructure (AccessToken, RefreshToken, TokenIndex, TokenManager)
- [x] Core endpoints (Login, Logout, Refresh, Verify, Revoke, CheckPermissions)
- [x] RBAC manager (refactored - permission checking only)
- [x] gRPC server implementation
- [x] Repository layers for Roles, Permissions, Tenants
- [x] Password hashing utilities

**TODOs to Fix:**
- [x] User operations — Auth service manages users directly (UserService, UserAPI, UserHandler). No Core User service call needed.
- [x] Import paths — All service code uses current `internal/infra/` paths.
- [x] Stale TODO comment in `internal/auth/api/user.go` line 143 (`// TODO: finish logic`) — logic is complete (`validateUserUpdateData` performs full diff); comment should be removed.
- [x] Future enhancement: `internal/auth/api/tenant.go` line 192 (`//TODO: Do diff and validate`) — UpdateTenant validates via `ValidateTenant()` but lacks the field-level diff that UpdateUser has. Functional and tested, but a hardening item.
- [ ] Audit logging via Events service (deferred to Priority 11)
- [ ] Pagination support for ListUsers, ListRoles, ListTenants (3 functional tests currently skipped)
- [ ] Unit test review after recent error handling and permission registry refactoring

**Notes:**
- User management is handled directly by Auth Service (UserService, UserAPI, UserHandler)
- RBAC verification handled by VerificationManager (wildcard permissions, cross-tenant support, admin short-circuits)
- Role and Permission CRUD is in Auth (RoleService, PermissionService via `rbac_role.go`, `rbac_permission.go`)
- Audit logs deferred to Events Service (Priority 11)
- Functional tests: 108 tests across 12 files in `internal/auth/functional/`
- mTLS support deferred to Config Service enhancement (Priority 8)

**Port:** 5000

---

#### 2. Generic gRPC Infrastructure (Priority 1.5 - NEW) 🔧
**Status:** ⬜ Not Started

**Why Now:** Services need to communicate via gRPC (Auth ↔ Core, Core ↔ Config, etc.). Building generic infrastructure avoids duplication and establishes consistent patterns.

**Prerequisites:**
- ✅ Pre-Phase gRPC infrastructure (proto generation, build scripts)

**Dependencies:**
- gRPC Go libraries (already installed)
- Proto definitions (from Pre-Phase)

**What to Build:**
- [-] Generic gRPC server infrastructure (`internal/infra/grpc/server/`)
  - [x] Server initialization utilities
  - [x] Graceful shutdown handling
  - [ ] Health check endpoints
  - [x] Server configuration struct
  - [x] Interceptor registration helpers
- [x] Generic gRPC client infrastructure (`internal/infra/grpc/client/`)
  - [x] Client connection management
  - [x] Connection pooling utilities (basic - advanced features deferred)
  - [x] Client configuration struct
  - [x] Interceptor registration helpers
- [-] Shared middleware/interceptors (`internal/infra/grpc/interceptors/`)
  - [x] Logging interceptor (request/response logging)
  - [ ] Error handling interceptor (standardized error conversion)
  - [ ] Authentication interceptor (JWT validation for service-to-service calls)
  - [ ] Metrics interceptor (basic request metrics)
- [x] Error handling utilities
  - [x] Standard gRPC status code mapping
  - [x] Error conversion helpers
  - [x] Integration with `internal/infra/errors/` package
- [ ] Documentation
  - [ ] Usage examples for server setup
  - [ ] Usage examples for client usage
  - [ ] Interceptor configuration guide

**Future Enhancements (Deferred):**
- [ ] Advanced connection pooling strategies
- [ ] Retry logic with exponential backoff
- [ ] Circuit breaker pattern
- [ ] Load balancing
- [ ] Service discovery integration
- [ ] Distributed tracing (OpenTelemetry)

**Deliverables:**
- Reusable gRPC server and client infrastructure in `internal/infra/grpc/`
- Shared interceptors for common concerns
- Documentation and usage examples
- Ready to use for Core User Service and cross-service communication

---

#### 3. Config Service - Phase 1 (Priority 5 - MOVED) ⚙️
**Status:** ⬜ Not Started

**Why Second:** Simple service, needed for feature flags and dynamic configuration. Required by User Service (Core module). Starting with .env approach for simplicity, structured for future mTLS enhancement.

**Prerequisites:**
- ✅ Pre-Phase infrastructure setup (gRPC infrastructure)
- ✅ Auth Service (for RBAC permission checks)

**Dependencies:**
- Auth Service (for CheckPermissions gRPC calls)
- .env files (config.env, secret.env, etc.)
- Redis (optional caching)
- gRPC infrastructure (from Pre-Phase)

**What to Build (Phase 1 - Simple Version):**
- [ ] .env file loading infrastructure
  - [ ] Support multiple .env files (config.env, secret.env, database.env, etc.)
  - [ ] Environment variable parsing and validation
  - [ ] Type-safe configuration structs
- [ ] gRPC server implementation
- [ ] Config service proto definitions
  - [ ] GetConfig RPC method
  - [ ] SetConfig RPC method
  - [ ] ListConfigs RPC method
- [ ] In-memory configuration storage (loaded from .env files)
- [ ] Configuration validation logic
- [ ] Simple RBAC via Auth service
  - [ ] Call Auth.CheckPermissions for user-level authorization
  - [ ] Verify user has permission to read/write specific configs
- [ ] Redis caching layer (optional for performance)
- [ ] Code structure prepared for future mTLS
  - [ ] Abstract authentication/authorization into interfaces
  - [ ] Separate transport security from business logic
  - [ ] Design for future certificate-based module authentication

**Key Design Decisions:**
- **Configuration Source**: .env files (not MongoDB) - industry standard, easier to manage
- **Security Model (Phase 1)**: User-level RBAC only (via Auth service)
- **Future Enhancement (Phase 5)**: Add mTLS + certificate-based module authentication

**Port:** 5002

---

### Phase 2: Core Business Logic 💼

#### 4. Core User Service (Priority 2 - MOVED UP) 👥
**Status:** ⬜ Not Started

**Why Second (After gRPC Infrastructure):** User management is core business logic (not authentication). Auth service depends on this for user operations. Required for functional testing infrastructure (test setup/cleanup).

**Prerequisites:**
- ✅ Auth Service (partially complete, will be finished after this)
- ✅ Generic gRPC Infrastructure (Priority 1.5)

**Dependencies:**
- Generic gRPC Infrastructure (client/server from `internal/infra/grpc/`)
- Auth Service (for CheckPermissions gRPC calls - optional for now)
- MongoDB (`core_db.users` collection)
- Proto definitions

**What to Build:**
- [ ] User service proto definitions (`internal/core/proto/user.proto`)
  - [ ] CreateUser RPC method
  - [ ] GetUser RPC method (by ID, by email)
  - [ ] UpdateUser RPC method
  - [ ] DeleteUser RPC method
  - [ ] ListUsers RPC method (with pagination, filtering by tenant)
- [ ] User repository (MongoDB: `core_db.users`)
  - [ ] CRUD operations with tenant isolation
  - [ ] User profile management
  - [ ] Metadata and preferences storage
  - [ ] Uses models from `internal/infra/models/core/`
- [ ] gRPC server implementation
  - [ ] Use generic gRPC server infrastructure from `internal/infra/grpc/`
  - [ ] Implement all User service RPC methods
  - [ ] Can be extended for future Core modules (Products, Orders, etc.)
- [ ] RBAC integration with Auth service (optional - can defer)
  - [ ] Call Auth.CheckPermissions before user operations
  - [ ] Verify user has permission to create/read/update/delete users
- [ ] System admin data seeding
  - [ ] Use CollectionHandlers directly with hard-coded data
  - [ ] Create default tenant
  - [ ] Create default roles (SuperAdmin, Admin, User) - via Auth service
  - [ ] Create default permissions - via Auth service
  - [ ] Create system admin user
  - [ ] Seeding script or initialization function

**Scope Notes:**
- This is ONLY user management, not full Core service
- Other Core modules (Products, Orders, Vendors, Inventory) come later (Priority 9)
- Minimal implementation to unblock Auth service fixes

**Port:** 5001 (shared with future Core service modules)

---

#### 5. Fix TODOs - Auth Service Dependencies (Priority 3 - NEW) 🔧
**Status:** ⬜ Not Started

**Why Third:** Auth service has TODOs that depend on Core User service. Must resolve these to complete Auth service.

**Prerequisites:**
- ✅ Generic gRPC Infrastructure (Priority 1.5)
- ✅ Core User Service (Priority 2)

**What to Fix:**
- [ ] Update Auth service to use Core User service gRPC client
  - [ ] Replace direct User collection access with gRPC calls to Core
  - [ ] Update Login endpoint to call Core.GetUser
  - [ ] Remove User repository from Auth service (moved to Core)
- [ ] Update RBAC manager
  - [ ] Already refactored to only check permissions (no user CRUD)
  - [ ] Verify it works with new architecture
- [ ] Audit logging integration
  - [ ] Add TODO comments for Events service integration (deferred to Priority 11)
  - [ ] Temporarily disable or comment out audit logging in Auth endpoints
- [ ] Review and resolve all remaining TODO comments in Auth service

**Deliverables:**
- Auth service fully functional with Core User service integration
- All critical TODOs resolved or documented for future work
- Auth service ready for unit testing

---

#### 6. Fix Unit Tests (Priority 4 - NEW) 🧪
**Status:** ⬜ Not Started

**Why Fourth:** Tests are broken after refactoring. Must fix before proceeding with new development.

**Prerequisites:**
- ✅ Architecture refactoring complete
- ✅ Auth service TODOs resolved (Priority 3)

**What to Fix:**
- [ ] Update import paths across all tests
  - [ ] Change `internal/{module}/models` → `internal/infra/models/{module}`
  - [ ] Change `internal/errors` → `internal/infra/errors`
  - [ ] Change `internal/db` → `internal/infra/db`
  - [ ] Change `internal/logging` → `internal/infra/logging`
  - [ ] Change `internal/events` → `internal/infra/events`
- [ ] Fix Auth service tests
  - [ ] Update tests to use Core User service mock/client
  - [ ] Fix repository tests (roles, permissions, tenants)
  - [ ] Fix RBAC manager tests
  - [ ] Fix token infrastructure tests
  - [ ] Preserve test quality (no gomock.Any(), custom matchers)
- [ ] Fix Core service tests (if any exist)
  - [ ] Update User repository tests to use new model locations
- [ ] Fix infrastructure tests
  - [ ] Update database handler tests (`internal/infra/db/`)
  - [ ] Update error handling tests (`internal/infra/errors/`)
- [ ] Run all tests and verify they pass
  - [ ] `make test` should succeed
  - [ ] All 100+ tests should pass
- [ ] Update test documentation if needed

**Test Quality Standards (MUST PRESERVE):**
- ✅ NO `gomock.Any()` usage
- ✅ Custom matchers only for dynamic timestamps
- ✅ Specific expected values in all test cases
- ✅ Table-driven tests where applicable

**Deliverables:**
- All unit tests passing
- Test coverage maintained
- Test quality standards preserved

---

#### 7. Core Service - Remaining Modules (Priority 9 - MOVED) 🏢
**Status:** ⬜ Not Started (Deferred after functional testing)

**Why Deferred:** User Service (Priority 2) provides enough functionality for initial testing. Other Core modules (Products, Orders, Vendors, Inventory) can wait until after functional testing infrastructure is proven.

**Prerequisites:**
- ✅ Auth Service (for RBAC permission checks)
- ✅ Config Service (for feature flags)
- ✅ Core User Service (shares gRPC server)
- ✅ Functional Testing Infrastructure (to ensure quality)

**Dependencies:**
- Auth Service (for RBAC permission checks)
- Config Service (for feature flags)
- MongoDB (`core_db` collections)
- Kafka (event publishing)
- gRPC infrastructure (from Pre-Phase)

**What to Build:**
- [ ] Products module
  - [ ] Products repository (MongoDB: `core_db.products`)
  - [ ] Product CRUD gRPC methods
  - [ ] Product validation and business rules
  - [ ] Event publishing: `product.created`, `product.updated`, `product.deleted`
- [ ] Orders module
  - [ ] Orders repository (MongoDB: `core_db.orders`)
  - [ ] CreateOrder, UpdateOrder, GetOrder, ListOrders gRPC methods
  - [ ] Order validation and business rules
  - [ ] Event publishing: `order.created`, `order.updated`, `order.cancelled`
- [ ] Vendors module
  - [ ] Vendors repository (MongoDB: `core_db.vendors`)
  - [ ] Vendor CRUD gRPC methods
  - [ ] Vendor approval workflow
  - [ ] Event publishing: `vendor.created`, `vendor.approved`, `vendor.rejected`
- [ ] Inventory module
  - [ ] Inventory repository (MongoDB: `core_db.inventory`)
  - [ ] Stock management operations
  - [ ] Inventory tracking and adjustments
  - [ ] Event publishing: `inventory.updated`, `inventory.low_stock`
- [ ] Shared infrastructure
  - [ ] Business rules and validation framework
  - [ ] Transaction management
  - [ ] Kafka event publisher integration
  - [ ] Multi-tenancy filtering (tenant_id in all queries)

**Completed Modules:**
- ✅ Users module (Priority 3)

**Port:** 5001 (shared with User Service)

---

### Phase 3: Quality Assurance 🧪

#### 8. Functional Testing Infrastructure (Priority 6 - MOVED) 🐍
**Status:** ✅ Complete

**Why Sixth:** Auth, Config, and User services are ready to test. Building test infrastructure now prevents technical debt and ensures quality before adding more services.

**Prerequisites:**
- ✅ Auth Service (to test authentication flows)
- ✅ Config Service (to test configuration management)
- ✅ Core User Service (to test user CRUD, needed for test setup/cleanup)

**Dependencies:**
- Python 3.10+ (for test framework)
- gRPC Python libraries (grpcio, grpcio-tools)
- Proto files (must generate Python code from .proto files)
- pytest (test runner)
- Docker (for running MongoDB and Redis during tests)

**What was Built:**
- [x] Python test framework structure
  - [x] Common test utilities module (`internal/infra/functional/`)
  - [x] Generic gRPC client utilities (`grpc_client.py`)
  - [x] Test fixtures and helpers (`conftest.py`, `helpers/db_injection.py`)
  - [x] Configuration management (`config.py`)
- [x] Proto generation for Python
  - [x] `make proto-python` target in root Makefile
  - [x] Python proto output directory (`internal/infra/functional/proto/`)
  - [x] Generated stubs: auth/v1, core/v1, config/v1, infra/v1, event/v1, gateway/v1
- [x] Test organization structure
  - [x] `internal/auth/functional/` — 12 test files, 108 tests
  - [ ] `internal/config/functional/` — pending (Config Service not yet built)
  - [ ] `internal/core/functional/` — pending (Core Service not yet built)
- [x] Test lifecycle structure (Pre-test → Test → Post-test)
  - [x] **Pre-test**: SystemSeeder seeds system tenant + admin user + role; `clean_database` fixture flushes between tests
  - [x] **Actual test**: Black-box gRPC calls via generated stubs
  - [x] **Post-test**: `clean_database` fixture handles cleanup automatically
- [x] Database clients
  - [x] `db/mongo_client.py` — MongoDB operations for test data injection and verification
  - [x] `db/redis_client.py` — Redis operations for token state verification
  - [x] `db/manager.py` — Database lifecycle management (setup, teardown, clean)
- [x] Seeders
  - [x] `seeders/system_seeder.py` — Seeds system tenant, TenantAdmin role, admin user
- [x] Test documentation
  - [x] `internal/infra/functional/README.md` — comprehensive guide

**Test Structure Example:**
```python
def test_user_login_flow():
    # Pre-test: Setup
    tenant = create_test_tenant()
    user = create_test_user(tenant_id=tenant.id)

    # Actual test
    response = auth_client.authenticate(
        email=user.email,
        password="test_password",
        tenant_id=tenant.id
    )
    assert response.access_token is not None
    assert response.user_id == user.id

    # Post-test: Cleanup
    delete_test_user(user.id)
    delete_test_tenant(tenant.id)
```

**Deliverables:**
- Python test framework with reusable utilities
- Proto generation pipeline for Python
- Functional test structure in each module
- Documentation and examples

---

#### 9. Functional Tests - Auth, Config, User (Priority 7 - MOVED) ✅
**Status:** ✅ Complete (3 pagination tests skipped)

**Why Seventh:** Tests the three completed services (Auth, Config, User) using the newly built testing infrastructure. Validates end-to-end flows work correctly.

**Prerequisites:**
- ✅ Functional Testing Infrastructure (Priority 4)
- ✅ Auth Service (service to test)
- ✅ Config Service (service to test)
- ✅ Core User Service (service to test)

**What was Built:**
- [x] Auth Service functional tests (`internal/auth/functional/`) — 108 tests across 12 files
  - [x] `test_auth.py` (6) — Login, logout, refresh, verify, revoke, revoke all tenant tokens
  - [x] `test_negative_auth.py` (13) — Invalid credentials, expired tokens, missing fields, inactive/suspended users
  - [x] `test_user.py` (5) — CreateUser, GetUser, ListUsers, UpdateUser, DeleteUser
  - [x] `test_negative_user.py` (20, 1 skipped) — Duplicate email/username, cross-tenant access, invalid roles, weak password, invalid email format; pagination skipped
  - [x] `test_role.py` (5) — CreateRole, GetRole, ListRoles, UpdateRole, DeleteRole
  - [x] `test_negative_role.py` (14, 1 skipped) — Duplicate name, cross-tenant, system role protection, invalid permissions; pagination skipped
  - [x] `test_permission.py` (3) — GetPermission, admin wildcard, ListPermissions
  - [x] `test_negative_permission.py` (5) — Nonexistent, invalid format, cross-tenant, protected role block
  - [x] `test_tenant.py` (6) — CreateTenant (with default seeding), GetTenant (by ID + name), ListTenants, UpdateTenant, DeleteTenant (with cascade)
  - [x] `test_negative_tenant.py` (12, 1 skipped) — Duplicate name, invalid status, nonexistent, missing fields; pagination skipped
  - [x] `test_rbac.py` (5) — CheckPermissions, HasPermission, GetUserPermissions, GetUserRoles, IsSystemTenantUser
  - [x] `test_negative_rbac.py` (14) — Nonexistent/inactive/suspended/deleted users, cross-tenant checks, invalid permission format
- [ ] Config Service functional tests (`internal/config/functional/`) — pending (Config Service not yet built)
- [ ] Core Service functional tests (`internal/core/functional/`) — pending (Core Service not yet built)

**Test Coverage Goals:**
- Positive test cases (happy path)
- Negative test cases (errors, invalid inputs)
- Edge cases (boundary conditions)
- Multi-tenancy isolation verification
- RBAC enforcement verification

**Deliverables:**
- Comprehensive functional test suite for Auth, Config, User services
- Test reports and coverage metrics
- CI/CD integration (optional)

---

### Phase 5: Security Enhancement 🔒

#### 10. Config Service - mTLS & Two-Tier RBAC (Priority 8 - MOVED) 🔐
**Status:** ⬜ Not Started

**Why Eighth:** Enhances Config service security with production-grade mTLS and certificate-based module authentication. Now that basic functionality is proven via functional tests, we can add enterprise-level security.

**Prerequisites:**
- ✅ Config Service Phase 1 (simple version working)
- ✅ Functional tests passing (Config service validated)

**Dependencies:**
- PKI infrastructure (certificate authority, cert generation tools)
- TLS/mTLS libraries (Go crypto/tls)
- Certificate storage (filesystem or secrets manager)

**What to Build:**
- [ ] PKI Infrastructure
  - [ ] Certificate Authority (CA) setup
  - [ ] Certificate generation scripts/tools
  - [ ] Certificate storage strategy
  - [ ] Certificate rotation mechanism
  - [ ] Expiration monitoring and alerts
- [ ] mTLS Implementation for Config Service
  - [ ] Server-side mTLS configuration
  - [ ] Client certificate validation
  - [ ] TLS handshake and mutual authentication
  - [ ] Graceful fallback for local development (disable mTLS in dev mode)
- [ ] Two-Tier RBAC Implementation
  - [ ] **Tier 1: Module-Level Authentication (Certificate Validation)**
    - [ ] Extract module identity from client certificate (Subject CN or SAN)
    - [ ] Validate module can access requested configuration scope
    - [ ] Example: Core module can only access `core.*` configs, not `auth.*`
    - [ ] Enforcement: Reject requests from unauthorized modules
  - [ ] **Tier 2: User-Level Authorization (Auth Service Integration)**
    - [ ] Extract user identity from request context (already implemented in Phase 1)
    - [ ] Call Auth.CheckPermissions to verify user permissions
    - [ ] Example: User must have `config:write` permission to update configs
    - [ ] Enforcement: Reject requests from users without proper permissions
  - [ ] Combined enforcement: Both tiers must pass for request to succeed
- [ ] Module Certificate Distribution
  - [ ] Auth service certificate
  - [ ] Core service certificate
  - [ ] Config service certificate (for self-access)
  - [ ] Gateway service certificate (future)
- [ ] Update Config proto
  - [ ] Add metadata fields for certificate-based auth
  - [ ] Version config API if needed
- [ ] Documentation
  - [ ] mTLS setup guide
  - [ ] Certificate generation and distribution procedures
  - [ ] Two-tier RBAC architecture diagram
  - [ ] Troubleshooting guide

**Security Model:**
```
Request Flow with Two-Tier RBAC:
1. Client connects with mTLS certificate
2. Config service validates certificate (Tier 1: Module authentication)
   - Extract module identity from cert
   - Check if module can access config scope (e.g., core.* configs)
3. Config service extracts user from request metadata
4. Config service calls Auth.CheckPermissions (Tier 2: User authorization)
   - Verify user has permission (e.g., config:read, config:write)
5. Both checks pass → Process request
6. Either check fails → Reject with appropriate error
```

**Benefits:**
- ✅ Production-grade security
- ✅ Module-level isolation (core can't read auth configs)
- ✅ User-level permissions (not all users can change configs)
- ✅ Mutual authentication via mTLS
- ✅ Certificate-based service identity

**Port:** 5002 (same as Config Service Phase 1)

---

### Phase 6: Integration Layer 🔗

#### 11. Gateway (Priority 10 - MOVED) 🌐
**Status:** ⬜ Not Started (Deferred after functional testing and security enhancements)

**Why Tenth:** Single entry point for WebUI. Depends on Auth and Core services being ready. Deferred until core services and testing are complete.

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

#### 12. Events Service (Priority 11 - MOVED) 📡
**Status:** ⬜ Not Started (Deferred)

**Why Eleventh:** Consumes events from Kafka and handles audit logging. Deferred until Core service modules are implemented and generating events.

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

### Phase 7: Frontend 🎨

#### 13. WebUI (Priority 12 - MOVED) 💻
**Status:** ⬜ Not Started (Deferred)

**Why Twelfth (Last):** Depends on Gateway being ready to provide GraphQL API. Final phase after all backend services are complete.

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
- ✅ Uses generic Repository pattern from `internal/infra/db/repository.go`

### Code Organization
- ✅ Starting as monorepo with multiple packages
- ✅ Will break down to microservices and shared Go modules later
- ✅ Infrastructure centralized in `internal/infra/` for easier management:
  - `internal/infra/models/` - Domain models by module (auth, core, config, gateway, events)
  - `internal/infra/grpc/` - Generic gRPC infrastructure (to be built)
  - `internal/infra/proto/` - Shared proto definitions
  - `internal/infra/db/` - Database interfaces (MongoDB, Redis)
  - `internal/infra/errors/` - Error handling utilities
  - `internal/infra/logging/` - Logging utilities
  - `internal/infra/events/` - Event publishing/consuming

### Infrastructure Notes
- ⚠️ MongoDB and Redis connection URIs are currently hardcoded in `internal/infra/db/mongo/mongo.go` and `internal/infra/db/redis/redis.go`
- ⚠️ Will be moved to environment configuration later (not blocking for initial development)

---

---

## Future Features (To Be Planned)

### Token Management Enhancements 🔐
**Status:** 📝 Planned (after Phase 4: Test Infrastructure Update)

Additional enhancements for the token management system:

#### Option B: Security Enhancements 🛡️
**Goal:** Add RBAC and audit improvements

**What to Build:**
- [ ] Implement RBAC check in `RevokeAllTenantTokens()` (currently has TODO)
- [ ] Add comprehensive audit logging for all token operations
  - [ ] Token creation, refresh, revocation events
  - [ ] User authentication attempts (success/failure)
  - [ ] Token validation failures
- [ ] Implement token rotation policies
  - [ ] Automatic token rotation after N days
  - [ ] Force rotation on security events
  - [ ] Grace period for old tokens during rotation
- [ ] Add rate limiting for token refresh
  - [ ] Prevent token refresh abuse
  - [ ] Per-user and per-tenant rate limits
  - [ ] Redis-based rate limiting counters

**Benefits:**
- Enhanced security posture
- Better compliance with audit requirements
- Prevention of token abuse
- Improved security monitoring

---

#### Option C: Performance Optimization ⚡
**Goal:** Optimize Redis operations

**What to Build:**
- [ ] Add Redis pipelining for bulk operations
  - [ ] Batch token validation requests
  - [ ] Optimize multi-token operations
  - [ ] Reduce network round trips
- [ ] Implement token caching strategies
  - [ ] In-memory cache for frequently accessed tokens
  - [ ] Cache invalidation on revocation
  - [ ] TTL-based cache expiration
- [ ] Add metrics/monitoring for token operations
  - [ ] Token creation rate
  - [ ] Token validation latency
  - [ ] Redis operation performance
  - [ ] Token cache hit/miss ratios
- [ ] Optimize SCAN batch sizes based on tenant size
  - [ ] Dynamic batch sizing
  - [ ] Tenant-specific optimization
  - [ ] Performance profiling and tuning

**Benefits:**
- Reduced latency for token operations
- Better scalability for large tenants
- Improved Redis resource utilization
- Better observability of token system

---

#### Option D: Feature Completion 🎯
**Goal:** Additional token management features

**What to Build:**
- [ ] Implement token blacklisting
  - [ ] Immediate token invalidation
  - [ ] Blacklist persistence
  - [ ] Automatic blacklist cleanup
- [ ] Add token metadata querying
  - [ ] List active tokens for user
  - [ ] Token creation timestamp and location
  - [ ] Last used information
  - [ ] Device/client information
- [ ] Implement session management UI endpoints
  - [ ] View active sessions
  - [ ] Revoke specific sessions
  - [ ] Session activity history
- [ ] Add token analytics/reporting
  - [ ] Token usage statistics
  - [ ] Authentication patterns
  - [ ] Security event reporting
  - [ ] Tenant-level analytics

**Benefits:**
- Better user experience (session management)
- Enhanced security visibility
- Improved administrative capabilities
- Data-driven security insights

---

### Advanced gRPC Infrastructure Features 🔧
**Status:** 📝 Planned (deferred from Priority 1.5)

Advanced features for production-grade gRPC infrastructure:
- Advanced connection pooling strategies
- Retry logic with exponential backoff
- Circuit breaker pattern
- Load balancing
- Service discovery integration
- Distributed tracing (OpenTelemetry)

*To be implemented after basic infrastructure is proven.*

---

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

