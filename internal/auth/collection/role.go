package collection

import (
	"erp.localhost/infra/db/mongo/collection"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_auth "erp.localhost/infra/model/auth"
	authv1 "erp.localhost/infra/model/auth/v1"
	model_mongo "erp.localhost/infra/model/db/mongo"
)

type RoleCollection struct {
	*collection.BaseCollectionHandler[authv1.Role]
}

func NewRoleCollection(logger logger.Logger) (*RoleCollection, *infra_error.AppError) {
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

func (c *RoleCollection) FindOne(filter map[string]any) (*authv1.Role, *infra_error.AppError) {
	role, err := c.BaseCollectionHandler.FindOne(filter)
	if err != nil {
		return nil, c.convertError(err, filter)
	}
	return role, nil
}

// convertError converts generic NOT_FOUND_ITEM to NOT_FOUND_ROLE
func (c *RoleCollection) convertError(err error, context any) *infra_error.AppError {
	if appErr, ok := infra_error.AsAppError(err); ok {
		if appErr.Code == infra_error.NotFoundItem.Code {
			return infra_error.NotFound(
				infra_error.NotFoundRole,
				model_auth.ResourceTypeRole,
				context,
			)
		}
		return appErr
	}
	return infra_error.Internal(infra_error.InternalUnexpectedError, err)
}
