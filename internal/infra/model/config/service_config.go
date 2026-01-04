package config

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ServiceConfig represents configuration for a service
// Stored in MongoDB config_db.service_configs collection
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
