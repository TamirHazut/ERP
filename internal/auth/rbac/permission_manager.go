package rbac

import (
	collection "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
)

type PermissionManager struct {
	permissionsCollection *collection.PermissionsCollection
	logger                logger.Logger
}

// NewPermissionManager creates a new PermissionManager instance
func NewPermissionManager(
	permissionsCollection *collection.PermissionsCollection,
	logger logger.Logger,
) *PermissionManager {
	return &PermissionManager{
		permissionsCollection: permissionsCollection,
		logger:                logger,
	}
}

// CreatePermission creates a new permission
func (pm *PermissionManager) CreatePermission(permission *model_auth.Permission) (string, error) {
	pm.logger.Debug("PermissionManager: Creating permission", "permission_name", permission.DisplayName, "tenant_id", permission.TenantID)
	return pm.permissionsCollection.CreatePermission(permission)
}

// UpdatePermission updates an existing permission
func (pm *PermissionManager) UpdatePermission(permission *model_auth.Permission) error {
	pm.logger.Debug("PermissionManager: Updating permission", "permission_id", permission.ID.Hex(), "tenant_id", permission.TenantID)
	return pm.permissionsCollection.UpdatePermission(permission)
}

// GetPermissionByID retrieves a permission by its ID
func (pm *PermissionManager) GetPermissionByID(tenantID, permissionID string) (*model_auth.Permission, error) {
	pm.logger.Debug("PermissionManager: Getting permission by ID", "permission_id", permissionID, "tenant_id", tenantID)
	return pm.permissionsCollection.GetPermissionByID(tenantID, permissionID)
}

// GetPermissionByName retrieves a permission by its name
func (pm *PermissionManager) GetPermissionByName(tenantID, name string) (*model_auth.Permission, error) {
	pm.logger.Debug("PermissionManager: Getting permission by name", "permission_name", name, "tenant_id", tenantID)
	return pm.permissionsCollection.GetPermissionByName(tenantID, name)
}

// ListPermissions retrieves all permissions for a tenant
func (pm *PermissionManager) ListPermissions(tenantID string) ([]*model_auth.Permission, error) {
	pm.logger.Debug("PermissionManager: Listing permissions", "tenant_id", tenantID)
	return pm.permissionsCollection.GetPermissionsByTenantID(tenantID)
}

// DeletePermission deletes a permission
func (pm *PermissionManager) DeletePermission(tenantID, permissionID string) error {
	pm.logger.Debug("PermissionManager: Deleting permission", "permission_id", permissionID, "tenant_id", tenantID)
	return pm.permissionsCollection.DeletePermission(tenantID, permissionID)
}

// DeletePermission deletes all the tenant permissions
func (pm *PermissionManager) DeleteTenantPermissions(tenantID string) error {
	pm.logger.Debug("PermissionManager: Deleting permission", "tenant_id", tenantID)
	return pm.permissionsCollection.DeleteTenantPermissions(tenantID)
}
