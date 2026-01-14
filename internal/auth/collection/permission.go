package collection

import (
	"time"

	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
)

type PermissionsCollection struct {
	collection collection.CollectionHandler[model_auth.Permission]
	logger     logger.Logger
}

func NewPermissionCollection(collection collection.CollectionHandler[model_auth.Permission], logger logger.Logger) *PermissionsCollection {
	if collection == nil {
		return nil
	}
	return &PermissionsCollection{
		collection: collection,
		logger:     logger,
	}
}

func (r *PermissionsCollection) CreatePermission(permission *model_auth.Permission) (string, error) {
	if err := permission.Validate(true); err != nil {
		return "", err
	}
	permission.CreatedAt = time.Now()
	permission.UpdatedAt = time.Now()
	r.logger.Debug("Creating permission", "permission", permission)
	return r.collection.Create(permission)
}

func (r *PermissionsCollection) GetPermissionByID(tenantID, permissionID string) (*model_auth.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       permissionID,
	}
	r.logger.Debug("Getting permission by id", "filter", filter)
	return r.findPermissionByFilter(filter)
}

func (r *PermissionsCollection) GetPermissionByName(tenantID, name string) (*model_auth.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"name":      name,
	}
	r.logger.Debug("Getting permission by name", "filter", filter)
	return r.findPermissionByFilter(filter)
}

func (r *PermissionsCollection) GetPermissionsByTenantID(tenantID string) ([]*model_auth.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting permissions by tenant id", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionsCollection) GetPermissionsByResource(tenantID, resource string) ([]*model_auth.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"resource":  resource,
	}
	r.logger.Debug("Getting permissions by resource", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionsCollection) GetPermissionsByAction(tenantID, action string) ([]*model_auth.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"action":    action,
	}
	r.logger.Debug("Getting permissions by action", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionsCollection) GetPermissionsByResourceAndAction(tenantID, resource, action string) ([]*model_auth.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"resource":  resource,
		"action":    action,
	}
	r.logger.Debug("Getting permissions by resource and action", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionsCollection) UpdatePermission(permission *model_auth.Permission) error {
	if err := permission.Validate(false); err != nil {
		return err
	}
	filter := map[string]any{
		"tenant_id": permission.TenantID,
		"_id":       permission.ID,
	}
	r.logger.Debug("Updating permission", "permission", permission)
	currentPermission, err := r.GetPermissionByID(permission.TenantID, permission.ID.Hex())
	if err != nil {
		return err
	}
	if permission.CreatedAt != currentPermission.CreatedAt {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields, "CreatedAt")
	}
	permission.UpdatedAt = time.Now()
	return r.collection.Update(filter, permission)
}

func (r *PermissionsCollection) DeletePermission(tenantID, permissionID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	if permissionID != "" {
		filter["_id"] = permissionID
	}
	r.logger.Debug("Deleting permission", "filter", filter)
	return r.collection.Delete(filter)
}

func (r *PermissionsCollection) findPermissionByFilter(filter map[string]any) (*model_auth.Permission, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	permission, err := r.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func (r *PermissionsCollection) findPermissionsByFilter(filter map[string]any) ([]*model_auth.Permission, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	permissions, err := r.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
