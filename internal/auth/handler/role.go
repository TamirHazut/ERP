package handler

import (
	"context"
	"errors"
	"strings"

	aggregation_auth "erp.localhost/internal/auth/aggregation"
	collection_auth "erp.localhost/internal/auth/collection"
	aggregation_mongo "erp.localhost/internal/infra/db/mongo/aggregation"
	collection_mongo "erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_auth "erp.localhost/internal/infra/model/auth/validator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RoleHandler struct {
	collection  collection_mongo.CollectionHandler[authv1.Role]
	aggregation aggregation_mongo.AggregationHandler[authv1.Role]
	logger      logger.Logger
}

func NewRoleHandler(logger logger.Logger) (*RoleHandler, error) {
	collection, err := collection_auth.NewRoleCollection(logger)
	if err != nil {
		logger.Error("failed to create user collection handler", "error", err)
		return nil, err
	}
	aggregation, err := aggregation_auth.NewRoleAggregationHandler(logger)
	if err != nil {
		logger.Error("failed to create user aggregation handler", "error", err)
		return nil, err
	}
	return &RoleHandler{
		collection:  collection,
		aggregation: aggregation,
		logger:      logger,
	}, nil
}

func (r *RoleHandler) CreateRole(role *authv1.Role) (string, error) {
	if err := validator_auth.ValidateRole(role, true); err != nil {
		return "", err
	}
	role.CreatedAt = timestamppb.Now()
	role.UpdatedAt = timestamppb.Now()
	r.logger.Debug("Creating role", "role", role)
	role.Name = strings.ToLower(role.Name)
	return r.collection.Create(role)
}

func (r *RoleHandler) GetRoleByID(tenantID, roleID string) (*authv1.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       roleID,
	}
	r.logger.Debug("Getting role by id", "filter", filter)
	return r.findRoleByFilter(filter)
}

func (r *RoleHandler) GetRoleByName(tenantID, name string) (*authv1.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"name":      name,
	}
	r.logger.Debug("Getting role by name", "filter", filter)
	return r.findRoleByFilter(filter)
}

func (r *RoleHandler) GetRolesByTenantID(tenantID string) ([]*authv1.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting roles by tenant id", "filter", filter)
	return r.findRolesByFilter(filter)
}

func (r *RoleHandler) GetRolesByPermissionsIDs(tenantID string, permissionsIDs []string) ([]*authv1.Role, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"permissions": map[string]any{
			"$all": permissionsIDs,
		},
	}
	r.logger.Debug("Getting roles by permissions ids", "filter", filter)
	return r.findRolesByFilter(filter)
}

func (r *RoleHandler) UpdateRole(role *authv1.Role) error {
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

func (r *RoleHandler) DeleteRole(tenantID, roleID string) error {
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

func (r *RoleHandler) DeleteTenantRoles(tenantID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId", "RoleId")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Deleting role", "filter", filter)
	return r.collection.Delete(filter)
}

func (r *RoleHandler) findRoleByFilter(filter map[string]any) (*authv1.Role, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	role, err := r.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleHandler) findRolesByFilter(filter map[string]any) ([]*authv1.Role, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	roles, err := r.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// =====================================================
// Aggregation Methods (Optimized Query Performance)
// =====================================================

// GetRolesByIDsAggregation retrieves multiple roles by IDs using aggregation
// This replaces N sequential queries with a single batch query using $in operator
func (r *RoleHandler) GetRolesByIDsAggregation(
	tenantID string,
	roleIDs []string,
	fields []string,
) ([]*authv1.Role, error) {
	if r.aggregation == nil {
		r.logger.Warn("aggregation handler not initialized, falling back to sequential queries")
		roles := make([]*authv1.Role, 0, len(roleIDs))
		for _, id := range roleIDs {
			role, err := r.GetRoleByID(tenantID, id)
			if err != nil {
				r.logger.Debug("role not found", "id", id)
				continue
			}
			roles = append(roles, role)
		}
		return roles, nil
	}

	return r.aggregation.BatchGetByIDs(context.Background(), tenantID, roleIDs, fields)
}

// GetUserRolesAggregation retrieves all roles for a user using aggregation
// This replaces the N query pattern (1 query per role)
func (r *RoleHandler) GetUserRolesAggregation(
	tenantID, userID string,
	fields []string,
) ([]*authv1.Role, error) {
	roleAggregation, ok := r.aggregation.(*aggregation_auth.RoleAggregationHandler)
	if !ok {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("missmatched types"))
	}

	return roleAggregation.GetUserRoles(context.Background(), tenantID, userID, fields)
}
