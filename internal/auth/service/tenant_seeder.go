package service

import (
	"context"
	"fmt"

	"erp.localhost/internal/auth/api"
	core_collection "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/auth/hash"
	"erp.localhost/internal/infra/db"
	mongo_collection "erp.localhost/internal/infra/db/mongo/collection"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_auth "erp.localhost/internal/infra/model/auth/validator"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TenantDefaults struct {
	PermissionID string // ID of "*:*" permission
	RoleId       string // ID of TenantAdmin role
	UserId       string // ID of initial admin user
}

type TenantSeeder struct {
	userCollection *core_collection.UserCollection
	rbacAPI        *api.RBACAPI
	logger         logger.Logger
}

func NewTenantSeeder(rbacAPI *api.RBACAPI, logger logger.Logger) *TenantSeeder {
	uc := mongo_collection.NewBaseCollectionHandler[authv1.User](model_mongo.AuthDB, model_mongo.UsersCollection, logger)
	if uc == nil {
		logger.Fatal("failed to create users collection handler")
		return nil
	}
	return &TenantSeeder{
		userCollection: core_collection.NewUserCollection(uc, logger),
		rbacAPI:        rbacAPI,
		logger:         logger,
	}
}

// SeedDefaults creates default permission, role, and admin user for a new tenant
func (ts *TenantSeeder) SeedDefaults(ctx context.Context, tenantID, adminEmail, createdBy string) (*TenantDefaults, error) {
	ts.logger.Info("Seeding defaults for new tenant", "tenant_id", tenantID)

	defaults := &TenantDefaults{}

	// Step 1: Create "*:*" permission via Auth gRPC
	permissionID, err := ts.createWildcardPermission(ctx, tenantID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create wildcard permission: %w", err)
	}
	defaults.PermissionID = permissionID
	ts.logger.Info("Wildcard permission created", "tenant_id", tenantID, "permission_id", permissionID)

	// Step 2: Create TenantAdmin role via Auth gRPC
	roleID, err := ts.createTenantAdminRole(ctx, tenantID, permissionID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create TenantAdmin role: %w", err)
	}
	defaults.RoleId = roleID
	ts.logger.Info("TenantAdmin role created", "tenant_id", tenantID, "role_id", roleID)

	// Step 3: Create initial admin user in Core
	userID, err := ts.createAdminUser(tenantID, db.TenantAdminUser, db.TenantAdminPassword, roleID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}
	defaults.UserId = userID
	ts.logger.Info("Admin user created", "tenant_id", tenantID, "user_id", userID, "email", adminEmail)

	ts.logger.Info("Tenant defaults seeded successfully", "tenant_id", tenantID)
	return defaults, nil
}

func (ts *TenantSeeder) createWildcardPermission(ctx context.Context, tenantID, createdBy string) (string, error) {
	permission := &authv1.Permission{
		TenantId:         tenantID,
		DisplayName:      "Full Access",
		PermissionString: "full_access",
		Description:      "Grants full access to all resources and actions",
		Resource:         model_auth.ResourceTypeAll,     // "*"
		Action:           model_auth.PermissionActionAll, // "*"
		Status:           authv1.PermissionStatus_PERMISSION_STATUS_ACTIVE,
		CreatedBy:        createdBy,
	}

	return ts.rbacAPI.Permissions.CreatePermission(tenantID, createdBy, permission, tenantID)
}

func (ts *TenantSeeder) createTenantAdminRole(ctx context.Context, tenantID, permissionID, createdBy string) (string, error) {
	role := &authv1.Role{
		TenantId:    tenantID,
		Name:        model_auth.RoleTenantAdmin,
		Description: "Tenant administrator with full access to all tenant resources",
		Type:        authv1.RoleType_ROLE_TYPE_SYSTEM,
		Permissions: []string{permissionID}, // Assign "*:*" permission
		Status:      authv1.RoleStatus_ROLE_STATUS_ACTIVE,
		CreatedBy:   createdBy,
	}

	return ts.rbacAPI.Roles.CreateRole(tenantID, createdBy, role, tenantID)
}

func (ts *TenantSeeder) createAdminUser(tenantID, username, plainPassword, roleID, createdBy string) (string, error) {
	// Hash password
	hashedPassword, err := hash.HashPassword(plainPassword)
	if err != nil {
		return "", err
	}

	user := &authv1.User{
		TenantId:     tenantID,
		Username:     username,
		PasswordHash: hashedPassword,
		Status:       authv1.UserStatus_USER_STATUS_ACTIVE,
		CreatedBy:    createdBy,
		Roles: []*authv1.UserRole{
			{
				TenantId:   tenantID,
				RoleId:     roleID,
				AssignedAt: timestamppb.Now(),
				AssignedBy: createdBy,
			},
		},
	}

	// Validate user
	if err := validator_auth.ValidateUser(user, true); err != nil {
		return "", err
	}

	// Create user via collection
	return ts.userCollection.CreateUser(user)
}

// RollbackDefaults deletes all seeded defaults (used when tenant creation fails)
func (ts *TenantSeeder) RollbackDefaults(ctx context.Context, tenantID string, defaults *TenantDefaults) error {
	ts.logger.Warn("Rolling back tenant defaults", "tenant_id", tenantID)

	var rollbackErrors []error

	// Delete admin user (local collection)
	if defaults.UserId != "" {
		if err := ts.userCollection.DeleteUser(tenantID, defaults.UserId); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to delete admin user: %w", err))
		}
	}

	// Delete role (via Auth gRPC)
	if defaults.RoleId != "" {
		if err := ts.rbacAPI.Roles.DeleteRole(tenantID, defaults.UserId, defaults.RoleId, tenantID); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to delete role via gRPC: %w", err))
		}
	}

	// Delete permission (via Auth gRPC)
	if defaults.PermissionID != "" {
		if err := ts.rbacAPI.Permissions.DeletePermission(tenantID, defaults.UserId, defaults.PermissionID, tenantID); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to delete permission via gRPC: %w", err))
		}
	}

	if len(rollbackErrors) > 0 {
		return fmt.Errorf("rollback partially failed: %v", rollbackErrors)
	}

	ts.logger.Info("Tenant defaults rolled back successfully", "tenant_id", tenantID)
	return nil
}
