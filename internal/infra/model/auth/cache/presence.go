package cache

import "time"

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
