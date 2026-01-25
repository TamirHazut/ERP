package handler

import (
	"time"

	"erp.localhost/internal/auth/token"
	"erp.localhost/internal/infra/db/redis"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1_cache "erp.localhost/internal/infra/model/auth/v1/cache"
	validator_auth_cache "erp.localhost/internal/infra/model/auth/validator/cache"
)

//go:generate mockgen -destination=mock/mock_token_handler.go -package=mock erp.localhost/internal/auth/handler TokenHandler
type TokenHandler[T any] interface {
	// Store stores a single token for a user (replaces existing if present)
	Store(tenantID string, userID string, value *T) error
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
	DeleteByPattern(tenantID string, pattern string) (int, error)
}

// AccessTokenHandler handles access token operations in Redis
// Single token per user design - Key pattern: tokens:{tenant_id}:{user_id}
// Stores only ONE access token per user - new logins replace existing tokens
type AccessTokenHandler struct {
	handler redis.KeyHandler[authv1_cache.TokenMetadata]
	logger  logger.Logger
}

func NewAccessTokenHandler(logger logger.Logger) (*AccessTokenHandler, error) {
	handler, err := token.NewAccessTokenKeyHandler(logger)
	if err != nil {
		return nil, err
	}
	return &AccessTokenHandler{
		handler: handler,
		logger:  logger,
	}, nil
}

// Store stores an access token in Redis (replaces existing token if present)
// Key: tokens:{tenant_id}:{user_id}
// Single token per user - automatically replaces any existing token
func (h *AccessTokenHandler) Store(tenantID string, userID string, metadata *authv1_cache.TokenMetadata) error {
	if err := validator_auth_cache.ValidateTokenMetaData(metadata); err != nil {
		h.logger.Error("Failed to validate token", "error", err)
		return err
	}

	// Ensure tenant_id matches
	if metadata.TenantId != tenantID || metadata.UserId != userID {
		h.logger.Warn("tenant_id or user_id mismatch", "tenantID", tenantID, "token_tenantID", metadata.GetTenantId(), "userID", userID, "token_userID", metadata.GetUserId())
		return infra_error.Validation(infra_error.ValidationInvalidFormat, "tenant_id or user_id mismatch")
	}

	ttl := time.Until(metadata.ExpiresAt.AsTime())
	opts := map[string]any{"ttl": ttl}

	// Store token using userID as key (automatically replaces old token)
	err := h.handler.Set(tenantID, userID, metadata, opts)
	if err != nil {
		h.logger.Error("Failed to store access token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	h.logger.Debug("Access token stored", "tenantID", tenantID, "userID", userID)
	return nil
}

// GetOne retrieves the single access token for a user from Redis
func (h *AccessTokenHandler) GetOne(tenantID string, userID string) (*authv1_cache.TokenMetadata, error) {
	token, err := h.handler.GetOne(tenantID, userID)
	if err != nil {
		h.logger.Debug("Access token not found", "tenantID", tenantID, "userID", userID)
		return nil, err
	}

	return token, nil
}

// Validate checks if a token is valid (exists, not revoked, not expired)
func (h *AccessTokenHandler) Validate(tenantID string, userID string) (*authv1_cache.TokenMetadata, error) {
	metadata, err := h.handler.GetOne(tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Check if revoked
	if metadata.Revoked {
		return nil, infra_error.Auth(infra_error.AuthTokenRevoked)
	}

	// Check if expired
	if time.Now().After(metadata.ExpiresAt.AsTime()) {
		return nil, infra_error.Auth(infra_error.AuthTokenExpired)
	}

	return metadata, nil
}

// Revoke revokes the single access token for a user
func (h *AccessTokenHandler) Revoke(tenantID string, userID string, revokedBy string) error {
	metadata, err := h.GetOne(tenantID, userID)
	if err != nil || metadata == nil {
		// No token to revoke
		h.logger.Debug("No access token to revoke", "tenantID", tenantID, "userID", userID)
		return nil
	}

	// metadata.Revoked = true
	// metadata.RevokedAt = timestamppb.Now()
	// metadata.RevokedBy = revokedBy

	// err = h.keyHandler.Update(tenantID, userID, metadata)
	err = h.Delete(tenantID, userID)
	if err != nil {
		h.logger.Error("Failed to revoke access token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	h.logger.Debug("Access token revoked", "tenantID", tenantID, "userID", userID, "revokedBy", revokedBy)
	return nil
}

// Delete permanently removes the access token from Redis (hard delete)
func (h *AccessTokenHandler) Delete(tenantID string, userID string) error {
	err := h.handler.Delete(tenantID, userID)
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
	keys, err := h.handler.ScanKeys(tenantID, "*")
	if err != nil {
		h.logger.Error("Failed to scan access token keys", "error", err, "tenantID", tenantID)
		return nil, err
	}

	h.logger.Debug("Access token keys scanned", "tenantID", tenantID, "keys_found", len(keys))
	return keys, nil
}

// DeleteByPattern deletes all access tokens for a tenant
// Returns the number of tokens deleted
func (h *AccessTokenHandler) DeleteByPattern(tenantID string, pattern string) (int, error) {
	// Pattern: all user IDs in this tenant (tenantID:*)
	count, err := h.handler.DeleteByPattern(tenantID, pattern)
	if err != nil {
		h.logger.Error("Failed to delete access tokens by pattern", "error", err, "tenantID", tenantID)
		return 0, err
	}

	h.logger.Info("Access tokens deleted for tenant", "tenantID", tenantID, "tokens_deleted", count)
	return count, nil
}
