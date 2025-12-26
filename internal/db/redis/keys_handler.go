package redis

import (
	"encoding/json"
	"fmt"

	db "erp.localhost/internal/db"
	erp_errors "erp.localhost/internal/errors"
	logging "erp.localhost/internal/logging"
)

type KeyHandler[T any] struct {
	dbHandler db.DBHandler
	logger    *logging.Logger
}

func NewKeyHandler[T any](keyPrefix KeyPrefix, logger *logging.Logger) *KeyHandler[T] {
	if logger == nil {
		logger = logging.NewLogger(logging.ModuleDB)
	}
	dbHandler := NewRedisHandler(keyPrefix)
	if dbHandler == nil {
		logger.Fatal("Failed to create redis handler")
		return nil
	}
	return &KeyHandler[T]{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

func (k *KeyHandler[T]) Set(tenantID string, key string, value any) error {
	k.logger.Debug("Setting key", "tenantID", tenantID, "key", key, "value", value)
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	_, err := k.dbHandler.Create(formattedKey, value)
	if err != nil {
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return nil
}

func (k *KeyHandler[T]) Get(tenantID string, key string) ([]T, error) {
	k.logger.Debug("Getting key", "tenantID", tenantID, "key", key)
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	values, err := k.dbHandler.Find(formattedKey, nil)
	if err != nil {
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	if len(values) == 0 {
		return nil, erp_errors.NotFound(erp_errors.NotFoundResource, key, nil)
	}
	var results []T
	for _, value := range values {
		var result T
		err = json.Unmarshal([]byte(value.(string)), &result)
		if err != nil {
			return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
		}
		results = append(results, result)
	}
	if err != nil {
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return results, nil
}

func (k *KeyHandler[T]) Update(tenantID string, key string, value any) error {
	k.logger.Debug("Updating key", "tenantID", tenantID, "key", key, "value", value)
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	err := k.dbHandler.Update(formattedKey, nil, value)
	if err != nil {
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return nil
}

func (k *KeyHandler[T]) Delete(tenantID string, key string) error {
	k.logger.Debug("Deleting key", "tenantID", tenantID, "key", key)
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	err := k.dbHandler.Delete(formattedKey, nil)
	if err != nil {
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return nil
}
