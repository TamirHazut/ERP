package token

import (
	"erp.localhost/infra/db/redis"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_auth "erp.localhost/infra/model/auth"
	authv1_cache "erp.localhost/infra/model/auth/v1/cache"
	model_redis "erp.localhost/infra/model/db/redis"
)

// RefreshTokenKeyHandler handles refresh token operations in Redis
// Single token per user design - Key pattern: refresh_tokens:{tenant_id}:{user_id}
// Stores only ONE refresh token per user - new logins replace existing tokens
type RefreshTokenKeyHandler struct {
	*redis.BaseKeyHandler[authv1_cache.RefreshToken]
}

// NewRefreshTokenKeyHandler creates a new RefreshTokenHandler
func NewRefreshTokenKeyHandler(logger logger.Logger) (*RefreshTokenKeyHandler, *infra_error.AppError) {
	keyHandler, err := redis.NewBaseKeyHandler[authv1_cache.RefreshToken](
		model_redis.RedisKeyRefreshToken,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &RefreshTokenKeyHandler{
		BaseKeyHandler: keyHandler,
	}, nil
}

func (c *RefreshTokenKeyHandler) GetOne(key string) (*authv1_cache.RefreshToken, *infra_error.AppError) {
	refreshToken, err := c.BaseKeyHandler.GetOne(key)
	if err != nil {
		return nil, c.convertError(err, key)
	}
	return refreshToken, nil
}

// convertError converts generic NOT_FOUND_ITEM to NOT_FOUND_REFRESH_TOKEN
func (c *RefreshTokenKeyHandler) convertError(err error, context any) *infra_error.AppError {
	if appErr, ok := infra_error.AsAppError(err); ok {
		if appErr.Code == infra_error.NotFoundItem.Code {
			return infra_error.NotFound(
				infra_error.NotFoundRefreshToken,
				model_auth.ResourceTypeToken,
				context,
			)
		}
		return appErr
	}
	return infra_error.Internal(infra_error.InternalUnexpectedError, err)
}
