package config_cache_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

// TenantFeatures represents cached tenant-specific features
type TenantFeatures struct {
	TenantID string             `json:"tenant_id"`
	Features map[string]bool    `json:"features"` // feature_key -> enabled
	Limits   SubscriptionLimits `json:"limits"`
	CachedAt time.Time          `json:"cached_at"`
}

type SubscriptionLimits struct {
	MaxUsers          int `bson:"max_users" json:"max_users"`
	MaxProducts       int `bson:"max_products" json:"max_products"`
	MaxOrdersPerMonth int `bson:"max_orders_per_month" json:"max_orders_per_month"`
	StorageGB         int `bson:"storage_gb" json:"storage_gb"`
}
