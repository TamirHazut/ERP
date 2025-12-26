package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ============================================================================
// CONFIG MODELS
// ============================================================================

// ServiceConfig represents configuration for a service
type ServiceConfig struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ConfigID    string                 `bson:"config_id" json:"config_id"`
	ServiceName string                 `bson:"service_name" json:"service_name"` // core, auth, gateway, events
	Environment string                 `bson:"environment" json:"environment"`   // development, staging, production
	TenantID    string                 `bson:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	Config      map[string]interface{} `bson:"config" json:"config"`
	Version     int                    `bson:"version" json:"version"`
	IsActive    bool                   `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
	UpdatedBy   string                 `bson:"updated_by" json:"updated_by"`
}

// CoreServiceConfig represents specific config for Core service
type CoreServiceConfig struct {
	MaxOrderItems      int    `json:"max_order_items"`
	OrderNumberPrefix  string `json:"order_number_prefix"`
	DefaultCurrency    string `json:"default_currency"`
	AllowBackorders    bool   `json:"allow_backorders"`
	AutoApproveVendors bool   `json:"auto_approve_vendors"`
}

type TokenConfig struct {
	AccessTokenDuration  time.Duration // Default: 15 minutes
	RefreshTokenDuration time.Duration // Default: 7 days
	JWTSecret            string        // Secret key for signing
	JWTIssuer            string        // Token issuer name
	AllowRefreshRotation bool          // Rotate refresh token on each use (RECOMMENDED)
	StrictIPCheck        bool          // Check if IP changes (optional)
	MaxActiveSessions    int           // Max concurrent sessions per user (0 = unlimited)
}

// AuthServiceConfig represents specific config for Auth service
type AuthServiceConfig struct {
	TokenConfig         TokenConfig    `json:"token_config"`
	MFARequired         bool           `json:"mfa_required"`
	PasswordPolicy      PasswordPolicy `json:"password_policy"`
	MaxLoginAttempts    int            `json:"max_login_attempts"`
	LockoutDurationMins int            `json:"lockout_duration_mins"`
}

type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSpecial   bool `json:"require_special"`
	ExpiryDays       int  `json:"expiry_days"`
}

// FeatureFlag represents a feature flag
type FeatureFlag struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	FlagID      string              `bson:"flag_id" json:"flag_id"`
	Name        string              `bson:"name" json:"name"`
	Key         string              `bson:"key" json:"key"`
	Description string              `bson:"description" json:"description"`
	Enabled     bool                `bson:"enabled" json:"enabled"`
	Rollout     FeatureRollout      `bson:"rollout" json:"rollout"`
	Metadata    FeatureFlagMetadata `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time           `bson:"updated_at" json:"updated_at"`
}

type FeatureRollout struct {
	Percentage int      `bson:"percentage" json:"percentage"` // 0-100
	TenantIDs  []string `bson:"tenant_ids,omitempty" json:"tenant_ids,omitempty"`
	UserIDs    []string `bson:"user_ids,omitempty" json:"user_ids,omitempty"`
}

type FeatureFlagMetadata struct {
	Category         string `bson:"category,omitempty" json:"category,omitempty"`
	OwnerTeam        string `bson:"owner_team,omitempty" json:"owner_team,omitempty"`
	DocumentationURL string `bson:"documentation_url,omitempty" json:"documentation_url,omitempty"`
}
