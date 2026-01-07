package cache

import "time"

// RateLimitBucket represents a rate limit bucket for tracking API usage
// Stored in: rate_limit:{tenant_id}:{user_id}:{endpoint} (Redis String with counter)
// Also can use: rate_limit_bucket:{bucket_id} (Redis Hash) for more complex scenarios
type RateLimitBucket struct {
	BucketID    string    `json:"bucket_id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id,omitempty"` // Optional - for user-specific limits
	IPAddress   string    `json:"ip_address,omitempty"` // Optional - for IP-based limits
	Endpoint    string    `json:"endpoint"` // API endpoint path
	Method      string    `json:"method"` // HTTP method
	TokensUsed  int       `json:"tokens_used"` // Current token count
	TokensLimit int       `json:"tokens_limit"` // Maximum tokens allowed
	WindowStart time.Time `json:"window_start"` // When current window started
	WindowEnd   time.Time `json:"window_end"` // When current window ends
	ResetAt     time.Time `json:"reset_at"` // When bucket resets
	Blocked     bool      `json:"blocked"` // Whether this bucket is currently blocked
	BlockedAt   time.Time `json:"blocked_at,omitempty"` // When the block started
}

// RateLimitRule represents a rate limiting rule configuration
// Stored in: rate_limit_rules:{tenant_id}:{rule_id} (Redis Hash)
// Or cached from database: rate_limit_rule_cache:{rule_id}
type RateLimitRule struct {
	RuleID      string        `json:"rule_id"`
	TenantID    string        `json:"tenant_id"`
	Name        string        `json:"name"`
	Endpoint    string        `json:"endpoint"` // Can use wildcards: /api/v1/*
	Method      string        `json:"method"` // GET, POST, *, etc.
	Limit       int           `json:"limit"` // Number of requests allowed
	Window      time.Duration `json:"window"` // Time window (e.g., 1 minute, 1 hour)
	Scope       string        `json:"scope"` // "user", "ip", "tenant", "global"
	Priority    int           `json:"priority"` // Higher priority rules evaluated first
	Enabled     bool          `json:"enabled"`
	CachedAt    time.Time     `json:"cached_at"`
}

// RateLimitViolation represents a rate limit violation event
// Stored in: rate_limit_violations:{tenant_id}:{user_id} (Redis Sorted Set, score = timestamp)
type RateLimitViolation struct {
	ViolationID string    `json:"violation_id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id,omitempty"`
	IPAddress   string    `json:"ip_address"`
	Endpoint    string    `json:"endpoint"`
	Method      string    `json:"method"`
	RuleID      string    `json:"rule_id"` // Which rule was violated
	Timestamp   time.Time `json:"timestamp"`
	RequestID   string    `json:"request_id,omitempty"`
}
