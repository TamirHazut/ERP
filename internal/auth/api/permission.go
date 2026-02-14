package api

import (
	"strings"

	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_auth "erp.localhost/infra/model/auth"
	authv1 "erp.localhost/infra/model/auth/v1"
	"erp.localhost/internal/auth/rbac"
)

// PermissionAPI provides permission read operations backed by the code-defined registry.
type PermissionAPI struct {
	verificationManager *rbac.VerificationManager
	logger              logger.Logger
}

// NewPermissionAPI creates a new PermissionAPI instance
func NewPermissionAPI(
	verificationManager *rbac.VerificationManager,
	logger logger.Logger,
) *PermissionAPI {
	return &PermissionAPI{
		verificationManager: verificationManager,
		logger:              logger,
	}
}

// GetPermissionByID retrieves a single permission from the registry by its permission string.
func (pa *PermissionAPI) GetPermissionByID(tenantID, requestorUserID, permissionID string, targetTenantID string) (*authv1.Permission, *infra_error.AppError) {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionRead)
	if err != nil {
		return nil, err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for GetPermissionByID", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return nil, err
	}

	if !model_auth.IsValidPermissionString(permissionID) {
		return nil, infra_error.NotFound(infra_error.NotFoundPermission, model_auth.ResourceTypePermission, permissionID)
	}

	// Find the definition to populate Resource/Action fields
	for _, def := range model_auth.GetAllPermissions() {
		if def.String == strings.ToLower(permissionID) {
			return defToProto(def, targetTenantID), nil
		}
	}

	return nil, infra_error.NotFound(infra_error.NotFoundPermission, model_auth.ResourceTypePermission, permissionID)
}

// ListPermissions retrieves all active permissions from the registry.
func (pa *PermissionAPI) ListPermissions(tenantID, requestorUserID string, targetTenantID string) ([]*authv1.Permission, *infra_error.AppError) {
	permissionStr, err := model_auth.CreatePermissionString(model_auth.ResourceTypePermission, model_auth.PermissionActionRead)
	if err != nil {
		return nil, err
	}

	if err := pa.verificationManager.HasPermission(tenantID, requestorUserID, permissionStr, targetTenantID); err != nil {
		pa.logger.Warn("Permission denied for ListPermissions", "tenant_id", tenantID, "user_id", requestorUserID, "permission", permissionStr)
		return nil, err
	}

	// Config service not yet built — pass all flags enabled (no gating in effect)
	defs := model_auth.GetActivePermissions(map[string]bool{})

	permissions := make([]*authv1.Permission, 0, len(defs))
	for _, def := range defs {
		permissions = append(permissions, defToProto(def, targetTenantID))
	}

	return permissions, nil
}

// defToProto converts a registry PermissionDef into an authv1.Permission proto.
func defToProto(def model_auth.PermissionDef, tenantID string) *authv1.Permission {
	return &authv1.Permission{
		Id:               def.String,
		TenantId:         tenantID,
		Resource:         def.Resource,
		Action:           def.Action,
		PermissionString: def.String,
		DisplayName:      titleCase(def.Resource) + " " + titleCase(def.Action),
		Status:           authv1.PermissionStatus_PERMISSION_STATUS_ACTIVE,
		IsDangerous:      def.String == model_auth.PermissionAdminString,
		Protected:        true,
	}
}

// titleCase capitalises the first letter of s.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
