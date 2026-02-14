package rbac

import (
	"errors"
	"strings"

	"erp.localhost/auth/handler"
	"erp.localhost/infra/db/redis"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_auth "erp.localhost/infra/model/auth"
	authv1 "erp.localhost/infra/model/auth/v1"
)

type VerificationManager struct {
	userHandler      *handler.UserHandler
	roleHandler      *handler.RoleHandler
	tenantHandler    *handler.TenantHandler
	systemKeyHandler *redis.SystemKeyHandler
	logger           logger.Logger
}

// NewVerificationManager creates a new VerificationManager instance
func NewVerificationManager(
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
	tenantHandler *handler.TenantHandler,
	logger logger.Logger,
) (*VerificationManager, *infra_error.AppError) {
	systemHandler, err := redis.NewSystemKeyHandler(logger)
	if err != nil {
		return nil, err
	}
	return &VerificationManager{
		userHandler:      userHandler,
		roleHandler:      roleHandler,
		tenantHandler:    tenantHandler,
		systemKeyHandler: systemHandler,
		logger:           logger,
	}, nil
}

// GetUserPermissions resolves all effective permission strings for a user.
// Roles are fetched via a single aggregation query; AdditionalPermissions
// and RevokedPermissions are read directly from the user document as permission strings.
func (vm *VerificationManager) GetUserPermissions(tenantID, userID string) (map[string]bool, *infra_error.AppError) {
	// Single aggregation: fetch roles with name + permissions fields
	roles, err := vm.roleHandler.GetUserRolesAggregation(tenantID, userID, []string{"name", "permissions"})
	if err != nil {
		vm.logger.Error("role aggregation failed in GetUserPermissions", "error", err)
		return nil, err
	}

	// Admin short-circuit: tenant_admin / system_admin → full wildcard
	for _, role := range roles {
		if role.Name == model_auth.RoleTenantAdmin || role.Name == model_auth.RoleSystemAdmin {
			return map[string]bool{model_auth.PermissionAdminString: true}, nil
		}
	}

	// Populate permission map from role permission strings
	userPermissions := make(map[string]bool)
	for _, role := range roles {
		for _, perm := range role.Permissions {
			userPermissions[perm] = true
		}
	}

	// Fetch user for additional / revoked permission strings
	user, err := vm.userHandler.GetUserByID(tenantID, userID)
	if err != nil {
		return nil, err
	}

	for _, perm := range user.AdditionalPermissions {
		userPermissions[perm] = true
	}
	for _, perm := range user.RevokedPermissions {
		userPermissions[perm] = false
	}

	return userPermissions, nil
}

/* System default checks */

func (vm *VerificationManager) IsSystemTenantUser(tenantID string) bool {
	systemTenantID, err := vm.systemKeyHandler.GetOne("tenant")
	if err != nil {
		return false
	}
	return tenantID == *systemTenantID
}

func (vm *VerificationManager) IsSystemAdminUser(userID string) bool {
	systemAdminUserID, err := vm.systemKeyHandler.GetOne("user")
	if err != nil {
		return false
	}
	return userID == *systemAdminUserID
}

func (vm *VerificationManager) IsSystemRole(roleID string) bool {
	systemRoleID, err := vm.systemKeyHandler.GetOne("role")
	if err != nil {
		return false
	}
	return roleID == *systemRoleID
}

// isTenantAdmin checks whether the user holds a tenant_admin or system_admin role.
func (vm *VerificationManager) isTenantAdmin(user *authv1.User) bool {
	roles, err := vm.roleHandler.GetUserRolesAggregation(user.TenantId, user.Id, []string{"name"})
	if err != nil {
		vm.logger.Warn("role aggregation failed in isTenantAdmin", "error", err)
		return false
	}
	for _, role := range roles {
		if role.Name == model_auth.RoleTenantAdmin || role.Name == model_auth.RoleSystemAdmin {
			return true
		}
	}
	return false
}

// GetUserRoles returns all role IDs assigned to a user
func (vm *VerificationManager) GetUserRoles(tenantID, userID string) ([]string, *infra_error.AppError) {
	user, err := vm.userHandler.GetUserByID(tenantID, userID)
	if err != nil {
		vm.logger.Error(err.Error())
		return nil, err
	}

	roleIDs := make([]string, 0, len(user.Roles))
	for _, userRole := range user.Roles {
		roleIDs = append(roleIDs, userRole.RoleId)
	}
	return roleIDs, nil
}

// checkPermissionsInternal performs the core permission checking logic.
// This is the single source of truth for all permission checks.
//
// Parameters:
//   - tenantID: The tenant ID of the requesting user
//   - userID: The user ID making the request
//   - permissions: List of permissions to check (e.g., "user:read", "role:write")
//   - targetTenantID: Optional target tenant for cross-tenant operations. If nil, defaults to tenantID
//
// Returns a map of permission → granted status (true/false)
func (vm *VerificationManager) checkPermissionsInternal(
	tenantID, userID string,
	permissions []string,
	targetTenantID *string,
) (map[string]bool, *infra_error.AppError) {
	effectiveTargetTenant := tenantID
	if targetTenantID != nil {
		effectiveTargetTenant = *targetTenantID
	}

	for _, permission := range permissions {
		if !model_auth.IsValidPermission(permission) {
			return nil, infra_error.Validation(infra_error.ValidationInvalidValue)
		}
	}

	// 1. Get user from database
	user, err := vm.userHandler.GetUserByID(tenantID, userID)
	if err != nil {
		vm.logger.Error(err.Error())
		return nil, err
	}

	if user.Status != authv1.UserStatus_USER_STATUS_ACTIVE {
		return nil, infra_error.Auth(infra_error.AuthAccountDisabled)
	}

	// 2. Validate target tenant exists
	if targetTenantID != nil {
		if err := vm.VerifyTenant(effectiveTargetTenant); err != nil {
			return nil, err
		}
	}

	// 3. Tenant admin for same tenant → grant all
	if tenantID == effectiveTargetTenant {
		if vm.isTenantAdmin(user) {
			result := make(map[string]bool)
			for _, perm := range permissions {
				result[perm] = true
			}
			return result, nil
		}
	}

	// 4. System tenant user performing cross-tenant operation
	if vm.IsSystemTenantUser(tenantID) && tenantID != effectiveTargetTenant {
		userPermissions, err := vm.GetUserPermissions(tenantID, userID)
		if err != nil {
			return nil, err
		}
		result := make(map[string]bool)
		for _, perm := range permissions {
			result[perm] = hasPermissionGranted(userPermissions, perm)
		}
		return result, nil
	}

	// 5. Non-system users cannot operate cross-tenant
	if tenantID != effectiveTargetTenant {
		return nil, infra_error.Auth(infra_error.AuthTenantAccessDenied)
		// result := make(map[string]bool)
		// for _, perm := range permissions {
		// 	result[perm] = false
		// }
		// return result, nil
	}

	// 6. Regular same-tenant permission check
	userPermissions, err := vm.GetUserPermissions(tenantID, userID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for _, perm := range permissions {
		result[perm] = hasPermissionGranted(userPermissions, perm)
	}
	return result, nil
}

// CheckPermissions checks if a user has the specified permissions.
// This method assumes same-tenant operations (targetTenantID = tenantID).
func (vm *VerificationManager) CheckPermissions(tenantID, userID string, permissions []string) (map[string]bool, *infra_error.AppError) {
	return vm.checkPermissionsInternal(tenantID, userID, permissions, nil)
}

// HasPermission checks if a user has a single permission for a target tenant.
// Returns nil if permission is granted, error if denied.
func (vm *VerificationManager) HasPermission(tenantID, userID, permission string, targetTenantID string) *infra_error.AppError {
	result, err := vm.checkPermissionsInternal(tenantID, userID, []string{permission}, &targetTenantID)
	if err != nil {
		return err
	}
	if !result[permission] {
		return infra_error.Auth(infra_error.AuthPermissionDenied)
	}
	return nil
}

func (vm *VerificationManager) VerifyTenant(tenantID string) *infra_error.AppError {
	_, err := vm.tenantHandler.GetTenantByID(tenantID)
	if err != nil {
		return infra_error.NotFound(infra_error.NotFoundTenant, model_auth.ResourceTypeTenant, tenantID)
	}
	return nil
}

func (vm *VerificationManager) VerifyRoleNotAssigned(tenantID, roleID string) *infra_error.AppError {
	users, err := vm.userHandler.GetUsersByRoleID(tenantID, roleID)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return infra_error.Auth(infra_error.AuthPermissionDenied).WithError(errors.New("role is assigned to one or more users"))
	}
	return nil
}

func (vm *VerificationManager) VerifyRolesNotAssigned(tenantID string) *infra_error.AppError {
	users, err := vm.userHandler.GetUsersWithRoles(tenantID)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return infra_error.Auth(infra_error.AuthPermissionDenied).WithError(errors.New("one or more roles are assigned to one or more users"))
	}
	return nil
}

// hasPermissionGranted checks whether a permission is granted by the user's effective
// permission map, taking wildcards into account.
//
// Three wildcard levels are checked in addition to the exact permission:
//   - *:* — full admin, covers everything
//   - *:action — covers all resources for that action
//   - resource:* — covers all actions for that resource
func hasPermissionGranted(userPermissions map[string]bool, permission string) bool {
	if userPermissions[permission] {
		return true
	}
	parts := strings.SplitN(permission, ":", 2)
	if len(parts) == 2 {
		if userPermissions["*:"+parts[1]] { // *:action
			return true
		}
		if userPermissions[parts[0]+":*"] { // resource:*
			return true
		}
	}
	return userPermissions[model_auth.PermissionAdminString] // *:*
}
