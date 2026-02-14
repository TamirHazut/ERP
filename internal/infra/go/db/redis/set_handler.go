package redis

import (
	"fmt"
	"time"

	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
)

//go:generate mockgen -destination=mock/mock_set_handler.go -package=mock erp.localhost/infra/db/redis SetHandler
type SetHandler interface {
	Add(tenantID string, key string, member string, opts ...map[string]any) *infra_error.AppError
	Remove(tenantID string, key string, member string) *infra_error.AppError
	Members(tenantID string, key string) ([]string, *infra_error.AppError)
	Clear(tenantID string, key string) *infra_error.AppError
}

type BaseSetHandler struct {
	redisHandler RedisHandler
	logger       logger.Logger
}

func NewBaseSetHandler(redisHandler RedisHandler, logger logger.Logger) *BaseSetHandler {
	return &BaseSetHandler{
		redisHandler: redisHandler,
		logger:       logger,
	}
}

func (h *BaseSetHandler) Add(tenantID string, key string, member string, opts ...map[string]any) *infra_error.AppError {
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	err := h.redisHandler.SAdd(formattedKey, member)
	if err != nil {
		h.logger.Error("Failed to add member to set", "error", err, "tenantID", tenantID, "key", key, "member", member)
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	if len(opts) > 0 {
		if ttl, ok := opts[0]["ttl"]; ok {
			if unitStr, ok := opts[0]["ttl_unit"]; ok {
				unit, err := time.ParseDuration(unitStr.(string))
				if err != nil {
					h.logger.Error("Failed to parse unit", "error", err, "tenantID", tenantID, "key", key, "member", member)
					return infra_error.Internal(infra_error.InternalInvalidArgument, err)
				}
				return h.redisHandler.Expire(formattedKey, ttl.(int), unit)
			}

		}
	}
	h.logger.Debug("Member added to set", "tenantID", tenantID, "key", key, "member", member)
	return nil
}

func (h *BaseSetHandler) Remove(tenantID string, key string, member string) *infra_error.AppError {
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	err := h.redisHandler.SRem(formattedKey, member)
	if err != nil {
		h.logger.Error("Failed to remove member from set", "error", err, "tenantID", tenantID, "key", key, "member", member)
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	h.logger.Debug("Member removed from set", "tenantID", tenantID, "key", key, "member", member)
	return nil
}

func (h *BaseSetHandler) Members(tenantID string, key string) ([]string, *infra_error.AppError) {
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	members, err := h.redisHandler.SMembers(formattedKey)
	if err != nil {
		h.logger.Error("Failed to get members from set", "error", err, "tenantID", tenantID, "key", key)
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return members, nil
}

func (h *BaseSetHandler) Clear(tenantID string, key string) *infra_error.AppError {
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	err := h.redisHandler.Clear(formattedKey)
	if err != nil {
		h.logger.Error("Failed to clear set", "error", err, "tenantID", tenantID, "key", key)
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	h.logger.Debug("Set cleared", "tenantID", tenantID, "key", key)
	return nil
}
