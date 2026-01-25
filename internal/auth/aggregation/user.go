package aggregation

import (
	"erp.localhost/internal/infra/db/mongo/aggregation"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
)

// UserAggregationHandler handles user-specific aggregations
type UserAggregationHandler struct {
	*aggregation.BaseAggregationHandler[authv1.User]
}

// NewUserAggregationHandler creates a new user aggregation handler
func NewUserAggregationHandler(logger logger.Logger) (*UserAggregationHandler, error) {
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
