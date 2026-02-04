package api

import (
	"erp.localhost/internal/auth/handler"
	"erp.localhost/internal/auth/rbac"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_auth "erp.localhost/internal/infra/model/auth/validator"
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
func (ra *RoleAPI) CreateRole(tenantID, requestorUserID string, role *authv1.Role, targetTenantID string) (string, *infra_error.AppError) {
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

	if role.Protected && !ra.verificationManager.IsSystemTenantUser(tenantID) {
		role.Protected = false
	}

	// 2. Validate and normalize permission strings
	if err := validator_auth.ValidatePermissionStrings(role.Permissions); err != nil {
		ra.logger.Warn("Invalid permission strings in role", "tenant_id", role.TenantId, "role_name", role.Name, "error", err)
		return "", err
	}
	role.Permissions = model_auth.NormalizePermissions(role.Permissions)

	// 3. Call business logic
	return ra.roleHandler.CreateRole(role)
}

// UpdateRole updates an existing role with authorization check
func (ra *RoleAPI) UpdateRole(tenantID, requestorUserID string, role *authv1.Role, targetTenantID string) *infra_error.AppError {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionUpdate)
	if err != nil {
		return err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for UpdateRole", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return err
	}

	// Validate and normalize permission strings
	if err := validator_auth.ValidatePermissionStrings(role.Permissions); err != nil {
		ra.logger.Warn("Invalid permission strings in role", "tenant_id", role.TenantId, "role_id", role.Id, "error", err)
		return err
	}
	role.Permissions = model_auth.NormalizePermissions(role.Permissions)

	return ra.roleHandler.UpdateRole(role)
}

// GetRoleByID retrieves a role by ID with authorization check
func (ra *RoleAPI) GetRoleByID(tenantID, requestorUserID, roleID string, targetTenantID string) (*authv1.Role, *infra_error.AppError) {
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
func (ra *RoleAPI) ListRoles(tenantID, requestorUserID string, targetTenantID string) ([]*authv1.Role, *infra_error.AppError) {
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
func (ra *RoleAPI) DeleteRole(tenantID, requestorUserID, roleID, targetTenantID string) *infra_error.AppError {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionDelete)
	if err != nil {
		return err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for DeleteRole", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return err
	}

	if err := ra.checkProtectedRole(tenantID, requestorUserID, roleID); err != nil {
		return err
	}

	if err := ra.verificationManager.VerifyRoleNotAssigned(targetTenantID, roleID); err != nil {
		return err
	}

	return ra.roleHandler.DeleteRole(targetTenantID, roleID)
}

func (ra *RoleAPI) DeleteTenantRoles(tenantID, requestorUserID, targetTenantID string) *infra_error.AppError {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeRole, model_auth.PermissionActionDelete)
	if err != nil {
		return err
	}

	if err := ra.verificationManager.HasPermission(tenantID, requestorUserID, permission, targetTenantID); err != nil {
		ra.logger.Warn("Permission denied for DeleteRole", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permission)
		return err
	}

	if err := ra.verificationManager.VerifyRolesNotAssigned(targetTenantID); err != nil {
		return err
	}

	return ra.roleHandler.DeleteTenantRoles(targetTenantID)
}

func (ra *RoleAPI) checkProtectedRole(tenantID, requestorUserID, roleID string) *infra_error.AppError {
	role, err := ra.roleHandler.GetRoleByID(tenantID, roleID)
	if err != nil {
		return err
	}
	if role.Protected {
		if !ra.verificationManager.IsSystemTenantUser(tenantID) ||
			!ra.verificationManager.IsSystemAdminUser(requestorUserID) ||
			ra.verificationManager.IsSystemRole(roleID) {
			return infra_error.Auth(infra_error.AuthPermissionDenied)
		}
	}
	return nil
}
