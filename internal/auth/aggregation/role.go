package aggregation

import (
	"context"

	"erp.localhost/internal/infra/db/mongo/aggregation"
	"erp.localhost/internal/infra/db/mongo/aggregation/pipeline"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
)

// RoleAggregationHandler handles role-specific aggregations
type RoleAggregationHandler struct {
	*aggregation.BaseAggregationHandler[authv1.Role]
	logger logger.Logger
}

// NewRoleAggregationHandler creates a new role aggregation handler
func NewRoleAggregationHandler(logger logger.Logger) (*RoleAggregationHandler, error) {
	aggregation, err := aggregation.NewBaseAggregationHandler[authv1.Role](
		model_mongo.AuthDB,
		model_mongo.RolesCollection,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &RoleAggregationHandler{
		BaseAggregationHandler: aggregation,
		logger:                 logger,
	}, nil
}

// GetUserRoles retrieves all roles for a user using aggregation
// This replaces the N query pattern (1 query per role)
func (h *RoleAggregationHandler) GetUserRoles(
	ctx context.Context,
	tenantID, userID string,
	fields []string,
) ([]*authv1.Role, error) {
	pipelineStages := pipeline.BuildUserRolesPipeline(tenantID, userID)
	return h.Aggregate(ctx, pipelineStages, fields)
}

// GetRoleWithPermissionsAggregation retrieves a role with all its permissions using aggregation
// This replaces the 1 + N pattern (1 role + N permissions)
func (h *RoleAggregationHandler) GetRoleWithPermissionsAggregation(
	tenantID, roleID string,
	fields []string,
) ([]*authv1.Permission, error) {
	// Create permission aggregation handler to get permissions for this role
	permHandler, err := NewPermissionAggregationHandler(h.logger)
	if err != nil {
		return nil, err
	}
	pipelineStages := pipeline.BuildRolePermissionsPipeline(tenantID, roleID)
	return permHandler.Aggregate(context.Background(), pipelineStages, fields)
}
