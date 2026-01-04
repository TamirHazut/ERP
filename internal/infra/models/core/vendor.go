package core_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Vendor represents a vendor/supplier
// Stored in MongoDB core_db.vendors collection
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
