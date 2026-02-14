package collection

import (
	db "erp.localhost/infra/db"
	"erp.localhost/infra/db/mongo"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_mongo "erp.localhost/infra/model/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

//go:generate mockgen -destination=mock/mock_collection_handler.go -package=mock erp.localhost/infra/db/mongo/collection CollectionHandler
type CollectionHandler[T any] interface {
	Create(item *T) (string, *infra_error.AppError)
	FindOne(filter map[string]any) (*T, *infra_error.AppError)
	FindAll(filter map[string]any) ([]*T, *infra_error.AppError)
	Update(filter map[string]any, item *T) *infra_error.AppError
	Delete(filter map[string]any) *infra_error.AppError
}

// Generic Collection
type BaseCollectionHandler[T any] struct {
	dbHandler  db.DBHandler
	collection string
	logger     logger.Logger
}

func NewBaseCollectionHandler[T any](dbName model_mongo.DBName, collection model_mongo.Collection, logger logger.Logger) (*BaseCollectionHandler[T], *infra_error.AppError) {
	if logger == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "logger")
	}
	dbHandler, err := mongo.NewMongoDBManager(dbName, logger)
	if err != nil {
		return nil, err
	}
	collectionHandler := &BaseCollectionHandler[T]{
		dbHandler:  dbHandler,
		collection: string(collection),
		logger:     logger,
	}
	if err := collectionHandler.createCollectionInDBIfNotExists(); err != nil {
		logger.Error(err.Error(), "collection", collection, "error", err)
		return nil, err
	}
	return collectionHandler, nil
}

func (r *BaseCollectionHandler[T]) createCollectionInDBIfNotExists() *infra_error.AppError {
	if dbHandler, ok := r.dbHandler.(*mongo.MongoDBManager); ok {
		return dbHandler.CreateCollectionInDBIfNotExists(r.collection)
	}
	return nil
}

func (r *BaseCollectionHandler[T]) Create(item *T) (string, *infra_error.AppError) {
	r.logger.Debug("Creating item", "collection", r.collection)
	id, err := r.dbHandler.Create(r.collection, item)
	if err != nil {
		r.logger.Error("failed to create item", "collection", r.collection, "error", err)
		return "", err
	}
	return id, nil
}

func (r *BaseCollectionHandler[T]) FindOne(filter map[string]any) (*T, *infra_error.AppError) {
	r.logger.Debug("Finding item", "collection", r.collection, "filter", filter)
	result := new(T)
	err := r.dbHandler.FindOne(r.collection, filter, result)
	if err != nil {
		r.logger.Error("failed to find item", "collection", r.collection, "filter", filter, "error", err)
		return nil, err
	}

	return result, nil
}

func (r *BaseCollectionHandler[T]) FindAll(filter map[string]any) ([]*T, *infra_error.AppError) {
	if filter == nil {
		r.logger.Debug("nil filter found", "collection", r.collection)
		filter = make(map[string]any)
	}
	r.logger.Debug("Finding items", "collection", r.collection, "filter", filter)
	result := make([]*T, 0)
	err := r.dbHandler.FindAll(r.collection, filter, &result)
	if err != nil {
		r.logger.Error("failed to find items", "collection", r.collection, "filter", filter, "error", err)
		return nil, err
	}
	return result, nil
}

func (r *BaseCollectionHandler[T]) Update(filter map[string]any, item *T) *infra_error.AppError {
	r.logger.Debug("Updating item", "collection", r.collection, "filter", filter, "item", item)

	// Convert item to BSON map and exclude _id field (immutable in MongoDB)
	updateData, err := r.prepareUpdateData(item)
	if err != nil {
		r.logger.Error("failed to prepare update item", "collection", r.collection, "filter", filter, "item", item, "error", err)
		return err
	}

	if err := r.dbHandler.Update(r.collection, filter, updateData); err != nil {
		r.logger.Error("failed to update item", "collection", r.collection, "filter", filter, "item", item, "error", err)
		return err
	}
	return nil
}

// prepareUpdateData converts item to BSON map and excludes the _id field
func (r *BaseCollectionHandler[T]) prepareUpdateData(item *T) (bson.M, *infra_error.AppError) {
	// Marshal to BSON bytes
	bytes, err := bson.Marshal(item)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}

	// Unmarshal to bson.M
	var updateMap bson.M
	if err := bson.Unmarshal(bytes, &updateMap); err != nil {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}

	// Remove _id field (immutable in MongoDB)
	delete(updateMap, "_id")

	return updateMap, nil
}

func (r *BaseCollectionHandler[T]) Delete(filter map[string]any) *infra_error.AppError {
	r.logger.Debug("Deleting items", "collection", r.collection, "filter", filter)
	if err := r.dbHandler.Delete(r.collection, filter); err != nil {
		r.logger.Error("failed to delete item", "collection", r.collection, "filter", filter, "error", err)
		return err
	}
	return nil
}
