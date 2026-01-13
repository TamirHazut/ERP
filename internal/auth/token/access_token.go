package token

import (
	"time"

	"erp.localhost/internal/infra/db/redis"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth_cache "erp.localhost/internal/infra/model/auth/cache"
)

//go:generate mockgen -destination=mock/mock_token_handler.go -package=mock erp.localhost/internal/auth/token TokenHandler
type TokenHandler[T any] interface {
	// Store stores a single token for a user (replaces existing if present)
	Store(tenantID string, userID string, value T) error
	// GetOne retrieves the single token for a user
	GetOne(tenantID string, userID string) (*T, error)
	// Validate checks if the token is valid (exists, not revoked, not expired)
	Validate(tenantID string, userID string) (*T, error)
	// Revoke revokes the single token for a user
	Revoke(tenantID string, userID string, revokedBy string) error
	// // RevokeAll revokes all the tokens that are related to a pattern
	// RevokeAll(pattern string, revokedBy string) error
	// ScanKeys finds all the keys that are related to a tenant
	ScanKeys(tenantID string) ([]string, error)
	// Delete permanently deletes the single token for a user
	Delete(tenantID string, userID string) error
	// Delete permanently deletes the tokens that match the pattern
	DeleteByPattern(tenantID string) (int, error)
}

// AccessTokenHandler handles access token operations in Redis
// Single token per user design - Key pattern: tokens:{tenant_id}:{user_id}
// Stores only ONE access token per user - new logins replace existing tokens
type AccessTokenHandler struct {
	keyHandler redis.KeyHandler[model_auth_cache.TokenMetadata]
	logger     logger.Logger
}

// NewAccessTokenHandler creates a new AccessTokenHandler
func NewAccessTokenHandler(keyHandler redis.KeyHandler[model_auth_cache.TokenMetadata], tokenIndex *TokenIndex, logger logger.Logger) *AccessTokenHandler {
	// tokenIndex parameter kept for backward compatibility but no longer used
	return &AccessTokenHandler{
		keyHandler: keyHandler,
		logger:     logger,
	}
}

// Store stores an access token in Redis (replaces existing token if present)
// Key: tokens:{tenant_id}:{user_id}
// Single token per user - automatically replaces any existing token
func (h *AccessTokenHandler) Store(tenantID string, userID string, metadata model_auth_cache.TokenMetadata) error {
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

	// Check if existing token exists (for logging)
	existingToken, _ := h.GetOne(tenantID, userID)
	if existingToken != nil {
		h.logger.Info("Replacing existing access token", "tenantID", tenantID, "userID", userID, "oldTokenID", existingToken.TokenID, "newTokenID", metadata.TokenID)
	}

	// Store token using userID as key (automatically replaces old token)
	err := h.keyHandler.Set(tenantID, userID, metadata)
	if err != nil {
		h.logger.Error("Failed to store access token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	h.logger.Debug("Access token stored", "tenantID", tenantID, "userID", userID, "tokenID", metadata.TokenID)
	return nil
}

// GetOne retrieves the single access token for a user from Redis
func (h *AccessTokenHandler) GetOne(tenantID string, userID string) (*model_auth_cache.TokenMetadata, error) {
	token, err := h.keyHandler.GetOne(tenantID, userID)
	if err != nil {
		h.logger.Debug("Access token not found", "tenantID", tenantID, "userID", userID)
		return nil, err
	}

	return token, nil
}

// Validate checks if a token is valid (exists, not revoked, not expired)
func (h *AccessTokenHandler) Validate(tenantID string, userID string) (*model_auth_cache.TokenMetadata, error) {
	metadata, err := h.GetOne(tenantID, userID)
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

// Revoke revokes the single access token for a user
func (h *AccessTokenHandler) Revoke(tenantID string, userID string, revokedBy string) error {
	metadata, err := h.GetOne(tenantID, userID)
	if err != nil {
		// No token to revoke
		h.logger.Debug("No access token to revoke", "tenantID", tenantID, "userID", userID)
		return nil
	}

	now := time.Now()
	metadata.Revoked = true
	metadata.RevokedAt = &now
	metadata.RevokedBy = revokedBy

	err = h.keyHandler.Update(tenantID, userID, *metadata)
	if err != nil {
		h.logger.Error("Failed to revoke access token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	h.logger.Debug("Access token revoked", "tenantID", tenantID, "userID", userID, "tokenID", metadata.TokenID, "revokedBy", revokedBy)
	return nil
}

// Delete permanently removes the access token from Redis (hard delete)
func (h *AccessTokenHandler) Delete(tenantID string, userID string) error {
	err := h.keyHandler.Delete(tenantID, userID)
	if err != nil {
		h.logger.Error("Failed to delete access token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	h.logger.Debug("Access token deleted", "tenantID", tenantID, "userID", userID)
	return nil
}

// ScanKeys returns all access token keys for a tenant
// Used for tenant-level token management (revoke/delete all tokens for a tenant)
func (h *AccessTokenHandler) ScanKeys(tenantID string) ([]string, error) {
	// Pattern: all user IDs in this tenant (tenantID:*)
	keys, err := h.keyHandler.ScanKeys(tenantID, "*")
	if err != nil {
		h.logger.Error("Failed to scan access token keys", "error", err, "tenantID", tenantID)
		return nil, err
	}

	h.logger.Debug("Access token keys scanned", "tenantID", tenantID, "keys_found", len(keys))
	return keys, nil
}

// DeleteByPattern deletes all access tokens for a tenant
// Returns the number of tokens deleted
func (h *AccessTokenHandler) DeleteByPattern(tenantID string) (int, error) {
	// Pattern: all user IDs in this tenant (tenantID:*)
	count, err := h.keyHandler.DeleteByPattern(tenantID, "*")
	if err != nil {
		h.logger.Error("Failed to delete access tokens by pattern", "error", err, "tenantID", tenantID)
		return 0, err
	}

	h.logger.Info("Access tokens deleted for tenant", "tenantID", tenantID, "tokens_deleted", count)
	return count, nil
}
