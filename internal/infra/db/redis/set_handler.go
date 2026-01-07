package redis

import (
	"fmt"
	"time"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
)

//go:generate mockgen -destination=mock/mock_set_handler.go -package=mock erp.localhost/internal/infra/db/redis SetHandler
type SetHandler interface {
	Add(tenantID string, key string, member string, opts ...map[string]any) error
	Remove(tenantID string, key string, member string) error
	Members(tenantID string, key string) ([]string, error)
	Clear(tenantID string, key string) error
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

func (h *BaseSetHandler) Add(tenantID string, key string, member string, opts ...map[string]any) error {
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
				err = h.redisHandler.Expire(formattedKey, ttl.(int), unit)
				if err != nil {
					h.logger.Error("Failed to set TTL on set", "error", err, "tenantID", tenantID, "key", key, "member", member)
					return infra_error.Internal(infra_error.InternalInvalidArgument, err)
				}
			}

		}
	}
	h.logger.Debug("Member added to set", "tenantID", tenantID, "key", key, "member", member)
	return nil
}

func (h *BaseSetHandler) Remove(tenantID string, key string, member string) error {
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	err := h.redisHandler.SRem(formattedKey, member)
	if err != nil {
		h.logger.Error("Failed to remove member from set", "error", err, "tenantID", tenantID, "key", key, "member", member)
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	h.logger.Debug("Member removed from set", "tenantID", tenantID, "key", key, "member", member)
	return nil
}

func (h *BaseSetHandler) Members(tenantID string, key string) ([]string, error) {
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	members, err := h.redisHandler.SMembers(formattedKey)
	if err != nil {
		h.logger.Error("Failed to get members from set", "error", err, "tenantID", tenantID, "key", key)
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return members, nil
}

func (h *BaseSetHandler) Clear(tenantID string, key string) error {
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	err := h.redisHandler.Clear(formattedKey)
	if err != nil {
		h.logger.Error("Failed to clear set", "error", err, "tenantID", tenantID, "key", key)
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	h.logger.Debug("Set cleared", "tenantID", tenantID, "key", key)
	return nil
}
