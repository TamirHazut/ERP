package handlers

import (
	"fmt"
	"time"

	token "erp.localhost/internal/auth/token"
	redis "erp.localhost/internal/infra/db/redis"
	redis_handlers "erp.localhost/internal/infra/db/redis/handlers"
	erp_errors "erp.localhost/internal/infra/error"
	logging "erp.localhost/internal/infra/logging"
	auth_cache_models "erp.localhost/internal/infra/model/auth/cache"
	redis_models "erp.localhost/internal/infra/model/db/redis"
	shared_models "erp.localhost/internal/infra/model/shared"
)

// AccessTokenHandler handles access token operations in Redis
// Key pattern: tokens:{tenant_id}:{token_id}
type AccessTokenHandler struct {
	keyHandler redis_handlers.KeyHandler[auth_cache_models.TokenMetadata]
	tokenIndex *token.TokenIndex
	logger     *logging.Logger
}

// NewAccessTokenHandler creates a new AccessTokenHandler
func NewAccessTokenHandler(keyHandler redis_handlers.KeyHandler[auth_cache_models.TokenMetadata], tokenIndex *token.TokenIndex, logger *logging.Logger) *AccessTokenHandler {
	if logger == nil {
		logger = logging.NewLogger(shared_models.ModuleAuth)
	}
	if keyHandler == nil {
		dbHandler := redis.NewBaseRedisHandler(redis_models.KeyPrefix(redis_models.RedisKeyToken))
		keyHandler = redis_handlers.NewBaseKeyHandler[auth_cache_models.TokenMetadata](dbHandler, logger)
	}
	return &AccessTokenHandler{
		keyHandler: keyHandler,
		tokenIndex: tokenIndex,
		logger:     logger,
	}
}

// Store stores an access token in Redis
// Key: tokens:{tenant_id}:{user_id}:{token_id}
func (h *AccessTokenHandler) Store(tenantID string, userID string, tokenID string, metadata auth_cache_models.TokenMetadata) error {
	// Basic validation
	if metadata.TokenID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "TokenID")
	}
	if metadata.UserID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "UserID")
	}
	if metadata.TenantID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "TenantID")
	}

	// Ensure tenant_id matches
	if metadata.TenantID != tenantID {
		return erp_errors.Validation(erp_errors.ValidationInvalidFormat, "tenant_id mismatch")
	}
	if metadata.UserID != userID {
		return erp_errors.Validation(erp_errors.ValidationInvalidFormat, "user_id mismatch")
	}

	// Ensure token_id matches
	if metadata.TokenID != tokenID {
		return erp_errors.Validation(erp_errors.ValidationInvalidFormat, "token_id mismatch")
	}

	err := h.keyHandler.Set(tenantID, tokenID, metadata)
	if err != nil {
		h.logger.Error("Failed to store access token", "error", err, "tenantID", tenantID, "tokenID", tokenID)
		return err
	}

	// Add to token index
	if h.tokenIndex != nil {
		if err := h.tokenIndex.AddAccessToken(tenantID, metadata.UserID, tokenID); err != nil {
			// Log error but don't fail - index is for optimization
			h.logger.Warn("Failed to add access token to index", "error", err, "tenantID", tenantID, "userID", metadata.UserID, "tokenID", tokenID)
		}
	}

	h.logger.Debug("Access token stored", "tenantID", tenantID, "tokenID", tokenID)
	return nil
}

// GetOne retrieves an access token from Redis
func (h *AccessTokenHandler) GetOne(tenantID string, userID string, tokenID string) (*auth_cache_models.TokenMetadata, error) {
	key := fmt.Sprintf("%s:%s", userID, tokenID)
	token, err := h.keyHandler.GetOne(tenantID, key)
	if err != nil {
		h.logger.Debug("Access token not found", "tenantID", tenantID, "tokenID", tokenID)
		return nil, err
	}

	return token, nil
}

// GetAll retrieves all access tokens from Redis
func (h *AccessTokenHandler) GetAll(tenantID string, userID string) ([]auth_cache_models.TokenMetadata, error) {
	tokens, err := h.keyHandler.GetAll(tenantID, userID)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// Validate checks if a token is valid (exists, not revoked, not expired)
func (h *AccessTokenHandler) Validate(tenantID string, userID string, tokenID string) (*auth_cache_models.TokenMetadata, error) {
	metadata, err := h.GetOne(tenantID, userID, tokenID)
	if err != nil {
		return nil, err
	}

	// Check if revoked
	if metadata.Revoked {
		return nil, erp_errors.Auth(erp_errors.AuthTokenRevoked)
	}

	// Check if expired
	if time.Now().After(metadata.ExpiresAt) {
		return nil, erp_errors.Auth(erp_errors.AuthTokenExpired)
	}

	return metadata, nil
}

// Revoke revokes a single access token
func (h *AccessTokenHandler) Revoke(tenantID string, userID string, tokenID string, revokedBy string) error {
	metadata, err := h.GetOne(tenantID, userID, tokenID)
	if err != nil {
		return err
	}

	now := time.Now()
	metadata.Revoked = true
	metadata.RevokedAt = &now
	metadata.RevokedBy = revokedBy

	err = h.keyHandler.Update(tenantID, tokenID, *metadata)
	if err != nil {
		h.logger.Error("Failed to revoke access token", "error", err, "tenantID", tenantID, "tokenID", tokenID)
		return err
	}

	h.logger.Debug("Access token revoked", "tenantID", tenantID, "tokenID", tokenID, "revokedBy", revokedBy)
	return nil
}

// RevokeAll revokes all access tokens for a user within a tenant
func (h *AccessTokenHandler) RevokeAll(tenantID string, userID string, revokedBy string) error {
	if h.tokenIndex == nil {
		return erp_errors.Internal(erp_errors.InternalUnexpectedError, fmt.Errorf("token index not initialized"))
	}

	// Get all token IDs from index
	tokenIDs, err := h.tokenIndex.GetAccessTokens(tenantID, userID)
	if err != nil {
		h.logger.Error("Failed to get access tokens from index", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	// Revoke each token
	for _, tokenID := range tokenIDs {
		if err := h.Revoke(tenantID, userID, tokenID, revokedBy); err != nil {
			// Log error but continue with other tokens
			h.logger.Warn("Failed to revoke access token", "error", err, "tenantID", tenantID, "tokenID", tokenID)
		}
	}

	// Clear the index
	if err := h.tokenIndex.ClearAccessTokens(tenantID, userID); err != nil {
		h.logger.Warn("Failed to clear access tokens index", "error", err, "tenantID", tenantID, "userID", userID)
	}

	h.logger.Debug("All access tokens revoked", "tenantID", tenantID, "userID", userID, "count", len(tokenIDs))
	return nil
}

// Delete removes a token from Redis (hard delete)
func (h *AccessTokenHandler) Delete(tenantID string, userID string, tokenID string) error {
	// Get token to find userID for index removal
	metadata, err := h.GetOne(tenantID, userID, tokenID)
	if err == nil && h.tokenIndex != nil {
		// Remove from index
		if err := h.tokenIndex.RemoveAccessToken(tenantID, metadata.UserID, tokenID); err != nil {
			h.logger.Warn("Failed to remove access token from index", "error", err, "tenantID", tenantID, "tokenID", tokenID)
		}
	}

	err = h.keyHandler.Delete(tenantID, tokenID)
	if err != nil {
		h.logger.Error("Failed to delete access token", "error", err, "tenantID", tenantID, "tokenID", tokenID)
		return err
	}

	h.logger.Debug("Access token deleted", "tenantID", tenantID, "tokenID", tokenID)
	return nil
}
