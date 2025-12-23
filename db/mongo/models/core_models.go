package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Order statuses
const (
	OrderStatusDraft     = "draft"
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
)

// Order types
const (
	OrderTypeSales    = "sales"
	OrderTypePurchase = "purchase"
	OrderTypeTransfer = "transfer"
)

// Payment statuses
const (
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusRefunded = "refunded"
	PaymentStatusFailed   = "failed"
)

// Product statuses
const (
	ProductStatusActive       = "active"
	ProductStatusInactive     = "inactive"
	ProductStatusDiscontinued = "discontinued"
)

// Vendor statuses
const (
	VendorStatusActive          = "active"
	VendorStatusInactive        = "inactive"
	VendorStatusPendingApproval = "pending_approval"
)

// ============================================================================
// CORE_DB MODELS
// ============================================================================

// Product represents a product in the system
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

// Vendor represents a vendor/supplier
type Vendor struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VendorID         string             `bson:"vendor_id" json:"vendor_id"`
	TenantID         string             `bson:"tenant_id" json:"tenant_id"`
	Name             string             `bson:"name" json:"name"`
	Code             string             `bson:"code" json:"code"`
	Contact          VendorContact      `bson:"contact" json:"contact"`
	Address          Address            `bson:"address" json:"address"`
	PaymentTerms     PaymentTerms       `bson:"payment_terms" json:"payment_terms"`
	Rating           float64            `bson:"rating" json:"rating"`
	Status           string             `bson:"status" json:"status"` // active, inactive, pending_approval
	ProductsSupplied []string           `bson:"products_supplied,omitempty" json:"products_supplied,omitempty"`
	Metadata         VendorMetadata     `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedBy        string             `bson:"created_by" json:"created_by"`
}

type VendorContact struct {
	Email         string `bson:"email" json:"email"`
	Phone         string `bson:"phone" json:"phone"`
	Website       string `bson:"website,omitempty" json:"website,omitempty"`
	ContactPerson string `bson:"contact_person,omitempty" json:"contact_person,omitempty"`
}

type PaymentTerms struct {
	Terms       string  `bson:"terms" json:"terms"`
	CreditLimit float64 `bson:"credit_limit" json:"credit_limit"`
	Currency    string  `bson:"currency" json:"currency"`
}

type VendorMetadata struct {
	TaxID        string `bson:"tax_id,omitempty" json:"tax_id,omitempty"`
	BusinessType string `bson:"business_type,omitempty" json:"business_type,omitempty"`
	Notes        string `bson:"notes,omitempty" json:"notes,omitempty"`
}

// Order represents an order
type Order struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	OrderID         string               `bson:"order_id" json:"order_id"`
	TenantID        string               `bson:"tenant_id" json:"tenant_id"`
	OrderNumber     string               `bson:"order_number" json:"order_number"`
	OrderType       string               `bson:"order_type" json:"order_type"` // sales, purchase, transfer
	CustomerID      string               `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	VendorID        string               `bson:"vendor_id,omitempty" json:"vendor_id,omitempty"`
	Status          string               `bson:"status" json:"status"` // draft, pending, confirmed, shipped, delivered, cancelled
	Items           []string             `bson:"items" json:"items"`   // References to order_items
	Totals          OrderTotals          `bson:"totals" json:"totals"`
	ShippingAddress Address              `bson:"shipping_address" json:"shipping_address"`
	BillingAddress  Address              `bson:"billing_address" json:"billing_address"`
	Payment         PaymentInfo          `bson:"payment" json:"payment"`
	Fulfillment     FulfillmentInfo      `bson:"fulfillment,omitempty" json:"fulfillment,omitempty"`
	Notes           string               `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt       time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `bson:"updated_at" json:"updated_at"`
	CreatedBy       string               `bson:"created_by" json:"created_by"`
	Timeline        []OrderTimelineEvent `bson:"timeline,omitempty" json:"timeline,omitempty"`
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

// Customer represents a customer
type Customer struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID     string             `bson:"customer_id" json:"customer_id"`
	TenantID       string             `bson:"tenant_id" json:"tenant_id"`
	Type           string             `bson:"type" json:"type"` // individual, business
	Name           string             `bson:"name" json:"name"`
	Email          string             `bson:"email" json:"email"`
	Phone          string             `bson:"phone" json:"phone"`
	Company        CompanyInfo        `bson:"company,omitempty" json:"company,omitempty"`
	Addresses      []CustomerAddress  `bson:"addresses" json:"addresses"`
	PaymentMethods []PaymentMethod    `bson:"payment_methods,omitempty" json:"payment_methods,omitempty"`
	CreditLimit    float64            `bson:"credit_limit" json:"credit_limit"`
	Status         string             `bson:"status" json:"status"` // active, inactive, blocked
	LifetimeValue  float64            `bson:"lifetime_value" json:"lifetime_value"`
	TotalOrders    int                `bson:"total_orders" json:"total_orders"`
	LastOrderDate  *time.Time         `bson:"last_order_date,omitempty" json:"last_order_date,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

type CompanyInfo struct {
	Name  string `bson:"name,omitempty" json:"name,omitempty"`
	TaxID string `bson:"tax_id,omitempty" json:"tax_id,omitempty"`
}

type CustomerAddress struct {
	AddressID string `bson:"address_id" json:"address_id"`
	Type      string `bson:"type" json:"type"` // billing, shipping
	IsDefault bool   `bson:"is_default" json:"is_default"`
	Street    string `bson:"street" json:"street"`
	City      string `bson:"city" json:"city"`
	State     string `bson:"state" json:"state"`
	Zip       string `bson:"zip" json:"zip"`
	Country   string `bson:"country" json:"country"`
}

type PaymentMethod struct {
	MethodID  string                 `bson:"method_id" json:"method_id"`
	Type      string                 `bson:"type" json:"type"`
	IsDefault bool                   `bson:"is_default" json:"is_default"`
	Details   map[string]interface{} `bson:"details,omitempty" json:"-"` // Encrypted
}

// Inventory represents inventory tracking
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

// Warehouse represents a warehouse/storage location
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

// Category represents a product category
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
