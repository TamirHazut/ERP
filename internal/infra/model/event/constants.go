package event

import "strings"

/* Audit log */
// Categories
const (
	CategoryAuth        = "auth"
	CategoryUserMgmt    = "user_mgmt"
	CategoryRoleMgmt    = "role_mgmt"
	CategoryOrder       = "order"
	CategoryProduct     = "product"
	CategoryInventory   = "inventory"
	CategoryVendor      = "vendor"
	CategoryCustomer    = "customer"
	CategoryConfig      = "config"
	CategoryTenant      = "tenant"
	CategorySecurity    = "security"
	CategoryDataAccess  = "data_access"
	CategoryIntegration = "integration"
	CategoryAPI         = "api"
)

func IsValidCategory(category string) bool {
	if category == "" {
		return false
	}
	category = strings.ToLower(category)
	validCategories := map[string]bool{
		CategoryAuth:        true,
		CategoryUserMgmt:    true,
		CategoryRoleMgmt:    true,
		CategoryOrder:       true,
		CategoryProduct:     true,
		CategoryInventory:   true,
		CategoryVendor:      true,
		CategoryCustomer:    true,
		CategoryConfig:      true,
		CategoryTenant:      true,
		CategorySecurity:    true,
		CategoryDataAccess:  true,
		CategoryIntegration: true,
		CategoryAPI:         true,
	}

	return validCategories[category]
}

/* Audit logs Actions */
// Auth Actions
const (
	ActionLogin           = "login"
	ActionLogout          = "logout"
	ActionLogoutAll       = "logout_all"
	ActionTokenRefresh    = "token_refresh"
	ActionTokenRevoke     = "token_revoke"
	ActionPasswordChanged = "password_changed"
	ActionPasswordReset   = "password_reset"
	ActionForcedLogout    = "forced_logout"
)

// User Management Actions
const (
	ActionUserCreated   = "user_created"
	ActionUserUpdated   = "user_updated"
	ActionUserDeleted   = "user_deleted"
	ActionUserSuspended = "user_suspended"
	ActionUserActivated = "user_activated"
	ActionUserLocked    = "user_locked"
	ActionUserUnlocked  = "user_unlocked"
)

// Role Management Actions
const (
	ActionRoleCreated        = "role_created"
	ActionRoleUpdated        = "role_updated"
	ActionRoleDeleted        = "role_deleted"
	ActionRoleAssigned       = "role_assigned"
	ActionRoleRevoked        = "role_revoked"
	ActionPermissionsAdded   = "permissions_added"
	ActionPermissionsRemoved = "permissions_removed"
)

// Order Actions
const (
	ActionOrderCreated   = "order_created"
	ActionOrderUpdated   = "order_updated"
	ActionOrderCancelled = "order_cancelled"
	ActionOrderFulfilled = "order_fulfilled"
	ActionOrderShipped   = "order_shipped"
	ActionOrderDelivered = "order_delivered"
	ActionOrderReturned  = "order_returned"
	ActionOrderRefunded  = "order_refunded"
)

// Product/Inventory Actions
const (
	ActionProductCreated   = "product_created"
	ActionProductUpdated   = "product_updated"
	ActionProductDeleted   = "product_deleted"
	ActionStockAdjusted    = "stock_adjusted"
	ActionStockTransferred = "stock_transferred"
	ActionPriceChanged     = "price_changed"
)

// Vendor/Customer Actions
const (
	ActionVendorCreated   = "vendor_created"
	ActionVendorUpdated   = "vendor_updated"
	ActionCustomerCreated = "customer_created"
	ActionCustomerUpdated = "customer_updated"
)

// Config Actions
const (
	ActionConfigCreated       = "config_created"
	ActionConfigUpdated       = "config_updated"
	ActionConfigDeleted       = "config_deleted"
	ActionFeatureFlagEnabled  = "feature_flag_enabled"
	ActionFeatureFlagDisabled = "feature_flag_disabled"
)

// Tenant Actions
const (
	ActionTenantCreated   = "tenant_created"
	ActionTenantUpdated   = "tenant_updated"
	ActionTenantSuspended = "tenant_suspended"
	ActionTenantActivated = "tenant_activated"
)

// Security Actions
const (
	ActionBruteForceDetected  = "brute_force_detected"
	ActionSuspiciousActivity  = "suspicious_activity"
	ActionUnauthorizedAccess  = "unauthorized_access"
	ActionTokenTheftSuspected = "token_theft_suspected"
	ActionMassDataExport      = "mass_data_export"
)

// Data Access Actions (GDPR/Compliance)
const (
	ActionPIIViewed          = "pii_viewed"
	ActionPIIExported        = "pii_exported"
	ActionPIIDeleted         = "pii_deleted"
	ActionGDPRDataExport     = "gdpr_data_export"
	ActionRightToBeForgotten = "right_to_be_forgotten"
)

func IsValidAuditAction(action string) bool {
	if action == "" {
		return false
	}
	action = strings.ToLower(action)
	validActions := map[string]bool{
		// ActionSystemWildcard:      true,
		ActionLogin:               true,
		ActionLogout:              true,
		ActionLogoutAll:           true,
		ActionTokenRefresh:        true,
		ActionPasswordChanged:     true,
		ActionPasswordReset:       true,
		ActionForcedLogout:        true,
		ActionUserCreated:         true,
		ActionUserUpdated:         true,
		ActionUserDeleted:         true,
		ActionUserSuspended:       true,
		ActionUserActivated:       true,
		ActionUserLocked:          true,
		ActionUserUnlocked:        true,
		ActionRoleCreated:         true,
		ActionRoleUpdated:         true,
		ActionRoleDeleted:         true,
		ActionRoleAssigned:        true,
		ActionRoleRevoked:         true,
		ActionPermissionsAdded:    true,
		ActionPermissionsRemoved:  true,
		ActionOrderCreated:        true,
		ActionOrderUpdated:        true,
		ActionOrderCancelled:      true,
		ActionOrderFulfilled:      true,
		ActionOrderShipped:        true,
		ActionOrderDelivered:      true,
		ActionOrderReturned:       true,
		ActionOrderRefunded:       true,
		ActionProductCreated:      true,
		ActionProductUpdated:      true,
		ActionProductDeleted:      true,
		ActionStockAdjusted:       true,
		ActionStockTransferred:    true,
		ActionPriceChanged:        true,
		ActionVendorCreated:       true,
		ActionVendorUpdated:       true,
		ActionCustomerCreated:     true,
		ActionCustomerUpdated:     true,
		ActionConfigCreated:       true,
		ActionConfigUpdated:       true,
		ActionConfigDeleted:       true,
		ActionFeatureFlagEnabled:  true,
		ActionFeatureFlagDisabled: true,
		ActionTenantCreated:       true,
		ActionTenantUpdated:       true,
		ActionTenantSuspended:     true,
		ActionTenantActivated:     true,
		ActionBruteForceDetected:  true,
		ActionSuspiciousActivity:  true,
		ActionUnauthorizedAccess:  true,
		ActionTokenTheftSuspected: true,
		ActionMassDataExport:      true,
		ActionPIIViewed:           true,
		ActionPIIExported:         true,
		ActionPIIDeleted:          true,
		ActionGDPRDataExport:      true,
		ActionRightToBeForgotten:  true,
	}

	return validActions[action]
}

// Actor Types
const (
	ActorTypeUser   = "user"
	ActorTypeSystem = "system"
	ActorTypeAPIKey = "api_key"
	ActorTypeCron   = "cron"
)

func IsValidActorType(actorType string) bool {
	if actorType == "" {
		return false
	}
	actorType = strings.ToLower(actorType)
	validActorTypes := map[string]bool{
		ActorTypeUser:   true,
		ActorTypeSystem: true,
		ActorTypeAPIKey: true,
		ActorTypeCron:   true,
	}

	return validActorTypes[actorType]
}

// Severities
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

func IsValidSeverity(severity string) bool {
	if severity == "" {
		return false
	}
	severity = strings.ToLower(severity)
	validSeverities := map[string]bool{
		SeverityInfo:     true,
		SeverityWarning:  true,
		SeverityError:    true,
		SeverityCritical: true,
	}

	return validSeverities[severity]
}

// Results
const (
	ResultSuccess = "success"
	ResultFailure = "failure"
	ResultPartial = "partial"
)

func IsValidResult(result string) bool {
	if result == "" {
		return false
	}
	result = strings.ToLower(result)
	validResults := map[string]bool{
		ResultSuccess: true,
		ResultFailure: true,
		ResultPartial: true,
	}

	return validResults[result]
}

// Target Types
const (
	TargetTypeUser       = "user"
	TargetTypeRole       = "role"
	TargetTypePermission = "permission"
	TargetTypeOrder      = "order"
	TargetTypeProduct    = "product"
	TargetTypeVendor     = "vendor"
	TargetTypeCustomer   = "customer"
	TargetTypeConfig     = "config"
	TargetTypeTenant     = "tenant"
	TargetTypeSession    = "session"
	TargetTypeToken      = "token"
)

func IsValidTargetType(targetType string) bool {
	if targetType == "" {
		return false
	}
	targetType = strings.ToLower(targetType)
	validTargetTypes := map[string]bool{
		TargetTypeUser:       true,
		TargetTypeRole:       true,
		TargetTypePermission: true,
		TargetTypeOrder:      true,
		TargetTypeProduct:    true,
		TargetTypeVendor:     true,
		TargetTypeCustomer:   true,
		TargetTypeConfig:     true,
		TargetTypeTenant:     true,
		TargetTypeSession:    true,
		TargetTypeToken:      true,
	}

	return validTargetTypes[targetType]
}
