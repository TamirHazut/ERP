package validator

import (
	infra_error "erp.localhost/internal/infra/error"
	infrav1 "erp.localhost/internal/infra/model/infra/v1"
)

func ValidateUserIdentifier(identifier *infrav1.UserIdentifier) error {
	if identifier == nil {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "identifier")
	}
	tenantID := identifier.GetTenantId()
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	userID := identifier.GetUserId()
	if userID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "user_id")
	}
	return nil
}
