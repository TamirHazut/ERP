package rbac

import (
	collection "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
)

type RoleManager struct {
	rolesCollection *collection.RolesCollection
	logger          logger.Logger
}

// NewRoleManager creates a new RoleManager instance
func NewRoleManager(
	rolesCollection *collection.RolesCollection,
	logger logger.Logger,
) *RoleManager {
	return &RoleManager{
		rolesCollection: rolesCollection,
		logger:          logger,
	}
}

// CreateRole creates a new role
func (rm *RoleManager) CreateRole(role *model_auth.Role) (string, error) {
	rm.logger.Debug("RoleManager: Creating role", "role_name", role.Name, "tenant_id", role.TenantID)
	return rm.rolesCollection.CreateRole(role)
}

// UpdateRole updates an existing role
func (rm *RoleManager) UpdateRole(role *model_auth.Role) error {
	rm.logger.Debug("RoleManager: Updating role", "role_id", role.ID.Hex(), "tenant_id", role.TenantID)
	return rm.rolesCollection.UpdateRole(role)
}

// GetRoleByID retrieves a role by its ID
func (rm *RoleManager) GetRoleByID(tenantID, roleID string) (*model_auth.Role, error) {
	rm.logger.Debug("RoleManager: Getting role by ID", "role_id", roleID, "tenant_id", tenantID)
	return rm.rolesCollection.GetRoleByID(tenantID, roleID)
}

// GetRoleByName retrieves a role by its name
func (rm *RoleManager) GetRoleByName(tenantID, name string) (*model_auth.Role, error) {
	rm.logger.Debug("RoleManager: Getting role by name", "role_name", name, "tenant_id", tenantID)
	return rm.rolesCollection.GetRoleByName(tenantID, name)
}

// ListRoles retrieves all roles for a tenant
func (rm *RoleManager) ListRoles(tenantID string) ([]*model_auth.Role, error) {
	rm.logger.Debug("RoleManager: Listing roles", "tenant_id", tenantID)
	return rm.rolesCollection.GetRolesByTenantID(tenantID)
}

// DeleteRole deletes a role
func (rm *RoleManager) DeleteRole(tenantID, roleID string) error {
	rm.logger.Debug("RoleManager: Deleting role", "role_id", roleID, "tenant_id", tenantID)
	return rm.rolesCollection.DeleteRole(tenantID, roleID)
}
