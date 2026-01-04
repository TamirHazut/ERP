package core_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Category represents a product category
// Stored in MongoDB core_db.categories collection
type Category struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CategoryID       string             `bson:"category_id" json:"category_id"`
	TenantID         string             `bson:"tenant_id" json:"tenant_id"`
	Name             string             `bson:"name" json:"name"`
	Slug             string             `bson:"slug" json:"slug"`
	Description      string             `bson:"description" json:"description"`
	ParentCategoryID string             `bson:"parent_category_id,omitempty" json:"parent_category_id,omitempty"`
	Level            int                `bson:"level" json:"level"` // 0 = root
	Path             string             `bson:"path" json:"path"`   // /electronics/computers/laptops
	ImageURL         string             `bson:"image_url,omitempty" json:"image_url,omitempty"`
	Status           string             `bson:"status" json:"status"` // active, inactive
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}
