package api

import (
	"erp.localhost/internal/auth/handler"
	"erp.localhost/internal/auth/rbac"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
)

// PermissionAPI provides permission management with authorization enforcement
type PermissionAPI struct {
	permissionHandler   *handler.PermissionHandler
	verificationManager *rbac.VerificationManager
	logger              logger.Logger
}

// NewPermissionAPI creates a new PermissionAPI instance
func NewPermissionAPI(
	permissionHandler *handler.PermissionHandler,
	verificationManager *rbac.VerificationManager,
	logger logger.Logger,
) *PermissionAPI {
	return &PermissionAPI{
		permissionHandler:   permissionHandler,
		verificationManager: verificationManager,
		logger:              logger,
	}
}

// CreatePermission creates a new permission with authorization check
func (pa *PermissionAPI) CreatePermission(tenantID, requestorUserID string, permission *authv1.Permission, targetTenantID string) (string, error) {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionCreate)
	if err != nil {
		return "", err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for CreatePermission", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return "", err
	}

	return pa.permissionHandler.CreatePermission(permission)
}

// UpdatePermission updates an existing permission with authorization check
func (pa *PermissionAPI) UpdatePermission(tenantID, requestorUserID string, permission *authv1.Permission, targetTenantID string) error {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionUpdate)
	if err != nil {
		return err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for UpdatePermission", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return err
	}

	return pa.permissionHandler.UpdatePermission(permission)
}

// GetPermissionByID retrieves a permission by ID with authorization check
func (pa *PermissionAPI) GetPermissionByID(tenantID, requestorUserID, permissionID string, targetTenantID string) (*authv1.Permission, error) {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionRead)
	if err != nil {
		return nil, err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for GetPermissionByID", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return nil, err
	}

	return pa.permissionHandler.GetPermissionByID(targetTenantID, permissionID)
}

// ListPermissions retrieves all permissions for a tenant with authorization check
func (pa *PermissionAPI) ListPermissions(tenantID, requestorUserID string, targetTenantID string) ([]*authv1.Permission, error) {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionRead)
	if err != nil {
		return nil, err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for ListPermissions", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return nil, err
	}

	return pa.permissionHandler.GetPermissionsByTenantID(targetTenantID)
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

	return pa.permissionHandler.DeletePermission(targetTenantID, permissionID)
}

// DeletePermission deletes a permission with authorization check
func (pa *PermissionAPI) DeleteTenantPermissions(tenantID, requestorUserID, targetTenantID string) error {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionDelete)
	if err != nil {
		return err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for DeleteTenantPermissions", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return err
	}

	return pa.permissionHandler.DeleteTenantPermissions(targetTenantID)
}
