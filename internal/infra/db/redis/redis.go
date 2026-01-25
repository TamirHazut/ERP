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

func NewBaseRedisHandler(keyPrefix model_redis.KeyPrefix, logger logger.Logger) (*BaseRedisHandler, error) {
	redisHandler := &BaseRedisHandler{
		logger:    logger,
		keyPrefix: keyPrefix,
	}
	if err := redisHandler.init(); err != nil {
		redisHandler.logger.Error("Failed to initialize Redis", "error", err)
		return nil, err
	}
	return redisHandler, nil
}

func (r *BaseRedisHandler) init() error {
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

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return "", infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}

	result := r.client.Set(redisContext, formattedKey, valueBytes, 0)
	if result.Err() != nil {
		return "", result.Err()
	}
	// Handle TTL if provided
	if len(opts) > 0 {
		if ttl, ok := opts[0]["ttl"].(time.Duration); ok && ttl > 0 {
			r.Expire(key, int(ttl.Seconds()), time.Second)
		}
	}
	return result.Val(), nil
}

func (r *BaseRedisHandler) FindOne(key string, filter map[string]any, result any) error {
	formattedKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	value, err := r.client.Get(redisContext, formattedKey).Bytes()
	if err != nil {
		return err
	}
	err = json.Unmarshal(value, result)
	if err != nil {
		return err
	}
	return nil
}

func (r *BaseRedisHandler) FindAll(key string, filter map[string]any, result any) error {
	formattedKey := fmt.Sprintf("%s:%s*", r.keyPrefix, key)

	resultVal := reflect.ValueOf(result)

	// Enforce *[]*T
	if resultVal.Kind() != reflect.Ptr ||
		resultVal.Elem().Kind() != reflect.Slice ||
		resultVal.Elem().Type().Elem().Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to a slice of pointers (e.g. *[]*T)")
	}

	// 1️⃣ SCAN keys
	var cursor uint64
	keys := make([]string, 0)

	for {
		batch, nextCursor, err := r.client.Scan(
			redisContext,
			cursor,
			formattedKey,
			100,
		).Result()
		if err != nil {
			return err
		}

		keys = append(keys, batch...)

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	// No keys found → return empty slice
	if len(keys) == 0 {
		return nil
	}

	// 2️⃣ MGET values
	vals, err := r.client.MGet(redisContext, keys...).Result()
	if err != nil {
		return err
	}

	sliceVal := resultVal.Elem()
	elemType := sliceVal.Type().Elem().Elem() // T

	// 3️⃣ Unmarshal + append
	for i := range keys {
		if vals[i] == nil {
			continue // expired between SCAN and MGET
		}

		var data []byte
		switch v := vals[i].(type) {
		case string:
			data = []byte(v)
		case []byte:
			data = v
		default:
			continue
		}

		newElem := reflect.New(elemType) // *T
		if err := json.Unmarshal(data, newElem.Interface()); err != nil {
			return err
		}

		sliceVal.Set(reflect.Append(sliceVal, newElem))
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

// Scan scans for keys matching a pattern
// Returns keys in batches to avoid blocking Redis
// Pattern should include the key prefix (e.g., "tokens:tenant-123:*")
func (r *BaseRedisHandler) Scan(pattern string, batchSize int64) ([]string, error) {
	var allKeys []string
	var cursor uint64 = 0

	// Format pattern with key prefix if not already included
	fullPattern := fmt.Sprintf("%s:%s", r.keyPrefix, pattern)

	for {
		keys, nextCursor, err := r.client.Scan(redisContext, cursor, fullPattern, batchSize).Result()
		if err != nil {
			r.logger.Error("Failed to scan Redis keys", "error", err, "pattern", fullPattern)
			return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
		}

		allKeys = append(allKeys, keys...)
		cursor = nextCursor

		// Cursor returns to 0 when iteration is complete
		if cursor == 0 {
			break
		}
	}

	r.logger.Debug("Redis SCAN completed", "pattern", fullPattern, "keys_found", len(allKeys))
	return allKeys, nil
}

// DeleteByPattern deletes all keys matching a pattern
// Uses SCAN to find keys and pipeline for efficient deletion
func (r *BaseRedisHandler) DeleteByPattern(pattern string) (int, error) {
	keys, err := r.Scan(pattern, 100)
	if err != nil {
		return 0, err
	}

	if len(keys) == 0 {
		r.logger.Debug("No keys found to delete", "pattern", pattern)
		return 0, nil
	}

	// Delete in pipeline for efficiency
	pipe := r.client.Pipeline()
	for _, key := range keys {
		pipe.Del(redisContext, key)
	}

	_, err = pipe.Exec(redisContext)
	if err != nil {
		r.logger.Error("Failed to delete keys by pattern", "error", err, "pattern", pattern, "keys_count", len(keys))
		return 0, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	r.logger.Info("Keys deleted by pattern", "pattern", pattern, "keys_deleted", len(keys))
	return len(keys), nil
}
