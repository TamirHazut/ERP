package collection

import (
	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_auth "erp.localhost/internal/infra/model/auth/validator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PermissionsCollection struct {
	collection collection.CollectionHandler[authv1.Permission]
	logger     logger.Logger
}

func NewPermissionCollection(collection collection.CollectionHandler[authv1.Permission], logger logger.Logger) *PermissionsCollection {
	if collection == nil {
		return nil
	}
	return &PermissionsCollection{
		collection: collection,
		logger:     logger,
	}
}

func (r *PermissionsCollection) CreatePermission(permission *authv1.Permission) (string, error) {
	if err := validator_auth.ValidatePermission(permission, true); err != nil {
		return "", err
	}
	permission.CreatedAt = timestamppb.Now()
	permission.UpdatedAt = timestamppb.Now()
	r.logger.Debug("Creating permission", "permission", permission)
	return r.collection.Create(permission)
}

func (r *PermissionsCollection) GetPermissionByID(tenantID, permissionID string) (*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       permissionID,
	}
	r.logger.Debug("Getting permission by id", "filter", filter)
	return r.findPermissionByFilter(filter)
}

func (r *PermissionsCollection) GetPermissionByName(tenantID, name string) (*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"name":      name,
	}
	r.logger.Debug("Getting permission by name", "filter", filter)
	return r.findPermissionByFilter(filter)
}

func (r *PermissionsCollection) GetPermissionsByTenantID(tenantID string) ([]*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting permissions by tenant id", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionsCollection) GetPermissionsByResource(tenantID, resource string) ([]*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"resource":  resource,
	}
	r.logger.Debug("Getting permissions by resource", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionsCollection) GetPermissionsByAction(tenantID, action string) ([]*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"action":    action,
	}
	r.logger.Debug("Getting permissions by action", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionsCollection) GetPermissionsByResourceAndAction(tenantID, resource, action string) ([]*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"resource":  resource,
		"action":    action,
	}
	r.logger.Debug("Getting permissions by resource and action", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionsCollection) UpdatePermission(permission *authv1.Permission) error {
	if err := validator_auth.ValidatePermission(permission, false); err != nil {
		return err
	}
	filter := map[string]any{
		"tenant_id": permission.TenantId,
		"_id":       permission.Id,
	}
	r.logger.Debug("Updating permission", "permission", permission)
	currentPermission, err := r.GetPermissionByID(permission.TenantId, permission.Id)
	if err != nil {
		return err
	}
	if permission.CreatedAt != currentPermission.CreatedAt {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields, "CreatedAt")
	}
	permission.UpdatedAt = timestamppb.Now()
	return r.collection.Update(filter, permission)
}

func (r *PermissionsCollection) DeletePermission(tenantID, permissionID string) error {
	if tenantID == "" || permissionID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId", "PermissionID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       permissionID,
	}
	r.logger.Debug("Deleting permission", "filter", filter)
	return r.collection.Delete(filter)
}

func (r *PermissionsCollection) DeleteTenantPermissions(tenantID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Deleting permission", "filter", filter)
	return r.collection.Delete(filter)
}

func (r *PermissionsCollection) findPermissionByFilter(filter map[string]any) (*authv1.Permission, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	permission, err := r.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func (r *PermissionsCollection) findPermissionsByFilter(filter map[string]any) ([]*authv1.Permission, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	permissions, err := r.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
