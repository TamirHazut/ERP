package validator

import (
	infra_error "erp.localhost/infra/error"
	authv1 "erp.localhost/infra/model/auth/v1"
)

func ValidateTenant(t *authv1.Tenant, createOperation bool) *infra_error.AppError {
	if t == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "tenant")
	}
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
	if _, ok := authv1.TenantStatus_name[int32(t.Status)]; !ok || t.Status == authv1.TenantStatus_TENANT_STATUS_UNSPECIFIED {
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
