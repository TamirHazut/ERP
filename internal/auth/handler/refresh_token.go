package handler

import (
	"time"

	"erp.localhost/internal/auth/token"
	"erp.localhost/internal/infra/db/redis"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1_cache "erp.localhost/internal/infra/model/auth/v1/cache"
	"erp.localhost/internal/infra/model/auth/validator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RefreshTokenHandler struct {
	handler redis.KeyHandler[authv1_cache.RefreshToken]
	logger  logger.Logger
}

func NewRefreshTokenHandler(logger logger.Logger) (*RefreshTokenHandler, error) {
	handler, err := token.NewRefreshTokenKeyHandler(logger)
	if err != nil {
		return nil, err
	}
	return &RefreshTokenHandler{
		handler: handler,
		logger:  logger,
	}, nil
}

// Store stores a refresh token in Redis (replaces existing token if present)
// Key: refresh_tokens:{tenant_id}:{user_id}
// Single token per user - automatically replaces any existing token
func (h *RefreshTokenHandler) Store(tenantID string, userID string, refreshToken *authv1_cache.RefreshToken) error {
	if err := validator.ValidateRefreshToken(refreshToken); err != nil {
		h.logger.Error("Failed to validate refresh token", "error", err)
		return err
	}

	// Ensure tenant_id and user_id match
	if refreshToken.GetTenantId() != tenantID || refreshToken.GetUserId() != userID {
		h.logger.Warn("tenant_id or user_id mismatch", "tenantID", tenantID, "refresh_token_tenantID", refreshToken.GetTenantId(), "userID", userID, "refresh_token_userID", refreshToken.GetUserId())
		return infra_error.Validation(infra_error.ValidationInvalidFormat, "tenant_id or user_id mismatch")
	}

	ttl := time.Until(refreshToken.ExpiresAt.AsTime())
	opts := map[string]any{"ttl": ttl}

	// Store token using userID as key (automatically replaces old token)
	err := h.handler.Set(tenantID, userID, refreshToken, opts)
	if err != nil {
		h.logger.Error("Failed to store refresh token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	h.logger.Debug("Refresh token stored", "tenantID", tenantID, "userID", userID)
	return nil
}

// GetOne retrieves the single refresh token for a user from Redis
func (h *RefreshTokenHandler) GetOne(tenantID string, userID string) (*authv1_cache.RefreshToken, error) {
	token, err := h.handler.GetOne(tenantID, userID)
	if err != nil {
		h.logger.Debug("Refresh token not found", "tenantID", tenantID, "userID", userID)
		return nil, err
	}
	return token, nil
}

// Validate checks if a refresh token is valid (exists, not revoked, not expired)
func (h *RefreshTokenHandler) Validate(tenantID string, userID string) (*authv1_cache.RefreshToken, error) {
	token, err := h.GetOne(tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Check if revoked
	if token.Revoked {
		return nil, infra_error.Auth(infra_error.AuthTokenRevoked)
	}

	// Check if expired
	if time.Now().After(token.ExpiresAt.AsTime()) {
		return nil, infra_error.Auth(infra_error.AuthRefreshTokenExpired)
	}

	return token, nil
}

// Revoke revokes the single refresh token for a user
func (h *RefreshTokenHandler) Revoke(tenantID string, userID string, revokedBy string) error {
	token, err := h.GetOne(tenantID, userID)
	if err != nil || token == nil {
		// No token to revoke
		h.logger.Debug("No refresh token to revoke", "tenantID", tenantID, "userID", userID)
		return nil
	}

	err = h.Delete(tenantID, userID)
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

	token.LastUsedAt = timestamppb.Now()

	err = h.handler.Update(tenantID, userID, token)
	if err != nil {
		h.logger.Error("Failed to update refresh token last used", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	return nil
}

// Delete permanently removes the refresh token from Redis (hard delete)
func (h *RefreshTokenHandler) Delete(tenantID string, userID string) error {
	err := h.handler.Delete(tenantID, userID)
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
	keys, err := h.handler.ScanKeys(tenantID, "*")
	if err != nil {
		h.logger.Error("Failed to scan refresh token keys", "error", err, "tenantID", tenantID)
		return nil, err
	}

	h.logger.Debug("Refresh token keys scanned", "tenantID", tenantID, "keys_found", len(keys))
	return keys, nil
}

// DeleteByPattern deletes all refresh tokens for a tenant
// Returns the number of tokens deleted
func (h *RefreshTokenHandler) DeleteByPattern(tenantID, pattern string) (int, error) {
	// Pattern: all user IDs in this tenant (tenantID:*)
	count, err := h.handler.DeleteByPattern(tenantID, pattern)
	if err != nil {
		h.logger.Error("Failed to delete refresh tokens by pattern", "error", err, "tenantID", tenantID)
		return 0, err
	}

	h.logger.Info("Refresh tokens deleted for tenant", "tenantID", tenantID, "tokens_deleted", count)
	return count, nil
}
