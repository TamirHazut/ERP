package models

// User statuses
const (
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
	UserStatusInvited   = "invited"
)

// Tenant statuses
const (
	TenantStatusActive    = "active"
	TenantStatusSuspended = "suspended"
	TenantStatusInactive  = "inactive"
	TenantStatusTrial     = "trial"
)

// Permission formats
const (
	PermissionWildcard = "*:*"
	PermissionFormat   = "resource:action[:scope]"
)

// Role types
const (
	RoleSystemAdmin = "system_admin"
	RoleTenantAdmin = "tenant_admin"
)
