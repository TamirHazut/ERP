package collection

import (
	"time"

	"erp.localhost/internal/infra/db/mongo"
	erp_errors "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging"
	auth_models "erp.localhost/internal/infra/model/auth"
	mongo_models "erp.localhost/internal/infra/model/db/mongo"
	shared_models "erp.localhost/internal/infra/model/shared"
)

type RolesCollection struct {
	collection mongo.CollectionHandler[auth_models.Role]
	logger     *logging.Logger
}

func NewRoleCollection(collection mongo.CollectionHandler[auth_models.Role]) *RolesCollection {
	logger := logging.NewLogger(shared_models.ModuleAuth)
	if collection == nil {
		collectionHandler := mongo.NewBaseCollectionHandler[auth_models.Role](string(mongo_models.RolesCollection), logger)
		if collectionHandler == nil {
			logger.Fatal("failed to create roles collection handler")
			return nil
		}
		collection = collectionHandler
	}
	return &RolesCollection{
		collection: collection,
		logger:     logger,
	}
}

func (r *RolesCollection) CreateRole(role auth_models.Role) (string, error) {
	if err := role.Validate(true); err != nil {
		return "", err
	}
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	r.logger.Debug("Creating role", "role", role)
	return r.collection.Create(role)
}

func (r *RolesCollection) GetRoleByID(tenantID, roleID string) (*auth_models.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       roleID,
	}
	r.logger.Debug("Getting role by id", "filter", filter)
	return r.findRoleByFilter(filter)
}

func (r *RolesCollection) GetRoleByName(tenantID, name string) (*auth_models.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"name":      name,
	}
	r.logger.Debug("Getting role by name", "filter", filter)
	return r.findRoleByFilter(filter)
}

func (r *RolesCollection) GetRolesByTenantID(tenantID string) ([]auth_models.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting roles by tenant id", "filter", filter)
	return r.findRolesByFilter(filter)
}

func (r *RolesCollection) GetRolesByPermissionsIDs(tenantID string, permissionsIDs []string) ([]auth_models.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"permissions": map[string]any{
			"$all": permissionsIDs,
		},
	}
	r.logger.Debug("Getting roles by permissions ids", "filter", filter)
	return r.findRolesByFilter(filter)
}

func (r *RolesCollection) UpdateRole(role auth_models.Role) error {
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
		return erp_errors.Validation(erp_errors.ValidationTryToChangeRestrictedFields, "CreatedAt")
	}
	role.UpdatedAt = time.Now()
	return r.collection.Update(filter, role)
}

func (r *RolesCollection) DeleteRole(tenantID, roleID string) error {
	if tenantID == "" || roleID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "TenantID", "RoleID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       roleID,
	}
	r.logger.Debug("Deleting role", "filter", filter)
	return r.collection.Delete(filter)
}

func (r *RolesCollection) findRoleByFilter(filter map[string]any) (*auth_models.Role, error) {
	role, err := r.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RolesCollection) findRolesByFilter(filter map[string]any) ([]auth_models.Role, error) {
	roles, err := r.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
