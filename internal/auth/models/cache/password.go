package cache

import "time"

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	Token     string     `json:"token"`
	UserID    string     `json:"user_id"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	Used      bool       `json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}
