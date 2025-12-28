package redis

import (
	"context"
	"fmt"

	erp_errors "erp.localhost/internal/errors"
	logging "erp.localhost/internal/logging"
	redis "github.com/redis/go-redis/v9"
)

var (
	redisContext = context.Background()
)

type RedisHandler struct {
	client    *redis.Client
	logger    *logging.Logger
	keyPrefix KeyPrefix
}

func NewRedisHandler(keyPrefix KeyPrefix) *RedisHandler {
	redisHandler := &RedisHandler{
		logger:    logging.NewLogger(logging.ModuleDB),
		keyPrefix: keyPrefix,
	}
	if err := redisHandler.init(); err != nil {
		redisHandler.logger.Error("Failed to initialize Redis", "error", err)
		return nil
	}
	return redisHandler
}

func (r *RedisHandler) init() error {
	r.logger = logging.NewLogger(logging.ModuleDB)
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

func (r *RedisHandler) Close() error {
	return r.client.Close()
}

func (r *RedisHandler) Create(key string, value any, opts ...map[string]any) (string, error) {
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

func (r *RedisHandler) FindOne(key string, filter map[string]any) (any, error) {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	value, err := r.client.Get(redisContext, formattedKey).Result()
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (r *RedisHandler) FindAll(key string, filter map[string]any) ([]any, error) {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	values, err := r.client.SMembers(redisContext, formattedKey).Result()
	if err != nil {
		return nil, err
	}
	var results []any
	for _, value := range values {
		results = append(results, value)
	}
	return results, nil
}

func (r *RedisHandler) Update(key string, filter map[string]any, value any, opts ...map[string]any) error {
	_, err := r.Create(key, value)
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisHandler) Delete(key string, filter map[string]any) error {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	return r.client.Del(redisContext, formattedKey).Err()
}
