package cache

import (
	"time"

	infra_error "erp.localhost/infra/error"
	authv1_cache "erp.localhost/infra/model/auth/v1/cache"
)

func ValidateRefreshToken(r *authv1_cache.RefreshToken) *infra_error.AppError {
	if r == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "token")
	}
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

// IsExpired - Check if token is expired
func IsExpired(r *authv1_cache.RefreshToken) bool {
	return r.ExpiresAt.AsTime().IsZero() || time.Now().After(r.ExpiresAt.AsTime())
}
