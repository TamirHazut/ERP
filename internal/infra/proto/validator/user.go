package validator

import (
	erp_errors "erp.localhost/internal/infra/error"
	infrav1 "erp.localhost/internal/infra/proto/infra/v1"
)

func ValidateUserIdentifier(identifier *infrav1.UserIdentifier) error {
	if identifier == nil {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "identifier")
	}
	tenantID := identifier.GetTenantId()
	if tenantID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenant_id")
	}
	userID := identifier.GetUserId()
	if userID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "user_id")
	}
	return nil
}
