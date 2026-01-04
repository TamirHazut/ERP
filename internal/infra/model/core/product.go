package core

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a product in the system
// Stored in MongoDB core_db.products collection
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProductID   string             `bson:"product_id" json:"product_id"`
	TenantID    string             `bson:"tenant_id" json:"tenant_id"`
	SKU         string             `bson:"sku" json:"sku"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	CategoryID  string             `bson:"category_id" json:"category_id"`
	Pricing     ProductPricing     `bson:"pricing" json:"pricing"`
	Inventory   ProductInventory   `bson:"inventory" json:"inventory"`
	Dimensions  ProductDimensions  `bson:"dimensions,omitempty" json:"dimensions,omitempty"`
	Images      []string           `bson:"images,omitempty" json:"images,omitempty"`
	Status      string             `bson:"status" json:"status"` // active, inactive, discontinued
	Metadata    ProductMetadata    `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedBy   string             `bson:"created_by" json:"created_by"`
}

type ProductPricing struct {
	Cost     float64 `bson:"cost" json:"cost"`
	Price    float64 `bson:"price" json:"price"`
	Currency string  `bson:"currency" json:"currency"`
	TaxRate  float64 `bson:"tax_rate" json:"tax_rate"`
}

type ProductInventory struct {
	TrackInventory  bool `bson:"track_inventory" json:"track_inventory"`
	Quantity        int  `bson:"quantity" json:"quantity"`
	Reserved        int  `bson:"reserved" json:"reserved"`
	Available       int  `bson:"available" json:"available"`
	ReorderPoint    int  `bson:"reorder_point" json:"reorder_point"`
	ReorderQuantity int  `bson:"reorder_quantity" json:"reorder_quantity"`
}

type ProductDimensions struct {
	Weight        float64 `bson:"weight" json:"weight"`
	WeightUnit    string  `bson:"weight_unit" json:"weight_unit"`
	Length        float64 `bson:"length" json:"length"`
	Width         float64 `bson:"width" json:"width"`
	Height        float64 `bson:"height" json:"height"`
	DimensionUnit string  `bson:"dimension_unit" json:"dimension_unit"`
}

type ProductMetadata struct {
	Barcode      string   `bson:"barcode,omitempty" json:"barcode,omitempty"`
	Manufacturer string   `bson:"manufacturer,omitempty" json:"manufacturer,omitempty"`
	Brand        string   `bson:"brand,omitempty" json:"brand,omitempty"`
	Tags         []string `bson:"tags,omitempty" json:"tags,omitempty"`
}
