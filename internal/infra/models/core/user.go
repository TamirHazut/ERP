package core_models

import (
	"time"

	erp_errors "erp.localhost/internal/infra/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
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

func (u *User) Validate(createOperation bool) error {
	missingFields := []string{}
	if !createOperation {
		if u.ID == primitive.NilObjectID {
			missingFields = append(missingFields, "ID")
		}
	}
	if u.TenantID == "" {
		missingFields = append(missingFields, "TenantID")
	}
	if u.Email == "" && u.Username == "" {
		missingFields = append(missingFields, "Email or Username")
	}
	if u.PasswordHash == "" {
		missingFields = append(missingFields, "PasswordHash")
	}
	if u.Status == "" {
		missingFields = append(missingFields, "Status")
	}
	if u.CreatedBy == "" {
		missingFields = append(missingFields, "CreatedBy")
	}
	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}
	return nil
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
