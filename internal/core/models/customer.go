package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Customer represents a customer
// Stored in MongoDB core_db.customers collection
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
