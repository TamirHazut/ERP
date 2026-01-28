package redis

import (
	"erp.localhost/internal/infra/logging/logger"
	model_redis "erp.localhost/internal/infra/model/db/redis"
)

type SystemKeyHandler struct {
	*BaseKeyHandler[string]
}

// NewSystemKeyHandler creates a new SystemKeyHandler
func NewSystemKeyHandler(logger logger.Logger) (*SystemKeyHandler, error) {
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
