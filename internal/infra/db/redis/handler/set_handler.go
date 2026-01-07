package handler

import (
	"fmt"
	"time"

	erp_errors "erp.localhost/internal/infra/error"
	logging "erp.localhost/internal/infra/logging"
	shared_models "erp.localhost/internal/infra/model/shared"
)

type BaseSetHandler struct {
	redisHandler RedisHandler
	logger       *logging.Logger
}

func NewBaseSetHandler(redisHandler RedisHandler, logger *logging.Logger) *BaseSetHandler {
	if logger == nil {
		logger = logging.NewLogger(shared_models.ModuleDB)
	}
	if redisHandler == nil {
		logger.Error("RedisHandler is nil")
		return nil
	}
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
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	if len(opts) > 0 {
		if ttl, ok := opts[0]["ttl"]; ok {
			if unitStr, ok := opts[0]["ttl_unit"]; ok {
				unit, err := time.ParseDuration(unitStr.(string))
				if err != nil {
					h.logger.Error("Failed to parse unit", "error", err, "tenantID", tenantID, "key", key, "member", member)
					return erp_errors.Internal(erp_errors.InternalInvalidArgument, err)
				}
				err = h.redisHandler.Expire(formattedKey, ttl.(int), unit)
				if err != nil {
					h.logger.Error("Failed to set TTL on set", "error", err, "tenantID", tenantID, "key", key, "member", member)
					return erp_errors.Internal(erp_errors.InternalInvalidArgument, err)
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
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	h.logger.Debug("Member removed from set", "tenantID", tenantID, "key", key, "member", member)
	return nil
}

func (h *BaseSetHandler) Members(tenantID string, key string) ([]string, error) {
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	members, err := h.redisHandler.SMembers(formattedKey)
	if err != nil {
		h.logger.Error("Failed to get members from set", "error", err, "tenantID", tenantID, "key", key)
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return members, nil
}

func (h *BaseSetHandler) Clear(tenantID string, key string) error {
	formattedKey := fmt.Sprintf("%s:%s", tenantID, key)
	err := h.redisHandler.Clear(formattedKey)
	if err != nil {
		h.logger.Error("Failed to clear set", "error", err, "tenantID", tenantID, "key", key)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	h.logger.Debug("Set cleared", "tenantID", tenantID, "key", key)
	return nil
}
