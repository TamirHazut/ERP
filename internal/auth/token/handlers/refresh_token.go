package handlers

import (
	"fmt"
	"time"

	token "erp.localhost/internal/auth/token"
	redis "erp.localhost/internal/infra/db/redis"
	redis_handlers "erp.localhost/internal/infra/db/redis/handlers"
	erp_errors "erp.localhost/internal/infra/errors"
	logging "erp.localhost/internal/infra/logging"
	auth_models "erp.localhost/internal/infra/models/auth"
	redis_models "erp.localhost/internal/infra/models/db/redis"
	shared_models "erp.localhost/internal/infra/models/shared"
)

// RefreshTokenHandler handles refresh token operations in Redis
// Key pattern: refresh_tokens:{tenant_id}:{user_id}:{token_id}
// Note: Multiple refresh tokens per user per tenant are allowed (for different devices/sessions)
type RefreshTokenHandler struct {
	keyHandler redis_handlers.KeyHandler[auth_models.RefreshToken]
	tokenIndex *token.TokenIndex
	logger     *logging.Logger
}

// NewRefreshTokenHandler creates a new RefreshTokenHandler
func NewRefreshTokenHandler(keyHandler redis_handlers.KeyHandler[auth_models.RefreshToken], tokenIndex *token.TokenIndex, logger *logging.Logger) *RefreshTokenHandler {
	if logger == nil {
		logger = logging.NewLogger(shared_models.ModuleAuth)
	}
	if keyHandler == nil {
		dbHandler := redis.NewBaseRedisHandler(redis_models.KeyPrefix(redis_models.RedisKeyRefreshToken))
		keyHandler = redis_handlers.NewBaseKeyHandler[auth_models.RefreshToken](dbHandler, logger)
	}
	return &RefreshTokenHandler{
		keyHandler: keyHandler,
		tokenIndex: tokenIndex,
		logger:     logger,
	}
}

// Store stores a refresh token in Redis
// Key: refresh_tokens:{tenant_id}:{user_id}:{token_id}
// tokenID should be unique (e.g., JTI from JWT or a UUID)
func (h *RefreshTokenHandler) Store(tenantID string, userID string, tokenID string, refreshToken auth_models.RefreshToken) error {
	if err := refreshToken.Validate(); err != nil {
		h.logger.Error("Failed to validate refresh token", "error", err)
		return err
	}

	// Ensure tenant_id and user_id match
	if refreshToken.TenantID != tenantID {
		return erp_errors.Validation(erp_errors.ValidationInvalidFormat, "tenant_id mismatch")
	}
	if refreshToken.UserID != userID {
		return erp_errors.Validation(erp_errors.ValidationInvalidFormat, "user_id mismatch")
	}

	// Use composite key: user_id:token_id
	key := userID + ":" + tokenID
	err := h.keyHandler.Set(tenantID, key, refreshToken)
	if err != nil {
		h.logger.Error("Failed to store refresh token", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return err
	}

	// Add to token index
	if h.tokenIndex != nil {
		if err := h.tokenIndex.AddRefreshToken(tenantID, userID, tokenID); err != nil {
			// Log error but don't fail - index is for optimization
			h.logger.Warn("Failed to add refresh token to index", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		}
	}

	h.logger.Debug("Refresh token stored", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}

func (h *RefreshTokenHandler) GetOne(tenantID string, userID string, tokenID string) (*auth_models.RefreshToken, error) {
	key := fmt.Sprintf("%s:%s", userID, tokenID)
	token, err := h.keyHandler.GetOne(tenantID, key)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Get retrieves a refresh token from Redis
func (h *RefreshTokenHandler) GetAll(tenantID string, userID string) ([]auth_models.RefreshToken, error) {
	tokens, err := h.keyHandler.GetAll(tenantID, userID)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// Validate checks if a refresh token is valid (exists, not revoked, not expired)
func (h *RefreshTokenHandler) Validate(tenantID string, userID string, tokenID string) (*auth_models.RefreshToken, error) {
	token, err := h.GetOne(tenantID, userID, tokenID)
	if err != nil {
		return nil, err
	}

	// Check if revoked
	if token.IsRevoked {
		return nil, erp_errors.Auth(erp_errors.AuthTokenRevoked)
	}

	// Check if expired
	if time.Now().After(token.ExpiresAt) {
		return nil, erp_errors.Auth(erp_errors.AuthRefreshTokenExpired)
	}

	return token, nil
}

// Revoke revokes a single refresh token
func (h *RefreshTokenHandler) Revoke(tenantID string, userID string, tokenID string, revokedBy string) error {
	token, err := h.GetOne(tenantID, userID, tokenID)
	if err != nil {
		return err
	}

	now := time.Now()
	token.IsRevoked = true
	token.RevokedAt = now
	token.RevokedBy = revokedBy
	key := userID + ":" + tokenID
	err = h.keyHandler.Update(tenantID, key, *token)
	if err != nil {
		h.logger.Error("Failed to revoke refresh token", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return err
	}

	h.logger.Debug("Refresh token revoked", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}

// RevokeAll revokes all refresh tokens for a user within a tenant
func (h *RefreshTokenHandler) RevokeAll(tenantID string, userID string, revokedBy string) error {
	if h.tokenIndex == nil {
		return erp_errors.Internal(erp_errors.InternalUnexpectedError, fmt.Errorf("token index not initialized"))
	}

	// Get all token IDs from index
	tokenIDs, err := h.tokenIndex.GetRefreshTokens(tenantID, userID)
	if err != nil {
		h.logger.Error("Failed to get refresh tokens from index", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	// Revoke each token
	for _, tokenID := range tokenIDs {
		if err := h.Revoke(tenantID, userID, tokenID, revokedBy); err != nil {
			// Log error but continue with other tokens
			h.logger.Warn("Failed to revoke refresh token", "error", err, "tenantID", tenantID, "tokenID", tokenID)
		}
	}

	// Clear the index
	if err := h.tokenIndex.ClearRefreshTokens(tenantID, userID); err != nil {
		h.logger.Warn("Failed to clear refresh tokens index", "error", err, "tenantID", tenantID, "userID", userID)
	}

	h.logger.Debug("All refresh tokens revoked", "tenantID", tenantID, "userID", userID, "count", len(tokenIDs))
	return nil
}

// UpdateLastUsed updates the LastUsedAt timestamp for a refresh token
func (h *RefreshTokenHandler) UpdateLastUsed(tenantID string, userID string, tokenID string) error {
	token, err := h.GetOne(tenantID, userID, tokenID)
	if err != nil {
		return err
	}

	token.LastUsedAt = time.Now()

	key := userID + ":" + tokenID
	err = h.keyHandler.Update(tenantID, key, *token)
	if err != nil {
		h.logger.Error("Failed to update refresh token last used", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return err
	}

	return nil
}

// Delete removes a refresh token from Redis (hard delete)
func (h *RefreshTokenHandler) Delete(tenantID string, userID string, tokenID string) error {
	key := userID + ":" + tokenID

	// Remove from index
	if h.tokenIndex != nil {
		if err := h.tokenIndex.RemoveRefreshToken(tenantID, userID, tokenID); err != nil {
			h.logger.Warn("Failed to remove refresh token from index", "error", err, "tenantID", tenantID, "tokenID", tokenID)
		}
	}

	err := h.keyHandler.Delete(tenantID, key)
	if err != nil {
		h.logger.Error("Failed to delete refresh token", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return err
	}

	h.logger.Debug("Refresh token deleted", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}
