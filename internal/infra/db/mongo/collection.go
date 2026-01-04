package mongo

import (
	"errors"

	db "erp.localhost/internal/infra/db"
	erp_errors "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging"
	mongo_models "erp.localhost/internal/infra/model/db/mongo"
	shared_models "erp.localhost/internal/infra/model/shared"
)

// Generic Collection
type BaseCollectionHandler[T any] struct {
	dbHandler  db.DBHandler
	collection string
	logger     *logging.Logger
}

func NewBaseCollectionHandler[T any](collection string, logger *logging.Logger) *BaseCollectionHandler[T] {
	if logger == nil {
		logger = logging.NewLogger(shared_models.ModuleDB)
	}
	dbName := mongo_models.GetDBNameFromCollection(collection)
	if dbName == "" {
		logger.Fatal("db not found", "collection", collection)
		return nil
	}
	dbHandler := NewMongoDBManager(mongo_models.DBName(dbName))
	if dbHandler == nil {
		logger.Fatal("failed to create db handler", "db", dbName, "collection", collection)
		return nil
	}
	collectionHandler := &BaseCollectionHandler[T]{
		dbHandler:  dbHandler,
		collection: collection,
		logger:     logger,
	}
	if err := collectionHandler.createCollectionInDBIfNotExists(); err != nil {
		logger.Error(err.Error(), "collection", collection, "error", err)
		return nil
	}
	return collectionHandler
}

func (r *BaseCollectionHandler[T]) createCollectionInDBIfNotExists() error {
	if dbHandler, ok := r.dbHandler.(*MongoDBManager); ok {
		return dbHandler.CreateCollectionInDBIfNotExists(r.collection)
	}
	return nil
}

func (r *BaseCollectionHandler[T]) Create(item T) (string, error) {
	r.logger.Debug("Creating item", "collection", r.collection)
	id, err := r.dbHandler.Create(r.collection, item)
	if err != nil {
		err = erp_errors.Internal(erp_errors.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "item", item)
		return "", err
	}
	return id, nil
}

func (r *BaseCollectionHandler[T]) FindOne(filter map[string]any) (*T, error) {
	r.logger.Debug("Finding item", "collection", r.collection, "filter", filter)
	item, err := r.dbHandler.FindOne(r.collection, filter)
	if err != nil {
		err = erp_errors.Internal(erp_errors.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
		return nil, err
	}
	if item == nil {
		err = erp_errors.NotFound(erp_errors.NotFoundResource, r.collection, filter)
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
		return nil, err
	}
	res, ok := item.(T)
	if !ok {
		err = erp_errors.Internal(erp_errors.InternalDatabaseError, errors.New("type assertion failed"))
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
		return nil, err
	}

	return &res, nil
}

func (r *BaseCollectionHandler[T]) FindAll(filter map[string]any) ([]T, error) {
	r.logger.Debug("Finding items", "collection", r.collection, "filter", filter)
	items, err := r.dbHandler.FindAll(r.collection, filter)
	if err != nil {
		err = erp_errors.Internal(erp_errors.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
		return nil, err
	}
	res := []T{}
	for _, item := range items {
		obj, ok := item.(T)
		if !ok {
			err = erp_errors.Internal(erp_errors.InternalDatabaseError, errors.New("type assertion failed"))
			r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
			return nil, err
		}
		res = append(res, obj)
	}
	return res, nil
}

func (r *BaseCollectionHandler[T]) Update(filter map[string]any, item T) error {
	r.logger.Debug("Updating item", "collection", r.collection, "filter", filter, "item", item)
	if filter == nil {
		err := erp_errors.Validation(erp_errors.ValidationRequiredFields, "filter")
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter, "item", item)
		return err
	}
	if err := r.dbHandler.Update(r.collection, filter, item); err != nil {
		err = erp_errors.Internal(erp_errors.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter, "item", item)
		return err
	}
	return nil
}

func (r *BaseCollectionHandler[T]) Delete(filter map[string]any) error {
	if filter == nil {
		err := erp_errors.Validation(erp_errors.ValidationRequiredFields, "filter")
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
		return err
	}
	r.logger.Debug("Deleting items", "collection", r.collection, "filter", filter)
	if err := r.dbHandler.Delete(r.collection, filter); err != nil {
		err = erp_errors.Internal(erp_errors.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "filter", filter)
		return err
	}
	return nil
}
