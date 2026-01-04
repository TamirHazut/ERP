package cache

import "time"

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	Token      string     `json:"token"`
	UserID     string     `json:"user_id"`
	Email      string     `json:"email"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	Verified   bool       `json:"verified"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}
