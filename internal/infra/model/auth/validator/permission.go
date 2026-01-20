package validator

import (
	infra_error "erp.localhost/internal/infra/error"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
)

func ValidatePermission(p *authv1.Permission, createOperation bool) error {
	missingFields := []string{}
	if !createOperation {
		if p.Id == "" {
			missingFields = append(missingFields, "Id")
		}
	}
	if p.TenantId == "" {
		missingFields = append(missingFields, "TenantId")
	}
	if p.Resource == "" {
		missingFields = append(missingFields, "Resource")
	}
	if p.Status == authv1.PermissionStatus_PERMISSION_STATUS_UNSPECIFIED {
		missingFields = append(missingFields, "Status")
	}
	if p.Action == "" {
		missingFields = append(missingFields, "Action")
	}
	if p.CreatedBy == "" {
		missingFields = append(missingFields, "CreatedBy")
	}
	if p.DisplayName == "" {
		missingFields = append(missingFields, "DisplayName")
	}
	if p.PermissionString == "" {
		missingFields = append(missingFields, "PermissionString")
	}
	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}
	return nil
}
