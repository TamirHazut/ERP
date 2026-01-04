package auth_models

import (
	"time"

	erp_errors "erp.localhost/internal/infra/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Permission represents a system permission
type Permission struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID         string             `bson:"tenant_id" json:"tenant_id"`
	Resource         string             `bson:"resource" json:"resource"`
	Action           string             `bson:"action" json:"action"`
	PermissionString string             `bson:"permission_string" json:"permission_string"`
	DisplayName      string             `bson:"display_name" json:"display_name"`
	Description      string             `bson:"description" json:"description"`
	Category         string             `bson:"category" json:"category"`
	IsDangerous      bool               `bson:"is_dangerous" json:"is_dangerous"`
	RequiresApproval bool               `bson:"requires_approval" json:"requires_approval"`
	Dependencies     []string           `bson:"dependencies,omitempty" json:"dependencies,omitempty"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedBy        string             `bson:"created_by" json:"created_by"`
	Metadata         PermissionMetadata `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

func (p *Permission) Validate(createOperation bool) error {
	missingFields := []string{}
	if !createOperation {
		if p.ID == primitive.NilObjectID {
			missingFields = append(missingFields, "ID")
		}
	}
	if p.TenantID == "" {
		missingFields = append(missingFields, "TenantID")
	}
	if p.Resource == "" {
		missingFields = append(missingFields, "Resource")
	}
	if p.Action == "" {
		missingFields = append(missingFields, "Action")
	}
	if p.CreatedBy == "" {
		missingFields = append(missingFields, "CreatedBy")
	}
	if p.DisplayName == "" {
		missingFields = append(missingFields, "DisplayName")
	}
	if p.PermissionString == "" {
		missingFields = append(missingFields, "PermissionString")
	}
	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}
	return nil
}

type PermissionMetadata struct {
	Module  string `bson:"module" json:"module"`
	UIGroup string `bson:"ui_group" json:"ui_group"`
}
