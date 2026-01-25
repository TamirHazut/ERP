package aggregation

import (
	"context"

	"erp.localhost/internal/infra/db/mongo/aggregation"
	"erp.localhost/internal/infra/db/mongo/aggregation/pipeline"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
)

// PermissionAggregationHandler handles permission-specific aggregations
type PermissionAggregationHandler struct {
	*aggregation.BaseAggregationHandler[authv1.Permission]
}

// NewPermissionAggregationHandler creates a new permission aggregation handler
func NewPermissionAggregationHandler(logger logger.Logger) (*PermissionAggregationHandler, error) {
	aggregation, err := aggregation.NewBaseAggregationHandler[authv1.Permission](
		model_mongo.AuthDB,
		model_mongo.PermissionsCollection,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &PermissionAggregationHandler{
		BaseAggregationHandler: aggregation,
	}, nil
}

// GetUserPermissions retrieves all permissions for a user using aggregation
// This replaces the N+1 query pattern (1 user + N roles + M permissions per role)
func (h *PermissionAggregationHandler) GetUserPermissions(
	ctx context.Context,
	tenantID, userID string,
	fields []string,
) ([]*authv1.Permission, error) {
	pipelineStages := pipeline.BuildUserPermissionsPipeline(tenantID, userID)
	return h.Aggregate(ctx, pipelineStages, fields)
}
