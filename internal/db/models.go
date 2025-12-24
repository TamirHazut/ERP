package db

import "time"

type RoleName string
type QuantityUnit string

const (
	RoleSystemAdmin RoleName = "Admin"
	RoleTenantAdmin RoleName = "TenantAdmin"
	RoleTenantUser  RoleName = "TenantUser"

	QuantityUnitGrams     QuantityUnit = "g"
	QuantityUnitKilograms QuantityUnit = "kg"
	QuantityUnitLiters    QuantityUnit = "l"
	QuantityUnitPieces    QuantityUnit = "piece"
)

// Tenant related models

type Tenant struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	Creator   string    `bson:"user_id" json:"user_id"`
}

// User related models

type User struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Username  string    `bson:"username" json:"username"`
	Password  string    `bson:"password" json:"password"`
	TenantID  string    `bson:"tenant_id" json:"tenant_id"`
	RoleID    string    `bson:"role_id" json:"role_id"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// RBAC related models

type Role struct {
	ID          string   `bson:"_id,omitempty" json:"id,omitempty"`
	Name        RoleName `bson:"name" json:"name"`
	Permissions []string `bson:"permissions" json:"permissions"`
}

type Permission struct {
	ID          string `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
}

// Inventory related models

type Vendor struct {
	ID       string `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string `bson:"name" json:"name"`
	TenantID string `bson:"tenant_id" json:"tenant_id"`
}

type Product struct {
	ID           string       `bson:"_id,omitempty" json:"id,omitempty"`
	Name         string       `bson:"name" json:"name"`
	Description  string       `bson:"description" json:"description"`
	TenantID     string       `bson:"tenant_id" json:"tenant_id"`
	VendorID     string       `bson:"vendor_id" json:"vendor_id"`
	Quantity     int          `bson:"quantity" json:"quantity"`
	QuantityUnit QuantityUnit `bson:"quantity_unit" json:"quantity_unit"`
	Price        float64      `bson:"price" json:"price"`
	Tags         []string     `bson:"tags" json:"tags"`
	CreatedAt    time.Time    `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `bson:"updated_at" json:"updated_at"`
}

type OrderProduct struct {
	ProductID string  `bson:"product_id" json:"product_id"`
	Quantity  int     `bson:"quantity" json:"quantity"`
	Discount  int     `bson:"discount" json:"discount"`
	Total     float64 `bson:"total" json:"total"`
}

type Order struct {
	ID        string         `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedAt time.Time      `bson:"created_at" json:"created_at"`
	Name      string         `bson:"name" json:"name"`
	TenantID  string         `bson:"tenant_id" json:"tenant_id"`
	UserID    string         `bson:"user_id" json:"user_id"`
	Products  []OrderProduct `bson:"products" json:"products"`
	Total     float64        `bson:"total" json:"total"`
}
