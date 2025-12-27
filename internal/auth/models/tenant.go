package models

import (
	"time"

	common_models "erp.localhost/internal/common/models"
	erp_errors "erp.localhost/internal/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Tenant represents an organization/company in the system
type Tenant struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Slug         string             `bson:"slug" json:"slug"`
	Domain       string             `bson:"domain,omitempty" json:"domain,omitempty"`
	Status       string             `bson:"status" json:"status"` // active, suspended, inactive, trial
	Subscription Subscription       `bson:"subscription" json:"subscription"`
	Settings     TenantSettings     `bson:"settings" json:"settings"`
	Contact      ContactInfo        `bson:"contact" json:"contact"`
	Branding     Branding           `bson:"branding,omitempty" json:"branding,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedBy    string             `bson:"created_by" json:"created_by"`
	Metadata     TenantMetadata     `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

func (t *Tenant) Validate(createOperation bool) error {
	missingFields := []string{}
	if !createOperation {
		if t.ID == primitive.NilObjectID {
			missingFields = append(missingFields, "ID")
		}
	}
	if t.Name == "" {
		missingFields = append(missingFields, "Name")
	}
	if t.Status == "" {
		missingFields = append(missingFields, "Status")
	}
	if t.CreatedBy == "" {
		missingFields = append(missingFields, "CreatedBy")
	}
	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}
	return nil
}

type Subscription struct {
	Plan      string             `bson:"plan" json:"plan"` // free, basic, professional, enterprise
	StartDate time.Time          `bson:"start_date" json:"start_date"`
	EndDate   time.Time          `bson:"end_date" json:"end_date"`
	Features  []string           `bson:"features" json:"features"`
	Limits    SubscriptionLimits `bson:"limits" json:"limits"`
}

type SubscriptionLimits struct {
	MaxUsers          int `bson:"max_users" json:"max_users"`
	MaxProducts       int `bson:"max_products" json:"max_products"`
	MaxOrdersPerMonth int `bson:"max_orders_per_month" json:"max_orders_per_month"`
	StorageGB         int `bson:"storage_gb" json:"storage_gb"`
}

type TenantSettings struct {
	Timezone      string           `bson:"timezone" json:"timezone"`
	Currency      string           `bson:"currency" json:"currency"`
	DateFormat    string           `bson:"date_format" json:"date_format"`
	Language      string           `bson:"language" json:"language"`
	BusinessHours map[string]Hours `bson:"business_hours,omitempty" json:"business_hours,omitempty"`
}

type Hours struct {
	Start string `bson:"start" json:"start"`
	End   string `bson:"end" json:"end"`
}

type ContactInfo struct {
	Email   string                `bson:"email" json:"email"`
	Phone   string                `bson:"phone" json:"phone"`
	Address common_models.Address `bson:"address" json:"address"`
}

type Branding struct {
	LogoURL      string `bson:"logo_url,omitempty" json:"logo_url,omitempty"`
	PrimaryColor string `bson:"primary_color,omitempty" json:"primary_color,omitempty"`
	CompanyName  string `bson:"company_name,omitempty" json:"company_name,omitempty"`
}

type TenantMetadata struct {
	OnboardingCompleted bool   `bson:"onboarding_completed" json:"onboarding_completed"`
	Industry            string `bson:"industry,omitempty" json:"industry,omitempty"`
	CompanySize         string `bson:"company_size,omitempty" json:"company_size,omitempty"`
}
