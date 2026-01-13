package token

import (
	"time"

	"erp.localhost/internal/infra/db/redis"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
)

// RefreshTokenHandler handles refresh token operations in Redis
// Single token per user design - Key pattern: refresh_tokens:{tenant_id}:{user_id}
// Stores only ONE refresh token per user - new logins replace existing tokens
type RefreshTokenHandler struct {
	keyHandler redis.KeyHandler[model_auth.RefreshToken]
	logger     logger.Logger
}

// NewRefreshTokenHandler creates a new RefreshTokenHandler
func NewRefreshTokenHandler(keyHandler redis.KeyHandler[model_auth.RefreshToken], tokenIndex *TokenIndex, logger logger.Logger) *RefreshTokenHandler {
	// tokenIndex parameter kept for backward compatibility but no longer used
	return &RefreshTokenHandler{
		keyHandler: keyHandler,
		logger:     logger,
	}
}

// Store stores a refresh token in Redis (replaces existing token if present)
// Key: refresh_tokens:{tenant_id}:{user_id}
// Single token per user - automatically replaces any existing token
func (h *RefreshTokenHandler) Store(tenantID string, userID string, refreshToken model_auth.RefreshToken) error {
	if err := refreshToken.Validate(); err != nil {
		h.logger.Error("Failed to validate refresh token", "error", err)
		return err
	}

	// Ensure tenant_id and user_id match
	if refreshToken.TenantID != tenantID {
		return infra_error.Validation(infra_error.ValidationInvalidFormat, "tenant_id mismatch")
	}
	if refreshToken.UserID != userID {
		return infra_error.Validation(infra_error.ValidationInvalidFormat, "user_id mismatch")
	}

	// Check if existing token exists (for logging)
	existingToken, _ := h.GetOne(tenantID, userID)
	if existingToken != nil {
		h.logger.Info("Replacing existing refresh token", "tenantID", tenantID, "userID", userID)
	}

	// Store token using userID as key (automatically replaces old token)
	err := h.keyHandler.Set(tenantID, userID, refreshToken)
	if err != nil {
		h.logger.Error("Failed to store refresh token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	h.logger.Debug("Refresh token stored", "tenantID", tenantID, "userID", userID)
	return nil
}

// GetOne retrieves the single refresh token for a user from Redis
func (h *RefreshTokenHandler) GetOne(tenantID string, userID string) (*model_auth.RefreshToken, error) {
	token, err := h.keyHandler.GetOne(tenantID, userID)
	if err != nil {
		h.logger.Debug("Refresh token not found", "tenantID", tenantID, "userID", userID)
		return nil, err
	}
	return token, nil
}

// Validate checks if a refresh token is valid (exists, not revoked, not expired)
func (h *RefreshTokenHandler) Validate(tenantID string, userID string) (*model_auth.RefreshToken, error) {
	token, err := h.GetOne(tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Check if revoked
	if token.IsRevoked {
		return nil, infra_error.Auth(infra_error.AuthTokenRevoked)
	}

	// Check if expired
	if time.Now().After(token.ExpiresAt) {
		return nil, infra_error.Auth(infra_error.AuthRefreshTokenExpired)
	}

	return token, nil
}

// Revoke revokes the single refresh token for a user
func (h *RefreshTokenHandler) Revoke(tenantID string, userID string, revokedBy string) error {
	token, err := h.GetOne(tenantID, userID)
	if err != nil {
		// No token to revoke
		h.logger.Debug("No refresh token to revoke", "tenantID", tenantID, "userID", userID)
		return nil
	}

	now := time.Now()
	token.IsRevoked = true
	token.RevokedAt = now
	token.RevokedBy = revokedBy
	err = h.keyHandler.Update(tenantID, userID, *token)
	if err != nil {
		h.logger.Error("Failed to revoke refresh token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	h.logger.Debug("Refresh token revoked", "tenantID", tenantID, "userID", userID, "revokedBy", revokedBy)
	return nil
}

// UpdateLastUsed updates the LastUsedAt timestamp for a refresh token
func (h *RefreshTokenHandler) UpdateLastUsed(tenantID string, userID string, tokenString string) error {
	token, err := h.GetOne(tenantID, userID)
	if err != nil {
		return err
	}

	token.LastUsedAt = time.Now()

	err = h.keyHandler.Update(tenantID, userID, *token)
	if err != nil {
		h.logger.Error("Failed to update refresh token last used", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	return nil
}

// Delete permanently removes the refresh token from Redis (hard delete)
func (h *RefreshTokenHandler) Delete(tenantID string, userID string) error {
	err := h.keyHandler.Delete(tenantID, userID)
	if err != nil {
		h.logger.Error("Failed to delete refresh token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	h.logger.Debug("Refresh token deleted", "tenantID", tenantID, "userID", userID)
	return nil
}

// ScanKeys returns all refresh token keys for a tenant
// Used for tenant-level token management (revoke/delete all tokens for a tenant)
func (h *RefreshTokenHandler) ScanKeys(tenantID string) ([]string, error) {
	// Pattern: all user IDs in this tenant (tenantID:*)
	keys, err := h.keyHandler.ScanKeys(tenantID, "*")
	if err != nil {
		h.logger.Error("Failed to scan refresh token keys", "error", err, "tenantID", tenantID)
		return nil, err
	}

	h.logger.Debug("Refresh token keys scanned", "tenantID", tenantID, "keys_found", len(keys))
	return keys, nil
}

// DeleteByPattern deletes all refresh tokens for a tenant
// Returns the number of tokens deleted
func (h *RefreshTokenHandler) DeleteByPattern(tenantID string) (int, error) {
	// Pattern: all user IDs in this tenant (tenantID:*)
	count, err := h.keyHandler.DeleteByPattern(tenantID, "*")
	if err != nil {
		h.logger.Error("Failed to delete refresh tokens by pattern", "error", err, "tenantID", tenantID)
		return 0, err
	}

	h.logger.Info("Refresh tokens deleted for tenant", "tenantID", tenantID, "tokens_deleted", count)
	return count, nil
}
