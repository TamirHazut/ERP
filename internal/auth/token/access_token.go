package token

import (
	"erp.localhost/internal/infra/db/redis"
	"erp.localhost/internal/infra/logging/logger"
	authv1_cache "erp.localhost/internal/infra/model/auth/v1/cache"
	model_redis "erp.localhost/internal/infra/model/db/redis"
)

// AccessTokenKeyHandler handles access token operations in Redis
// Single token per user design - Key pattern: tokens:{tenant_id}:{user_id}
// Stores only ONE access token per user - new logins replace existing tokens
type AccessTokenKeyHandler struct {
	*redis.BaseKeyHandler[authv1_cache.TokenMetadata]
}

// NewAccessTokenKeyHandler creates a new AccessTokenHandler
func NewAccessTokenKeyHandler(logger logger.Logger) (*AccessTokenKeyHandler, error) {
	keyHandler, err := redis.NewBaseKeyHandler[authv1_cache.TokenMetadata](
		model_redis.RedisKeyToken,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &AccessTokenKeyHandler{
		BaseKeyHandler: keyHandler,
	}, nil
}
