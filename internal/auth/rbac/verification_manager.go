package rbac

import (
	collection "erp.localhost/internal/auth/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
)

type VerificationManager struct {
	userCollection        *collection.UserCollection
	rolesCollection       *collection.RolesCollection
	permissionsCollection *collection.PermissionsCollection
	tenantCollection      *collection.TenantCollection
	systemTenantID        string // System tenant ID (from config or constant)
	logger                logger.Logger
}

// NewVerificationManager creates a new VerificationManager instance
func NewVerificationManager(
	userCollection *collection.UserCollection,
	rolesCollection *collection.RolesCollection,
	permissionsCollection *collection.PermissionsCollection,
	tenantCollection *collection.TenantCollection,
	logger logger.Logger,
) *VerificationManager {
	return &VerificationManager{
		userCollection:        userCollection,
		rolesCollection:       rolesCollection,
		permissionsCollection: permissionsCollection,
		tenantCollection:      tenantCollection,
		systemTenantID:        model_auth.SystemTenantID,
		logger:                logger,
	}
}

// Core implementation - uncomment and adapt from rbac_manager.go:106-129
func (vm *VerificationManager) GetUserPermissions(tenantID, userID string) (map[string]bool, error) {
	// 1. Get user from UserCollection
	user, err := vm.userCollection.GetUserByID(tenantID, userID)
	if err != nil {
		vm.logger.Error(err.Error())
		return nil, err
	}

	// 2. Check if user is tenant admin → grant ALL permissions
	if vm.isTenantAdmin(user) {
		return vm.getAllPermissions(), nil // Returns map with all possible permissions
	}

	// 3. Resolve permissions from user.Roles
	userPermissions := make(map[string]bool)
	for _, userRole := range user.Roles {
		role, err := vm.rolesCollection.GetRoleByID(tenantID, userRole.RoleID)
		if err != nil {
			vm.logger.Error(err.Error())
			return nil, err
		}
		for _, permission := range role.Permissions {
			userPermissions[permission] = true
		}
	}

	// 4. Apply user.AdditionalPermissions
	for _, permission := range user.AdditionalPermissions {
		userPermissions[permission] = true
	}

	// 5. Apply user.RevokedPermissions
	for _, permission := range user.RevokedPermissions {
		userPermissions[permission] = false
	}

	return userPermissions, nil
}

// Check if user belongs to system tenant
func (vm *VerificationManager) IsSystemTenantUser(tenantID string) bool {
	return tenantID == vm.systemTenantID
}

// Check if user has tenant admin role
func (vm *VerificationManager) isTenantAdmin(user *model_auth.User) bool {
	for _, userRole := range user.Roles {
		role, err := vm.rolesCollection.GetRoleByID(user.TenantID, userRole.RoleID)
		if err != nil {
			continue
		}
		if role.Name == model_auth.RoleTenantAdmin || role.IsTenantAdmin {
			return true
		}
	}
	return false
}

// Get all possible permissions (for tenant admin)
func (vm *VerificationManager) getAllPermissions() map[string]bool {
	// Query all permissions from PermissionsCollection
	// Or return a predefined set of all possible permissions
	return map[string]bool{
		// All possible permissions are granted
		"*:*": true, // Wildcard permission
	}
}

// GetUserRoles returns all role IDs assigned to a user
func (vm *VerificationManager) GetUserRoles(tenantID, userID string) ([]string, error) {
	// Get user from UserCollection
	user, err := vm.userCollection.GetUserByID(tenantID, userID)
	if err != nil {
		vm.logger.Error(err.Error())
		return nil, err
	}

	// Extract role IDs
	roleIDs := make([]string, 0, len(user.Roles))
	for _, userRole := range user.Roles {
		roleIDs = append(roleIDs, userRole.RoleID)
	}

	return roleIDs, nil
}

// CheckPermissions with system tenant and tenant admin logic
func (vm *VerificationManager) CheckPermissions(tenantID, userID string, permissions []string) (map[string]bool, error) {
	// 1. Get user
	user, err := vm.userCollection.GetUserByID(tenantID, userID)
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
	user, err := vm.userCollection.GetUserByID(tenantID, userID)
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
