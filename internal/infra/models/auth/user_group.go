package auth_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserGroup represents a group of users
type UserGroup struct {
	ID                    primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	GroupID               string                 `bson:"group_id" json:"group_id"`
	TenantID              string                 `bson:"tenant_id" json:"tenant_id"`
	Name                  string                 `bson:"name" json:"name"`
	Slug                  string                 `bson:"slug" json:"slug"`
	Description           string                 `bson:"description" json:"description"`
	Type                  string                 `bson:"type" json:"type"` // department, team, project, location
	Members               []GroupMember          `bson:"members" json:"members"`
	Roles                 []string               `bson:"roles" json:"roles"`
	AdditionalPermissions []string               `bson:"additional_permissions,omitempty" json:"additional_permissions,omitempty"`
	ParentGroupID         string                 `bson:"parent_group_id,omitempty" json:"parent_group_id,omitempty"`
	Metadata              map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt             time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time              `bson:"updated_at" json:"updated_at"`
}

type GroupMember struct {
	UserID  string    `bson:"user_id" json:"user_id"`
	AddedAt time.Time `bson:"added_at" json:"added_at"`
	AddedBy string    `bson:"added_by" json:"added_by"`
}
