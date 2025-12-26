package auth

import (
	"erp.localhost/internal/auth/models"
	redis_models "erp.localhost/internal/db/redis/models"
)

// AccessTokenHandler interface for access token operations
type AccessTokenHandler interface {
	Store(tenantID string, tokenID string, metadata redis_models.TokenMetadata) error
	Get(tenantID string, tokenID string) (*redis_models.TokenMetadata, error)
	Validate(tenantID string, tokenID string) (*redis_models.TokenMetadata, error)
	Revoke(tenantID string, tokenID string, revokedBy string) error
	RevokeAll(tenantID string, userID string, revokedBy string) error
	Delete(tenantID string, tokenID string) error
}

// RefreshTokenHandler interface for refresh token operations
type RefreshTokenHandler interface {
	Store(tenantID string, userID string, tokenID string, refreshToken models.RefreshToken) error
	Get(tenantID string, userID string, tokenID string) (*models.RefreshToken, error)
	Validate(tenantID string, userID string, tokenID string) (*models.RefreshToken, error)
	Revoke(tenantID string, userID string, tokenID string) error
	RevokeAll(tenantID string, userID string) error
	UpdateLastUsed(tenantID string, userID string, tokenID string) error
	Delete(tenantID string, userID string, tokenID string) error
}

