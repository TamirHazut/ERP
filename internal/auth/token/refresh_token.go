package token

import (
	"erp.localhost/internal/infra/db/redis"
	"erp.localhost/internal/infra/logging/logger"
	authv1_cache "erp.localhost/internal/infra/model/auth/v1/cache"
	model_redis "erp.localhost/internal/infra/model/db/redis"
)

// RefreshTokenKeyHandler handles refresh token operations in Redis
// Single token per user design - Key pattern: refresh_tokens:{tenant_id}:{user_id}
// Stores only ONE refresh token per user - new logins replace existing tokens
type RefreshTokenKeyHandler struct {
	*redis.BaseKeyHandler[authv1_cache.RefreshToken]
}

// NewRefreshTokenKeyHandler creates a new RefreshTokenHandler
func NewRefreshTokenKeyHandler(logger logger.Logger) (*RefreshTokenKeyHandler, error) {
	keyHandler, err := redis.NewBaseKeyHandler[authv1_cache.RefreshToken](
		model_redis.RedisKeyToken,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &RefreshTokenKeyHandler{
		BaseKeyHandler: keyHandler,
	}, nil
}
