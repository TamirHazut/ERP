package token

import (
	"fmt"
	"time"

	"erp.localhost/internal/infra/db/redis"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth_cache "erp.localhost/internal/infra/model/auth/cache"
)

//go:generate mockgen -destination=mock/mock_token_handler.go -package=mock erp.localhost/internal/auth/token TokenHandler
type TokenHandler[T any] interface {
	Store(tenantID string, userID string, tokenID string, value T) error
	GetOne(tenantID string, userID string, tokenID string) (*T, error)
	GetAll(tenantID string, userID string) ([]T, error)
	Validate(tenantID string, userID string, tokenID string) (*T, error)
	Revoke(tenantID string, userID string, tokenID string, revokedBy string) error
	RevokeAll(tenantID string, userID string, revokedBy string) error
	Delete(tenantID string, userID string, tokenID string) error
}

// AccessTokenHandler handles access token operations in Redis
// Key pattern: tokens:{tenant_id}:{token_id}
type AccessTokenHandler struct {
	keyHandler redis.KeyHandler[model_auth_cache.TokenMetadata]
	tokenIndex *TokenIndex
	logger     logger.Logger
}

// NewAccessTokenHandler creates a new AccessTokenHandler
func NewAccessTokenHandler(keyHandler redis.KeyHandler[model_auth_cache.TokenMetadata], tokenIndex *TokenIndex, logger logger.Logger) *AccessTokenHandler {
	return &AccessTokenHandler{
		keyHandler: keyHandler,
		tokenIndex: tokenIndex,
		logger:     logger,
	}
}

// Store stores an access token in Redis
// Key: tokens:{tenant_id}:{user_id}:{token_id}
func (h *AccessTokenHandler) Store(tenantID string, userID string, tokenID string, metadata model_auth_cache.TokenMetadata) error {
	// Basic validation
	if metadata.TokenID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TokenID")
	}
	if metadata.UserID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "UserID")
	}
	if metadata.TenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantID")
	}

	// Ensure tenant_id matches
	if metadata.TenantID != tenantID {
		return infra_error.Validation(infra_error.ValidationInvalidFormat, "tenant_id mismatch")
	}
	if metadata.UserID != userID {
		return infra_error.Validation(infra_error.ValidationInvalidFormat, "user_id mismatch")
	}

	// Ensure token_id matches
	if metadata.TokenID != tokenID {
		return infra_error.Validation(infra_error.ValidationInvalidFormat, "token_id mismatch")
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
func (h *AccessTokenHandler) GetOne(tenantID string, userID string, tokenID string) (*model_auth_cache.TokenMetadata, error) {
	key := fmt.Sprintf("%s:%s", userID, tokenID)
	token, err := h.keyHandler.GetOne(tenantID, key)
	if err != nil {
		h.logger.Debug("Access token not found", "tenantID", tenantID, "tokenID", tokenID)
		return nil, err
	}

	return token, nil
}

// GetAll retrieves all access tokens from Redis
func (h *AccessTokenHandler) GetAll(tenantID string, userID string) ([]model_auth_cache.TokenMetadata, error) {
	tokens, err := h.keyHandler.GetAll(tenantID, userID)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// Validate checks if a token is valid (exists, not revoked, not expired)
func (h *AccessTokenHandler) Validate(tenantID string, userID string, tokenID string) (*model_auth_cache.TokenMetadata, error) {
	metadata, err := h.GetOne(tenantID, userID, tokenID)
	if err != nil {
		return nil, err
	}

	// Check if revoked
	if metadata.Revoked {
		return nil, infra_error.Auth(infra_error.AuthTokenRevoked)
	}

	// Check if expired
	if time.Now().After(metadata.ExpiresAt) {
		return nil, infra_error.Auth(infra_error.AuthTokenExpired)
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
		return infra_error.Internal(infra_error.InternalUnexpectedError, fmt.Errorf("token index not initialized"))
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
