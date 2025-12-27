package cache

import "time"

// RateLimitInfo represents rate limit data in Redis
type RateLimitInfo struct {
	Key          string    `json:"key"`
	UserID       string    `json:"user_id,omitempty"`
	TenantID     string    `json:"tenant_id,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	Endpoint     string    `json:"endpoint"`
	RequestCount int       `json:"request_count"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
	Limit        int       `json:"limit"`
	Remaining    int       `json:"remaining"`
	ResetAt      time.Time `json:"reset_at"`
}

// TenantRateLimit represents tenant-level rate limiting
type TenantRateLimit struct {
	TenantID     string    `json:"tenant_id"`
	Endpoint     string    `json:"endpoint"`
	RequestCount int       `json:"request_count"`
	Limit        int       `json:"limit"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
}

// IPRateLimit represents IP-based rate limiting
type IPRateLimit struct {
	IPAddress    string     `json:"ip_address"`
	Endpoint     string     `json:"endpoint"`
	RequestCount int        `json:"request_count"`
	Limit        int        `json:"limit"`
	WindowStart  time.Time  `json:"window_start"`
	Blocked      bool       `json:"blocked"`
	BlockedUntil *time.Time `json:"blocked_until,omitempty"`
}
