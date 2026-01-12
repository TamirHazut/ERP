package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_redis "erp.localhost/internal/infra/model/db/redis"
	model_shared "erp.localhost/internal/infra/model/shared"
	redis "github.com/redis/go-redis/v9"
)

//go:generate mockgen -destination=mock/mock_redis_handler.go -package=mock erp.localhost/internal/infra/db/redis RedisHandler
type RedisHandler interface {
	SAdd(key string, members ...any) error
	SRem(key string, members ...any) error
	SMembers(key string) ([]string, error)
	Expire(key string, ttl int, unit time.Duration) error
	Clear(key string) error
}

var (
	redisContext = context.Background()
)

type BaseRedisHandler struct {
	client    *redis.Client
	logger    logger.Logger
	keyPrefix model_redis.KeyPrefix
}

func NewBaseRedisHandler(keyPrefix model_redis.KeyPrefix) *BaseRedisHandler {
	redisHandler := &BaseRedisHandler{
		logger:    logger.NewBaseLogger(model_shared.ModuleDB),
		keyPrefix: keyPrefix,
	}
	if err := redisHandler.init(); err != nil {
		redisHandler.logger.Error("Failed to initialize Redis", "error", err)
		return nil
	}
	return redisHandler
}

func (r *BaseRedisHandler) init() error {
	r.logger = logger.NewBaseLogger(model_shared.ModuleDB)
	uri := "redis://:supersecretredis@localhost:6379"
	options, err := redis.ParseURL(uri)
	if err != nil {
		return err
	}

	client := redis.NewClient(options)
	if err := client.Ping(redisContext).Err(); err != nil {
		return err
	}
	r.client = client

	return nil
}

func (r *BaseRedisHandler) Close() error {
	return r.client.Close()
}

func (r *BaseRedisHandler) Create(key string, value any, opts ...map[string]any) (string, error) {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)

	exists, err := r.client.Exists(redisContext, key).Result()
	if err != nil {
		return "", infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	if exists > 0 {
		return "", infra_error.Conflict(infra_error.ConflictDuplicateResource)
	}

	result := r.client.Set(redisContext, formattedKey, value, 0)
	if result.Err() != nil {
		return "", result.Err()
	}
	return result.Val(), nil
}

func (r *BaseRedisHandler) FindOne(key string, filter map[string]any, result any) error {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	value, err := r.client.Get(redisContext, formattedKey).Result()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(value), result)
	if err != nil {
		return err
	}
	return nil
}

func (r *BaseRedisHandler) FindAll(key string, filter map[string]any, result any) error {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	values, err := r.SMembers(formattedKey)
	if err != nil {
		return err
	}
	// Get the reflect.Value of the pointer
	resultVal := reflect.ValueOf(result)

	// Make sure it's a pointer to a slice
	if resultVal.Kind() != reflect.Ptr || resultVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result must be a pointer to a slice")
	}

	// Get the slice itself
	sliceVal := resultVal.Elem()

	// For each value, unmarshal and append
	for _, value := range values {
		// Create a new element of the slice's element type
		elemType := sliceVal.Type().Elem()
		newElem := reflect.New(elemType.Elem()).Interface()

		// Unmarshal the JSON into the new element
		if err := json.Unmarshal([]byte(value), newElem); err != nil {
			return err
		}

		// Append the pointer to the slice
		sliceVal.Set(reflect.Append(sliceVal, reflect.ValueOf(newElem)))
	}

	return nil
}

func (r *BaseRedisHandler) Update(key string, filter map[string]any, value any, opts ...map[string]any) error {
	_, err := r.Create(key, value)
	if err != nil {
		return err
	}
	return nil
}

func (r *BaseRedisHandler) Delete(key string, filter map[string]any) error {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	return r.client.Del(redisContext, formattedKey).Err()
}

func (r *BaseRedisHandler) SAdd(key string, members ...any) error {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	return r.client.SAdd(redisContext, formattedKey, members...).Err()
}

func (r *BaseRedisHandler) SRem(key string, members ...any) error {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	return r.client.SRem(redisContext, formattedKey, members...).Err()
}

func (r *BaseRedisHandler) Expire(key string, ttl int, unit time.Duration) error {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	return r.client.Expire(redisContext, formattedKey, time.Duration(ttl)*unit).Err()
}

func (r *BaseRedisHandler) SMembers(key string) ([]string, error) {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	return r.client.SMembers(redisContext, formattedKey).Result()
}

func (r *BaseRedisHandler) Clear(key string) error {
	return r.Delete(key, nil)
}
