package cache

import "time"

// TokenMetadata represents token metadata stored in Redis
type TokenMetadata struct {
	TokenID   string     `json:"token_id"`
	JTI       string     `json:"jti"` // JWT ID
	UserID    string     `json:"user_id"`
	TenantID  string     `json:"tenant_id"`
	TokenType string     `json:"token_type"` // access, refresh
	IssuedAt  time.Time  `json:"issued_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	Revoked   bool       `json:"revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	RevokedBy string     `json:"revoked_by,omitempty"`
	IPAddress string     `json:"ip_address"`
	UserAgent string     `json:"user_agent"`
	Scopes    []string   `json:"scopes,omitempty"`
}

// RevokedToken represents a revoked token (blacklist)
type RevokedToken struct {
	TokenID   string    `json:"token_id"`
	JTI       string    `json:"jti"`
	UserID    string    `json:"user_id"`
	RevokedAt time.Time `json:"revoked_at"`
	RevokedBy string    `json:"revoked_by"`
	Reason    string    `json:"reason,omitempty"`
	ExpiresAt time.Time `json:"expires_at"` // Original token expiry
}
