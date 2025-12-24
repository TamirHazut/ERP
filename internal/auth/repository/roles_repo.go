package repository

import (
	"time"

	"erp.localhost/internal/auth/models"
	"erp.localhost/internal/db"
	"erp.localhost/internal/db/mongo"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

type RoleRepository struct {
	repository *db.Repository[models.Role]
	logger     *logging.Logger
}

func NewRoleRepository(dbHandler db.DBHandler) *RoleRepository {
	logger := logging.NewLogger(logging.ModuleAuth)
	repository := db.NewRepository[models.Role](dbHandler, string(mongo.RolesCollection), logger)
	return &RoleRepository{
		repository: repository,
		logger:     logger,
	}
}

func (r *RoleRepository) CreateRole(role models.Role) (string, error) {
	if err := role.Validate(true); err != nil {
		return "", err
	}
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	r.logger.Debug("Creating role", "role", role)
	return r.repository.Create(role)
}

func (r *RoleRepository) GetRoleByID(tenantID, roleID string) (models.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       roleID,
	}
	r.logger.Debug("Getting role by id", "filter", filter)
	roles, err := r.repository.Find(filter)
	if err != nil {
		return models.Role{}, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	if len(roles) == 0 {
		return models.Role{}, erp_errors.NotFound(erp_errors.NotFoundRole, "Role", roleID)
	}
	return roles[0], nil
}

func (r *RoleRepository) GetRoleByName(tenantID, name string) (models.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"name":      name,
	}
	r.logger.Debug("Getting role by name", "filter", filter)
	roles, err := r.repository.Find(filter)
	if err != nil {
		return models.Role{}, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	if len(roles) == 0 {
		return models.Role{}, erp_errors.NotFound(erp_errors.NotFoundRole, "Role", name)
	}
	return roles[0], nil
}

func (r *RoleRepository) GetRolesByTenantID(tenantID string) ([]models.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting roles by tenant id", "filter", filter)
	roles, err := r.repository.Find(filter)
	if err != nil {
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return roles, nil
}

func (r *RoleRepository) GetRolesByPermissionsIDs(tenantID string, permissionsIDs []string) ([]models.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"permissions": map[string]any{
			"$all": permissionsIDs,
		},
	}
	r.logger.Debug("Getting roles by permissions ids", "filter", filter)
	roles, err := r.repository.Find(filter)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepository) UpdateRole(role models.Role) error {
	if err := role.Validate(false); err != nil {
		return err
	}
	filter := map[string]any{
		"tenant_id": role.TenantID,
		"_id":       role.ID,
	}
	r.logger.Debug("Updating role", "role", role)
	currentRole, err := r.GetRoleByID(role.TenantID, role.ID.String())
	if err != nil {
		return err
	}
	if role.CreatedAt != currentRole.CreatedAt {
		return erp_errors.Validation(erp_errors.ValidationTryToChangeRestrictedFields, []string{"CreatedAt"})
	}
	role.UpdatedAt = time.Now()
	return r.repository.Update(filter, role)
}

func (r *RoleRepository) DeleteRole(tenantID, roleID string) error {
	if tenantID == "" || roleID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, []string{"TenantID", "RoleID"})
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       roleID,
	}
	r.logger.Debug("Deleting role", "filter", filter)
	return r.repository.Delete(filter)
}
