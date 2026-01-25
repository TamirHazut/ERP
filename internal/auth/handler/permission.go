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

type PermissionHandler struct {
	collection  collection_mongo.CollectionHandler[authv1.Permission]
	aggregation aggregation_mongo.AggregationHandler[authv1.Permission]
	logger      logger.Logger
}

func NewPermissionHandler(logger logger.Logger) (*PermissionHandler, error) {
	collection, err := collection_auth.NewPermissionCollection(logger)
	if err != nil {
		logger.Error("failed to create user collection handler", "error", err)
		return nil, err
	}
	aggregation, err := aggregation_auth.NewPermissionAggregationHandler(logger)
	if err != nil {
		logger.Error("failed to create user aggregation handler", "error", err)
		return nil, err
	}
	return &PermissionHandler{
		collection:  collection,
		aggregation: aggregation,
		logger:      logger,
	}, nil
}

func (p *PermissionHandler) CreatePermission(permission *authv1.Permission) (string, error) {
	if err := validator_auth.ValidatePermission(permission, true); err != nil {
		return "", err
	}
	permission.CreatedAt = timestamppb.Now()
	permission.UpdatedAt = timestamppb.Now()
	p.logger.Debug("Creating permission", "permission", permission)
	permission.DisplayName = strings.ToLower(permission.DisplayName)
	permission.PermissionString = strings.ToLower(permission.PermissionString)
	return p.collection.Create(permission)
}

func (p *PermissionHandler) GetPermissionByID(tenantID, permissionID string) (*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       permissionID,
	}
	p.logger.Debug("Getting permission by id", "filter", filter)
	return p.findPermissionByFilter(filter)
}

func (p *PermissionHandler) GetPermissionByName(tenantID, name string) (*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id":         tenantID,
		"permission_string": name,
	}
	p.logger.Debug("Getting permission by name", "filter", filter)
	return p.findPermissionByFilter(filter)
}

func (p *PermissionHandler) GetPermissionsByTenantID(tenantID string) ([]*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	p.logger.Debug("Getting permissions by tenant id", "filter", filter)
	return p.findPermissionsByFilter(filter)
}

func (p *PermissionHandler) GetPermissionsByResource(tenantID, resource string) ([]*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"resource":  resource,
	}
	p.logger.Debug("Getting permissions by resource", "filter", filter)
	return p.findPermissionsByFilter(filter)
}

func (p *PermissionHandler) GetPermissionsByAction(tenantID, action string) ([]*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"action":    action,
	}
	p.logger.Debug("Getting permissions by action", "filter", filter)
	return p.findPermissionsByFilter(filter)
}

func (p *PermissionHandler) GetPermissionsByResourceAndAction(tenantID, resource, action string) ([]*authv1.Permission, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"resource":  resource,
		"action":    action,
	}
	p.logger.Debug("Getting permissions by resource and action", "filter", filter)
	return p.findPermissionsByFilter(filter)
}

func (p *PermissionHandler) UpdatePermission(permission *authv1.Permission) error {
	if err := validator_auth.ValidatePermission(permission, false); err != nil {
		return err
	}
	filter := map[string]any{
		"tenant_id": permission.TenantId,
		"_id":       permission.Id,
	}
	p.logger.Debug("Updating permission", "permission", permission)
	currentPermission, err := p.GetPermissionByID(permission.TenantId, permission.Id)
	if err != nil {
		return err
	}
	if permission.CreatedAt != currentPermission.CreatedAt {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields, "CreatedAt")
	}
	permission.UpdatedAt = timestamppb.Now()
	return p.collection.Update(filter, permission)
}

func (p *PermissionHandler) DeletePermission(tenantID, permissionID string) error {
	if tenantID == "" || permissionID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId", "PermissionID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       permissionID,
	}
	p.logger.Debug("Deleting permission", "filter", filter)
	return p.collection.Delete(filter)
}

func (p *PermissionHandler) DeleteTenantPermissions(tenantID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	p.logger.Debug("Deleting permission", "filter", filter)
	return p.collection.Delete(filter)
}

func (p *PermissionHandler) findPermissionByFilter(filter map[string]any) (*authv1.Permission, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	permission, err := p.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func (p *PermissionHandler) findPermissionsByFilter(filter map[string]any) ([]*authv1.Permission, error) {
	if tenant_id, ok := filter["tenant_id"]; !ok || tenant_id == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	permissions, err := p.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// =====================================================
// Aggregation Methods (Optimized Query Performance)
// =====================================================

// GetPermissionsByIDsAggregation retrieves multiple permissions by IDs using aggregation
// This replaces N sequential queries with a single batch query using $in operator
func (p *PermissionHandler) GetPermissionsByIDsAggregation(
	tenantID string,
	permissionIDs []string,
	fields []string,
) ([]*authv1.Permission, error) {
	if p.aggregation == nil {
		p.logger.Warn("aggregation handler not initialized, falling back to sequential queries")
		// Fallback to sequential queries if aggregation handler not available
		permissions := make([]*authv1.Permission, 0, len(permissionIDs))
		for _, id := range permissionIDs {
			perm, err := p.GetPermissionByID(tenantID, id)
			if err != nil {
				p.logger.Debug("permission not found", "id", id)
				continue
			}
			permissions = append(permissions, perm)
		}
		return permissions, nil
	}

	return p.aggregation.BatchGetByIDs(context.Background(), tenantID, permissionIDs, fields)
}

// GetUserPermissionsAggregation retrieves all permissions for a user using aggregation
// This replaces the N+1 query pattern (1 user + N roles + M permissions per role)
// with a single aggregation pipeline
func (p *PermissionHandler) GetUserPermissionsAggregation(
	tenantID, userID string,
	fields []string,
) ([]*authv1.Permission, error) {
	if p.aggregation == nil {
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, nil)
	}
	permissionAggregation, ok := p.aggregation.(*aggregation_auth.PermissionAggregationHandler)
	if !ok {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("missmatched types"))
	}
	return permissionAggregation.GetUserPermissions(context.Background(), tenantID, userID, fields)
}
