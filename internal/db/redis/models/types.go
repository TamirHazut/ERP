package redis

import "time"

// ============================================================================
// REDIS INFRASTRUCTURE TYPES
// ============================================================================

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
