package api

import (
	"erp.localhost/internal/auth/handler"
	"erp.localhost/internal/auth/rbac"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
)

// RoleAPI provides role management with authorization enforcement
type RoleAPI struct {
	roleHandler         *handler.RoleHandler
	verificationManager *rbac.VerificationManager
	logger              logger.Logger
}

// NewRoleAPI creates a new RoleAPI instance
func NewRoleAPI(
	roleHandler *handler.RoleHandler,
	verificationManager *rbac.VerificationManager,
	logger logger.Logger,
) *RoleAPI {
	return &RoleAPI{
		roleHandler:         roleHandler,
		verificationManager: verificationManager,
		logger:              logger,
	}
}

// CreateRole creates a new role with authorization check
func (ra *RoleAPI) CreateRole(tenantID, requestorUserID string, role *authv1.Role, targetTenantID string) (string, error) {
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
	return ra.roleHandler.CreateRole(role)
}

// UpdateRole updates an existing role with authorization check
func (ra *RoleAPI) UpdateRole(tenantID, requestorUserID string, role *authv1.Role, targetTenantID string) error {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionUpdate)
	if err != nil {
		return err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for UpdateRole", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return err
	}

	return ra.roleHandler.UpdateRole(role)
}

// GetRoleByID retrieves a role by ID with authorization check
func (ra *RoleAPI) GetRoleByID(tenantID, requestorUserID, roleID string, targetTenantID string) (*authv1.Role, error) {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionRead)
	if err != nil {
		return nil, err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for GetRoleByID", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return nil, err
	}

	return ra.roleHandler.GetRoleByID(targetTenantID, roleID)
}

// ListRoles retrieves all roles for a tenant with authorization check
func (ra *RoleAPI) ListRoles(tenantID, requestorUserID string, targetTenantID string) ([]*authv1.Role, error) {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionRead)
	if err != nil {
		return nil, err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for ListRoles", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return nil, err
	}

	return ra.roleHandler.GetRolesByTenantID(targetTenantID)
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

	return ra.roleHandler.DeleteRole(targetTenantID, roleID)
}

func (ra *RoleAPI) DeleteTenantRoles(tenantID, requestorUserID, targetTenantID string) error {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionDelete)
	if err != nil {
		return err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for DeleteRole", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return err
	}

	return ra.roleHandler.DeleteTenantRoles(targetTenantID)
}
