package mongo

import (
	"errors"

	db "erp.localhost/internal/db"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

// Generic Collection
type CollectionHandler[T any] struct {
	dbHandler  db.DBHandler
	collection string
	logger     *logging.Logger
}

func NewCollectionHandler[T any](dbHandler db.DBHandler, collection string, logger *logging.Logger) *CollectionHandler[T] {
	if logger == nil {
		logger = logging.NewLogger(logging.ModuleDB)
	}
	collectionHandler := &CollectionHandler[T]{
		dbHandler:  dbHandler,
		collection: collection,
		logger:     logger,
	}
	if err := collectionHandler.CreateCollectionInDBIfNotExists(); err != nil {
		logger.Error(err.Error(), "collection", collection, "error", err)
		return nil
	}
	return collectionHandler
}

func (r *CollectionHandler[T]) CreateCollectionInDBIfNotExists() error {
	if dbHandler, ok := r.dbHandler.(*MongoDBManager); ok {
		return dbHandler.CreateCollectionInDBIfNotExists(r.collection)
	}
	// If not a MongoDBManager (e.g., mock), skip collection creation
	// This allows tests to work without a real MongoDB connection
	return nil
}

func (r *CollectionHandler[T]) Create(item T) (string, error) {
	r.logger.Debug("Creating item", "collection", r.collection)
	id, err := r.dbHandler.Create(r.collection, item)
	if err != nil {
		err = erp_errors.Internal(erp_errors.InternalDatabaseError, err)
		r.logger.Error(err.Error(), "collection", r.collection, "item", item)
		return "", err
	}
	return id, nil
}

func (r *CollectionHandler[T]) FindOne(filter map[string]any) (*T, error) {
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

func (r *CollectionHandler[T]) FindAll(filter map[string]any) ([]T, error) {
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

func (r *CollectionHandler[T]) Update(filter map[string]any, item T) error {
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

func (r *CollectionHandler[T]) Delete(filter map[string]any) error {
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
