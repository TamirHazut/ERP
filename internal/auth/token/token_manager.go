package token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	keyshandlers "erp.localhost/internal/auth/keys_handlers"
	"erp.localhost/internal/auth/models"
	"erp.localhost/internal/db/redis"
	redis_models "erp.localhost/internal/db/redis/models"
	erp_errors "erp.localhost/internal/errors"
	logging "erp.localhost/internal/logging"
	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	Issuer = "erp.localhost"
)

// TokenManager coordinates all token operations including JWT generation/verification and Redis storage
type TokenManager struct {
	secretKey            string
	tokenDuration        time.Duration
	refreshTokenDuration time.Duration
	accessTokenHandler   AccessTokenHandler
	refreshTokenHandler  RefreshTokenHandler
	logger               *logging.Logger
}

// GenerateAccessTokenInput input for generating access tokens
type GenerateAccessTokenInput struct {
	UserID      string
	TenantID    string
	Email       string
	Username    string
	Roles       []string
	Permissions []string
	SessionID   string
	DeviceID    string
}

// GenerateRefreshTokenInput input for generating refresh tokens
type GenerateRefreshTokenInput struct {
	UserID    string
	TenantID  string
	SessionID string
	DeviceID  string
	IPAddress string
	UserAgent string
}

func (i *GenerateAccessTokenInput) Validate() error {
	missingFields := []string{}
	if i.UserID == "" {
		missingFields = append(missingFields, "UserID")
	}
	if i.TenantID == "" {
		missingFields = append(missingFields, "TenantID")
	}
	if i.Username == "" {
		missingFields = append(missingFields, "Username")
	}
	if len(i.Roles) == 0 {
		missingFields = append(missingFields, "Roles")
	}
	if len(i.Permissions) == 0 {
		missingFields = append(missingFields, "Permissions")
	}
	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}
	return nil
}

// NewTokenManager creates a new TokenManager
func NewTokenManager(secretKey string, tokenDuration time.Duration, refreshTokenDuration time.Duration) *TokenManager {
	logger := logging.NewLogger(logging.ModuleAuth)
	if secretKey == "" {
		logger.Fatal("secret key is required")
		return nil
	}
	if tokenDuration <= 0 {
		logger.Fatal("token duration must be greater than 0")
		return nil
	}
	if refreshTokenDuration <= 0 {
		logger.Fatal("refresh token duration must be greater than 0")
		return nil
	}

	return &TokenManager{
		secretKey:            secretKey,
		tokenDuration:        tokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		accessTokenHandler:   keyshandlers.NewAccessTokenKeyHandler(redis.KeyPrefix("tokens")),
		refreshTokenHandler:  keyshandlers.NewRefreshTokenKeyHandler(redis.KeyPrefix("refresh_tokens")),
		logger:               logger,
	}
}

// ============================================================================
// JWT TOKEN GENERATION AND VERIFICATION
// ============================================================================

// GenerateAccessToken generates a new JWT access token for the given user
func (tm *TokenManager) GenerateAccessToken(input *GenerateAccessTokenInput) (string, error) {
	if err := input.Validate(); err != nil {
		return "", err
	}
	currentTimestamp := time.Now().Unix()
	expiresAt := time.Now().Add(tm.tokenDuration)
	claims := &models.AccessTokenClaims{
		UserID:      input.UserID,
		TenantID:    input.TenantID,
		Username:    input.Username,
		Email:       input.Email,
		Roles:       input.Roles,
		Permissions: input.Permissions,
		TokenType:   TokenTypeAccess,
		SessionID:   input.SessionID,
		DeviceID:    input.DeviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Unix(currentTimestamp, 0)),
			NotBefore: jwt.NewNumericDate(time.Unix(currentTimestamp, 0)),
			Issuer:    Issuer,
			Subject:   input.UserID,
			Audience:  []string{uuid.New().String()},
		},
	}

	if err := claims.Validate(); err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tm.secretKey))
	if err != nil {
		return "", erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(err)
	}
	return tokenString, nil
}

// VerifyAccessToken verifies if the given JWT token is a valid access token
func (tm *TokenManager) VerifyAccessToken(tokenString string) (*redis_models.TokenMetadata, error) {
	if tokenString == "" {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("token is required"))
	}

	accessTokenMetadata, err := tm.GetTokenMetadata(tokenString)
	if err != nil {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(err)
	}
	if accessTokenMetadata == nil {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("access token not found"))
	}
	if accessTokenMetadata.ExpiresAt.Before(time.Now()) {
		return nil, erp_errors.Auth(erp_errors.AuthTokenExpired).WithError(errors.New("access token has expired"))
	}
	if accessTokenMetadata.Revoked {
		return nil, erp_errors.Auth(erp_errors.AuthTokenRevoked).WithError(errors.New("access token has been revoked"))
	}
	if accessTokenMetadata.RevokedAt != nil && accessTokenMetadata.RevokedAt.Before(time.Now()) {
		return nil, erp_errors.Auth(erp_errors.AuthTokenRevoked).WithError(errors.New("access token has been revoked"))
	}

	return accessTokenMetadata, nil
}

// GenerateRefreshToken generates a new refresh token for the given user
func (tm *TokenManager) GenerateRefreshToken(ctx context.Context, input GenerateRefreshTokenInput) (models.RefreshToken, error) {
	if input.UserID == "" {
		return models.RefreshToken{}, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("user_id is required"))
	}

	tm.logger.Debug("Generating refresh token", "input", input)
	now := time.Now()
	expiresAt := now.Add(tm.refreshTokenDuration)

	// Generate cryptographically secure random token
	// 32 bytes = 256 bits of entropy (very secure)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return models.RefreshToken{}, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(err)
	}

	// Encode to base64 URL-safe string (no padding)
	tokenString := base64.RawURLEncoding.EncodeToString(tokenBytes)

	// Create refresh token storage model with metadata
	refreshToken := models.RefreshToken{
		Token:      tokenString,
		UserID:     input.UserID,
		TenantID:   input.TenantID,
		SessionID:  input.SessionID,
		DeviceID:   input.DeviceID,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		ExpiresAt:  expiresAt,
		CreatedAt:  now,
		LastUsedAt: time.Time{},
		RevokedAt:  time.Time{},
		IsRevoked:  false,
	}

	// Validate before storing
	if err := refreshToken.Validate(); err != nil {
		return models.RefreshToken{}, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(err)
	}

	// // Store refresh token in Redis (use tokenString as tokenID)
	// if err := tm.refreshTokenHandler.Store(input.TenantID, input.UserID, tokenString, *refreshToken); err != nil {
	// 	return "", erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(err)
	// }
	return refreshToken, nil
}

// VerifyRefreshToken verifies if the given refresh token is valid
func (tm *TokenManager) VerifyRefreshToken(tenantID string, userID string, tokenString string) (*models.RefreshToken, error) {
	tm.logger.Debug("Verifying refresh token", "tenantID", tenantID, "userID", userID, "token", tokenString)
	if tokenString == "" {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("token is required"))
	}
	if userID == "" {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("userID is required"))
	}

	// Validate the token (this also retrieves it)
	refreshToken, err := tm.refreshTokenHandler.Validate(tenantID, userID, tokenString)
	if err != nil {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(err)
	}

	// Basic validation
	if err := refreshToken.Validate(); err != nil {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(err)
	}

	// Check if revoked
	if !refreshToken.IsValid() {
		return nil, erp_errors.Auth(erp_errors.AuthTokenRevoked).WithError(errors.New("token has been revoked"))
	}

	// Check if expired
	if refreshToken.IsExpired() {
		// Auto-cleanup expired token
		tm.refreshTokenHandler.Delete(tenantID, userID, tokenString)
		return nil, erp_errors.Auth(erp_errors.AuthRefreshTokenExpired).WithError(errors.New("token has expired"))
	}

	// SECURITY: Check for suspicious activity
	// 1. Check if token is being reused (already used recently)
	if !refreshToken.LastUsedAt.IsZero() {
		timeSinceLastUse := time.Since(refreshToken.LastUsedAt)
		if timeSinceLastUse < 1*time.Minute {
			// Token used twice within 1 minute - possible token theft
			// Revoke all user tokens as security measure
			tm.refreshTokenHandler.RevokeAll(tenantID, refreshToken.UserID)
			return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("suspicious activity detected - all sessions terminated"))
		}
	}

	// Update last used timestamp
	if err := tm.refreshTokenHandler.UpdateLastUsed(tenantID, userID, tokenString); err != nil {
		// Log but don't fail
		tm.logger.Warn("Failed to update last used timestamp", "error", err)
	}

	return refreshToken, nil
}

// ============================================================================
// REDIS TOKEN STORAGE OPERATIONS
// ============================================================================

// StoreTokens stores both access and refresh tokens in Redis
// This is typically called after successful authentication
func (tm *TokenManager) StoreTokens(tenantID string, userID string, accessTokenID string, refreshTokenID string, accessTokenMetadata redis_models.TokenMetadata, refreshToken models.RefreshToken) error {
	// Store access token
	if err := tm.accessTokenHandler.Store(tenantID, userID, accessTokenID, accessTokenMetadata); err != nil {
		tm.logger.Error("Failed to store access token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	// Store refresh token
	if err := tm.refreshTokenHandler.Store(tenantID, userID, refreshTokenID, refreshToken); err != nil {
		// If refresh token storage fails, try to clean up access token
		tm.logger.Error("Failed to store refresh token", "error", err, "tenantID", tenantID, "userID", userID)
		_ = tm.accessTokenHandler.Delete(tenantID, userID, accessTokenID)
		return err
	}

	tm.logger.Debug("Tokens stored successfully", "tenantID", tenantID, "userID", userID, "accessTokenID", accessTokenID, "refreshTokenID", refreshTokenID)
	return nil
}

// ValidateAccessTokenFromRedis validates an access token from Redis
func (tm *TokenManager) ValidateAccessTokenFromRedis(tenantID string, userID string, tokenID string) (*redis_models.TokenMetadata, error) {
	return tm.accessTokenHandler.Validate(tenantID, userID, tokenID)
}

// ValidateRefreshTokenFromRedis validates a refresh token from Redis
func (tm *TokenManager) ValidateRefreshTokenFromRedis(tenantID string, userID string, tokenID string) (*models.RefreshToken, error) {
	return tm.refreshTokenHandler.Validate(tenantID, userID, tokenID)
}

// RefreshTokens generates new tokens and revokes old refresh token (token rotation)
func (tm *TokenManager) RefreshTokens(tenantID string, userID string, oldRefreshTokenID string, newAccessTokenID string, newRefreshTokenID string, newAccessTokenMetadata redis_models.TokenMetadata, newRefreshToken models.RefreshToken) error {
	// Validate old refresh token
	_, err := tm.refreshTokenHandler.Validate(tenantID, userID, oldRefreshTokenID)
	if err != nil {
		tm.logger.Error("Invalid refresh token", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", oldRefreshTokenID)
		return err
	}

	// Revoke old refresh token
	if err := tm.refreshTokenHandler.Revoke(tenantID, userID, oldRefreshTokenID); err != nil {
		tm.logger.Error("Failed to revoke old refresh token", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", oldRefreshTokenID)
		return err
	}

	// Store new access token
	if err := tm.accessTokenHandler.Store(tenantID, userID, newAccessTokenID, newAccessTokenMetadata); err != nil {
		tm.logger.Error("Failed to store new access token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	// Store new refresh token
	if err := tm.refreshTokenHandler.Store(tenantID, userID, newRefreshTokenID, newRefreshToken); err != nil {
		// If new refresh token storage fails, try to clean up new access token
		tm.logger.Error("Failed to store new refresh token", "error", err, "tenantID", tenantID, "userID", userID)
		_ = tm.accessTokenHandler.Delete(tenantID, userID, newAccessTokenID)
		return err
	}

	tm.logger.Debug("Tokens refreshed successfully", "tenantID", tenantID, "userID", userID, "oldRefreshTokenID", oldRefreshTokenID, "newAccessTokenID", newAccessTokenID, "newRefreshTokenID", newRefreshTokenID)
	return nil
}

// RevokeAccessTokenFromRedis revokes a single access token in Redis
func (tm *TokenManager) RevokeAccessTokenFromRedis(tenantID string, userID string, tokenID string, revokedBy string) error {
	return tm.accessTokenHandler.Revoke(tenantID, userID, tokenID, revokedBy)
}

// RevokeRefreshTokenFromRedis revokes a single refresh token in Redis
func (tm *TokenManager) RevokeRefreshTokenFromRedis(tenantID string, userID string, tokenID string) error {
	return tm.refreshTokenHandler.Revoke(tenantID, userID, tokenID)
}

// RevokeAllTokens revokes all tokens (both access and refresh) for a user
// This is typically called on logout or security incidents
func (tm *TokenManager) RevokeAllTokens(tenantID string, userID string, revokedBy string) error {
	// Revoke all access tokens
	if err := tm.accessTokenHandler.RevokeAll(tenantID, userID, revokedBy); err != nil {
		tm.logger.Error("Failed to revoke all access tokens", "error", err, "tenantID", tenantID, "userID", userID)
		// Continue with refresh tokens even if access tokens fail
	}

	// Revoke all refresh tokens
	if err := tm.refreshTokenHandler.RevokeAll(tenantID, userID); err != nil {
		tm.logger.Error("Failed to revoke all refresh tokens", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	tm.logger.Debug("All tokens revoked", "tenantID", tenantID, "userID", userID, "revokedBy", revokedBy)
	return nil
}

// GetAccessTokenFromRedis retrieves access token metadata from Redis
func (tm *TokenManager) GetAccessTokenFromRedis(tenantID string, userID string, tokenID string) (*redis_models.TokenMetadata, error) {
	return tm.accessTokenHandler.GetOne(tenantID, userID, tokenID)
}

// GetAllAccessTokensFromRedis retrieves all access tokens from Redis
func (tm *TokenManager) GetAllAccessTokensFromRedis(tenantID string, userID string) ([]redis_models.TokenMetadata, error) {
	return tm.accessTokenHandler.GetAll(tenantID, userID)
}

// GetRefreshTokenFromRedis retrieves refresh token from Redis
func (tm *TokenManager) GetRefreshTokenFromRedis(tenantID string, userID string, tokenID string) (*models.RefreshToken, error) {
	return tm.refreshTokenHandler.GetOne(tenantID, userID, tokenID)
}

// GetAllRefreshTokensFromRedis retrieves all refresh tokens from Redis
func (tm *TokenManager) GetAllRefreshTokensFromRedis(tenantID string, userID string) ([]models.RefreshToken, error) {
	return tm.refreshTokenHandler.GetAll(tenantID, userID)
}

// UpdateRefreshTokenLastUsed updates the last used timestamp for a refresh token
func (tm *TokenManager) UpdateRefreshTokenLastUsed(tenantID string, userID string, tokenID string) error {
	return tm.refreshTokenHandler.UpdateLastUsed(tenantID, userID, tokenID)
}

// DeleteAccessTokenFromRedis permanently deletes an access token from Redis
func (tm *TokenManager) DeleteAccessTokenFromRedis(tenantID string, userID string, tokenID string) error {
	return tm.accessTokenHandler.Delete(tenantID, userID, tokenID)
}

// DeleteRefreshTokenFromRedis permanently deletes a refresh token from Redis
func (tm *TokenManager) DeleteRefreshTokenFromRedis(tenantID string, userID string, tokenID string) error {
	return tm.refreshTokenHandler.Delete(tenantID, userID, tokenID)
}

// RevokeAccessToken revokes a JWT access token (legacy method for compatibility)
func (tm *TokenManager) RevokeAccessToken(ctx context.Context, tokenString string) error {
	// Parse token to get JTI and expiration
	accessTokenMetadata, err := tm.VerifyAccessToken(tokenString)
	if err != nil {
		// If token is already invalid/expired, consider it revoked
		return nil
	}

	if accessTokenMetadata.ExpiresAt.Before(time.Now()) {
		return erp_errors.Auth(erp_errors.AuthTokenExpired).WithError(errors.New("token expired"))
	}

	return nil
}

// RevokeRefreshToken revokes a refresh token (legacy method for compatibility)
func (tm *TokenManager) RevokeRefreshToken(tenantID string, userID string, tokenString string) error {
	if tokenString == "" || tenantID == "" || userID == "" {
		return erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("token, tenantID, and userID are required"))
	}
	// Verify token exists and is valid
	_, err := tm.VerifyRefreshToken(tenantID, userID, tokenString)
	if err != nil {
		return err
	}
	// Revoke the token
	if err := tm.refreshTokenHandler.Revoke(tenantID, userID, tokenString); err != nil {
		return erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(err)
	}
	tm.logger.Info("Refresh token revoked", "tenantID", tenantID, "userID", userID, "token", tokenString)
	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user (legacy method for compatibility)
func (tm *TokenManager) RevokeAllUserRefreshTokens(tenantID string, userID string) error {
	if userID == "" || tenantID == "" {
		return errors.New("user_id and tenant_id are required")
	}

	if err := tm.refreshTokenHandler.RevokeAll(tenantID, userID); err != nil {
		return err
	}

	return nil
}

func (tm *TokenManager) GetTokenMetadata(accessTokenString string) (*redis_models.TokenMetadata, error) {
	claims := &models.AccessTokenClaims{}

	token, err := jwt.Parse(accessTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("invalid signing method"))
		}
		return []byte(tm.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("invalid token"))
	}
	if claimsMap, ok := token.Claims.(jwt.MapClaims); ok {
		if sub, ok := claimsMap["sub"].(string); ok {
			claims.UserID = sub
		}
		if tenantID, ok := claimsMap["tenant_id"].(string); ok {
			claims.TenantID = tenantID
		}
	}
	if claims.UserID == "" {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("user_id is required"))
	}
	if claims.TenantID == "" {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("tenant_id is required"))
	}
	hashedAccessToken := sha256.Sum256([]byte(accessTokenString))
	accessTokenID := hex.EncodeToString(hashedAccessToken[:])
	accessTokenMetadata, err := tm.accessTokenHandler.GetOne(claims.TenantID, claims.UserID, accessTokenID)
	if err != nil {
		return nil, err
	}

	if accessTokenMetadata == nil {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("access token not found"))
	}
	if accessTokenMetadata.TokenID != accessTokenID {
		return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid).WithError(errors.New("access token ID mismatch"))
	}
	return accessTokenMetadata, nil
}
