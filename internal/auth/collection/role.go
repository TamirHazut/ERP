package collection

import (
	"erp.localhost/internal/infra/db/mongo/collection"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
)

type RoleCollection struct {
	*collection.BaseCollectionHandler[authv1.Role]
}

func NewRoleCollection(logger logger.Logger) (*RoleCollection, error) {
	collection, err := collection.NewBaseCollectionHandler[authv1.Role](
		model_mongo.AuthDB,
		model_mongo.RolesCollection,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &RoleCollection{
		BaseCollectionHandler: collection,
	}, nil
}
