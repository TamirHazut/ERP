package core

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Warehouse represents a warehouse/storage location
// Stored in MongoDB core_db.warehouses collection
type Warehouse struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	WarehouseID string             `bson:"warehouse_id" json:"warehouse_id"`
	TenantID    string             `bson:"tenant_id" json:"tenant_id"`
	Name        string             `bson:"name" json:"name"`
	Code        string             `bson:"code" json:"code"`
	Address     Address            `bson:"address" json:"address"`
	Contact     WarehouseContact   `bson:"contact" json:"contact"`
	Capacity    WarehouseCapacity  `bson:"capacity" json:"capacity"`
	Status      string             `bson:"status" json:"status"` // active, inactive, maintenance
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type WarehouseContact struct {
	Manager string `bson:"manager" json:"manager"`
	Email   string `bson:"email" json:"email"`
	Phone   string `bson:"phone" json:"phone"`
}

type WarehouseCapacity struct {
	TotalSpace int    `bson:"total_space" json:"total_space"`
	UsedSpace  int    `bson:"used_space" json:"used_space"`
	Unit       string `bson:"unit" json:"unit"` // sqft, sqm
}
