package cache

import "time"

// GraphQLQueryCache represents cached GraphQL query results
// Stored in: query_cache:{tenant_id}:{query_hash} (Redis String with JSON-encoded response)
type GraphQLQueryCache struct {
	QueryHash    string                 `json:"query_hash"` // SHA256 of query + variables
	TenantID     string                 `json:"tenant_id"`
	UserID       string                 `json:"user_id,omitempty"` // Optional - for user-specific caching
	Query        string                 `json:"query"` // The GraphQL query
	Variables    map[string]interface{} `json:"variables,omitempty"` // Query variables
	Response     interface{}            `json:"response"` // Cached response data
	CachedAt     time.Time              `json:"cached_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	HitCount     int                    `json:"hit_count"` // Number of times this cache was hit
	LastAccessAt time.Time              `json:"last_access_at"`
}

// PersistedQuery represents a persisted GraphQL query (APQ - Automatic Persisted Queries)
// Stored in: persisted_queries:{query_hash} (Redis String with query text)
// These are shared across all tenants for common queries
type PersistedQuery struct {
	QueryHash   string    `json:"query_hash"` // SHA256 hash of the query
	QueryText   string    `json:"query_text"` // The full GraphQL query text
	OperationName string  `json:"operation_name,omitempty"`
	StoredAt    time.Time `json:"stored_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	UseCount    int       `json:"use_count"` // How many times this was used
	Version     int       `json:"version"` // Version number for cache invalidation
}

// ResolverCache represents cached field resolver results
// Stored in: resolver_cache:{tenant_id}:{type}:{field}:{key} (Redis Hash)
// Used for caching expensive field resolvers (e.g., computed fields, external API calls)
type ResolverCache struct {
	CacheKey    string                 `json:"cache_key"` // Unique key for this resolver result
	TenantID    string                 `json:"tenant_id"`
	TypeName    string                 `json:"type_name"` // GraphQL type name (e.g., "User", "Product")
	FieldName   string                 `json:"field_name"` // Field name (e.g., "orders", "recommendations")
	ParentID    string                 `json:"parent_id"` // ID of parent object
	Arguments   map[string]interface{} `json:"arguments,omitempty"` // Field arguments
	Result      interface{}            `json:"result"` // Cached resolver result
	CachedAt    time.Time              `json:"cached_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Invalidated bool                   `json:"invalidated"` // Whether this cache entry was manually invalidated
}
