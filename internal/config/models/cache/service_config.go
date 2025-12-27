package cache

import "time"

// ServiceConfigCache represents cached service configuration
type ServiceConfigCache struct {
	ServiceName string                 `json:"service_name"`
	Environment string                 `json:"environment"`
	Config      map[string]interface{} `json:"config"`
	Version     int                    `json:"version"`
	CachedAt    time.Time              `json:"cached_at"`
}
