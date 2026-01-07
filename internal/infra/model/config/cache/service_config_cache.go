package cache

import "time"

// ServiceConfigCache represents cached service configuration
// Stored in: service_config_cache:{service_name}:{environment}:{tenant_id} (Redis Hash)
// TTL: 30 minutes
type ServiceConfigCache struct {
	ConfigID    string                 `json:"config_id"`
	ServiceName string                 `json:"service_name"` // core, auth, gateway, event
	Environment string                 `json:"environment"`  // development, staging, production
	TenantID    string                 `json:"tenant_id,omitempty"`
	Config      map[string]interface{} `json:"config"`
	Version     int                    `json:"version"`
	CachedAt    time.Time              `json:"cached_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
}

// ConfigVersionCache tracks the latest version of each config
// Stored in: config_versions:{service_name}:{environment} (Redis Hash)
// Used for cache invalidation when config changes
type ConfigVersionCache struct {
	ServiceName    string         `json:"service_name"`
	Environment    string         `json:"environment"`
	LatestVersion  int            `json:"latest_version"`
	TenantVersions map[string]int `json:"tenant_versions,omitempty"` // tenant_id -> version
	UpdatedAt      time.Time      `json:"updated_at"`
}
