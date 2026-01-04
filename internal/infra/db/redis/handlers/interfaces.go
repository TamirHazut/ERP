package handlers

import "time"

//go:generate mockgen -destination=mocks/mock_redis_handler.go -package=mocks erp.localhost/internal/infra/db/redis/handlers RedisHandler
//go:generate mockgen -destination=mocks/mock_key_handler.go -package=mocks erp.localhost/internal/infra/db/redis/handlers KeyHandler
//go:generate mockgen -destination=mocks/mock_set_handler.go -package=mocks erp.localhost/internal/infra/db/redis/handlers SetHandler

type RedisHandler interface {
	SAdd(key string, members ...any) error
	SRem(key string, members ...any) error
	SMembers(key string) ([]string, error)
	Expire(key string, ttl int, unit time.Duration) error
	Clear(key string) error
}

type KeyHandler[T any] interface {
	Set(tenantID string, key string, value T, opts ...map[string]any) error
	GetOne(tenantID string, key string) (*T, error)
	GetAll(tenantID string, userID string) ([]T, error)
	Update(tenantID string, key string, value T, opts ...map[string]any) error
	Delete(tenantID string, key string) error
}

type SetHandler interface {
	Add(tenantID string, key string, member string, opts ...map[string]any) error
	Remove(tenantID string, key string, member string) error
	Members(tenantID string, key string) ([]string, error)
	Clear(tenantID string, key string) error
}
