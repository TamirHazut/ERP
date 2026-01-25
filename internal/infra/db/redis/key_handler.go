package redis

import (
	"fmt"

	db "erp.localhost/internal/infra/db"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_redis "erp.localhost/internal/infra/model/db/redis"
)

//go:generate mockgen -destination=mock/mock_key_handler.go -package=mock erp.localhost/internal/infra/db/redis KeyHandler
type KeyHandler[T any] interface {
	Set(tenantID string, key string, value *T, opts ...map[string]any) error
	GetOne(tenantID string, key string) (*T, error)
	GetAll(tenantID string, userID string) ([]*T, error)
	Update(tenantID string, key string, value *T, opts ...map[string]any) error
	Delete(tenantID string, key string) error
	// ScanKeys scans for keys matching a pattern for a specific tenant
	ScanKeys(tenantID string, pattern string) ([]string, error)
	// DeleteByPattern deletes all keys matching a pattern for a specific tenant
	DeleteByPattern(tenantID string, pattern string) (int, error)
}

type BaseKeyHandler[T any] struct {
	dbHandler db.DBHandler
	logger    logger.Logger
}

func NewBaseKeyHandler[T any](keyPrefix model_redis.KeyPrefix, logger logger.Logger) (*BaseKeyHandler[T], error) {
	dbHandler, err := NewBaseRedisHandler(keyPrefix, logger)
	if err != nil {
		return nil, err
	}
	return &BaseKeyHandler[T]{
		dbHandler: dbHandler,
		logger:    logger,
	}, nil
}

func (k *BaseKeyHandler[T]) Set(tenantID string, key string, value *T, opts ...map[string]any) error {
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
	result := new(T) // create a non-nil pointer for type T
	err := k.dbHandler.FindOne(formattedKey, nil, result)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	// // Handle case where value is nil (not found)
	// if result == nil {
	// 	return nil, infra_error.NotFound(infra_error.NotFoundResource, "key", formattedKey)
	// }
	return result, nil
}

func (k *BaseKeyHandler[T]) GetAll(tenantID string, userID string) ([]*T, error) {
	k.logger.Debug("Getting key", "tenantID", tenantID, "userID", userID)
	result := make([]*T, 0)
	formattedKey := fmt.Sprintf("%s:%s", tenantID, userID)
	err := k.dbHandler.FindAll(formattedKey, nil, &result)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return result, nil
}

func (k *BaseKeyHandler[T]) Update(tenantID string, key string, value *T, opts ...map[string]any) error {
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

// ScanKeys scans for keys matching a pattern for a specific tenant
// Pattern is relative to tenant (e.g., "*" for all keys in tenant, "user-123" for specific user)
func (k *BaseKeyHandler[T]) ScanKeys(tenantID string, pattern string) ([]string, error) {
	k.logger.Debug("Scanning keys", "tenantID", tenantID, "pattern", pattern)

	// Type assert to get BaseRedisHandler
	redisHandler, ok := k.dbHandler.(*BaseRedisHandler)
	if !ok {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, fmt.Errorf("dbHandler is not a BaseRedisHandler"))
	}

	// Build full pattern: tenant_id:pattern
	fullPattern := fmt.Sprintf("%s:%s", tenantID, pattern)
	keys, err := redisHandler.Scan(fullPattern, 100)
	if err != nil {
		return nil, err
	}

	k.logger.Debug("Keys scanned", "tenantID", tenantID, "pattern", pattern, "keys_found", len(keys))
	return keys, nil
}

// DeleteByPattern deletes all keys matching a pattern for a specific tenant
// Returns the number of keys deleted
func (k *BaseKeyHandler[T]) DeleteByPattern(tenantID string, pattern string) (int, error) {
	k.logger.Debug("Deleting keys by pattern", "tenantID", tenantID, "pattern", pattern)

	// Type assert to get BaseRedisHandler
	redisHandler, ok := k.dbHandler.(*BaseRedisHandler)
	if !ok {
		return 0, infra_error.Internal(infra_error.InternalUnexpectedError, fmt.Errorf("dbHandler is not a BaseRedisHandler"))
	}

	// Build full pattern: tenant_id:pattern
	fullPattern := fmt.Sprintf("%s:%s*", tenantID, pattern)
	count, err := redisHandler.DeleteByPattern(fullPattern)
	if err != nil {
		return 0, err
	}

	k.logger.Info("Keys deleted by pattern", "fullPattern", fullPattern, "keys_deleted", count)
	return count, nil
}
