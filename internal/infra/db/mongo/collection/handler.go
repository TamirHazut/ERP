package collection

import (
	db "erp.localhost/internal/infra/db"
	"erp.localhost/internal/infra/db/mongo"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

//go:generate mockgen -destination=mock/mock_collection_handler.go -package=mock erp.localhost/internal/infra/db/mongo/collection CollectionHandler
type CollectionHandler[T any] interface {
	Create(item *T) (string, error)
	FindOne(filter map[string]any) (*T, error)
	FindAll(filter map[string]any) ([]*T, error)
	Update(filter map[string]any, item *T) error
	Delete(filter map[string]any) error
}

// Generic Collection
type BaseCollectionHandler[T any] struct {
	dbHandler  db.DBHandler
	collection string
	logger     logger.Logger
}

func NewBaseCollectionHandler[T any](dbName model_mongo.DBName, collection model_mongo.Collection, logger logger.Logger) (*BaseCollectionHandler[T], error) {
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

func (r *BaseCollectionHandler[T]) createCollectionInDBIfNotExists() error {
	if dbHandler, ok := r.dbHandler.(*mongo.MongoDBManager); ok {
		return dbHandler.CreateCollectionInDBIfNotExists(r.collection)
	}
	return nil
}

func (r *BaseCollectionHandler[T]) Create(item *T) (string, error) {
	r.logger.Debug("Creating item", "collection", r.collection)
	id, err := r.dbHandler.Create(r.collection, item)
	if err != nil {
		err = infra_error.Internal(infra_error.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "item", item)
		return "", err
	}
	return id, nil
}

func (r *BaseCollectionHandler[T]) FindOne(filter map[string]any) (*T, error) {
	r.logger.Debug("Finding item", "collection", r.collection, "filter", filter)
	result := new(T)
	err := r.dbHandler.FindOne(r.collection, filter, result)
	if err != nil {
		err = infra_error.Internal(infra_error.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
		return nil, err
	}
	// if result == nil {
	// 	err = infra_error.NotFound(infra_error.NotFoundResource, r.collection, filter)
	// 	r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
	// 	return nil, err
	// }

	return result, nil
}

func (r *BaseCollectionHandler[T]) FindAll(filter map[string]any) ([]*T, error) {
	if filter == nil {
		r.logger.Debug("nil filter found", "collection", r.collection)
		filter = make(map[string]any)
	}
	r.logger.Debug("Finding items", "collection", r.collection, "filter", filter)
	result := make([]*T, 0)
	err := r.dbHandler.FindAll(r.collection, filter, &result)
	if err != nil {
		err = infra_error.Internal(infra_error.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
		return nil, err
	}
	return result, nil
}

func (r *BaseCollectionHandler[T]) Update(filter map[string]any, item *T) error {
	r.logger.Debug("Updating item", "collection", r.collection, "filter", filter, "item", item)
	if filter == nil {
		err := infra_error.Validation(infra_error.ValidationRequiredFields, "filter")
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter, "item", item)
		return err
	}

	// Convert item to BSON map and exclude _id field (immutable in MongoDB)
	updateData, err := r.prepareUpdateData(item)
	if err != nil {
		err = infra_error.Internal(infra_error.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter, "item", item)
		return err
	}

	if err := r.dbHandler.Update(r.collection, filter, updateData); err != nil {
		err = infra_error.Internal(infra_error.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter, "item", item)
		return err
	}
	return nil
}

// prepareUpdateData converts item to BSON map and excludes the _id field
func (r *BaseCollectionHandler[T]) prepareUpdateData(item *T) (bson.M, error) {
	// Marshal to BSON bytes
	bytes, err := bson.Marshal(item)
	if err != nil {
		return nil, err
	}

	// Unmarshal to bson.M
	var updateMap bson.M
	if err := bson.Unmarshal(bytes, &updateMap); err != nil {
		return nil, err
	}

	// Remove _id field (immutable in MongoDB)
	delete(updateMap, "_id")

	return updateMap, nil
}

func (r *BaseCollectionHandler[T]) Delete(filter map[string]any) error {
	if filter == nil {
		err := infra_error.Validation(infra_error.ValidationRequiredFields, "filter")
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
		return err
	}
	r.logger.Debug("Deleting items", "collection", r.collection, "filter", filter)
	if err := r.dbHandler.Delete(r.collection, filter); err != nil {
		err = infra_error.Internal(infra_error.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
		return err
	}
	return nil
}
