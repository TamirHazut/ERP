package validator

import (
	"strings"

	infra_error "erp.localhost/internal/infra/error"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
)

func ValidatePermission(p *authv1.Permission, createOperation bool) *infra_error.AppError {
	if p == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "permission")
	}
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
	return ValidatePermissionString(p.PermissionString, p.Resource, p.Action)
}

func ValidatePermissionString(permissionString, resourceType, action string) *infra_error.AppError {
	if !model_auth.IsValidResourceType(resourceType) {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "resource")
	}
	if !model_auth.IsValidPermissionAction(action) {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "action")
	}
	if !model_auth.IsValidPermission(permissionString) {
		return infra_error.Validation(infra_error.ValidationInvalidFormat, "permission")
	}
	permissionSplt := strings.Split(permissionString, ":")
	if permissionSplt[0] != resourceType || permissionSplt[1] != action {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "permission")
	}
	return nil
}
