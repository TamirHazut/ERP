package token

import (
	"erp.localhost/internal/infra/db/redis"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
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
func NewAccessTokenKeyHandler(logger logger.Logger) (*AccessTokenKeyHandler, *infra_error.AppError) {
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

func (c *AccessTokenKeyHandler) GetOne(key string) (*authv1_cache.TokenMetadata, *infra_error.AppError) {
	token, err := c.BaseKeyHandler.GetOne(key)
	if err != nil {
		return nil, c.convertError(err, key)
	}
	return token, nil
}

// convertError converts generic NOT_FOUND_ITEM to NOT_FOUND_TOKEN
func (c *AccessTokenKeyHandler) convertError(err error, context any) *infra_error.AppError {
	if appErr, ok := infra_error.AsAppError(err); ok {
		if appErr.Code == infra_error.NotFoundItem.Code {
			return infra_error.NotFound(
				infra_error.NotFoundToken,
				model_auth.ResourceTypeToken,
				context,
			)
		}
		return appErr
	}
	return infra_error.Internal(infra_error.InternalUnexpectedError, err)
}
