package api

import (
	"erp.localhost/internal/auth/handler"
	"erp.localhost/internal/auth/rbac"
	"erp.localhost/internal/infra/logging/logger"
)

// RBACAPI combines all RBAC APIs for easy initialization
type RBACAPI struct {
	Roles        *RoleAPI
	Permissions  *PermissionAPI
	Verification *VerificationAPI
}

// NewRBACAPI creates a new RBACAPI instance with all sub-APIs
func NewRBACAPI(
	roleHandler *handler.RoleHandler,
	permissionHandler *handler.PermissionHandler,
	verificationManager *rbac.VerificationManager,
	logger logger.Logger,
) *RBACAPI {
	return &RBACAPI{
		Roles:        NewRoleAPI(roleHandler, verificationManager, logger),
		Permissions:  NewPermissionAPI(permissionHandler, verificationManager, logger),
		Verification: NewVerificationAPI(verificationManager, logger),
	}
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
func (va *VerificationAPI) GetUserPermissionsIDs(tenantID, userID string) (map[string]bool, error) {
	return va.verificationManager.GetUserPermissionsIDs(tenantID, userID)
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
