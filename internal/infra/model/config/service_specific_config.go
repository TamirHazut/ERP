package config

import "time"

// CoreServiceConfig represents specific config for Core service
type CoreServiceConfig struct {
	MaxOrderItems      int    `json:"max_order_items"`
	OrderNumberPrefix  string `json:"order_number_prefix"`
	DefaultCurrency    string `json:"default_currency"`
	AllowBackorders    bool   `json:"allow_backorders"`
	AutoApproveVendors bool   `json:"auto_approve_vendors"`
}

// TokenConfig represents token configuration for Auth service
type TokenConfig struct {
	AccessTokenDuration  time.Duration `json:"access_token_duration"`  // Default: 15 minutes
	RefreshTokenDuration time.Duration `json:"refresh_token_duration"` // Default: 7 days
	JWTSecret            string        `json:"jwt_secret"`             // Secret key for signing
	JWTIssuer            string        `json:"jwt_issuer"`             // Token issuer name
	AllowRefreshRotation bool          `json:"allow_refresh_rotation"` // Rotate refresh token on each use (RECOMMENDED)
	StrictIPCheck        bool          `json:"strict_ip_check"`        // Check if IP changes (optional)
	MaxActiveSessions    int           `json:"max_active_sessions"`    // Max concurrent sessions per user (0 = unlimited)
}

// PasswordPolicy represents password policy configuration
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSpecial   bool `json:"require_special"`
	ExpiryDays       int  `json:"expiry_days"`
}

// AuthServiceConfig represents specific config for Auth service
type AuthServiceConfig struct {
	TokenConfig         TokenConfig    `json:"token_config"`
	MFARequired         bool           `json:"mfa_required"`
	PasswordPolicy      PasswordPolicy `json:"password_policy"`
	MaxLoginAttempts    int            `json:"max_login_attempts"`
	LockoutDurationMins int            `json:"lockout_duration_mins"`
}
