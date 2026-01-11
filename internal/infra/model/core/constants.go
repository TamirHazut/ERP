package core

import "strings"

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

// Order statuses
const (
	OrderStatusDraft     = "draft"
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
)

// Order types
const (
	OrderTypeSales    = "sales"
	OrderTypePurchase = "purchase"
	OrderTypeTransfer = "transfer"
)

// Payment statuses
const (
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusRefunded = "refunded"
	PaymentStatusFailed   = "failed"
)

// Product statuses
const (
	ProductStatusActive       = "active"
	ProductStatusInactive     = "inactive"
	ProductStatusDiscontinued = "discontinued"
)

// Vendor statuses
const (
	VendorStatusActive          = "active"
	VendorStatusInactive        = "inactive"
	VendorStatusPendingApproval = "pending_approval"
)
