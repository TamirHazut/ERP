package db

import "time"

const (
	SystemTenant          = "system"
	SystemAdminUser       = "SystemAdmin"
	SystemAdminEmail      = "system@system.com"
	SystemAdminPassword   = "ERP@SystemAdmin.Secret5"
	TenantAdminUser       = "admin"
	TenantAdminRole       = "admin"
	TenantAdminPermission = "*:*"
	TenantAdminPassword   = "admin"
)

var (
	SystemTenantID          = ""
	SystemAdminUserID       = ""
	SystemAdminRoleID       = ""
	SystemAdminPermissionID = ""
)

// ============================================================================
// HELPER STRUCTS AND TYPES
// ============================================================================

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Skip     int `json:"-"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

// FilterParams represents infra filter parameters
type FilterParams struct {
	TenantId  string                 `json:"tenant_id"`
	Status    string                 `json:"status,omitempty"`
	StartDate *time.Time             `json:"start_date,omitempty"`
	EndDate   *time.Time             `json:"end_date,omitempty"`
	Search    string                 `json:"search,omitempty"`
	SortBy    string                 `json:"sort_by,omitempty"`
	SortOrder string                 `json:"sort_order,omitempty"` // asc, desc
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
