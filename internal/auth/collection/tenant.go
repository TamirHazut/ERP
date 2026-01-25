package collection

import (
	"erp.localhost/internal/infra/db/mongo/collection"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
)

type TenantCollection struct {
	*collection.BaseCollectionHandler[authv1.Tenant]
}

func NewTenantCollection(logger logger.Logger) (*TenantCollection, error) {
	collection, err := collection.NewBaseCollectionHandler[authv1.Tenant](
		model_mongo.AuthDB,
		model_mongo.TenantsCollection,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &TenantCollection{
		BaseCollectionHandler: collection,
	}, nil
}
