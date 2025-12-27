package cache

import "time"

// FeatureFlagCache represents cached feature flag configuration
// Stored in: feature_flags:{tenant_id}:{flag_key} (Redis Hash)
// Or: feature_flags:{tenant_id} (Redis Hash with multiple flags)
// TTL: 10 minutes
type FeatureFlagCache struct {
	FlagKey     string                 `json:"flag_key"`
	TenantID    string                 `json:"tenant_id,omitempty"` // Empty for global flags
	Enabled     bool                   `json:"enabled"`
	Value       interface{}            `json:"value,omitempty"` // For flags with values (not just boolean)
	Rollout     *RolloutConfig         `json:"rollout,omitempty"` // Percentage-based rollout
	Targeting   *TargetingRules        `json:"targeting,omitempty"` // User/group targeting
	Environment string                 `json:"environment"` // development, staging, production
	CachedAt    time.Time              `json:"cached_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Version     int                    `json:"version"` // For cache invalidation
}

// RolloutConfig represents percentage-based feature rollout
type RolloutConfig struct {
	Percentage int      `json:"percentage"` // 0-100
	Buckets    []string `json:"buckets,omitempty"` // User IDs or groups in rollout
}

// TargetingRules represents feature flag targeting rules
type TargetingRules struct {
	UserIDs   []string          `json:"user_ids,omitempty"` // Specific users
	GroupIDs  []string          `json:"group_ids,omitempty"` // Specific groups
	Rules     []TargetingRule   `json:"rules,omitempty"` // Complex targeting rules
}

// TargetingRule represents a single targeting rule
type TargetingRule struct {
	Attribute string      `json:"attribute"` // e.g., "role", "department", "country"
	Operator  string      `json:"operator"` // "equals", "contains", "in", etc.
	Value     interface{} `json:"value"` // Comparison value
}

// TenantFeatureFlagsCache represents all feature flags for a tenant (bulk cache)
// Stored in: tenant_feature_flags:{tenant_id} (Redis Hash)
// TTL: 10 minutes
type TenantFeatureFlagsCache struct {
	TenantID    string                    `json:"tenant_id"`
	Environment string                    `json:"environment"`
	Flags       map[string]FeatureFlagCache `json:"flags"` // flag_key -> flag config
	CachedAt    time.Time                 `json:"cached_at"`
	ExpiresAt   time.Time                 `json:"expires_at"`
}
