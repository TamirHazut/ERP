package auth_cache_models

import "time"

// InviteToken represents a user invitation token
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
