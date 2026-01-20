package collection

import (
	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_auth "erp.localhost/internal/infra/model/auth/validator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RolesCollection struct {
	collection collection.CollectionHandler[authv1.Role]
	logger     logger.Logger
}

func NewRoleCollection(collection collection.CollectionHandler[authv1.Role], logger logger.Logger) *RolesCollection {
	if collection == nil {
		return nil
	}
	return &RolesCollection{
		collection: collection,
		logger:     logger,
	}
}

func (r *RolesCollection) CreateRole(role *authv1.Role) (string, error) {
	if err := validator_auth.ValidateRole(role, true); err != nil {
		return "", err
	}
	role.CreatedAt = timestamppb.Now()
	role.UpdatedAt = timestamppb.Now()
	r.logger.Debug("Creating role", "role", role)
	return r.collection.Create(role)
}

func (r *RolesCollection) GetRoleByID(tenantID, roleID string) (*authv1.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       roleID,
	}
	r.logger.Debug("Getting role by id", "filter", filter)
	return r.findRoleByFilter(filter)
}

func (r *RolesCollection) GetRoleByName(tenantID, name string) (*authv1.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"name":      name,
	}
	r.logger.Debug("Getting role by name", "filter", filter)
	return r.findRoleByFilter(filter)
}

func (r *RolesCollection) GetRolesByTenantID(tenantID string) ([]*authv1.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting roles by tenant id", "filter", filter)
	return r.findRolesByFilter(filter)
}

func (r *RolesCollection) GetRolesByPermissionsIDs(tenantID string, permissionsIDs []string) ([]*authv1.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"permissions": map[string]any{
			"$all": permissionsIDs,
		},
	}
	r.logger.Debug("Getting roles by permissions ids", "filter", filter)
	return r.findRolesByFilter(filter)
}

func (r *RolesCollection) UpdateRole(role *authv1.Role) error {
	if err := validator_auth.ValidateRole(role, false); err != nil {
		return err
	}
	filter := map[string]any{
		"tenant_id": role.TenantId,
		"_id":       role.Id,
	}
	r.logger.Debug("Updating role", "role", role)
	currentRole, err := r.GetRoleByID(role.TenantId, role.Id)
	if err != nil {
		return err
	}
	if role.CreatedAt != currentRole.CreatedAt {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields, "CreatedAt")
	}
	role.UpdatedAt = timestamppb.Now()
	return r.collection.Update(filter, role)
}

func (r *RolesCollection) DeleteRole(tenantID, roleID string) error {
	if tenantID == "" || roleID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId", "RoleId")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       roleID,
	}
	r.logger.Debug("Deleting role", "filter", filter)
	return r.collection.Delete(filter)
}

func (r *RolesCollection) DeleteTenantRoles(tenantID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId", "RoleId")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Deleting role", "filter", filter)
	return r.collection.Delete(filter)
}

func (r *RolesCollection) findRoleByFilter(filter map[string]any) (*authv1.Role, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	role, err := r.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RolesCollection) findRolesByFilter(filter map[string]any) ([]*authv1.Role, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	roles, err := r.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
