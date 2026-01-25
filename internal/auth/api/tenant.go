package api

import (
	"context"
	"errors"
	"fmt"

	"erp.localhost/internal/auth/handler"
	"erp.localhost/internal/auth/hash"
	"erp.localhost/internal/infra/db"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_auth "erp.localhost/internal/infra/model/auth/validator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TenantDefaults struct {
	PermissionID string // ID of "*:*" permission
	RoleId       string // ID of TenantAdmin role
	UserId       string // ID of initial admin user
}

type TenantAPI struct {
	logger        logger.Logger
	tenantHandler *handler.TenantHandler
	authAPI       *AuthAPI
	rbacAPI       *RBACAPI
	userAPI       *UserAPI
}

func NewTenantAPI(authAPI *AuthAPI, rbacAPI *RBACAPI, userAPI *UserAPI, logger logger.Logger) (*TenantAPI, error) {
	tenantHandler, err := handler.NewTenantHandler(logger)
	if err != nil {
		logger.Error("failed to create new user handler", "error", err)
		return nil, err
	}
	return &TenantAPI{
		logger:        logger,
		tenantHandler: tenantHandler,
		authAPI:       authAPI,
		rbacAPI:       rbacAPI,
		userAPI:       userAPI,
	}, nil
}

func (t *TenantAPI) CreateTenant(tenantID, userID string, newTenant *authv1.Tenant) (string, error) {
	// Step 1: validate input
	if tenantID == "" || userID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		t.logger.Error("failed to create tenant", "error", err)
		return "", err
	}
	if err := validator_auth.ValidateTenant(newTenant, true); err != nil {
		t.logger.Error("failed to create tenant", "error", err)
		return "", err
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(tenantID, userID, model_auth.ResourceTypeTenant, model_auth.PermissionActionCreate); err != nil {
		return "", err
	}
	// Step 3: Check for duplication
	tenant, err := t.tenantHandler.GetTenantByName(newTenant.Name)
	if err != nil {
		t.logger.Error("failed to get temamt for verification", "tenant_id", tenantID, "error", err)
		return "", err
	}
	if tenant != nil {
		err := infra_error.Validation(infra_error.ConflictDuplicateEmail)
		t.logger.Error("failed to create new tenant", "tenantID", tenantID, "error", err.Error())
		return "", err
	}

	adminEmail := newTenant.GetContact().GetEmail()

	// Step 4: Create tenant in MongoDB
	newTenantID, err := t.tenantHandler.CreateTenant(newTenant)
	if err != nil {
		t.logger.Error("failed to create tenant", "error", err)
		return "", err
	}
	t.logger.Info("tenant created in database", "tenant_id", tenantID)

	// Step 5: Seed defaults (permission, role, admin user)
	defaults, err := t.seedDefaults(tenantID, adminEmail, userID)
	if err != nil {
		t.logger.Error("failed to seed tenant defaults", "tenant_id", tenantID, "error", err)

		// Rollback: Delete tenant
		if deleteErr := t.tenantHandler.DeleteTenant(tenantID); deleteErr != nil {
			t.logger.Error("failed to rollback tenant creation", "tenant_id", tenantID, "error", deleteErr)
		}

		return "", err
	}
	t.logger.Info("tenant defaults seeded", "tenant_id", tenantID, "permission_id", defaults.PermissionID, "role_id", defaults.RoleId, "user_id", defaults.UserId)

	return newTenantID, nil
}

func (t *TenantAPI) GetTenant(tenantID, userID, targetTenantID, targetTenantName string) (*authv1.Tenant, error) {

	if tenantID == "" || userID == "" || (targetTenantID == "" && targetTenantName == "") {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id, target_tenant_name"))
		t.logger.Error("failed to get tenant", "error", err)
		return nil, err
	}

	if err := t.checkPermission(tenantID, userID, model_auth.ResourceTypeTenant, model_auth.PermissionActionRead); err != nil {
		return nil, err
	}

	if targetTenantID != "" {
		t.logger.Debug("getting tenant by id", "tenant_id", targetTenantID)
		return t.tenantHandler.GetTenantByID(targetTenantID)
	} else {
		t.logger.Debug("getting tenant by name", "name", targetTenantName)
		return t.tenantHandler.GetTenantByName(targetTenantName)
	}
}

func (t *TenantAPI) ListTenants(tenantID, userID, status string) ([]*authv1.Tenant, error) {
	// Step 1: validate input
	if tenantID == "" || userID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		t.logger.Error("failed to get tenants", "error", err)
		return nil, err
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(tenantID, userID, model_auth.ResourceTypeTenant, model_auth.PermissionActionRead); err != nil {
		return nil, err
	}

	if status != "" {
		t.logger.Debug("getting tenants by status", "status", status)
		return t.tenantHandler.GetTenantsByStatus(status)
	} else {
		t.logger.Debug("getting all tenants")
		return t.tenantHandler.GetTenants()
	}

}

func (t *TenantAPI) UpdateTenant(tenantID, userID string, tenant *authv1.Tenant) error {
	// Step 1: validate input
	if tenantID == "" || userID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		t.logger.Error("failed to update tenant", "error", err)
		return err
	}

	if err := validator_auth.ValidateTenant(tenant, false); err != nil {
		t.logger.Error("failed to update tenant", "error", err)
		return err
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(tenantID, userID, model_auth.ResourceTypeTenant, model_auth.PermissionActionUpdate); err != nil {
		return err
	}

	t.logger.Info("updating tenant", "tenant_id", tenant, "requested_by", userID, "target_tenant_id", tenant.GetId())

	// Step 4: Get existing tenant
	existingTenant, err := t.tenantHandler.GetTenantByID(tenant.GetId())
	if err != nil || existingTenant == nil {
		t.logger.Error("failed to get existing tenant", "tenant_id", tenant.Id, "error", err)
		return err
	}

	//TODO: Do diff and validate
	return t.tenantHandler.UpdateTenant(tenant)
}

func (t *TenantAPI) DeleteTenant(tenantID, userID, targetTenantID string) error {
	// Step 1: validate input
	if tenantID == "" || userID == "" || targetTenantID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id"))
		t.logger.Error("failed to delete tenant", "error", err)
		return err
	}

	// Step 2: Verify tenant exists
	_, err := t.tenantHandler.GetTenantByID(targetTenantID)
	if err != nil {
		t.logger.Error("tenant not found", "target_tenant_id", targetTenantID, "error", err)
		return err
	}

	// Step 3: Revoke all tenant users tokens
	t.logger.Info("starting tenant deletion cascade", "tenant_id", tenantID, "requested_by", userID, "target_tenant_id", targetTenantID)
	if _, _, err := t.authAPI.RevokeAllTenantTokens(tenantID, userID, targetTenantID); err != nil {
		t.logger.Error("failed to revoke tokens for tenant", "tenant_id", tenantID, "error", err)
		// Continue with deletion even if this fails
	} else {
		t.logger.Info("revoked all tokens for tenant", "target_tenant_id", targetTenantID)
	}

	// STEP 4: Delete ALL users for this tenant (bulk operation)
	// This deletes all user documents with matching tenant_id in one operation
	t.logger.Info("deleting all users for tenant", "target_tenant_id", targetTenantID)
	if err := t.userAPI.DeleteTenantUsers(tenantID, userID, targetTenantID); err != nil {
		t.logger.Error("failed to delete roles for tenant", "target_tenant_id", targetTenantID, "error", err)
		return err
	} else {
		t.logger.Info("deleted all roles for tenant", "target_tenant_id", targetTenantID)
	}

	// STEP 5: Delete ALL roles for this tenant (bulk operation)
	// This deletes all role documents with matching tenant_id in one operation
	t.logger.Info("deleting all roles for tenant", "target_tenant_id", targetTenantID)
	if err := t.rbacAPI.Roles.DeleteTenantRoles(tenantID, userID, targetTenantID); err != nil {
		t.logger.Error("failed to delete roles for tenant", "target_tenant_id", targetTenantID, "error", err)
		// Continue with deletion even if this fails
	} else {
		t.logger.Info("deleted all roles for tenant", "target_tenant_id", targetTenantID)
	}

	// STEP 6: Delete ALL permissions for this tenant (bulk operation)
	// This deletes all permission documents with matching tenant_id in one operation
	t.logger.Info("deleting all permissions for tenant", "target_tenant_id", targetTenantID)
	if err := t.rbacAPI.Permissions.DeleteTenantPermissions(tenantID, userID, targetTenantID); err != nil {
		t.logger.Error("failed to delete permissions for tenant", "target_tenant_id", targetTenantID, "error", err)
		// Continue with deletion even if this fails
	} else {
		t.logger.Info("deleted all permissions for tenant", "target_tenant_id", targetTenantID)
	}

	// STEP 7 Delete the tenant itself
	t.logger.Info("deleting tenant", "target_tenant_id", targetTenantID)
	return t.tenantHandler.DeleteTenant(targetTenantID)
}

/* Helper functions */

// checkPermission verifies if a user has the required permission
func (t *TenantAPI) checkPermission(tenantID, userID, resource, action string) error {
	// Create permission string using helper
	permString, err := model_auth.CreatePermissionString(resource, action)
	if err != nil {
		t.logger.Error("invalid permission format", "resource", resource, "action", action, "error", err)
		return err
	}

	permissions := []string{permString}
	res, err := t.rbacAPI.Verification.CheckPermissions(tenantID, userID, permissions)
	if err != nil {
		return err
	}
	// Check result
	if !res[permString] {
		t.logger.Warn("permission denied", "user_id", userID, "tenant_id", tenantID, "permission", permString)
		return infra_error.Auth(infra_error.AuthPermissionDenied)
	}

	t.logger.Debug("permission check passed", "user_id", userID, "permission", permString)
	return nil
}

/* Seeding functions */

// SeedDefaults creates default permission, role, and admin user for a new tenant
func (t *TenantAPI) seedDefaults(tenantID, adminEmail, createdBy string) (*TenantDefaults, error) {
	t.logger.Info("Seeding defaults for new tenant", "tenant_id", tenantID)

	defaults := &TenantDefaults{}

	// Step 1: Create "*:*" permission
	permissionID, err := t.createWildcardPermission(tenantID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create wildcard permission: %w", err)
	}
	defaults.PermissionID = permissionID
	t.logger.Info("Wildcard permission created", "tenant_id", tenantID, "permission_id", permissionID)

	// Step 2: Create TenantAdmin role
	roleID, err := t.createTenantAdminRole(tenantID, permissionID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create TenantAdmin role: %w", err)
	}
	defaults.RoleId = roleID
	t.logger.Info("TenantAdmin role created", "tenant_id", tenantID, "role_id", roleID)

	// Step 3: Create initial admin user in Core
	userID, err := t.createAdminUser(tenantID, db.TenantAdminUser, db.TenantAdminPassword, roleID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}
	defaults.UserId = userID
	t.logger.Info("Admin user created", "tenant_id", tenantID, "user_id", userID, "email", adminEmail)

	t.logger.Info("Tenant defaults seeded successfully", "tenant_id", tenantID)
	return defaults, nil
}

func (t *TenantAPI) createWildcardPermission(tenantID, createdBy string) (string, error) {

	permission := &authv1.Permission{
		TenantId:         tenantID,
		DisplayName:      "Full Access",
		PermissionString: db.SystemAdminPermissionID,
		Description:      "Grants full access to all resources and actions",
		Resource:         model_auth.ResourceTypeAll,     // "*"
		Action:           model_auth.PermissionActionAll, // "*"
		Status:           authv1.PermissionStatus_PERMISSION_STATUS_ACTIVE,
		CreatedBy:        createdBy,
		IsDangerous:      true,
	}

	return t.rbacAPI.Permissions.CreatePermission(tenantID, createdBy, permission, tenantID)
}

func (t *TenantAPI) createTenantAdminRole(tenantID, permissionID, createdBy string) (string, error) {
	role := &authv1.Role{
		TenantId:    tenantID,
		Name:        model_auth.RoleTenantAdmin,
		Description: "Tenant administrator with full access to all tenant resources",
		Type:        authv1.RoleType_ROLE_TYPE_SYSTEM,
		Permissions: []string{permissionID}, // Assign "*:*" permission
		Status:      authv1.RoleStatus_ROLE_STATUS_ACTIVE,
		CreatedBy:   createdBy,
	}

	return t.rbacAPI.Roles.CreateRole(tenantID, createdBy, role, tenantID)
}

func (t *TenantAPI) createAdminUser(tenantID, username, plainPassword, roleID, createdBy string) (string, error) {
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
	return t.userAPI.userHandler.CreateUser(user)
}

// RollbackDefaults deletes all seeded defaults (used when tenant creation fails)
func (t *TenantAPI) RollbackDefaults(ctx context.Context, tenantID string, defaults *TenantDefaults) error {
	t.logger.Warn("Rolling back tenant defaults", "tenant_id", tenantID)

	var rollbackErrors []error

	// Delete admin user (local collection)
	if defaults.UserId != "" {
		if err := t.userAPI.userHandler.DeleteUser(tenantID, defaults.UserId); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to delete admin user: %w", err))
		}
	}

	// Delete role (via Auth gRPC)
	if defaults.RoleId != "" {
		if err := t.rbacAPI.Roles.DeleteRole(tenantID, defaults.UserId, defaults.RoleId, tenantID); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to delete role via gRPC: %w", err))
		}
	}

	// Delete permission (via Auth gRPC)
	if defaults.PermissionID != "" {
		if err := t.rbacAPI.Permissions.DeletePermission(tenantID, defaults.UserId, defaults.PermissionID, tenantID); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to delete permission via gRPC: %w", err))
		}
	}

	if len(rollbackErrors) > 0 {
		return fmt.Errorf("rollback partially failed: %v", rollbackErrors)
	}

	t.logger.Info("Tenant defaults rolled back successfully", "tenant_id", tenantID)
	return nil
}
