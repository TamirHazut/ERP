package redis

import (
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_redis "erp.localhost/infra/model/db/redis"
)

type SystemKeyHandler struct {
	*BaseKeyHandler[string]
}

// NewSystemKeyHandler creates a new SystemKeyHandler
func NewSystemKeyHandler(logger logger.Logger) (*SystemKeyHandler, *infra_error.AppError) {
	keyHandler, err := NewBaseKeyHandler[string](
		model_redis.RedisKeySystem,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &SystemKeyHandler{
		BaseKeyHandler: keyHandler,
	}, nil
}
