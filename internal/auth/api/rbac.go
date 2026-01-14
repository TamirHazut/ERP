package api

import (
	"erp.localhost/internal/auth/rbac"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
)

// RoleAPI provides role management with authorization enforcement
type RoleAPI struct {
	roleManager         *rbac.RoleManager
	verificationManager *rbac.VerificationManager
	logger              logger.Logger
}

// NewRoleAPI creates a new RoleAPI instance
func NewRoleAPI(
	roleManager *rbac.RoleManager,
	verificationManager *rbac.VerificationManager,
	logger logger.Logger,
) *RoleAPI {
	return &RoleAPI{
		roleManager:         roleManager,
		verificationManager: verificationManager,
		logger:              logger,
	}
}

// CreateRole creates a new role with authorization check
func (ra *RoleAPI) CreateRole(tenantID, requestorUserID string, role *model_auth.Role, targetTenantID string) (string, error) {
	// 1. Check permission (with cross-tenant support)
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionCreate)
	if err != nil {
		return "", err
	}

	// targetTenantID is the tenant where the role will be created
	// If requestor is system tenant user, they can create roles in any tenant
	// If requestor is tenant admin, they can create roles in their own tenant
	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for CreateRole", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return "", err
	}

	// 2. Call business logic
	return ra.roleManager.CreateRole(role)
}

// UpdateRole updates an existing role with authorization check
func (ra *RoleAPI) UpdateRole(tenantID, requestorUserID string, role *model_auth.Role, targetTenantID string) error {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionUpdate)
	if err != nil {
		return err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for UpdateRole", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return err
	}

	return ra.roleManager.UpdateRole(role)
}

// GetRoleByID retrieves a role by ID with authorization check
func (ra *RoleAPI) GetRoleByID(tenantID, requestorUserID, roleID string, targetTenantID string) (*model_auth.Role, error) {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionRead)
	if err != nil {
		return nil, err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for GetRoleByID", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return nil, err
	}

	return ra.roleManager.GetRoleByID(targetTenantID, roleID)
}

// ListRoles retrieves all roles for a tenant with authorization check
func (ra *RoleAPI) ListRoles(tenantID, requestorUserID string, targetTenantID string) ([]*model_auth.Role, error) {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionRead)
	if err != nil {
		return nil, err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for ListRoles", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return nil, err
	}

	return ra.roleManager.ListRoles(targetTenantID)
}

// DeleteRole deletes a role with authorization check
func (ra *RoleAPI) DeleteRole(tenantID, requestorUserID, roleID string, targetTenantID string) error {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionDelete)
	if err != nil {
		return err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for DeleteRole", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return err
	}

	return ra.roleManager.DeleteRole(targetTenantID, roleID)
}

// PermissionAPI provides permission management with authorization enforcement
type PermissionAPI struct {
	permissionManager   *rbac.PermissionManager
	verificationManager *rbac.VerificationManager
	logger              logger.Logger
}

// NewPermissionAPI creates a new PermissionAPI instance
func NewPermissionAPI(
	permissionManager *rbac.PermissionManager,
	verificationManager *rbac.VerificationManager,
	logger logger.Logger,
) *PermissionAPI {
	return &PermissionAPI{
		permissionManager:   permissionManager,
		verificationManager: verificationManager,
		logger:              logger,
	}
}

// CreatePermission creates a new permission with authorization check
func (pa *PermissionAPI) CreatePermission(tenantID, requestorUserID string, permission *model_auth.Permission, targetTenantID string) (string, error) {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionCreate)
	if err != nil {
		return "", err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for CreatePermission", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return "", err
	}

	return pa.permissionManager.CreatePermission(permission)
}

// UpdatePermission updates an existing permission with authorization check
func (pa *PermissionAPI) UpdatePermission(tenantID, requestorUserID string, permission *model_auth.Permission, targetTenantID string) error {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionUpdate)
	if err != nil {
		return err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for UpdatePermission", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return err
	}

	return pa.permissionManager.UpdatePermission(permission)
}

// GetPermissionByID retrieves a permission by ID with authorization check
func (pa *PermissionAPI) GetPermissionByID(tenantID, requestorUserID, permissionID string, targetTenantID string) (*model_auth.Permission, error) {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionRead)
	if err != nil {
		return nil, err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for GetPermissionByID", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return nil, err
	}

	return pa.permissionManager.GetPermissionByID(targetTenantID, permissionID)
}

// ListPermissions retrieves all permissions for a tenant with authorization check
func (pa *PermissionAPI) ListPermissions(tenantID, requestorUserID string, targetTenantID string) ([]*model_auth.Permission, error) {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionRead)
	if err != nil {
		return nil, err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for ListPermissions", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return nil, err
	}

	return pa.permissionManager.ListPermissions(targetTenantID)
}

// DeletePermission deletes a permission with authorization check
func (pa *PermissionAPI) DeletePermission(tenantID, requestorUserID, permissionID string, targetTenantID string) error {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionDelete)
	if err != nil {
		return err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for DeletePermission", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return err
	}

	return pa.permissionManager.DeletePermission(targetTenantID, permissionID)
}

// VerificationAPI provides permission verification operations (no authorization needed)
type VerificationAPI struct {
	verificationManager *rbac.VerificationManager
	logger              logger.Logger
}

// NewVerificationAPI creates a new VerificationAPI instance
func NewVerificationAPI(
	verificationManager *rbac.VerificationManager,
	logger logger.Logger,
) *VerificationAPI {
	return &VerificationAPI{
		verificationManager: verificationManager,
		logger:              logger,
	}
}

// GetUserPermissions retrieves all permissions for a user
func (va *VerificationAPI) GetUserPermissions(tenantID, userID string) (map[string]bool, error) {
	return va.verificationManager.GetUserPermissions(tenantID, userID)
}

// GetUserRoles retrieves all role IDs for a user
func (va *VerificationAPI) GetUserRoles(tenantID, userID string) ([]string, error) {
	return va.verificationManager.GetUserRoles(tenantID, userID)
}

// CheckPermissions checks if a user has specific permissions
func (va *VerificationAPI) CheckPermissions(tenantID, userID string, permissions []string) (map[string]bool, error) {
	return va.verificationManager.CheckPermissions(tenantID, userID, permissions)
}

// HasPermission checks if a user has a specific permission (with cross-tenant support)
func (va *VerificationAPI) HasPermission(tenantID, userID, permission string, targetTenantID string) error {
	return va.verificationManager.HasPermission(tenantID, userID, permission, targetTenantID)
}

// IsSystemTenantUser checks if a user belongs to the system tenant
func (va *VerificationAPI) IsSystemTenantUser(tenantID string) bool {
	return va.verificationManager.IsSystemTenantUser(tenantID)
}

// RBACAPI combines all RBAC APIs for easy initialization
type RBACAPI struct {
	Roles        *RoleAPI
	Permissions  *PermissionAPI
	Verification *VerificationAPI
}

// NewRBACAPI creates a new RBACAPI instance with all sub-APIs
func NewRBACAPI(
	roleManager *rbac.RoleManager,
	permissionManager *rbac.PermissionManager,
	verificationManager *rbac.VerificationManager,
	logger logger.Logger,
) *RBACAPI {
	return &RBACAPI{
		Roles:        NewRoleAPI(roleManager, verificationManager, logger),
		Permissions:  NewPermissionAPI(permissionManager, verificationManager, logger),
		Verification: NewVerificationAPI(verificationManager, logger),
	}
}
