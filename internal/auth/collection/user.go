package collection

import (
	"erp.localhost/infra/db/mongo/collection"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_auth "erp.localhost/infra/model/auth"
	authv1 "erp.localhost/infra/model/auth/v1"
	model_mongo "erp.localhost/infra/model/db/mongo"
)

type UserCollection struct {
	*collection.BaseCollectionHandler[authv1.User]
}

func NewUserCollection(logger logger.Logger) (*UserCollection, *infra_error.AppError) {
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

func (c *UserCollection) FindOne(filter map[string]any) (*authv1.User, *infra_error.AppError) {
	user, err := c.BaseCollectionHandler.FindOne(filter)
	if err != nil {
		return nil, c.convertError(err, filter)
	}
	return user, nil
}

// convertError converts generic NOT_FOUND_ITEM to NOT_FOUND_USER
func (c *UserCollection) convertError(err error, context any) *infra_error.AppError {
	if appErr, ok := infra_error.AsAppError(err); ok {
		if appErr.Code == infra_error.NotFoundItem.Code {
			return infra_error.NotFound(
				infra_error.NotFoundUser,
				model_auth.ResourceTypeUser,
				context,
			)
		}
		return appErr
	}
	return infra_error.Internal(infra_error.InternalUnexpectedError, err)
}
