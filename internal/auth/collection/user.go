package collection

import (
	"erp.localhost/internal/infra/db/mongo/collection"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
)

type UserCollection struct {
	*collection.BaseCollectionHandler[authv1.User]
}

func NewUserCollection(logger logger.Logger) (*UserCollection, error) {
	collection, err := collection.NewBaseCollectionHandler[authv1.User](
		model_mongo.AuthDB,
		model_mongo.UsersCollection,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &UserCollection{
		BaseCollectionHandler: collection,
	}, nil
}
