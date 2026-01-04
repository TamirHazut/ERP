package auth_models

import (
	"time"

	erp_errors "erp.localhost/internal/infra/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role represents a role with permissions
type Role struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID     string             `bson:"tenant_id,omitempty" json:"tenant_id,omitempty"` // null for system-wide
	Name         string             `bson:"name" json:"name"`
	Slug         string             `bson:"slug" json:"slug"`
	Description  string             `bson:"description" json:"description"`
	Type         string             `bson:"type" json:"type"` // system, custom
	IsSystemRole bool               `bson:"is_system_role" json:"is_system_role"`
	Permissions  []string           `bson:"permissions" json:"permissions"`
	Priority     int                `bson:"priority" json:"priority"`
	Status       string             `bson:"status" json:"status"` // active, inactive
	Metadata     RoleMetadata       `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedBy    string             `bson:"created_by" json:"created_by"`
}

func (r *Role) Validate(createOperation bool) error {
	missingFields := []string{}
	if !createOperation {
		if r.ID == primitive.NilObjectID {
			missingFields = append(missingFields, "ID")
		}
	}
	if r.TenantID == "" {
		missingFields = append(missingFields, "TenantID")
	}
	if r.Name == "" {
		missingFields = append(missingFields, "Name")
	}
	if r.Status == "" {
		missingFields = append(missingFields, "Status")
	}
	if r.CreatedBy == "" {
		missingFields = append(missingFields, "CreatedBy")
	}
	if r.Permissions == nil {
		missingFields = append(missingFields, "Permissions")
	}
	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}
	return nil
}

type RoleMetadata struct {
	Color         string `bson:"color,omitempty" json:"color,omitempty"`
	Icon          string `bson:"icon,omitempty" json:"icon,omitempty"`
	MaxAssignable int    `bson:"max_assignable,omitempty" json:"max_assignable,omitempty"`
}
