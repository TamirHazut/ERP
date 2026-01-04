package core

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Inventory represents inventory tracking
// Stored in MongoDB core_db.inventory collection
type Inventory struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	InventoryID   string             `bson:"inventory_id" json:"inventory_id"`
	TenantID      string             `bson:"tenant_id" json:"tenant_id"`
	ProductID     string             `bson:"product_id" json:"product_id"`
	WarehouseID   string             `bson:"warehouse_id" json:"warehouse_id"`
	Quantity      int                `bson:"quantity" json:"quantity"`
	Reserved      int                `bson:"reserved" json:"reserved"`
	Available     int                `bson:"available" json:"available"`
	Location      InventoryLocation  `bson:"location,omitempty" json:"location,omitempty"`
	BatchNumber   string             `bson:"batch_number,omitempty" json:"batch_number,omitempty"`
	SerialNumbers []string           `bson:"serial_numbers,omitempty" json:"serial_numbers,omitempty"`
	ExpiryDate    *time.Time         `bson:"expiry_date,omitempty" json:"expiry_date,omitempty"`
	ReceivedDate  time.Time          `bson:"received_date" json:"received_date"`
	Cost          float64            `bson:"cost" json:"cost"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type InventoryLocation struct {
	Aisle string `bson:"aisle,omitempty" json:"aisle,omitempty"`
	Shelf string `bson:"shelf,omitempty" json:"shelf,omitempty"`
	Bin   string `bson:"bin,omitempty" json:"bin,omitempty"`
}
