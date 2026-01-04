package auth_models

import (
	"time"

	erp_errors "erp.localhost/internal/infra/errors"
	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenClaims - Full user information for API requests
type AccessTokenClaims struct {
	UserID               string   `json:"sub"`                 // Subject (user ID)
	TenantID             string   `json:"tenant_id"`           // Tenant ID
	Email                string   `json:"email"`               // Email
	Username             string   `json:"username"`            // Username
	Roles                []string `json:"roles"`               // User roles
	Permissions          []string `json:"permissions"`         // Granular permissions
	TokenType            string   `json:"type"`                // "access"
	SessionID            string   `json:"sid,omitempty"`       // Session ID (optional)
	DeviceID             string   `json:"device_id,omitempty"` // Device identifier
	jwt.RegisteredClaims          // Standard JWT claims (exp, iat, iss, sub, etc.)
}

// Validate - Validates access token claims
func (c *AccessTokenClaims) Validate() error {
	missingFields := []string{}
	if c.UserID == "" {
		missingFields = append(missingFields, "UserID")
	}
	if c.TenantID == "" {
		missingFields = append(missingFields, "TenantID")
	}
	if c.Username == "" {
		missingFields = append(missingFields, "Username")
	}
	if c.Permissions == nil {
		missingFields = append(missingFields, "Permissions")
	}
	if len(c.Roles) == 0 {
		missingFields = append(missingFields, "Roles")
	}
	if c.TokenType != "access" {
		missingFields = append(missingFields, "TokenType")
	}
	// Validate expiration
	if c.ExpiresAt == nil {
		missingFields = append(missingFields, "ExpiresAt")
	}
	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}
	if c.IsExpired() {
		return erp_errors.Auth(erp_errors.AuthTokenExpired)
	}
	return nil
}

func (c *AccessTokenClaims) IsExpired() bool {
	return time.Now().After(c.ExpiresAt.Time)
}

// RefreshTokenClaims - Minimal information for token refresh
type RefreshTokenClaims struct {
	UserID               string `json:"sub"`                 // Subject (user ID)
	TokenType            string `json:"type"`                // "refresh"
	SessionID            string `json:"sid,omitempty"`       // Session ID for tracking
	DeviceID             string `json:"device_id,omitempty"` // Device identifier
	jwt.RegisteredClaims        // Standard JWT claims (exp, iat, jti, etc.)
}

// Validate - Validates refresh token claims
func (c *RefreshTokenClaims) Validate() error {
	missingFields := []string{}
	if c.UserID == "" {
		missingFields = append(missingFields, "UserID")
	}
	if c.TokenType != "refresh" {
		missingFields = append(missingFields, "TokenType")
	}
	// Validate expiration
	if c.ExpiresAt == nil {
		missingFields = append(missingFields, "ExpiresAt")
	} else if time.Now().After(c.ExpiresAt.Time) {
		// Only check if expired if ExpiresAt is not nil
		return erp_errors.Auth(erp_errors.AuthTokenExpired)
	}
	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}
	return nil
}
