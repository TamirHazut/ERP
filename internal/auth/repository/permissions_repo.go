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
	repository *db.Repository[models.Permission]
	logger     *logging.Logger
}

func NewPermissionRepository(dbHandler db.DBHandler) *PermissionRepository {
	logger := logging.NewLogger(logging.ModuleAuth)
	repository := db.NewRepository[models.Permission](dbHandler, string(mongo.PermissionsCollection), logger)
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

func (r *PermissionRepository) GetPermissionByID(tenantID, permissionID string) (models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       permissionID,
	}
	r.logger.Debug("Getting permission by id", "filter", filter)
	permissions, err := r.repository.Find(filter)
	if err != nil {
		return models.Permission{}, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	if len(permissions) == 0 {
		return models.Permission{}, erp_errors.NotFound(erp_errors.NotFoundPermission, "Permission", permissionID)
	}
	return permissions[0], nil
}

func (r *PermissionRepository) GetPermissionByName(tenantID, name string) (models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"name":      name,
	}
	r.logger.Debug("Getting permission by name", "filter", filter)
	permissions, err := r.repository.Find(filter)
	if err != nil {
		return models.Permission{}, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	if len(permissions) == 0 {
		return models.Permission{}, erp_errors.NotFound(erp_errors.NotFoundPermission, "Permission", name)
	}
	return permissions[0], nil
}

func (r *PermissionRepository) GetPermissionsByTenantID(tenantID string) ([]models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting permissions by tenant id", "filter", filter)
	return r.getPermissionsByFilter(filter)
}

func (r *PermissionRepository) GetPermissionsByResource(tenantID, resource string) ([]models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"resource":  resource,
	}
	r.logger.Debug("Getting permissions by resource", "filter", filter)
	return r.getPermissionsByFilter(filter)
}

func (r *PermissionRepository) GetPermissionsByAction(tenantID, action string) ([]models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"action":    action,
	}
	r.logger.Debug("Getting permissions by action", "filter", filter)
	return r.getPermissionsByFilter(filter)
}

func (r *PermissionRepository) GetPermissionsByResourceAndAction(tenantID, resource, action string) ([]models.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"resource":  resource,
		"action":    action,
	}
	r.logger.Debug("Getting permissions by resource and action", "filter", filter)
	return r.getPermissionsByFilter(filter)
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
		return erp_errors.Validation(erp_errors.ValidationTryToChangeRestrictedFields, []string{"CreatedAt"})
	}
	permission.UpdatedAt = time.Now()
	return r.repository.Update(filter, permission)
}

func (r *PermissionRepository) DeletePermission(tenantID, permissionID string) error {
	if tenantID == "" || permissionID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, []string{"TenantID", "PermissionID"})
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       permissionID,
	}
	r.logger.Debug("Deleting permission", "filter", filter)
	return r.repository.Delete(filter)
}

func (r *PermissionRepository) getPermissionsByFilter(filter map[string]any) ([]models.Permission, error) {
	permissions, err := r.repository.Find(filter)
	if err != nil {
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return permissions, nil
}
