package collection

import (
	"time"

	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
)

type RolesCollection struct {
	collection collection.CollectionHandler[model_auth.Role]
	logger     logger.Logger
}

func NewRoleCollection(collection collection.CollectionHandler[model_auth.Role], logger logger.Logger) *RolesCollection {
	return &RolesCollection{
		collection: collection,
		logger:     logger,
	}
}

func (r *RolesCollection) CreateRole(role *model_auth.Role) (string, error) {
	if err := role.Validate(true); err != nil {
		return "", err
	}
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	r.logger.Debug("Creating role", "role", role)
	return r.collection.Create(role)
}

func (r *RolesCollection) GetRoleByID(tenantID, roleID string) (*model_auth.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       roleID,
	}
	r.logger.Debug("Getting role by id", "filter", filter)
	return r.findRoleByFilter(filter)
}

func (r *RolesCollection) GetRoleByName(tenantID, name string) (*model_auth.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"name":      name,
	}
	r.logger.Debug("Getting role by name", "filter", filter)
	return r.findRoleByFilter(filter)
}

func (r *RolesCollection) GetRolesByTenantID(tenantID string) ([]*model_auth.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting roles by tenant id", "filter", filter)
	return r.findRolesByFilter(filter)
}

func (r *RolesCollection) GetRolesByPermissionsIDs(tenantID string, permissionsIDs []string) ([]*model_auth.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"permissions": map[string]any{
			"$all": permissionsIDs,
		},
	}
	r.logger.Debug("Getting roles by permissions ids", "filter", filter)
	return r.findRolesByFilter(filter)
}

func (r *RolesCollection) UpdateRole(role *model_auth.Role) error {
	if err := role.Validate(false); err != nil {
		return err
	}
	filter := map[string]any{
		"tenant_id": role.TenantID,
		"_id":       role.ID,
	}
	r.logger.Debug("Updating role", "role", role)
	currentRole, err := r.GetRoleByID(role.TenantID, role.ID.Hex())
	if err != nil {
		return err
	}
	if role.CreatedAt != currentRole.CreatedAt {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields, "CreatedAt")
	}
	role.UpdatedAt = time.Now()
	return r.collection.Update(filter, role)
}

func (r *RolesCollection) DeleteRole(tenantID, roleID string) error {
	if tenantID == "" || roleID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantID", "RoleID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       roleID,
	}
	r.logger.Debug("Deleting role", "filter", filter)
	return r.collection.Delete(filter)
}

func (r *RolesCollection) findRoleByFilter(filter map[string]any) (*model_auth.Role, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	role, err := r.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RolesCollection) findRolesByFilter(filter map[string]any) ([]*model_auth.Role, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	roles, err := r.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
