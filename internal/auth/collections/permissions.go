package repository

import (
	"time"

	"erp.localhost/internal/auth/models"
	"erp.localhost/internal/db"
	"erp.localhost/internal/db/mongo"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

type PermissionRepository struct {
	repository *mongo.CollectionHandler[models.Permission]
	logger     *logging.Logger
}

func NewPermissionRepository(dbHandler db.DBHandler) *PermissionRepository {
	logger := logging.NewLogger(logging.ModuleAuth)
	repository := mongo.NewCollectionHandler[models.Permission](dbHandler, string(mongo.PermissionsCollection), logger)
	return &PermissionRepository{
		repository: repository,
		logger:     logger,
	}
}

func (r *PermissionRepository) CreatePermission(permission models.Permission) (string, error) {
	if err := permission.Validate(true); err != nil {
		return "", err
	}
	permission.CreatedAt = time.Now()
	permission.UpdatedAt = time.Now()
	r.logger.Debug("Creating permission", "permission", permission)
	return r.repository.Create(permission)
}

func (r *PermissionRepository) GetPermissionByID(tenantID, permissionID string) (*models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       permissionID,
	}
	r.logger.Debug("Getting permission by id", "filter", filter)
	return r.findPermissionByFilter(filter)
}

func (r *PermissionRepository) GetPermissionByName(tenantID, name string) (*models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"name":      name,
	}
	r.logger.Debug("Getting permission by name", "filter", filter)
	return r.findPermissionByFilter(filter)
}

func (r *PermissionRepository) GetPermissionsByTenantID(tenantID string) ([]models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting permissions by tenant id", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionRepository) GetPermissionsByResource(tenantID, resource string) ([]models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"resource":  resource,
	}
	r.logger.Debug("Getting permissions by resource", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionRepository) GetPermissionsByAction(tenantID, action string) ([]models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"action":    action,
	}
	r.logger.Debug("Getting permissions by action", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionRepository) GetPermissionsByResourceAndAction(tenantID, resource, action string) ([]models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"resource":  resource,
		"action":    action,
	}
	r.logger.Debug("Getting permissions by resource and action", "filter", filter)
	return r.findPermissionsByFilter(filter)
}

func (r *PermissionRepository) UpdatePermission(permission models.Permission) error {
	if err := permission.Validate(false); err != nil {
		return err
	}
	filter := map[string]any{
		"tenant_id": permission.TenantID,
		"_id":       permission.ID,
	}
	r.logger.Debug("Updating permission", "permission", permission)
	currentPermission, err := r.GetPermissionByID(permission.TenantID, permission.ID.String())
	if err != nil {
		return err
	}
	if permission.CreatedAt != currentPermission.CreatedAt {
		return erp_errors.Validation(erp_errors.ValidationTryToChangeRestrictedFields, "CreatedAt")
	}
	permission.UpdatedAt = time.Now()
	return r.repository.Update(filter, permission)
}

func (r *PermissionRepository) DeletePermission(tenantID, permissionID string) error {
	if tenantID == "" || permissionID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "TenantID", "PermissionID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       permissionID,
	}
	r.logger.Debug("Deleting permission", "filter", filter)
	return r.repository.Delete(filter)
}

func (r *PermissionRepository) findPermissionByFilter(filter map[string]any) (*models.Permission, error) {
	permission, err := r.repository.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func (r *PermissionRepository) findPermissionsByFilter(filter map[string]any) ([]models.Permission, error) {
	permissions, err := r.repository.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
