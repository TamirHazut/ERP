package token

import (
	"time"

	"erp.localhost/internal/infra/db/redis"
	redis_handler "erp.localhost/internal/infra/db/redis/handler"
	erp_errors "erp.localhost/internal/infra/error"
	logging "erp.localhost/internal/infra/logging"
	redis_models "erp.localhost/internal/infra/model/db/redis"
	shared_models "erp.localhost/internal/infra/model/shared"
)

var (
	accessTokenTTLUnit  = time.Minute
	refreshTokenTTLUnit = time.Hour
	accessTokenTTL      = 15
	refreshTokenTTL     = 2
)

// TokenIndex manages token indices using Redis Sets
// Key pattern: user_access_tokens:{tenant_id}:{user_id} -> Set of token_ids
// Key pattern: user_refresh_tokens:{tenant_id}:{user_id} -> Set of token_ids
type TokenIndex struct {
	accessTokenSetHandler  redis_handler.SetHandler
	refreshTokenSetHandler redis_handler.SetHandler
	logger                 *logging.Logger
}

// NewTokenIndex creates a new TokenIndex
func NewTokenIndex(accessTokenSetHandler redis_handler.SetHandler, refreshTokenSetHandler redis_handler.SetHandler) *TokenIndex {
	logger := logging.NewLogger(shared_models.ModuleAuth)
	if accessTokenSetHandler == nil {
		accessTokenRedisHandler := redis.NewBaseRedisHandler(redis_models.KeyPrefix(redis_models.RedisKeyUserAccessTokens))
		accessTokenSetHandler = redis_handler.NewBaseSetHandler(accessTokenRedisHandler, logger)
	}
	if refreshTokenSetHandler == nil {
		refreshTokenRedisHandler := redis.NewBaseRedisHandler(redis_models.KeyPrefix(redis_models.RedisKeyUserRefreshTokens))
		refreshTokenSetHandler = redis_handler.NewBaseSetHandler(refreshTokenRedisHandler, logger)
	}

	return &TokenIndex{
		accessTokenSetHandler:  accessTokenSetHandler,
		refreshTokenSetHandler: refreshTokenSetHandler,
		logger:                 logger,
	}
}

// AddAccessToken adds a token_id to the user's access token index
func (t *TokenIndex) AddAccessToken(tenantID string, userID string, tokenID string) error {
	opts := map[string]any{
		"ttl":      accessTokenTTL,
		"ttl_unit": accessTokenTTLUnit,
	}
	err := t.accessTokenSetHandler.Add(tenantID, userID, tokenID, opts)
	if err != nil {
		t.logger.Error("Failed to add access token to index", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	t.logger.Debug("Access token added to index", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}

// RemoveAccessToken removes a token_id from the user's access token index
func (t *TokenIndex) RemoveAccessToken(tenantID string, userID string, tokenID string) error {
	err := t.accessTokenSetHandler.Remove(tenantID, userID, tokenID)
	if err != nil {
		t.logger.Error("Failed to remove access token from index", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}

	t.logger.Debug("Access token removed from index", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}

// GetAccessTokens returns all token_ids for a user's access tokens
func (t *TokenIndex) GetAccessTokens(tenantID string, userID string) ([]string, error) {
	tokenIDs, err := t.accessTokenSetHandler.Members(tenantID, userID)
	if err != nil {
		t.logger.Error("Failed to get access tokens from index", "error", err, "tenantID", tenantID, "userID", userID)
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}

	return tokenIDs, nil
}

// AddRefreshToken adds a token_id to the user's refresh token index
func (t *TokenIndex) AddRefreshToken(tenantID string, userID string, tokenID string) error {
	opts := map[string]any{
		"ttl":      refreshTokenTTL,
		"ttl_unit": refreshTokenTTLUnit,
	}
	err := t.refreshTokenSetHandler.Add(tenantID, userID, tokenID, opts)
	if err != nil {
		t.logger.Error("Failed to add refresh token to index", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}

	t.logger.Debug("Refresh token added to index", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}

// RemoveRefreshToken removes a token_id from the user's refresh token index
func (t *TokenIndex) RemoveRefreshToken(tenantID string, userID string, tokenID string) error {
	err := t.refreshTokenSetHandler.Remove(tenantID, userID, tokenID)
	if err != nil {
		t.logger.Error("Failed to remove refresh token from index", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}

	t.logger.Debug("Refresh token removed from index", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}

// GetRefreshTokens returns all token_ids for a user's refresh tokens
func (t *TokenIndex) GetRefreshTokens(tenantID string, userID string) ([]string, error) {
	tokenIDs, err := t.refreshTokenSetHandler.Members(tenantID, userID)
	if err != nil {
		t.logger.Error("Failed to get refresh tokens from index", "error", err, "tenantID", tenantID, "userID", userID)
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}

	return tokenIDs, nil
}

// ClearAccessTokens removes all access tokens from the index (used when revoking all)
func (t *TokenIndex) ClearAccessTokens(tenantID string, userID string) error {
	err := t.accessTokenSetHandler.Clear(tenantID, userID)
	if err != nil {
		t.logger.Error("Failed to clear access tokens index", "error", err, "tenantID", tenantID, "userID", userID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}

	t.logger.Debug("Access tokens index cleared", "tenantID", tenantID, "userID", userID)
	return nil
}

// ClearRefreshTokens removes all refresh tokens from the index (used when revoking all)
func (t *TokenIndex) ClearRefreshTokens(tenantID string, userID string) error {
	err := t.refreshTokenSetHandler.Clear(tenantID, userID)
	if err != nil {
		t.logger.Error("Failed to clear refresh tokens index", "error", err, "tenantID", tenantID, "userID", userID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}

	t.logger.Debug("Refresh tokens index cleared", "tenantID", tenantID, "userID", userID)
	return nil
}
