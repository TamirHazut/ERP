package models

import (
	"time"

	common_models "erp.localhost/internal/common/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Order represents an order
// Stored in MongoDB core_db.orders collection
type Order struct {
	ID              primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	OrderID         string                `bson:"order_id" json:"order_id"`
	TenantID        string                `bson:"tenant_id" json:"tenant_id"`
	OrderNumber     string                `bson:"order_number" json:"order_number"`
	OrderType       string                `bson:"order_type" json:"order_type"` // sales, purchase, transfer
	CustomerID      string                `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	VendorID        string                `bson:"vendor_id,omitempty" json:"vendor_id,omitempty"`
	Status          string                `bson:"status" json:"status"` // draft, pending, confirmed, shipped, delivered, cancelled
	Items           []string              `bson:"items" json:"items"`   // References to order_items
	Totals          OrderTotals           `bson:"totals" json:"totals"`
	ShippingAddress common_models.Address `bson:"shipping_address" json:"shipping_address"`
	BillingAddress  common_models.Address `bson:"billing_address" json:"billing_address"`
	Payment         PaymentInfo           `bson:"payment" json:"payment"`
	Fulfillment     FulfillmentInfo       `bson:"fulfillment,omitempty" json:"fulfillment,omitempty"`
	Notes           string                `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt       time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time             `bson:"updated_at" json:"updated_at"`
	CreatedBy       string                `bson:"created_by" json:"created_by"`
	Timeline        []OrderTimelineEvent  `bson:"timeline,omitempty" json:"timeline,omitempty"`
}

type OrderTotals struct {
	Subtotal float64 `bson:"subtotal" json:"subtotal"`
	Tax      float64 `bson:"tax" json:"tax"`
	Shipping float64 `bson:"shipping" json:"shipping"`
	Discount float64 `bson:"discount" json:"discount"`
	Total    float64 `bson:"total" json:"total"`
	Currency string  `bson:"currency" json:"currency"`
}

type PaymentInfo struct {
	Method        string     `bson:"method" json:"method"` // credit_card, bank_transfer, cash, check
	Status        string     `bson:"status" json:"status"` // pending, paid, refunded, failed
	PaidAt        *time.Time `bson:"paid_at,omitempty" json:"paid_at,omitempty"`
	TransactionID string     `bson:"transaction_id,omitempty" json:"transaction_id,omitempty"`
}

type FulfillmentInfo struct {
	WarehouseID    string     `bson:"warehouse_id,omitempty" json:"warehouse_id,omitempty"`
	ShippedAt      *time.Time `bson:"shipped_at,omitempty" json:"shipped_at,omitempty"`
	DeliveredAt    *time.Time `bson:"delivered_at,omitempty" json:"delivered_at,omitempty"`
	TrackingNumber string     `bson:"tracking_number,omitempty" json:"tracking_number,omitempty"`
	Carrier        string     `bson:"carrier,omitempty" json:"carrier,omitempty"`
}

type OrderTimelineEvent struct {
	Status    string    `bson:"status" json:"status"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	UserID    string    `bson:"user_id" json:"user_id"`
	Notes     string    `bson:"notes,omitempty" json:"notes,omitempty"`
}

// OrderItem represents an item in an order
// Stored in MongoDB core_db.order_items collection
type OrderItem struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ItemID    string             `bson:"item_id" json:"item_id"`
	OrderID   string             `bson:"order_id" json:"order_id"`
	TenantID  string             `bson:"tenant_id" json:"tenant_id"`
	ProductID string             `bson:"product_id" json:"product_id"`
	SKU       string             `bson:"sku" json:"sku"`
	Name      string             `bson:"name" json:"name"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	UnitPrice float64            `bson:"unit_price" json:"unit_price"`
	TaxRate   float64            `bson:"tax_rate" json:"tax_rate"`
	Discount  float64            `bson:"discount" json:"discount"`
	Subtotal  float64            `bson:"subtotal" json:"subtotal"`
	Tax       float64            `bson:"tax" json:"tax"`
	Total     float64            `bson:"total" json:"total"`
	Status    string             `bson:"status" json:"status"` // pending, fulfilled, cancelled, returned
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
