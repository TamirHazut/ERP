package aggregation

import (
	"erp.localhost/infra/db/mongo/aggregation"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	authv1 "erp.localhost/infra/model/auth/v1"
	model_mongo "erp.localhost/infra/model/db/mongo"
)

// UserAggregationHandler handles user-specific aggregations
type UserAggregationHandler struct {
	*aggregation.BaseAggregationHandler[authv1.User]
}

// NewUserAggregationHandler creates a new user aggregation handler
func NewUserAggregationHandler(logger logger.Logger) (*UserAggregationHandler, *infra_error.AppError) {
	aggregation, err := aggregation.NewBaseAggregationHandler[authv1.User](
		model_mongo.AuthDB,
		model_mongo.UsersCollection,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &UserAggregationHandler{
		BaseAggregationHandler: aggregation,
	}, nil
}
