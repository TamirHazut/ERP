package collection

import (
	"erp.localhost/infra/db/mongo/collection"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_mongo "erp.localhost/infra/model/db/mongo"
	eventv1 "erp.localhost/infra/model/event/v1"
)

type DlqCollection struct {
	*collection.BaseCollectionHandler[eventv1.DlqEntry]
}

func NewDlqCollection(logger logger.Logger) (*DlqCollection, *infra_error.AppError) {
	collection, err := collection.NewBaseCollectionHandler[eventv1.DlqEntry](
		model_mongo.EventDB,
		model_mongo.DlqCollection,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &DlqCollection{
		BaseCollectionHandler: collection,
	}, nil
}
