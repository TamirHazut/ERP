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

## Code Quality Initiative: Model Reorganization 📦

**Status:** 🟡 In Progress (Phase 1: 90% ✅, Phases 2-5: ⬜)

**Why Important:** Monolithic `models.go` files (500+ lines) are hard to navigate, test, and maintain. Breaking them down improves code organization and developer productivity.

**What Was Done:**

### ✅ Phase 1: Auth Models (90% Complete)

**Domain Models Breakdown:**
- [x] Split `internal/auth/models/models.go` (534 lines) into 9 focused files:
  - [x] `constants.go` - All status constants, role types, permission formats
  - [x] `tenant.go` - Tenant + 7 related structs + `Validate()` method
  - [x] `user.go` - User + 5 related structs + `Validate()` method
  - [x] `role.go` - Role + RoleMetadata + `Validate()` method
  - [x] `permission.go` - Permission + PermissionMetadata + `Validate()` method
  - [x] `user_group.go` - UserGroup + GroupMember
  - [x] `audit.go` - AuditLog + 3 related structs
  - [x] `token_claims.go` - AccessTokenClaims + RefreshTokenClaims + validation methods
  - [x] `refresh_token.go` - RefreshToken + validation methods

**Unit Tests:**
- [x] Created 6 comprehensive test files with table-driven tests:
  - [x] `tenant_test.go` - Tests for Tenant.Validate()
  - [x] `user_test.go` - Tests for User.Validate()
  - [x] `role_test.go` - Tests for Role.Validate()
  - [x] `permission_test.go` - Tests for Permission.Validate()
  - [x] `token_claims_test.go` - Tests for Claims validation and IsExpired()
  - [x] `refresh_token_test.go` - Tests for RefreshToken validation and helper methods

**Cache Models (Moved from Redis):**
- [x] Created `internal/auth/models/cache/` subdirectory
- [x] Moved 14 auth-related cache models from `internal/db/redis/models/`:
  - [x] `session.go` - Session + DeviceInfo
  - [x] `token.go` - TokenMetadata + RevokedToken
  - [x] `rbac.go` - UserPermissionsCache, UserRolesCache, RoleSummary, RolePermissionsCache
  - [x] `password.go` - PasswordResetToken
  - [x] `verification.go` - EmailVerificationToken
  - [x] `mfa.go` - MFACode
  - [x] `invitation.go` - InviteToken
  - [x] `security.go` - LoginAttempts
  - [x] `presence.go` - ActiveUser

**Remaining Tasks (Phase 1 - 10%):**
- [x] Update imports in auth service files:
  - [x] `internal/auth/keys_handlers/access_token.go`
  - [x] `internal/auth/keys_handlers/refresh_token.go`
  - [x] `internal/auth/token/token_manager.go`
  - [x] `internal/auth/service/auth.go`
- [x] Delete old `internal/auth/models/models.go` (after import verification)
- [x] Delete old `internal/auth/models/models_test.go`
- [x] Delete moved cache models from `internal/db/redis/models/models.go`
- [x] Run tests to verify everything works

### ✅ Phase 2: Gateway Cache Models (Completed)
- [x] Create `internal/gateway/models/cache/` directory
- [x] Move 4 gateway-related cache models from Redis:
  - [x] `rate_limit.go` - RateLimitInfo, TenantRateLimit, IPRateLimit
  - [x] `query_cache.go` - QueryCache

### ✅ Phase 3: Config Models (Completed)
- [x] Break down `internal/config/models/models.go` into:
  - [x] `service_config.go` - 5 structs
  - [x] `feature_flag.go` - 3 structs
- [x] Create `internal/config/models/cache/` directory
- [x] Move 3 config-related cache models from Redis:
  - [x] `feature_flags.go` - FeatureFlagCache, TenantFeatures
  - [x] `service_config.go` - ServiceConfigCache

### ✅ Phase 4: Core Models (Completed)
- [x] Break down `internal/core/models/models.go` into:
  - [x] `constants.go` - All status/type constants
  - [x] `product.go` - 5 structs
  - [x] `vendor.go` - 4 structs
  - [x] `order.go` - 6 structs
  - [x] `customer.go` - 4 structs
  - [x] `inventory.go` - 2 structs
  - [x] `warehouse.go` - 3 structs
  - [x] `category.go` - 1 struct

### ✅ Phase 5: Redis Infrastructure Cleanup (Completed)
- [x] Create `internal/db/redis/types.go` - Generic infrastructure types (RedisKeyOptions, CacheEntry, DistributedLock)
- [x] Create `internal/db/redis/cross_service_cache.go` - Cross-service caches (UserCache, TenantCache, ProductCache, OrderCache)
- [x] Delete `internal/db/redis/models/models.go` (after all moves complete)

**Documentation:**
- [x] `MODEL_BREAKDOWN_PLAN.md` - Complete reorganization plan
- [x] `MODEL_REORGANIZATION.md` - Cache model relocation strategy
- [x] `DUPLICATES_ANALYSIS.md` - Duplicate code analysis
- [x] `IMPLEMENTATION_STATUS.md` - Current progress tracking
- [x] Updated `CLAUDE.md` - Model organization guidelines

**Benefits Achieved:**
- ✅ Better code organization (27 focused files vs 1 monolithic file)
- ✅ Easier navigation (find User model in `user.go` instead of searching 534-line file)
- ✅ Improved testing (colocated test files, comprehensive coverage)
- ✅ Reduced merge conflicts (different developers work on different entity files)
- ✅ Clear ownership (each service owns its models and caches)

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

### ✅ Completed: Redis Handler Tests (internal/db/redis/handlers/)

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
**Status:** ✅ Complete (100%)

**Why First:** Required by all other services for authentication/authorization. Foundation for the entire system.

**Note:** User Service was moved to Core module as it's part of business logic, not authentication.

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
- [x] Audit logs collection (`internal/auth/collections/audit_logs.go`)
  - [x] CRUD operations with tenant isolation
  - [x] Enhanced audit models with detailed change tracking
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
- `POST /auth/login` → gRPC `Authenticate()` ✅
- `POST /auth/logout` → gRPC `Logout()` ✅
- `POST /auth/refresh` → gRPC `RefreshToken()` ✅
- `GET /auth/verify` → gRPC `VerifyToken()` ✅
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
- ✅ All unit tests passing and stable (100+ tests across 10 packages)
- ✅ Collection tests (permissions, roles, tenants, users, audit_logs) - Refactored with custom matchers
- ✅ Model validation tests (permission, role, tenant, user, token_claims, refresh_token)
- ✅ Key handler tests (access_token, refresh_token, token_index) - Complete coverage
- ✅ Token manager tests
- ✅ RBAC manager tests (comprehensive coverage of all operations - 19 test cases)
- ✅ Redis handler tests (set_handler) - Complete coverage
- ✅ Utils tests (password hashing)
- ✅ Test refactoring complete - NO gomock.Any() usage anywhere
- ✅ All tests use specific expected values and custom matchers where needed

**Completed Tasks:**
- [x] Create `internal/auth/cmd/main.go` entry point to start the server
- [x] Complete RBAC manager implementation with comprehensive tests
- [x] All core endpoints (Login, Logout, Refresh, Verify, Revoke, CheckPermissions)
- [x] gRPC server implementation with all methods
- [x] Token infrastructure (AccessToken, RefreshToken, TokenIndex, TokenManager)
- [x] Repository layers for Users, Roles, Permissions, Tenants, AuditLogs

**Notes:**
- User management moved to Core Service (part of business logic, not authentication)
- End-to-end functional testing infrastructure will be built in Phase 3
- Audit logging in Logout can be re-enabled during functional testing phase
- mTLS support deferred to Config Service enhancement phase

**Port:** 5000

---

#### 2. Config Service - Phase 1 (Priority 2) ⚙️
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

#### 3. Core User Service (Priority 3) 👥
**Status:** ⬜ Not Started

**Why Third:** User management is core business logic (not authentication). Required for functional testing infrastructure (test setup/cleanup). Minimal implementation to unblock testing.

**Prerequisites:**
- ✅ Auth Service (for RBAC permission checks)
- ✅ Config Service Phase 1 (for feature flags - optional)

**Dependencies:**
- Auth Service (for CheckPermissions gRPC calls)
- Config Service (for feature flags - optional)
- MongoDB (`core_db.users` collection)
- gRPC infrastructure (from Pre-Phase)

**What to Build:**
- [ ] gRPC server implementation (can share with future Core service modules)
- [ ] User service proto definitions
  - [ ] CreateUser RPC method
  - [ ] GetUser RPC method
  - [ ] UpdateUser RPC method
  - [ ] DeleteUser RPC method
  - [ ] ListUsers RPC method (with pagination, filtering by tenant)
- [ ] User repository (MongoDB: `core_db.users`)
  - [ ] CRUD operations with tenant isolation
  - [ ] User profile management
  - [ ] Metadata and preferences storage
- [ ] RBAC integration with Auth service
  - [ ] Call Auth.CheckPermissions before user operations
  - [ ] Verify user has permission to create/read/update/delete users
- [ ] Config integration (optional for Phase 1)
  - [ ] Feature flag support (if Config service ready)
  - [ ] Can defer to later if Config not ready
- [ ] System admin data seeding
  - [ ] Use CollectionHandlers directly with hard-coded data
  - [ ] Create default tenant
  - [ ] Create default roles (SuperAdmin, Admin, User)
  - [ ] Create default permissions
  - [ ] Create system admin user
  - [ ] Seeding script or initialization function

**Scope Notes:**
- This is ONLY user management, not full Core service
- Other Core modules (Products, Orders, Vendors, Inventory) come later
- Minimal implementation to enable functional testing

**Port:** 5001 (shared with future Core service modules)

---

#### 4. Core Service - Remaining Modules (Priority 7) 🏢
**Status:** ⬜ Not Started (Deferred after functional testing)

**Why Deferred:** User Service (Priority 3) provides enough functionality for initial testing. Other Core modules (Products, Orders, Vendors, Inventory) can wait until after functional testing infrastructure is proven.

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

#### 5. Functional Testing Infrastructure (Priority 4) 🐍
**Status:** ⬜ Not Started

**Why Fourth:** Auth, Config, and User services are ready to test. Building test infrastructure now prevents technical debt and ensures quality before adding more services.

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

**What to Build:**
- [ ] Python test framework structure
  - [ ] Common test utilities module (`tests/common/`)
  - [ ] Generic gRPC client utilities
  - [ ] Test fixtures and helpers
  - [ ] Configuration management for test environments
- [ ] Proto generation for Python
  - [ ] Modify `make proto` to generate Go + Python proto files
  - [ ] Add `make proto-python` target
  - [ ] Python proto output directory (`tests/proto/`)
  - [ ] Update `.gitignore` for Python generated files
- [ ] Test organization structure
  - [ ] Create `functional/` folder in each module:
    - [ ] `internal/auth/functional/`
    - [ ] `internal/config/functional/`
    - [ ] `internal/core/functional/`
  - [ ] Each module has its own test suite
- [ ] Test lifecycle structure (Pre-test → Test → Post-test)
  - [ ] **Pre-test**: Setup test data (create users, tenants, roles via gRPC)
  - [ ] **Actual test**: Execute the flow being tested (black-box approach)
  - [ ] **Post-test**: Cleanup test data (delete created resources)
  - [ ] Fixtures for common setup/teardown patterns
- [ ] Generic gRPC client implementation
  - [ ] Connection management (connect to services)
  - [ ] Credential handling (JWT tokens for authenticated calls)
  - [ ] Request/response helpers
  - [ ] Error handling utilities
- [ ] Test documentation
  - [ ] How to run functional tests
  - [ ] How to add new test cases
  - [ ] Test environment setup guide

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

#### 6. Functional Tests - Auth, Config, User (Priority 5) ✅
**Status:** ⬜ Not Started

**Why Fifth:** Tests the three completed services (Auth, Config, User) using the newly built testing infrastructure. Validates end-to-end flows work correctly.

**Prerequisites:**
- ✅ Functional Testing Infrastructure (Priority 4)
- ✅ Auth Service (service to test)
- ✅ Config Service (service to test)
- ✅ Core User Service (service to test)

**What to Build:**
- [ ] Auth Service functional tests (`internal/auth/functional/`)
  - [ ] `test_authentication.py` - Login, logout, token validation
  - [ ] `test_token_refresh.py` - Token refresh and rotation
  - [ ] `test_token_revocation.py` - Single token revoke, revoke all
  - [ ] `test_permissions.py` - CheckPermissions endpoint
  - [ ] `test_rbac.py` - Role-based access control flows
  - [ ] `test_multi_tenant.py` - Multi-tenant isolation (user in tenant A cannot access tenant B resources)
- [ ] Config Service functional tests (`internal/config/functional/`)
  - [ ] `test_get_config.py` - Get configuration values
  - [ ] `test_set_config.py` - Update configuration (with permissions)
  - [ ] `test_config_permissions.py` - RBAC for configuration access
  - [ ] `test_config_validation.py` - Invalid configuration rejection
- [ ] User Service functional tests (`internal/core/functional/`)
  - [ ] `test_user_crud.py` - Create, read, update, delete users
  - [ ] `test_user_permissions.py` - RBAC for user operations
  - [ ] `test_user_multi_tenant.py` - User isolation per tenant
  - [ ] `test_user_pagination.py` - List users with pagination
- [ ] Integration/End-to-end tests
  - [ ] `test_full_user_lifecycle.py` - Login → Create User → Logout → Cleanup
  - [ ] `test_permission_enforcement.py` - User without permission cannot perform action

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

#### 7. Config Service - mTLS & Two-Tier RBAC (Priority 6) 🔐
**Status:** ⬜ Not Started

**Why Sixth:** Enhances Config service security with production-grade mTLS and certificate-based module authentication. Now that basic functionality is proven via functional tests, we can add enterprise-level security.

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

#### 8. Gateway (Priority 8) 🌐
**Status:** ⬜ Not Started (Deferred after functional testing and security enhancements)

**Why Eighth:** Single entry point for WebUI. Depends on Auth and Core services being ready. Deferred until core services and testing are complete.

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

#### 9. Events Service (Priority 9) 📡
**Status:** ⬜ Not Started (Deferred)

**Why Ninth:** Consumes events from Kafka. Deferred until Core service modules are implemented and generating events.

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

#### 10. WebUI (Priority 10) 💻
**Status:** ⬜ Not Started (Deferred)

**Why Last:** Depends on Gateway being ready to provide GraphQL API. Final phase after all backend services are complete.

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

