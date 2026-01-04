package auth_cache_models

import "time"

// LoginAttempts tracks failed login attempts
type LoginAttempts struct {
	Identifier   string     `json:"identifier"` // user_id or ip_address
	Attempts     int        `json:"attempts"`
	FirstAttempt time.Time  `json:"first_attempt"`
	LastAttempt  time.Time  `json:"last_attempt"`
	Locked       bool       `json:"locked"`
	LockedUntil  *time.Time `json:"locked_until,omitempty"`
	FailedIPs    []string   `json:"failed_ips,omitempty"`
}
