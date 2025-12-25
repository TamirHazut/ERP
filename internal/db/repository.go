package db

import (
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

// Generic Repository
type Repository[T any] struct {
	dbHandler DBHandler
	dbName    string
	logger    *logging.Logger
}

func NewRepository[T any](dbHandler DBHandler, dbName string, logger *logging.Logger) *Repository[T] {
	return &Repository[T]{
		dbHandler: dbHandler,
		dbName:    dbName,
		logger:    logging.NewLogger(logging.ModuleDB),
	}
}

func (r *Repository[T]) Create(item T) (string, error) {
	r.logger.Debug("Creating item", "dbName", r.dbName)
	id, err := r.dbHandler.Create(r.dbName, item)
	if err != nil {
		return "", erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return id, nil
}

func (r *Repository[T]) Find(filter map[string]any) ([]T, error) {
	r.logger.Debug("Finding items", "dbName", r.dbName, "filter", filter)
	items, err := r.dbHandler.Find(r.dbName, filter)
	if err != nil {
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	res := []T{}
	for _, item := range items {
		res = append(res, item.(T))
	}
	return res, nil
}

func (r *Repository[T]) Update(filter map[string]any, item T) error {
	if filter == nil {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "filter")
	}
	r.logger.Debug("Updating item", "dbName", r.dbName, "filter", filter)
	err := r.dbHandler.Update(r.dbName, filter, item)
	if err != nil {
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return nil
}

func (r *Repository[T]) Delete(filter map[string]any) error {
	if filter == nil {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "filter")
	}
	r.logger.Debug("Deleting items", "dbName", r.dbName, "filter", filter)
	err := r.dbHandler.Delete(r.dbName, filter)
	if err != nil {
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return nil
}
