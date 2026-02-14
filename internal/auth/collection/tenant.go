package collection

import (
	"erp.localhost/infra/db/mongo/collection"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_auth "erp.localhost/infra/model/auth"
	authv1 "erp.localhost/infra/model/auth/v1"
	model_mongo "erp.localhost/infra/model/db/mongo"
)

type TenantCollection struct {
	*collection.BaseCollectionHandler[authv1.Tenant]
}

func NewTenantCollection(logger logger.Logger) (*TenantCollection, *infra_error.AppError) {
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

func (c *TenantCollection) FindOne(filter map[string]any) (*authv1.Tenant, *infra_error.AppError) {
	tenant, err := c.BaseCollectionHandler.FindOne(filter)
	if err != nil {
		return nil, c.convertError(err, filter)
	}
	return tenant, nil
}

// convertError converts generic NOT_FOUND_ITEM to NOT_FOUND_TENANT
func (c *TenantCollection) convertError(err error, context any) *infra_error.AppError {
	if appErr, ok := infra_error.AsAppError(err); ok {
		if appErr.Code == infra_error.NotFoundItem.Code {
			return infra_error.NotFound(
				infra_error.NotFoundTenant,
				model_auth.ResourceTypeTenant,
				context,
			)
		}
		return appErr
	}
	return infra_error.Internal(infra_error.InternalUnexpectedError, err)
}
