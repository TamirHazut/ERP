package core_models

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
