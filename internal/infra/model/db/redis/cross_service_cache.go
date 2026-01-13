package redis

import (
	"time"

	model_auth "erp.localhost/internal/infra/model/auth"
	model_core "erp.localhost/internal/infra/model/core"
)

// ============================================================================
// CROSS-SERVICE CACHE MODELS
// These are cache models used by multiple services or for service-to-service caching
// ============================================================================

// UserCache represents cached user data
// Key: user_cache:{user_id}
// TTL: 10 minutes
// Used by: Gateway (for auth checks), Core (for user lookups), Events (for notifications)
type UserCache struct {
	UserID   string                 `json:"user_id"`
	TenantID string                 `json:"tenant_id"`
	Email    string                 `json:"email"`
	Username string                 `json:"username"`
	Profile  model_auth.UserProfile `json:"profile"`
	Status   string                 `json:"status"`
	Roles    []string               `json:"roles"`
	CachedAt time.Time              `json:"cached_at"`
}

// TenantCache represents cached tenant data
// Key: tenant_cache:{tenant_id}
// TTL: 30 minutes
// Used by: All services for tenant validation and settings
type TenantCache struct {
	TenantID     string                    `json:"tenant_id"`
	Name         string                    `json:"name"`
	Status       string                    `json:"status"`
	Subscription model_auth.Subscription   `json:"subscription"`
	Settings     model_auth.TenantSettings `json:"settings"`
	CachedAt     time.Time                 `json:"cached_at"`
}

// ProductCache represents cached product data
// Key: product_cache:{product_id}
// TTL: 15 minutes
// Used by: Core (product lookups), Gateway (GraphQL resolvers), Events (inventory updates)
type ProductCache struct {
	ProductID string                      `json:"product_id"`
	TenantID  string                      `json:"tenant_id"`
	SKU       string                      `json:"sku"`
	Name      string                      `json:"name"`
	Price     float64                     `json:"price"`
	Inventory model_core.ProductInventory `json:"inventory"`
	Status    string                      `json:"status"`
	CachedAt  time.Time                   `json:"cached_at"`
}

// OrderCache represents cached order data
// Key: order_cache:{order_id}
// TTL: 5 minutes
// Used by: Core (order lookups), Gateway (GraphQL resolvers), Events (order status updates)
type OrderCache struct {
	OrderID     string    `json:"order_id"`
	TenantID    string    `json:"tenant_id"`
	OrderNumber string    `json:"order_number"`
	CustomerID  string    `json:"customer_id"`
	Status      string    `json:"status"`
	Total       float64   `json:"total"`
	CachedAt    time.Time `json:"cached_at"`
}
