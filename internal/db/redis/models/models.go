package redis

import (
	"time"

	auth "erp.localhost/internal/auth/models"
	core "erp.localhost/internal/core/models"
)

// ============================================================================
// REDIS MODELS
// ============================================================================

// UserCache represents cached user data
// Key: user_cache:{user_id}
// TTL: 10 minutes
type UserCache struct {
	UserID   string           `json:"user_id"`
	TenantID string           `json:"tenant_id"`
	Email    string           `json:"email"`
	Username string           `json:"username"`
	Profile  auth.UserProfile `json:"profile"`
	Status   string           `json:"status"`
	Roles    []string         `json:"roles"`
	CachedAt time.Time        `json:"cached_at"`
}

// TenantCache represents cached tenant data
// Key: tenant_cache:{tenant_id}
// TTL: 30 minutes
type TenantCache struct {
	TenantID     string              `json:"tenant_id"`
	Name         string              `json:"name"`
	Status       string              `json:"status"`
	Subscription auth.Subscription   `json:"subscription"`
	Settings     auth.TenantSettings `json:"settings"`
	CachedAt     time.Time           `json:"cached_at"`
}

// ProductCache represents cached product data
// Key: product_cache:{product_id}
// TTL: 15 minutes
type ProductCache struct {
	ProductID string                `json:"product_id"`
	TenantID  string                `json:"tenant_id"`
	SKU       string                `json:"sku"`
	Name      string                `json:"name"`
	Price     float64               `json:"price"`
	Inventory core.ProductInventory `json:"inventory"`
	Status    string                `json:"status"`
	CachedAt  time.Time             `json:"cached_at"`
}

// OrderCache represents cached order data
// Key: order_cache:{order_id}
// TTL: 5 minutes
type OrderCache struct {
	OrderID     string    `json:"order_id"`
	TenantID    string    `json:"tenant_id"`
	OrderNumber string    `json:"order_number"`
	CustomerID  string    `json:"customer_id"`
	Status      string    `json:"status"`
	Total       float64   `json:"total"`
	CachedAt    time.Time `json:"cached_at"`
}

// DistributedLock represents a distributed lock
// Key: lock:{resource_id}
// TTL: 30 seconds (should be short)
type DistributedLock struct {
	ResourceID string            `json:"resource_id"`
	LockID     string            `json:"lock_id"`  // Unique identifier for this lock instance
	OwnerID    string            `json:"owner_id"` // Service/process that owns the lock
	AcquiredAt time.Time         `json:"acquired_at"`
	ExpiresAt  time.Time         `json:"expires_at"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// ============================================================================
// REDIS HELPER TYPES
// ============================================================================

// RedisKeyOptions represents options for building Redis keys
type RedisKeyOptions struct {
	Prefix    string
	Separator string
	TTL       time.Duration
}

// CacheEntry is a generic cache entry wrapper
type CacheEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	TTL       int         `json:"ttl"` // seconds
	CachedAt  time.Time   `json:"cached_at"`
	ExpiresAt time.Time   `json:"expires_at"`
	Version   int         `json:"version,omitempty"`
	Tags      []string    `json:"tags,omitempty"` // For cache invalidation
}
