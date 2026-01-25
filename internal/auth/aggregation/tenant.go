package aggregation

import (
	"erp.localhost/internal/infra/db/mongo/aggregation"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
)

// TenantAggregationHandler handles tenant-specific aggregations
type TenantAggregationHandler struct {
	*aggregation.BaseAggregationHandler[authv1.Tenant]
}

// NewTenantAggregationHandler creates a new tenant aggregation handler
func NewTenantAggregationHandler(logger logger.Logger) (*TenantAggregationHandler, error) {
	aggregation, err := aggregation.NewBaseAggregationHandler[authv1.Tenant](
		model_mongo.AuthDB,
		model_mongo.TenantsCollection,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &TenantAggregationHandler{
		BaseAggregationHandler: aggregation,
	}, nil
}
