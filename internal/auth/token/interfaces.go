package token

import (
	"erp.localhost/internal/auth/models"
	redis_models "erp.localhost/internal/db/redis/models"
)

// AccessTokenHandler interface for access token operations
type AccessTokenHandler interface {
	Store(tenantID string, userID string, tokenID string, metadata redis_models.TokenMetadata) error
	GetOne(tenantID string, userID string, tokenID string) (*redis_models.TokenMetadata, error)
	GetAll(tenantID string, userID string) ([]redis_models.TokenMetadata, error)
	Validate(tenantID string, userID string, tokenID string) (*redis_models.TokenMetadata, error)
	Revoke(tenantID string, userID string, tokenID string, revokedBy string) error
	RevokeAll(tenantID string, userID string, revokedBy string) error
	Delete(tenantID string, userID string, tokenID string) error
}

// RefreshTokenHandler interface for refresh token operations
type RefreshTokenHandler interface {
	Store(tenantID string, userID string, tokenID string, refreshToken models.RefreshToken) error
	GetOne(tenantID string, userID string, tokenID string) (*models.RefreshToken, error)
	GetAll(tenantID string, userID string) ([]models.RefreshToken, error)
	Validate(tenantID string, userID string, tokenID string) (*models.RefreshToken, error)
	Revoke(tenantID string, userID string, tokenID string) error
	RevokeAll(tenantID string, userID string) error
	UpdateLastUsed(tenantID string, userID string, tokenID string) error
	Delete(tenantID string, userID string, tokenID string) error
}
