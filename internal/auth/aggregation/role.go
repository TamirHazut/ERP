package aggregation

import (
	"context"

	"erp.localhost/infra/db/mongo/aggregation"
	"erp.localhost/infra/db/mongo/aggregation/pipeline"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	authv1 "erp.localhost/infra/model/auth/v1"
	model_mongo "erp.localhost/infra/model/db/mongo"
)

// RoleAggregationHandler handles role-specific aggregations
type RoleAggregationHandler struct {
	*aggregation.BaseAggregationHandler[authv1.Role]
	logger logger.Logger
}

// NewRoleAggregationHandler creates a new role aggregation handler
func NewRoleAggregationHandler(logger logger.Logger) (*RoleAggregationHandler, *infra_error.AppError) {
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
func (h *RoleAggregationHandler) GetUserRoles(ctx context.Context, tenantID, userID string, fields []string) ([]*authv1.Role, *infra_error.AppError) {
	pipelineStages := pipeline.BuildUserRolesPipeline(tenantID, userID)
	return h.AggregateFrom(ctx, string(model_mongo.UsersCollection), pipelineStages, fields)
}
