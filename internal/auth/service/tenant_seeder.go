package service

import (
	"context"
	"fmt"
	"time"

	core_collection "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/auth/password"
	mongo_collection "erp.localhost/internal/infra/db/mongo/collection"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
	proto_infra "erp.localhost/internal/infra/proto/infra/v1"
)

type TenantDefaults struct {
	PermissionID string // ID of "*:*" permission
	RoleID       string // ID of TenantAdmin role
	UserID       string // ID of initial admin user
}

type TenantSeeder struct {
	rbacGRPCClient proto_auth.RBACServiceClient
	userCollection *core_collection.UserCollection
	logger         logger.Logger
}

func NewTenantSeeder(rbacGRPCClient proto_auth.RBACServiceClient, logger logger.Logger) *TenantSeeder {
	uc := mongo_collection.NewBaseCollectionHandler[model_auth.User](model_mongo.CoreDB, model_mongo.UsersCollection, logger)
	if uc == nil {
		logger.Fatal("failed to create users collection handler")
		return nil
	}
	return &TenantSeeder{
		rbacGRPCClient: rbacGRPCClient,
		userCollection: core_collection.NewUserCollection(uc, logger),
		logger:         logger,
	}
}

// SeedDefaults creates default permission, role, and admin user for a new tenant
func (ts *TenantSeeder) SeedDefaults(ctx context.Context, tenantID, adminEmail, adminPassword, createdBy string) (*TenantDefaults, error) {
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
	defaults.RoleID = roleID
	ts.logger.Info("TenantAdmin role created", "tenant_id", tenantID, "role_id", roleID)

	// Step 3: Create initial admin user in Core
	userID, err := ts.createAdminUser(tenantID, adminEmail, adminPassword, roleID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}
	defaults.UserID = userID
	ts.logger.Info("Admin user created", "tenant_id", tenantID, "user_id", userID, "email", adminEmail)

	ts.logger.Info("Tenant defaults seeded successfully", "tenant_id", tenantID)
	return defaults, nil
}

func (ts *TenantSeeder) createWildcardPermission(ctx context.Context, tenantID, createdBy string) (string, error) {
	req := &proto_auth.CreateResourceRequest{
		Identifier: &proto_infra.UserIdentifier{
			TenantId: tenantID,
			UserId:   createdBy,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
		Resource: &proto_auth.CreateResourceRequest_Permission{
			Permission: &proto_auth.CreatePermissionData{
				TenantId:    tenantID,
				Name:        "Full Access",
				Slug:        "full_access",
				Description: "Grants full access to all resources and actions",
				Resource:    model_auth.ResourceTypeAll,     // "*"
				Action:      model_auth.PermissionActionAll, // "*"
				Status:      model_auth.PermissionStatusActive,
				CreatedBy:   createdBy,
			},
		},
	}

	resp, err := ts.rbacGRPCClient.CreateResource(ctx, req)
	if err != nil {
		return "", fmt.Errorf("gRPC call to create permission failed: %w", err)
	}

	return resp.ResourceId, nil
}

func (ts *TenantSeeder) createTenantAdminRole(ctx context.Context, tenantID, permissionID, createdBy string) (string, error) {
	req := &proto_auth.CreateResourceRequest{
		Identifier: &proto_infra.UserIdentifier{
			TenantId: tenantID,
			UserId:   createdBy,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_ROLE,
		Resource: &proto_auth.CreateResourceRequest_Role{
			Role: &proto_auth.CreateRoleData{
				TenantId:     tenantID,
				Name:         "Tenant Administrator",
				Slug:         model_auth.RoleTenantAdmin, // "tenant_admin"
				Description:  "Tenant administrator with full access to all tenant resources",
				Type:         model_auth.RoleTenantAdmin,
				IsSystemRole: false,
				Permissions:  []string{permissionID}, // Assign "*:*" permission
				Status:       model_auth.RoleStatusActive,
				CreatedBy:    createdBy,
			},
		},
	}

	resp, err := ts.rbacGRPCClient.CreateResource(ctx, req)
	if err != nil {
		return "", fmt.Errorf("gRPC call to create role failed: %w", err)
	}

	return resp.ResourceId, nil
}

func (ts *TenantSeeder) createAdminUser(tenantID, email, plainPassword, roleID, createdBy string) (string, error) {
	// Hash password
	hashedPassword, err := password.HashPassword(plainPassword)
	if err != nil {
		return "", fmt.Errorf("failed to hash admin password: %w", err)
	}

	user := &model_auth.User{
		TenantID:     tenantID,
		Email:        email,
		Username:     email, // Use email as username
		PasswordHash: hashedPassword,
		Status:       model_auth.UserStatusActive,
		CreatedBy:    createdBy,
		Roles: []model_auth.UserRole{
			{
				TenantID:   tenantID,
				RoleID:     roleID,
				AssignedAt: time.Now(),
				AssignedBy: createdBy,
			},
		},
	}

	// Validate user
	if err := user.Validate(true); err != nil {
		return "", fmt.Errorf("user validation failed: %w", err)
	}

	// Create user via collection
	userID, err := ts.userCollection.CreateUser(user)
	if err != nil {
		return "", fmt.Errorf("failed to create admin user: %w", err)
	}

	return userID, nil
}

// RollbackDefaults deletes all seeded defaults (used when tenant creation fails)
func (ts *TenantSeeder) RollbackDefaults(ctx context.Context, tenantID string, defaults *TenantDefaults) error {
	ts.logger.Warn("Rolling back tenant defaults", "tenant_id", tenantID)

	var rollbackErrors []error

	// Delete admin user (local collection)
	if defaults.UserID != "" {
		if err := ts.userCollection.DeleteUser(tenantID, defaults.UserID); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to delete admin user: %w", err))
		}
	}

	// Delete role (via Auth gRPC)
	if defaults.RoleID != "" {
		req := &proto_auth.DeleteResourceRequest{
			Identifier: &proto_infra.UserIdentifier{
				TenantId: tenantID,
			},
			ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_ROLE,
			Resource: &proto_auth.DeleteResourceRequest_ResourceId{
				ResourceId: defaults.RoleID,
			},
		}
		if _, err := ts.rbacGRPCClient.DeleteResource(ctx, req); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to delete role via gRPC: %w", err))
		}
	}

	// Delete permission (via Auth gRPC)
	if defaults.PermissionID != "" {
		req := &proto_auth.DeleteResourceRequest{
			Identifier: &proto_infra.UserIdentifier{
				TenantId: tenantID,
			},
			ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
			Resource: &proto_auth.DeleteResourceRequest_ResourceId{
				ResourceId: defaults.PermissionID,
			},
		}
		if _, err := ts.rbacGRPCClient.DeleteResource(ctx, req); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to delete permission via gRPC: %w", err))
		}
	}

	if len(rollbackErrors) > 0 {
		return fmt.Errorf("rollback partially failed: %v", rollbackErrors)
	}

	ts.logger.Info("Tenant defaults rolled back successfully", "tenant_id", tenantID)
	return nil
}
