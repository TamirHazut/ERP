package collection

import (
	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
)

type PermissionCollection struct {
	*collection.BaseCollectionHandler[authv1.Permission]
}

func NewPermissionCollection(logger logger.Logger) (*PermissionCollection, *infra_error.AppError) {
	collection, err := collection.NewBaseCollectionHandler[authv1.Permission](
		model_mongo.AuthDB,
		model_mongo.PermissionsCollection,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &PermissionCollection{
		BaseCollectionHandler: collection,
	}, nil
}

func (c *PermissionCollection) FindOne(filter map[string]any) (*authv1.Permission, *infra_error.AppError) {
	permission, err := c.BaseCollectionHandler.FindOne(filter)
	if err != nil {
		return nil, c.convertError(err, filter)
	}
	return permission, nil
}

// convertError converts generic NOT_FOUND_ITEM to NOT_FOUND_PERMISSION
func (c *PermissionCollection) convertError(err error, context any) *infra_error.AppError {
	if appErr, ok := infra_error.AsAppError(err); ok {
		if appErr.Code == infra_error.NotFoundItem.Code {
			return infra_error.NotFound(
				infra_error.NotFoundPermission,
				model_auth.ResourceTypePermission,
				context,
			)
		}
		return appErr
	}
	return infra_error.Internal(infra_error.InternalUnexpectedError, err)
}
