package rbac

import (
	"erp.localhost/internal/auth/handler"
	"erp.localhost/internal/infra/db"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
)

type VerificationManager struct {
	userHandler       *handler.UserHandler
	roleHandler       *handler.RoleHandler
	permissionHandler *handler.PermissionHandler
	tenantHandler     *handler.TenantHandler
	systemTenantID    string // System tenant ID (from config or constant)
	logger            logger.Logger
}

// NewVerificationManager creates a new VerificationManager instance
func NewVerificationManager(
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
	permissionHandler *handler.PermissionHandler,
	tenantHandler *handler.TenantHandler,
	logger logger.Logger,
) *VerificationManager {
	return &VerificationManager{
		userHandler:       userHandler,
		roleHandler:       roleHandler,
		permissionHandler: permissionHandler,
		tenantHandler:     tenantHandler,
		systemTenantID:    db.SystemTenantID,
		logger:            logger,
	}
}

// GetUserPermissionsIDs retrieves all the users permissions in a map with the format <id> -> <has permission (true/false)>
func (vm *VerificationManager) GetUserPermissionsIDs(tenantID, userID string) (map[string]bool, error) {
	// 1. Get user from UserCollection
	user, err := vm.userHandler.GetUserByID(tenantID, userID)
	if err != nil {
		vm.logger.Error(err.Error())
		return nil, err
	}

	if vm.isTenantAdmin(user) {
		// Return all permission IDs from database
		return vm.getAllPermissionIDs(tenantID), nil
	}

	// 3. Resolve permissions from user.Roles
	userPermissions := make(map[string]bool)
	for _, userRole := range user.Roles {
		role, err := vm.roleHandler.GetRoleByID(tenantID, userRole.RoleId)
		if err != nil {
			vm.logger.Error(err.Error())
			return nil, err
		}
		for _, permission := range role.Permissions {
			perm, err := vm.permissionHandler.GetPermissionByID(tenantID, permission)
			if err != nil {
				continue
			}
			switch perm.Status {
			case authv1.PermissionStatus_PERMISSION_STATUS_ACTIVE:
				userPermissions[perm.PermissionString] = true
			default:
				userPermissions[perm.PermissionString] = false
			}
		}
	}

	// 4. Apply user.AdditionalPermissions
	for _, permission := range user.AdditionalPermissions {
		perm, err := vm.permissionHandler.GetPermissionByID(tenantID, permission)
		if err != nil {
			continue
		}
		switch perm.Status {
		case authv1.PermissionStatus_PERMISSION_STATUS_ACTIVE:
			userPermissions[perm.PermissionString] = true
		default:
			userPermissions[perm.PermissionString] = false
		}
	}

	// 5. Apply user.RevokedPermissions
	for _, permission := range user.RevokedPermissions {
		userPermissions[permission] = false
	}

	return userPermissions, nil
}

// Returns permission strings (for RBAC checks like "users:read")
// OPTIMIZED: Uses MongoDB aggregation to replace 70+ queries with 1-2 queries
func (vm *VerificationManager) GetUserPermissions(tenantID, userID string) (map[string]bool, error) {
	// OPTIMIZATION: Check admin status using aggregation (1 query instead of N)
	roles, err := vm.roleHandler.GetUserRolesAggregation(tenantID, userID, []string{"name"})
	if err != nil {
		// Fallback to original method if aggregation fails
		vm.logger.Warn("role aggregation failed, falling back to original method", "error", err)
		return vm.getUserPermissionsLegacy(tenantID, userID)
	}

	// Check if user has admin role
	for _, role := range roles {
		if role.Name == model_auth.RoleTenantAdmin || role.Name == model_auth.RoleSystemAdmin {
			return vm.getAllPermissions(), nil
		}
	}

	// OPTIMIZATION: Get all permissions in single aggregation (1 query instead of 50+)
	permissions, err := vm.permissionHandler.GetUserPermissionsAggregation(tenantID, userID, nil)
	if err != nil {
		vm.logger.Warn("permission aggregation failed, falling back to original method", "error", err)
		return vm.getUserPermissionsLegacy(tenantID, userID)
	}

	// Process results into permission map
	userPermissions := make(map[string]bool)
	for _, perm := range permissions {
		if perm.Status == authv1.PermissionStatus_PERMISSION_STATUS_ACTIVE {
			userPermissions[perm.PermissionString] = true
		}
	}

	// Handle additional and revoked permissions
	// These are much smaller sets, so individual queries are acceptable
	user, err := vm.userHandler.GetUserByID(tenantID, userID)
	if err == nil {
		// Apply additional permissions
		for _, permissionID := range user.AdditionalPermissions {
			perm, err := vm.permissionHandler.GetPermissionByID(tenantID, permissionID)
			if err != nil {
				continue
			}
			if perm.Status == authv1.PermissionStatus_PERMISSION_STATUS_ACTIVE {
				userPermissions[perm.PermissionString] = true
			}
		}

		// Apply revoked permissions
		for _, permissionID := range user.RevokedPermissions {
			perm, err := vm.permissionHandler.GetPermissionByID(tenantID, permissionID)
			if err != nil {
				continue
			}
			userPermissions[perm.PermissionString] = false
		}
	}

	return userPermissions, nil
}

// getUserPermissionsLegacy is the original implementation kept as fallback
func (vm *VerificationManager) getUserPermissionsLegacy(tenantID, userID string) (map[string]bool, error) {
	user, err := vm.userHandler.GetUserByID(tenantID, userID)
	if err != nil {
		return nil, err
	}

	if vm.isTenantAdmin(user) {
		return vm.getAllPermissions(), nil
	}

	userPermissions := make(map[string]bool)

	// Resolve from roles
	for _, userRole := range user.Roles {
		role, err := vm.roleHandler.GetRoleByID(tenantID, userRole.RoleId)
		if err != nil {
			continue
		}
		for _, permissionID := range role.Permissions {
			perm, err := vm.permissionHandler.GetPermissionByID(tenantID, permissionID)
			if err != nil {
				continue
			}
			userPermissions[perm.PermissionString] = true
		}
	}

	// Apply additional permissions
	for _, permissionID := range user.AdditionalPermissions {
		perm, err := vm.permissionHandler.GetPermissionByID(tenantID, permissionID)
		if err != nil {
			continue
		}
		switch perm.Status {
		case authv1.PermissionStatus_PERMISSION_STATUS_ACTIVE:
			userPermissions[perm.PermissionString] = true
		default:
			userPermissions[perm.PermissionString] = false
		}
	}

	// Apply revoked permissions
	for _, permissionID := range user.RevokedPermissions {
		perm, err := vm.permissionHandler.GetPermissionByID(tenantID, permissionID)
		if err != nil {
			continue
		}
		userPermissions[perm.PermissionString] = false
	}

	return userPermissions, nil
}

// Check if user belongs to system tenant
func (vm *VerificationManager) IsSystemTenantUser(tenantID string) bool {
	return tenantID == vm.systemTenantID
}

// Check if user has tenant admin role
// OPTIMIZED: Uses MongoDB aggregation to replace N queries with 1 query
func (vm *VerificationManager) isTenantAdmin(user *authv1.User) bool {
	roles, err := vm.roleHandler.GetUserRolesAggregation(user.TenantId, user.Id, []string{"name"})
	if err != nil {
		// Fallback to original method if aggregation fails
		vm.logger.Warn("role aggregation failed in isTenantAdmin, falling back", "error", err)
		return vm.isTenantAdminLegacy(user)
	}

	for _, role := range roles {
		if role.Name == model_auth.RoleTenantAdmin || role.Name == model_auth.RoleSystemAdmin {
			return true
		}
	}
	return false
}

// isTenantAdminLegacy is the original implementation kept as fallback
func (vm *VerificationManager) isTenantAdminLegacy(user *authv1.User) bool {
	for _, userRole := range user.Roles {
		role, err := vm.roleHandler.GetRoleByID(user.TenantId, userRole.RoleId)
		if err != nil {
			continue
		}
		if role.Name == model_auth.RoleTenantAdmin || role.Name == model_auth.RoleSystemAdmin {
			return true
		}
	}
	return false
}

// Get all permission IDs (for tenant admin)
func (vm *VerificationManager) getAllPermissionIDs(tenantID string) map[string]bool {
	// Query all permissions from database
	permissions, err := vm.permissionHandler.GetPermissionsByTenantID(tenantID)
	if err != nil {
		vm.logger.Error("failed to get all permissions", "error", err)
		return map[string]bool{}
	}

	result := make(map[string]bool)
	for _, perm := range permissions {
		result[perm.Id] = true
	}

	return result
}

// Get all possible permissions (for tenant admin)
func (vm *VerificationManager) getAllPermissions() map[string]bool {
	wildCard, _ := model_auth.CreatePermissionString(model_auth.ResourceTypeAll, model_auth.PermissionActionAll)
	// Query all permissions from PermissionsCollection
	// Or return a predefined set of all possible permissions
	return map[string]bool{
		// All possible permissions are granted
		wildCard: true, // Wildcard permission
	}
}

// GetUserRoles returns all role IDs assigned to a user
func (vm *VerificationManager) GetUserRoles(tenantID, userID string) ([]string, error) {
	// Get user from UserCollection
	user, err := vm.userHandler.GetUserByID(tenantID, userID)
	if err != nil {
		vm.logger.Error(err.Error())
		return nil, err
	}

	// Extract role IDs
	roleIDs := make([]string, 0, len(user.Roles))
	for _, userRole := range user.Roles {
		roleIDs = append(roleIDs, userRole.RoleId)
	}

	return roleIDs, nil
}

// CheckPermissions with system tenant and tenant admin logic
func (vm *VerificationManager) CheckPermissions(tenantID, userID string, permissions []string) (map[string]bool, error) {
	// 1. Get user
	user, err := vm.userHandler.GetUserByID(tenantID, userID)
	if err != nil {
		vm.logger.Error(err.Error())
		return nil, err
	}
	// 2. Check if tenant admin → grant all
	if vm.isTenantAdmin(user) {
		result := make(map[string]bool)
		for _, perm := range permissions {
			result[perm] = true
		}
		return result, nil
	}

	// 3. Get user permissions
	userPermissions, err := vm.GetUserPermissions(tenantID, userID)
	if err != nil {
		return nil, err
	}

	// 4. Check each permission
	result := make(map[string]bool)
	for _, perm := range permissions {
		userPerm, ok := userPermissions[perm]
		if !ok {
			userPerm = false
		}
		result[perm] = userPerm
	}

	return result, nil
}

// HasPermission with cross-tenant check for system tenant users
func (vm *VerificationManager) HasPermission(tenantID, userID, permission string, targetTenantID string) error {
	// 1. Get user
	user, err := vm.userHandler.GetUserByID(tenantID, userID)
	if err != nil {
		return err
	}

	// 2. Check if tenant admin (for same tenant operations)
	if tenantID == targetTenantID && vm.isTenantAdmin(user) {
		return nil // Tenant admin has all permissions in their tenant
	}

	// 3. Check if system tenant user (cross-tenant operations)
	if vm.IsSystemTenantUser(tenantID) {
		// System tenant users can operate on all tenants
		// Just check if they have the permission (no tenant restriction)
		userPermissions, err := vm.GetUserPermissions(tenantID, userID)
		if err != nil {
			return err
		}
		if userPermissions[permission] {
			return nil // System user has permission for cross-tenant operation
		}
		return infra_error.Auth(infra_error.AuthPermissionDenied)
	}

	// 4. Regular permission check (same tenant only)
	if tenantID != targetTenantID {
		return infra_error.Auth(infra_error.AuthPermissionDenied)
	}

	userPermissions, err := vm.GetUserPermissions(tenantID, userID)
	if err != nil {
		return err
	}

	if !userPermissions[permission] {
		return infra_error.Auth(infra_error.AuthPermissionDenied)
	}

	return nil
}
