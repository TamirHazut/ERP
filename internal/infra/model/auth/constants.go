package auth

import (
	"fmt"
	"strings"

	infra_error "erp.localhost/internal/infra/error"
)

/* User */
// User statuses
const (
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
	UserStatusInvited   = "invited"
)

func IsValidUserStatus(userStatus string) bool {
	if userStatus == "" {
		return false
	}
	userStatus = strings.ToLower(userStatus)
	validUserStatuses := map[string]bool{
		UserStatusActive:    true,
		UserStatusInactive:  true,
		UserStatusSuspended: true,
		UserStatusInvited:   true,
	}

	return validUserStatuses[userStatus]
}

/* Tenant */
// System tenant ID for cross-tenant operations
const (
	SystemTenantID = "system"
)

// Tenant statuses
const (
	TenantStatusActive    = "active"
	TenantStatusSuspended = "suspended"
	TenantStatusInactive  = "inactive"
	TenantStatusTrial     = "trial"
)

func IsValidTenantStatus(tenantStatus string) bool {
	if tenantStatus == "" {
		return false
	}
	tenantStatus = strings.ToLower(tenantStatus)
	validTenantStatuses := map[string]bool{
		TenantStatusActive:    true,
		TenantStatusSuspended: true,
		TenantStatusInactive:  true,
		TenantStatusTrial:     true,
	}

	return validTenantStatuses[tenantStatus]
}

/* RBAC */

func CreatePermissionString(resource string, action string) (string, error) {
	resource = strings.ToLower(resource)
	if !IsValidResourceType(resource) {
		return "", infra_error.Validation(infra_error.ValidationInvalidType, "resource")
	}
	action = strings.ToLower(action)
	if !IsValidPermissionAction(action) {
		return "", infra_error.Validation(infra_error.ValidationInvalidType, "action")
	}

	return fmt.Sprintf("%s:%s", resource, action), nil
}

// Permission status
const (
	PermissionStatusActive   = "active"
	PermissionStatusInactive = "inactive"
)

func IsValidPermissionStatus(permissionStatus string) bool {
	if permissionStatus == "" {
		return false
	}
	permissionStatus = strings.ToLower(permissionStatus)
	validPermissionStatus := map[string]bool{
		PermissionStatusActive:   true,
		PermissionStatusInactive: true,
	}
	return validPermissionStatus[permissionStatus]
}

// Permission formats
const (
	PermissionFormat = "[resource]:[action]"
)

func IsValidPermissionFormat(permissionFormat string) bool {
	if permissionFormat == "" {
		return false
	}
	permissionFormat = strings.ToLower(permissionFormat)

	permissionBreakDown := strings.Split(permissionFormat, ":")

	if len(permissionBreakDown) != 2 ||
		!IsValidResourceType(permissionBreakDown[0]) ||
		!IsValidPermissionAction(permissionBreakDown[1]) {
		return false
	}
	return true
}

// Permission actions
const (
	PermissionActionAll              = "*"
	PermissionActionCreate           = "create"
	PermissionActionRead             = "read"
	PermissionActionUpdate           = "update"
	PermissionActionDelete           = "delete"
	PermissionActionModifyPermission = "permission"
	PermissionActionModifyRole       = "role"
)

func IsValidPermissionAction(permissionAction string) bool {
	if permissionAction == "" {
		return false
	}
	permissionAction = strings.ToLower(permissionAction)
	validPermissionActions := map[string]bool{
		PermissionActionCreate:           true,
		PermissionActionRead:             true,
		PermissionActionUpdate:           true,
		PermissionActionDelete:           true,
		PermissionActionModifyPermission: true,
		PermissionActionModifyRole:       true,
	}
	return validPermissionActions[permissionAction]
}

// Role types
const (
	RoleStatusActive   = "active"
	RoleStatusInactive = "inactive"
)

func IsValidRoleStatus(roleStatus string) bool {
	if roleStatus == "" {
		return false
	}
	roleStatus = strings.ToLower(roleStatus)
	validRoleStatus := map[string]bool{
		RoleStatusActive:   true,
		RoleStatusInactive: true,
	}
	return validRoleStatus[roleStatus]
}

const (
	RoleSystemAdmin = "system_admin"
	RoleTenantAdmin = "tenant_admin"
)

func IsValidRoleType(roleType string) bool {
	if roleType == "" {
		return false
	}
	roleType = strings.ToLower(roleType)
	validRoleTypes := map[string]bool{
		RoleSystemAdmin: true,
		RoleTenantAdmin: true,
	}

	return validRoleTypes[roleType]
}

// Resource types
const (
	ResourceTypeAll        = "*"
	ResourceTypeUser       = "user"
	ResourceTypeRole       = "role"
	ResourceTypePermission = "permission"
	ResourceTypeOrder      = "order"
	ResourceTypeProduct    = "product"
	ResourceTypeVendor     = "vendor"
	ResourceTypeCustomer   = "customer"
	ResourceTypeConfig     = "config"
	ResourceTypeTenant     = "tenant"
	ResourceTypeToken      = "token"
)

func IsValidResourceType(resourceType string) bool {
	if resourceType == "" {
		return false
	}
	resourceType = strings.ToLower(resourceType)
	validResourceTypes := map[string]bool{
		ResourceTypeAll:        true,
		ResourceTypeUser:       true,
		ResourceTypeRole:       true,
		ResourceTypePermission: true,
		ResourceTypeOrder:      true,
		ResourceTypeProduct:    true,
		ResourceTypeVendor:     true,
		ResourceTypeCustomer:   true,
		ResourceTypeConfig:     true,
		ResourceTypeTenant:     true,
		ResourceTypeToken:      true,
	}

	return validResourceTypes[resourceType]
}
