package mongo

import (
	db "erp.localhost/internal/db"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

// Generic Repository
type CollectionHandler[T any] struct {
	dbHandler  db.DBHandler
	collection string
	logger     *logging.Logger
}

func NewCollectionHandler[T any](dbHandler db.DBHandler, collection string, logger *logging.Logger) *CollectionHandler[T] {
	if logger == nil {
		logger = logging.NewLogger(logging.ModuleDB)
	}
	return &CollectionHandler[T]{
		dbHandler:  dbHandler,
		collection: collection,
		logger:     logger,
	}
}

func (r *CollectionHandler[T]) Create(item T) (string, error) {
	r.logger.Debug("Creating item", "collection", r.collection)
	id, err := r.dbHandler.Create(r.collection, item)
	if err != nil {
		return "", erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return id, nil
}

func (r *CollectionHandler[T]) Find(filter map[string]any) ([]T, error) {
	r.logger.Debug("Finding items", "collection", r.collection, "filter", filter)
	items, err := r.dbHandler.Find(r.collection, filter)
	if err != nil {
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	res := []T{}
	for _, item := range items {
		res = append(res, item.(T))
	}
	return res, nil
}

func (r *CollectionHandler[T]) Update(filter map[string]any, item T) error {
	if filter == nil {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "filter")
	}
	r.logger.Debug("Updating item", "collection", r.collection, "filter", filter)
	err := r.dbHandler.Update(r.collection, filter, item)
	if err != nil {
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return nil
}

func (r *CollectionHandler[T]) Delete(filter map[string]any) error {
	if filter == nil {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "filter")
	}
	r.logger.Debug("Deleting items", "collection", r.collection, "filter", filter)
	err := r.dbHandler.Delete(r.collection, filter)
	if err != nil {
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return nil
}
