package cache

import (
	infra_error "erp.localhost/infra/error"
	authv1_cache "erp.localhost/infra/model/auth/v1/cache"
)

func ValidateTokenMetaData(tm *authv1_cache.TokenMetadata) *infra_error.AppError {
	if tm == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "token")
	}
	missingFields := []string{}
	if tm.TenantId == "" {
		missingFields = append(missingFields, "TenantId")
	}
	if tm.UserId == "" {
		missingFields = append(missingFields, "UserId")
	}
	if tm.Jti == "" {
		missingFields = append(missingFields, "Jti")
	}
	if tm.IssuedAt.AsTime().IsZero() {
		missingFields = append(missingFields, "IssuedAt")
	}
	if tm.ExpiresAt.AsTime().IsZero() {
		missingFields = append(missingFields, "ExpiresAt")
	}
	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}
	return nil
}
