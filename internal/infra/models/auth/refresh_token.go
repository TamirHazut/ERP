package auth_models

import (
	"time"

	erp_errors "erp.localhost/internal/infra/errors"
)

// RefreshToken - Refresh token stored in database/Redis
type RefreshToken struct {
	Token      string    `bson:"token" json:"token"`                                   // The actual token string
	UserID     string    `bson:"user_id" json:"user_id"`                               // Owner of the token
	TenantID   string    `bson:"tenant_id" json:"tenant_id"`                           // Tenant ID
	SessionID  string    `bson:"session_id" json:"session_id"`                         // Session ID
	DeviceID   string    `bson:"device_id,omitempty" json:"device_id,omitempty"`       // Device identifier
	IPAddress  string    `bson:"ip_address,omitempty" json:"ip_address,omitempty"`     // IP when token was created
	UserAgent  string    `bson:"user_agent,omitempty" json:"user_agent,omitempty"`     // Browser/app info
	ExpiresAt  time.Time `bson:"expires_at" json:"expires_at"`                         // When token expires
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`                         // When token was created
	LastUsedAt time.Time `bson:"last_used_at,omitempty" json:"last_used_at,omitempty"` // Last time token was used
	RevokedAt  time.Time `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`     // When token was revoked
	IsRevoked  bool      `bson:"is_revoked" json:"is_revoked"`                         // Quick check if revoked
	RevokedBy  string    `bson:"revoked_by,omitempty" json:"revoked_by,omitempty"`     // Who revoked the token
}

func (r *RefreshToken) Validate() error {
	missingFields := []string{}
	if r.Token == "" {
		missingFields = append(missingFields, "Token")
	}
	if r.TenantID == "" {
		missingFields = append(missingFields, "TenantID")
	}
	if r.UserID == "" {
		missingFields = append(missingFields, "UserID")
	}
	if r.ExpiresAt.IsZero() {
		missingFields = append(missingFields, "ExpiresAt")
	}
	if r.CreatedAt.IsZero() {
		missingFields = append(missingFields, "CreatedAt")
	}
	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}
	if r.IsExpired() {
		return erp_errors.Auth(erp_errors.AuthRefreshTokenExpired)
	}
	return nil
}

// IsValid - Check if refresh token is still valid
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked && !rt.IsExpired()
}

// IsExpired - Check if token is expired
func (rt *RefreshToken) IsExpired() bool {
	return rt.ExpiresAt.IsZero() || time.Now().After(rt.ExpiresAt)
}
