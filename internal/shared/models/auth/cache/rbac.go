package auth_cache_models

import "time"

// UserPermissionsCache represents cached permissions in Redis
type UserPermissionsCache struct {
	UserID      string    `json:"user_id"`
	TenantID    string    `json:"tenant_id"`
	Permissions []string  `json:"permissions"`
	CachedAt    time.Time `json:"cached_at"`
	Version     int       `json:"version"` // Increment to invalidate
}

// UserRolesCache represents cached user roles
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
type RolePermissionsCache struct {
	RoleID      string    `json:"role_id"`
	TenantID    string    `json:"tenant_id"`
	Permissions []string  `json:"permissions"`
	CachedAt    time.Time `json:"cached_at"`
	Version     int       `json:"version"`
}
