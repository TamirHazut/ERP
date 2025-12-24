package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User statuses
const (
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
	UserStatusInvited   = "invited"
)

// Tenant statuses
const (
	TenantStatusActive    = "active"
	TenantStatusSuspended = "suspended"
	TenantStatusInactive  = "inactive"
	TenantStatusTrial     = "trial"
)

// Permission formats
const (
	PermissionWildcard = "*:*"
	PermissionFormat   = "resource:action[:scope]"
)

// Role types
const (
	RoleSystemAdmin = "system_admin"
	RoleTenantAdmin = "tenant_admin"
)

// ============================================================================
// AUTH_DB MODELS
// ============================================================================

// Tenant represents an organization/company in the system
type Tenant struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID     string             `bson:"tenant_id" json:"tenant_id"`
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
	Email   string  `bson:"email" json:"email"`
	Phone   string  `bson:"phone" json:"phone"`
	Address Address `bson:"address" json:"address"`
}

type Address struct {
	Street  string `bson:"street" json:"street"`
	City    string `bson:"city" json:"city"`
	State   string `bson:"state" json:"state"`
	Zip     string `bson:"zip" json:"zip"`
	Country string `bson:"country" json:"country"`
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

// User represents a user in the system
type User struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID                string             `bson:"user_id" json:"user_id"`
	TenantID              string             `bson:"tenant_id" json:"tenant_id"`
	Email                 string             `bson:"email" json:"email"`
	Username              string             `bson:"username" json:"username"`
	PasswordHash          string             `bson:"password_hash" json:"-"`
	Profile               UserProfile        `bson:"profile" json:"profile"`
	Roles                 []UserRole         `bson:"roles" json:"roles"`
	AdditionalPermissions []string           `bson:"additional_permissions,omitempty" json:"additional_permissions,omitempty"`
	RevokedPermissions    []string           `bson:"revoked_permissions,omitempty" json:"revoked_permissions,omitempty"`
	Status                string             `bson:"status" json:"status"` // active, inactive, suspended, invited
	EmailVerified         bool               `bson:"email_verified" json:"email_verified"`
	PhoneVerified         bool               `bson:"phone_verified" json:"phone_verified"`
	MFAEnabled            bool               `bson:"mfa_enabled" json:"mfa_enabled"`
	MFASecret             string             `bson:"mfa_secret,omitempty" json:"-"`
	LastLogin             *time.Time         `bson:"last_login,omitempty" json:"last_login,omitempty"`
	LastPasswordChange    time.Time          `bson:"last_password_change" json:"last_password_change"`
	PasswordResetToken    string             `bson:"password_reset_token,omitempty" json:"-"`
	PasswordResetExpires  *time.Time         `bson:"password_reset_expires,omitempty" json:"-"`
	Preferences           UserPreferences    `bson:"preferences" json:"preferences"`
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedBy             string             `bson:"created_by" json:"created_by"`
	LastActivity          time.Time          `bson:"last_activity" json:"last_activity"`
	LoginHistory          []LoginRecord      `bson:"login_history,omitempty" json:"login_history,omitempty"`
}

type UserProfile struct {
	FirstName   string `bson:"first_name" json:"first_name"`
	LastName    string `bson:"last_name" json:"last_name"`
	DisplayName string `bson:"display_name" json:"display_name"`
	AvatarURL   string `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	Phone       string `bson:"phone,omitempty" json:"phone,omitempty"`
	Title       string `bson:"title,omitempty" json:"title,omitempty"`
	Department  string `bson:"department,omitempty" json:"department,omitempty"`
}

type UserRole struct {
	RoleID     string     `bson:"role_id" json:"role_id"`
	TenantID   string     `bson:"tenant_id" json:"tenant_id"`
	AssignedAt time.Time  `bson:"assigned_at" json:"assigned_at"`
	AssignedBy string     `bson:"assigned_by" json:"assigned_by"`
	ExpiresAt  *time.Time `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
}

type UserPreferences struct {
	Language        string                 `bson:"language" json:"language"`
	Timezone        string                 `bson:"timezone" json:"timezone"`
	Theme           string                 `bson:"theme" json:"theme"`
	Notifications   NotificationSettings   `bson:"notifications" json:"notifications"`
	DashboardLayout map[string]interface{} `bson:"dashboard_layout,omitempty" json:"dashboard_layout,omitempty"`
}

type NotificationSettings struct {
	Email bool `bson:"email" json:"email"`
	Push  bool `bson:"push" json:"push"`
	SMS   bool `bson:"sms" json:"sms"`
}

type LoginRecord struct {
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	IPAddress string    `bson:"ip_address" json:"ip_address"`
	UserAgent string    `bson:"user_agent" json:"user_agent"`
	Success   bool      `bson:"success" json:"success"`
}

// Role represents a role with permissions
type Role struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoleID       string             `bson:"role_id" json:"role_id"`
	TenantID     string             `bson:"tenant_id,omitempty" json:"tenant_id,omitempty"` // null for system-wide
	Name         string             `bson:"name" json:"name"`
	Slug         string             `bson:"slug" json:"slug"`
	Description  string             `bson:"description" json:"description"`
	Type         string             `bson:"type" json:"type"` // system, custom
	IsSystemRole bool               `bson:"is_system_role" json:"is_system_role"`
	Permissions  []string           `bson:"permissions" json:"permissions"`
	Priority     int                `bson:"priority" json:"priority"`
	Status       string             `bson:"status" json:"status"` // active, inactive
	Metadata     RoleMetadata       `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedBy    string             `bson:"created_by" json:"created_by"`
}

type RoleMetadata struct {
	Color         string `bson:"color,omitempty" json:"color,omitempty"`
	Icon          string `bson:"icon,omitempty" json:"icon,omitempty"`
	MaxAssignable int    `bson:"max_assignable,omitempty" json:"max_assignable,omitempty"`
}

// Permission represents a system permission
type Permission struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PermissionID     string             `bson:"permission_id" json:"permission_id"`
	Resource         string             `bson:"resource" json:"resource"`
	Action           string             `bson:"action" json:"action"`
	PermissionString string             `bson:"permission_string" json:"permission_string"`
	DisplayName      string             `bson:"display_name" json:"display_name"`
	Description      string             `bson:"description" json:"description"`
	Category         string             `bson:"category" json:"category"`
	IsDangerous      bool               `bson:"is_dangerous" json:"is_dangerous"`
	RequiresApproval bool               `bson:"requires_approval" json:"requires_approval"`
	Dependencies     []string           `bson:"dependencies,omitempty" json:"dependencies,omitempty"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	Metadata         PermissionMetadata `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

type PermissionMetadata struct {
	Module  string `bson:"module" json:"module"`
	UIGroup string `bson:"ui_group" json:"ui_group"`
}

// UserGroup represents a group of users
type UserGroup struct {
	ID                    primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	GroupID               string                 `bson:"group_id" json:"group_id"`
	TenantID              string                 `bson:"tenant_id" json:"tenant_id"`
	Name                  string                 `bson:"name" json:"name"`
	Slug                  string                 `bson:"slug" json:"slug"`
	Description           string                 `bson:"description" json:"description"`
	Type                  string                 `bson:"type" json:"type"` // department, team, project, location
	Members               []GroupMember          `bson:"members" json:"members"`
	Roles                 []string               `bson:"roles" json:"roles"`
	AdditionalPermissions []string               `bson:"additional_permissions,omitempty" json:"additional_permissions,omitempty"`
	ParentGroupID         string                 `bson:"parent_group_id,omitempty" json:"parent_group_id,omitempty"`
	Metadata              map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt             time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time              `bson:"updated_at" json:"updated_at"`
}

type GroupMember struct {
	UserID  string    `bson:"user_id" json:"user_id"`
	AddedAt time.Time `bson:"added_at" json:"added_at"`
	AddedBy string    `bson:"added_by" json:"added_by"`
}

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LogID        string             `bson:"log_id" json:"log_id"`
	TenantID     string             `bson:"tenant_id" json:"tenant_id"`
	Actor        Actor              `bson:"actor" json:"actor"`
	Action       string             `bson:"action" json:"action"`
	ResourceType string             `bson:"resource_type" json:"resource_type"`
	ResourceID   string             `bson:"resource_id" json:"resource_id"`
	Changes      Changes            `bson:"changes,omitempty" json:"changes,omitempty"`
	Status       string             `bson:"status" json:"status"` // success, failure
	ErrorMessage string             `bson:"error_message,omitempty" json:"error_message,omitempty"`
	Metadata     AuditMetadata      `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Timestamp    time.Time          `bson:"timestamp" json:"timestamp"`
}

type Actor struct {
	UserID    string `bson:"user_id" json:"user_id"`
	Username  string `bson:"username" json:"username"`
	IPAddress string `bson:"ip_address" json:"ip_address"`
	UserAgent string `bson:"user_agent" json:"user_agent"`
}

type Changes struct {
	Before map[string]interface{} `bson:"before,omitempty" json:"before,omitempty"`
	After  map[string]interface{} `bson:"after,omitempty" json:"after,omitempty"`
}

type AuditMetadata struct {
	RequestID  string `bson:"request_id,omitempty" json:"request_id,omitempty"`
	SessionID  string `bson:"session_id,omitempty" json:"session_id,omitempty"`
	APIVersion string `bson:"api_version,omitempty" json:"api_version,omitempty"`
}
