package validator

import (
	infra_error "erp.localhost/internal/infra/error"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
)

func ValidateTenant(t *authv1.Tenant, createOperation bool) error {
	missingFields := []string{}
	if !createOperation {
		if t.Id == "" {
			missingFields = append(missingFields, "Id")
		}
	}
	if t.Name == "" {
		missingFields = append(missingFields, "Name")
	}
	if t.CreatedBy == "" {
		missingFields = append(missingFields, "CreatedBy")
	}
	if t.Status == authv1.TenantStatus_TENANT_STATUS_UNSPECIFIED {
		missingFields = append(missingFields, "Status")
	}
	if t.GetContact().GetEmail() == "" {
		missingFields = append(missingFields, "EMail")
	}
	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}
	return nil
}
