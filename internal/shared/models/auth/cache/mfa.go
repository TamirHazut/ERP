package auth_cache_models

import "time"

// MFACode represents a temporary MFA code
type MFACode struct {
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	Method    string    `json:"method"` // totp, sms, email
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Attempts  int       `json:"attempts"`
	Verified  bool      `json:"verified"`
}
