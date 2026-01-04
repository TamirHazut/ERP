package auth_cache_models

import "time"

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
