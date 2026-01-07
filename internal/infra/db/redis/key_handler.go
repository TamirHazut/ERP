package redis

import (
	"encoding/json"
	"fmt"

	db "erp.localhost/internal/infra/db"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
)

//go:generate mockgen -destination=mock/mock_key_handler.go -package=mock erp.localhost/internal/infra/db/redis KeyHandler
type KeyHandler[T any] interface {
	Set(tenantID string, key string, value T, opts ...map[string]any) error
	GetOne(tenantID string, key string) (*T, error)
	GetAll(tenantID string, userID string) ([]T, error)
	Update(tenantID string, key string, value T, opts ...map[string]any) error
	Delete(tenantID string, key string) error
}

type BaseKeyHandler[T any] struct {
	dbHandler db.DBHandler
	logger    logger.Logger
}

func NewBaseKeyHandler[T any](dbHandler db.DBHandler, logger logger.Logger) *BaseKeyHandler[T] {
	return &BaseKeyHandler[T]{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

func (k *BaseKeyHandler[T]) Set(tenantID string, key string, value T, opts ...map[string]any) error {
	k.logger.Debug("Setting key", "tenantID", tenantID, "key", key, "value", value)
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	_, err := k.dbHandler.Create(formattedKey, value, opts...)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return nil
}

func (k *BaseKeyHandler[T]) GetOne(tenantID string, key string) (*T, error) {
	k.logger.Debug("Getting key", "tenantID", tenantID, "key", key)
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	value, err := k.dbHandler.FindOne(formattedKey, nil)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	// Handle case where value is nil (not found)
	if value == nil {
		return nil, infra_error.NotFound(infra_error.NotFoundResource, "key", formattedKey)
	}

	// Handle case where mock returns struct directly
	if typedValue, ok := value.(T); ok {
		return &typedValue, nil
	}

	// Handle case where Redis returns JSON string
	var result T
	err = json.Unmarshal([]byte(value.(string)), &result)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return &result, nil
}

func (k *BaseKeyHandler[T]) GetAll(tenantID string, userID string) ([]T, error) {
	k.logger.Debug("Getting key", "tenantID", tenantID, "userID", userID)
	formattedKey := fmt.Sprintf("%s:%s", tenantID, userID)
	values, err := k.dbHandler.FindAll(formattedKey, nil)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	results := make([]T, len(values))
	for i, value := range values {
		// Handle case where mock returns struct directly
		if typedValue, ok := value.(T); ok {
			results[i] = typedValue
			continue
		}

		// Handle case where Redis returns JSON string
		var result T
		err = json.Unmarshal([]byte(value.(string)), &result)
		if err != nil {
			return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
		}
		results[i] = result
	}
	return results, nil
}

func (k *BaseKeyHandler[T]) Update(tenantID string, key string, value T, opts ...map[string]any) error {
	k.logger.Debug("Updating key", "tenantID", tenantID, "key", key, "value", value)
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	err := k.dbHandler.Update(formattedKey, nil, value, opts...)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return nil
}

func (k *BaseKeyHandler[T]) Delete(tenantID string, key string) error {
	k.logger.Debug("Deleting key", "tenantID", tenantID, "key", key)
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	err := k.dbHandler.Delete(formattedKey, nil)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return nil
}
