package redis

import (
	"context"
	"fmt"
	"time"

	erp_errors "erp.localhost/internal/infra/error"
	logging "erp.localhost/internal/infra/logging"
	redis_models "erp.localhost/internal/infra/model/db/redis"
	shared_models "erp.localhost/internal/infra/model/shared"
	redis "github.com/redis/go-redis/v9"
)

var (
	redisContext = context.Background()
)

type BaseRedisHandler struct {
	client    *redis.Client
	logger    *logging.Logger
	keyPrefix redis_models.KeyPrefix
}

func NewBaseRedisHandler(keyPrefix redis_models.KeyPrefix) *BaseRedisHandler {
	redisHandler := &BaseRedisHandler{
		logger:    logging.NewLogger(shared_models.ModuleDB),
		keyPrefix: keyPrefix,
	}
	if err := redisHandler.init(); err != nil {
		redisHandler.logger.Error("Failed to initialize Redis", "error", err)
		return nil
	}
	return redisHandler
}

func (r *BaseRedisHandler) init() error {
	r.logger = logging.NewLogger(shared_models.ModuleDB)
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
	if _, err := r.FindOne(key, nil); err == nil {
		return "", erp_errors.Conflict(erp_errors.ConflictDuplicateResource)
	}
	result := r.client.Set(redisContext, formattedKey, value, 0)
	if result.Err() != nil {
		return "", result.Err()
	}
	return result.Val(), nil
}

func (r *BaseRedisHandler) FindOne(key string, filter map[string]any) (any, error) {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	value, err := r.client.Get(redisContext, formattedKey).Result()
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (r *BaseRedisHandler) FindAll(key string, filter map[string]any) ([]any, error) {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	values, err := r.SMembers(formattedKey)
	if err != nil {
		return nil, err
	}
	var results []any
	for _, value := range values {
		results = append(results, value)
	}
	return results, nil
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
