package validator

import (
	infra_error "erp.localhost/internal/infra/error"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
)

func ValidateRole(r *authv1.Role, createOperation bool) error {
	missingFields := []string{}
	if !createOperation {
		if r.Id == "" {
			missingFields = append(missingFields, "Id")
		}
	}
	if r.TenantId == "" {
		missingFields = append(missingFields, "TenantId")
	}
	if r.Name == "" {
		missingFields = append(missingFields, "Name")
	}
	if r.Status == authv1.RoleStatus_ROLE_STATUS_UNSPECIFIED {
		missingFields = append(missingFields, "Status")
	}
	if r.CreatedBy == "" {
		missingFields = append(missingFields, "CreatedBy")
	}
	if r.Permissions == nil {
		missingFields = append(missingFields, "Permissions")
	}
	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}
	return nil
}
