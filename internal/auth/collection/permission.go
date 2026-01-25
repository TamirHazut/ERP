package collection

import (
	"erp.localhost/internal/infra/db/mongo/collection"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
)

type PermissionCollection struct {
	*collection.BaseCollectionHandler[authv1.Permission]
}

func NewPermissionCollection(logger logger.Logger) (*PermissionCollection, error) {
	collection, err := collection.NewBaseCollectionHandler[authv1.Permission](
		model_mongo.AuthDB,
		model_mongo.UsersCollection,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &PermissionCollection{
		BaseCollectionHandler: collection,
	}, nil
}
