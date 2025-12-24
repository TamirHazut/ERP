package redis

import (
	"context"
	"sync"

	logging "erp.localhost/internal/logging"
	redis "github.com/redis/go-redis/v9"
)

var (
	initRedisOnce sync.Once
	redisHandler  *RedisHandler
	redisContext  = context.Background()
)

type RedisHandler struct {
	client *redis.Client
	logger *logging.Logger
}

func GetRedisHandler() *RedisHandler {
	initRedisOnce.Do(func() {
		redisHandler = &RedisHandler{}
		err := redisHandler.init()
		if err != nil {
			redisHandler = nil
		}
	})
	return redisHandler
}

func (r *RedisHandler) init() error {
	r.logger = logging.NewLogger(logging.ModuleDB)
	uri := "redis://:supersecretredis@localhost:6379"
	options, err := redis.ParseURL(uri)
	if err != nil {
		r.logger.Error("Failed to parse Redis URL", "error", err)
		return err
	}

	client := redis.NewClient(options)
	if err := client.Ping(redisContext).Err(); err != nil {
		r.logger.Error("Failed to ping Redis", "error", err)
		return err
	}
	r.client = client

	return nil
}

func (r *RedisHandler) Close() error {
	return r.client.Close()
}

func (r *RedisHandler) Create(key string, value any) error {
	return r.client.Set(redisContext, key, value, 0).Err()
}

func (r *RedisHandler) Get(key string) (string, error) {
	return r.client.Get(redisContext, key).Result()
}

func (r *RedisHandler) Update(key string, value any) error {
	return r.client.Set(redisContext, key, value, 0).Err()
}

func (r *RedisHandler) Delete(key string) error {
	return r.client.Del(redisContext, key).Err()
}
