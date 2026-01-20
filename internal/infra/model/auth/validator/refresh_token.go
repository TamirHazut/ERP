package validator

import (
	"time"

	infra_error "erp.localhost/internal/infra/error"
	authv1_cache "erp.localhost/internal/infra/model/auth/v1/cache"
)

func ValidateRefreshToken(r *authv1_cache.RefreshToken) error {
	missingFields := []string{}
	if r.TokenHash == "" {
		missingFields = append(missingFields, "Token")
	}
	if r.TenantId == "" {
		missingFields = append(missingFields, "TenantID")
	}
	if r.UserId == "" {
		missingFields = append(missingFields, "UserID")
	}
	if r.ExpiresAt.AsTime().IsZero() {
		missingFields = append(missingFields, "ExpiresAt")
	}
	if r.CreatedAt.AsTime().IsZero() {
		missingFields = append(missingFields, "CreatedAt")
	}
	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}
	if IsExpired(r) {
		return infra_error.Auth(infra_error.AuthRefreshTokenExpired)
	}
	return nil
}

// IsValid - Check if refresh token is still valid
func IsValidRefreshToken(r *authv1_cache.RefreshToken) bool {
	return !r.Revoked && !IsExpired(r)
}

// IsExpired - Check if token is expired
func IsExpired(r *authv1_cache.RefreshToken) bool {
	return r.ExpiresAt.AsTime().IsZero() || time.Now().After(r.ExpiresAt.AsTime())
}
