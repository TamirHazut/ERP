package cache

import "time"

// QueryCache represents a cached GraphQL query response
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
