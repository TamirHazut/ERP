package redis

import (
	"time"

	mongo "erp.localhost/internal/db/mongo/models"
)

// ============================================================================
// REDIS MODELS
// ============================================================================

// Session represents a user session stored in Redis
// Key: sessions:{session_id}
// TTL: 24 hours (configurable)
type Session struct {
	SessionID    string            `json:"session_id"`
	UserID       string            `json:"user_id"`
	TenantID     string            `json:"tenant_id"`
	Email        string            `json:"email"`
	Username     string            `json:"username"`
	Roles        []string          `json:"roles"`
	Permissions  []string          `json:"permissions,omitempty"`
	ExpiresAt    time.Time         `json:"expires_at"`
	CreatedAt    time.Time         `json:"created_at"`
	LastActivity time.Time         `json:"last_activity"`
	IPAddress    string            `json:"ip_address"`
	UserAgent    string            `json:"user_agent"`
	DeviceInfo   DeviceInfo        `json:"device_info,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type DeviceInfo struct {
	DeviceID   string `json:"device_id,omitempty"`
	DeviceType string `json:"device_type,omitempty"` // mobile, desktop, tablet
	OS         string `json:"os,omitempty"`
	Browser    string `json:"browser,omitempty"`
}

// TokenMetadata represents token metadata stored in Redis
// Key: tokens:{token_id}
// TTL: Matches JWT expiry
type TokenMetadata struct {
	TokenID   string     `json:"token_id"`
	JTI       string     `json:"jti"` // JWT ID
	UserID    string     `json:"user_id"`
	TenantID  string     `json:"tenant_id"`
	TokenType string     `json:"token_type"` // access, refresh
	IssuedAt  time.Time  `json:"issued_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	Revoked   bool       `json:"revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	RevokedBy string     `json:"revoked_by,omitempty"`
	IPAddress string     `json:"ip_address"`
	UserAgent string     `json:"user_agent"`
	Scopes    []string   `json:"scopes,omitempty"`
}

// RefreshToken represents a refresh token
// Key: refresh_tokens:{user_id}
// TTL: 7 days (configurable)
type RefreshToken struct {
	Token      string     `json:"token"`
	UserID     string     `json:"user_id"`
	TenantID   string     `json:"tenant_id"`
	TokenID    string     `json:"token_id"`
	IssuedAt   time.Time  `json:"issued_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	LastUsed   time.Time  `json:"last_used"`
	DeviceInfo DeviceInfo `json:"device_info,omitempty"`
}

// RevokedToken represents a revoked token (blacklist)
// Key: revoked_tokens:{token_id} or blacklist:{jti}
// TTL: Until original token expiry
type RevokedToken struct {
	TokenID   string    `json:"token_id"`
	JTI       string    `json:"jti"`
	UserID    string    `json:"user_id"`
	RevokedAt time.Time `json:"revoked_at"`
	RevokedBy string    `json:"revoked_by"`
	Reason    string    `json:"reason,omitempty"`
	ExpiresAt time.Time `json:"expires_at"` // Original token expiry
}

// UserPermissionsCache represents cached permissions in Redis
// Key: permissions:{user_id}
// TTL: 5 minutes
type UserPermissionsCache struct {
	UserID      string    `json:"user_id"`
	TenantID    string    `json:"tenant_id"`
	Permissions []string  `json:"permissions"`
	CachedAt    time.Time `json:"cached_at"`
	Version     int       `json:"version"` // Increment to invalidate
}

// UserRolesCache represents cached user roles
// Key: roles:{user_id}
// TTL: 5 minutes
type UserRolesCache struct {
	UserID   string        `json:"user_id"`
	TenantID string        `json:"tenant_id"`
	Roles    []RoleSummary `json:"roles"`
	CachedAt time.Time     `json:"cached_at"`
	Version  int           `json:"version"`
}

type RoleSummary struct {
	RoleID   string `json:"role_id"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

// RolePermissionsCache represents cached role permissions
// Key: role_perms:{role_id}
// TTL: 10 minutes
type RolePermissionsCache struct {
	RoleID      string    `json:"role_id"`
	TenantID    string    `json:"tenant_id"`
	Permissions []string  `json:"permissions"`
	CachedAt    time.Time `json:"cached_at"`
	Version     int       `json:"version"`
}

// RateLimitInfo represents rate limit data in Redis
// Key: rate_limit:{user_id}:{endpoint}
// TTL: Window duration (e.g., 60 seconds)
type RateLimitInfo struct {
	Key          string    `json:"key"`
	UserID       string    `json:"user_id,omitempty"`
	TenantID     string    `json:"tenant_id,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	Endpoint     string    `json:"endpoint"`
	RequestCount int       `json:"request_count"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
	Limit        int       `json:"limit"`
	Remaining    int       `json:"remaining"`
	ResetAt      time.Time `json:"reset_at"`
}

// TenantRateLimit represents tenant-level rate limiting
// Key: tenant_limit:{tenant_id}:{endpoint}
type TenantRateLimit struct {
	TenantID     string    `json:"tenant_id"`
	Endpoint     string    `json:"endpoint"`
	RequestCount int       `json:"request_count"`
	Limit        int       `json:"limit"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
}

// IPRateLimit represents IP-based rate limiting
// Key: ip_limit:{ip_address}:{endpoint}
type IPRateLimit struct {
	IPAddress    string     `json:"ip_address"`
	Endpoint     string     `json:"endpoint"`
	RequestCount int        `json:"request_count"`
	Limit        int        `json:"limit"`
	WindowStart  time.Time  `json:"window_start"`
	Blocked      bool       `json:"blocked"`
	BlockedUntil *time.Time `json:"blocked_until,omitempty"`
}

// QueryCache represents a cached GraphQL query response
// Key: query_cache:{query_hash}
// TTL: 1-60 minutes (depends on query type)
type QueryCache struct {
	QueryHash string                 `json:"query_hash"`
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	Response  interface{}            `json:"response"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	CachedAt  time.Time              `json:"cached_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	HitCount  int                    `json:"hit_count"`
}

// UserCache represents cached user data
// Key: user_cache:{user_id}
// TTL: 10 minutes
type UserCache struct {
	UserID   string            `json:"user_id"`
	TenantID string            `json:"tenant_id"`
	Email    string            `json:"email"`
	Username string            `json:"username"`
	Profile  mongo.UserProfile `json:"profile"`
	Status   string            `json:"status"`
	Roles    []string          `json:"roles"`
	CachedAt time.Time         `json:"cached_at"`
}

// TenantCache represents cached tenant data
// Key: tenant_cache:{tenant_id}
// TTL: 30 minutes
type TenantCache struct {
	TenantID     string               `json:"tenant_id"`
	Name         string               `json:"name"`
	Status       string               `json:"status"`
	Subscription mongo.Subscription   `json:"subscription"`
	Settings     mongo.TenantSettings `json:"settings"`
	CachedAt     time.Time            `json:"cached_at"`
}

// ProductCache represents cached product data
// Key: product_cache:{product_id}
// TTL: 15 minutes
type ProductCache struct {
	ProductID string                 `json:"product_id"`
	TenantID  string                 `json:"tenant_id"`
	SKU       string                 `json:"sku"`
	Name      string                 `json:"name"`
	Price     float64                `json:"price"`
	Inventory mongo.ProductInventory `json:"inventory"`
	Status    string                 `json:"status"`
	CachedAt  time.Time              `json:"cached_at"`
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

// PasswordResetToken represents a password reset token
// Key: pwd_reset:{token}
// TTL: 1 hour
type PasswordResetToken struct {
	Token     string     `json:"token"`
	UserID    string     `json:"user_id"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	Used      bool       `json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

// EmailVerificationToken represents an email verification token
// Key: email_verify:{token}
// TTL: 24 hours
type EmailVerificationToken struct {
	Token      string     `json:"token"`
	UserID     string     `json:"user_id"`
	Email      string     `json:"email"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	Verified   bool       `json:"verified"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

// MFACode represents a temporary MFA code
// Key: mfa_code:{user_id}
// TTL: 5 minutes
type MFACode struct {
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	Method    string    `json:"method"` // totp, sms, email
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Attempts  int       `json:"attempts"`
	Verified  bool      `json:"verified"`
}

// InviteToken represents a user invitation token
// Key: invite:{token}
// TTL: 7 days
type InviteToken struct {
	Token      string     `json:"token"`
	Email      string     `json:"email"`
	TenantID   string     `json:"tenant_id"`
	RoleIDs    []string   `json:"role_ids"`
	InvitedBy  string     `json:"invited_by"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	Accepted   bool       `json:"accepted"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
}

// LoginAttempts tracks failed login attempts
// Key: login_attempts:{user_id} or login_attempts:{ip_address}
// TTL: 15 minutes
type LoginAttempts struct {
	Identifier   string     `json:"identifier"` // user_id or ip_address
	Attempts     int        `json:"attempts"`
	FirstAttempt time.Time  `json:"first_attempt"`
	LastAttempt  time.Time  `json:"last_attempt"`
	Locked       bool       `json:"locked"`
	LockedUntil  *time.Time `json:"locked_until,omitempty"`
	FailedIPs    []string   `json:"failed_ips,omitempty"`
}

// ActiveUser represents an active user in the system
// Stored in: active_users:{tenant_id} (Redis Set)
// Also in: online_users (Redis Sorted Set, score = last_activity timestamp)
type ActiveUser struct {
	UserID       string    `json:"user_id"`
	TenantID     string    `json:"tenant_id"`
	SessionID    string    `json:"session_id"`
	LastActivity time.Time `json:"last_activity"`
	Status       string    `json:"status"` // online, idle, away
}

// FeatureFlagCache represents a cached feature flag
// Key: feature_flag:{flag_key}
// TTL: 5 minutes
type FeatureFlagCache struct {
	FlagKey  string               `json:"flag_key"`
	Enabled  bool                 `json:"enabled"`
	Rollout  mongo.FeatureRollout `json:"rollout"`
	CachedAt time.Time            `json:"cached_at"`
	Version  int                  `json:"version"`
}

// TenantFeatures represents cached tenant-specific features
// Key: tenant_features:{tenant_id}
// TTL: 10 minutes
type TenantFeatures struct {
	TenantID string                   `json:"tenant_id"`
	Features map[string]bool          `json:"features"` // feature_key -> enabled
	Limits   mongo.SubscriptionLimits `json:"limits"`
	CachedAt time.Time                `json:"cached_at"`
}

// ServiceConfigCache represents cached service configuration
// Key: config:{service_name}:{environment}
// TTL: 15 minutes
type ServiceConfigCache struct {
	ServiceName string                 `json:"service_name"`
	Environment string                 `json:"environment"`
	Config      map[string]interface{} `json:"config"`
	Version     int                    `json:"version"`
	CachedAt    time.Time              `json:"cached_at"`
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
