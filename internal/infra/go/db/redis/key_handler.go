package redis

import (
	"fmt"

	db "erp.localhost/infra/db"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_redis "erp.localhost/infra/model/db/redis"
)

//go:generate mockgen -destination=mock/mock_key_handler.go -package=mock erp.localhost/infra/db/redis KeyHandler
type KeyHandler[T any] interface {
	Set(key string, value *T, opts ...map[string]any) *infra_error.AppError
	GetOne(key string) (*T, *infra_error.AppError)
	GetAll(key string) ([]*T, *infra_error.AppError)
	Update(key string, value *T, opts ...map[string]any) *infra_error.AppError
	Delete(key string) *infra_error.AppError
	// ScanKeys scans for keys matching a pattern for a specific tenant
	ScanKeys(pattern string) ([]string, *infra_error.AppError)
	// DeleteByPattern deletes all keys matching a pattern for a specific tenant
	DeleteByPattern(pattern string) (int, *infra_error.AppError)
}

type BaseKeyHandler[T any] struct {
	dbHandler db.DBHandler
	logger    logger.Logger
}

func NewBaseKeyHandler[T any](keyPrefix model_redis.KeyPrefix, logger logger.Logger) (*BaseKeyHandler[T], *infra_error.AppError) {
	dbHandler, err := NewBaseRedisHandler(keyPrefix, logger)
	if err != nil {
		return nil, err
	}
	return &BaseKeyHandler[T]{
		dbHandler: dbHandler,
		logger:    logger,
	}, nil
}

func (k *BaseKeyHandler[T]) Set(key string, value *T, opts ...map[string]any) *infra_error.AppError {
	k.logger.Debug("Setting key", "key", key, "value", value)
	_, err := k.dbHandler.Create(key, value, opts...)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return nil
}

func (k *BaseKeyHandler[T]) GetOne(key string) (*T, *infra_error.AppError) {
	k.logger.Debug("Getting key", "key", key)
	result := new(T) // create a non-nil pointer for type T
	err := k.dbHandler.FindOne(key, nil, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *BaseKeyHandler[T]) GetAll(key string) ([]*T, *infra_error.AppError) {
	k.logger.Debug("Getting key", "key", key)
	result := make([]*T, 0)
	err := k.dbHandler.FindAll(key, nil, &result)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return result, nil
}

func (k *BaseKeyHandler[T]) Update(key string, value *T, opts ...map[string]any) *infra_error.AppError {
	k.logger.Debug("Updating key", "key", key, "value", value)
	err := k.dbHandler.Update(key, nil, value, opts...)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return nil
}

func (k *BaseKeyHandler[T]) Delete(key string) *infra_error.AppError {
	k.logger.Debug("Deleting key", "key", key)
	err := k.dbHandler.Delete(key, nil)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return nil
}

// ScanKeys scans for keys matching a pattern for a specific tenant
// Pattern is relative to tenant (e.g., "*" for all keys in tenant, "user-123" for specific user)
func (k *BaseKeyHandler[T]) ScanKeys(pattern string) ([]string, *infra_error.AppError) {
	k.logger.Debug("Scanning keys", "pattern", pattern)

	// Type assert to get BaseRedisHandler
	redisHandler, ok := k.dbHandler.(*BaseRedisHandler)
	if !ok {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, fmt.Errorf("dbHandler is not a BaseRedisHandler"))
	}

	keys, err := redisHandler.Scan(pattern, 100)
	if err != nil {
		return nil, err
	}

	k.logger.Debug("Keys scanned", "pattern", pattern, "keys_found", len(keys))
	return keys, nil
}

// DeleteByPattern deletes all keys matching a pattern for a specific tenant
// Returns the number of keys deleted
func (k *BaseKeyHandler[T]) DeleteByPattern(pattern string) (int, *infra_error.AppError) {
	k.logger.Debug("Deleting keys by pattern", "pattern", pattern)

	// Type assert to get BaseRedisHandler
	redisHandler, ok := k.dbHandler.(*BaseRedisHandler)
	if !ok {
		return 0, infra_error.Internal(infra_error.InternalUnexpectedError, fmt.Errorf("dbHandler is not a BaseRedisHandler"))
	}

	count, err := redisHandler.DeleteByPattern(pattern)
	if err != nil {
		return 0, err
	}

	k.logger.Info("Keys deleted by pattern", "keys_deleted", count)
	return count, nil
}
