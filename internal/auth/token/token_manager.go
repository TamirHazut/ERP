package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_auth_cache "erp.localhost/internal/infra/model/auth/cache"
	model_shared "erp.localhost/internal/infra/model/shared"
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
	accessTokenHandler   TokenHandler[model_auth_cache.TokenMetadata]
	refreshTokenHandler  TokenHandler[model_auth.RefreshToken]
	logger               logger.Logger
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
	CreatedAt time.Time
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
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}
	return nil
}

// NewTokenManager creates a new TokenManager
func NewTokenManager(secretKey string, tokenDuration time.Duration, refreshTokenDuration time.Duration) *TokenManager {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)
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
		accessTokenHandler:   NewAccessTokenHandler(nil, nil, logger),
		refreshTokenHandler:  NewRefreshTokenHandler(nil, nil, logger),
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
	claims := &model_auth.AccessTokenClaims{
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
		return "", infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}
	return tokenString, nil
}

// VerifyAccessToken verifies if the given JWT token is a valid access token
func (tm *TokenManager) VerifyAccessToken(tokenString string) (*model_auth_cache.TokenMetadata, error) {
	if tokenString == "" {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("token is required"))
	}

	accessTokenMetadata, err := tm.GetTokenMetadata(tokenString)
	if err != nil {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}
	if accessTokenMetadata == nil {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("access token not found"))
	}
	if accessTokenMetadata.ExpiresAt.Before(time.Now()) {
		return nil, infra_error.Auth(infra_error.AuthTokenExpired).WithError(errors.New("access token has expired"))
	}
	if accessTokenMetadata.Revoked {
		return nil, infra_error.Auth(infra_error.AuthTokenRevoked).WithError(errors.New("access token has been revoked"))
	}
	if accessTokenMetadata.RevokedAt != nil && accessTokenMetadata.RevokedAt.Before(time.Now()) {
		return nil, infra_error.Auth(infra_error.AuthTokenRevoked).WithError(errors.New("access token has been revoked"))
	}

	return accessTokenMetadata, nil
}

// GenerateRefreshToken generates a new refresh token for the given user
func (tm *TokenManager) GenerateRefreshToken(input GenerateRefreshTokenInput) (model_auth.RefreshToken, error) {
	if input.UserID == "" {
		return model_auth.RefreshToken{}, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("user_id is required"))
	}

	tm.logger.Debug("Generating refresh token", "input", input)
	if input.CreatedAt.IsZero() {
		input.CreatedAt = time.Now()
	}
	now := input.CreatedAt
	expiresAt := now.Add(tm.refreshTokenDuration)

	// Generate cryptographically secure random token
	// 32 bytes = 256 bits of entropy (very secure)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return model_auth.RefreshToken{}, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}

	// Encode to base64 URL-safe string (no padding)
	tokenString := base64.RawURLEncoding.EncodeToString(tokenBytes)

	// Create refresh token storage model with metadata
	refreshToken := model_auth.RefreshToken{
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
		return model_auth.RefreshToken{}, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}

	// // Store refresh token in Redis (use tokenString as tokenID)
	// if err := tm.refreshTokenHandler.Store(input.TenantID, input.UserID, tokenString, *refreshToken); err != nil {
	// 	return "", infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	// }
	return refreshToken, nil
}

// VerifyRefreshToken verifies if the given refresh token is valid
func (tm *TokenManager) VerifyRefreshToken(tenantID string, userID string, tokenString string) (*model_auth.RefreshToken, error) {
	if tenantID == "" {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("tenantID is required"))
	}
	if tokenString == "" {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("token is required"))
	}
	if userID == "" {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("userID is required"))
	}

	tm.logger.Debug("Verifying refresh token", "tenantID", tenantID, "userID", userID, "token", tokenString)

	// Validate the token (this also retrieves it)
	refreshToken, err := tm.refreshTokenHandler.Validate(tenantID, userID)
	if err != nil {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}

	// SECURITY: Verify the stored token matches the provided token
	// This is critical for detecting stolen/old tokens
	if refreshToken.Token != tokenString {
		tm.logger.Warn("Attempted use of invalid refresh token", "tenantID", tenantID, "userID", userID)
		// Revoke the current valid token (security measure)
		_ = tm.RevokeAllTokens(tenantID, userID, "system")
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("token mismatch - possible theft detected"))
	}

	// Basic validation
	if err := refreshToken.Validate(); err != nil {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}

	// Check if revoked
	if !refreshToken.IsValid() {
		return nil, infra_error.Auth(infra_error.AuthTokenRevoked).WithError(errors.New("token has been revoked"))
	}

	// Check if expired
	if refreshToken.IsExpired() {
		// Auto-cleanup expired token
		if err := tm.refreshTokenHandler.Delete(tenantID, userID); err != nil {
			return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
		}
		return nil, infra_error.Auth(infra_error.AuthRefreshTokenExpired).WithError(errors.New("token has expired"))
	}

	// SECURITY: Check for suspicious activity
	// 1. Check if token is being reused (already used recently)
	if !refreshToken.LastUsedAt.IsZero() {
		timeSinceLastUse := time.Since(refreshToken.LastUsedAt)
		if timeSinceLastUse < 1*time.Minute {
			// Token used twice within 1 minute - possible token theft
			// Revoke all user tokens as security measure
			tm.logger.Warn("Suspicious: Token reused within 1 minute", "tenantID", tenantID, "userID", userID)
			if err := tm.RevokeAllTokens(tenantID, refreshToken.UserID, "system"); err != nil {
				return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
			}
			return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("suspicious activity detected - all sessions terminated"))
		}
	}

	// Update last used timestamp with safe type assertion
	if refreshTokenHandler, ok := tm.refreshTokenHandler.(*RefreshTokenHandler); ok {
		if err := refreshTokenHandler.UpdateLastUsed(tenantID, userID, tokenString); err != nil {
			tm.logger.Warn("Failed to update last used timestamp", "error", err)
		}
	} else {
		tm.logger.Debug("UpdateLastUsed not available for this token handler implementation")
	}

	return refreshToken, nil
}

// ============================================================================
// REDIS TOKEN STORAGE OPERATIONS
// ============================================================================

// StoreTokens stores both access and refresh tokens in Redis
// This is typically called after successful authentication
// Single token per user - automatically replaces any existing tokens
func (tm *TokenManager) StoreTokens(tenantID string, userID string, accessTokenMetadata model_auth_cache.TokenMetadata, refreshToken model_auth.RefreshToken) error {
	tm.logger.Info("Storing token pair (single token per user - replaces existing)", "tenantID", tenantID, "userID", userID)

	// Store access token (automatically replaces existing)
	if err := tm.accessTokenHandler.Store(tenantID, userID, accessTokenMetadata); err != nil {
		tm.logger.Error("Failed to store access token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	// Store refresh token (automatically replaces existing)
	if err := tm.refreshTokenHandler.Store(tenantID, userID, refreshToken); err != nil {
		// If refresh token storage fails, try to clean up access token
		tm.logger.Error("Failed to store refresh token", "error", err, "tenantID", tenantID, "userID", userID)
		_ = tm.accessTokenHandler.Delete(tenantID, userID)
		return err
	}

	tm.logger.Info("Token pair stored successfully", "tenantID", tenantID, "userID", userID)
	return nil
}

// ValidateAccessTokenFromRedis validates an access token from Redis
func (tm *TokenManager) ValidateAccessTokenFromRedis(tenantID string, userID string) (*model_auth_cache.TokenMetadata, error) {
	return tm.accessTokenHandler.Validate(tenantID, userID)
}

// ValidateRefreshTokenFromRedis validates a refresh token from Redis
func (tm *TokenManager) ValidateRefreshTokenFromRedis(tenantID string, userID string) (*model_auth.RefreshToken, error) {
	return tm.refreshTokenHandler.Validate(tenantID, userID)
}

// // RevokeAccessTokenFromRedis revokes a single access token in Redis
// func (tm *TokenManager) RevokeAccessTokenFromRedis(tenantID string, userID string, tokenID string, revokedBy string) error {
// 	return tm.accessTokenHandler.Revoke(tenantID, userID, tokenID, revokedBy)
// }

// // RevokeRefreshTokenFromRedis revokes a single refresh token in Redis
// func (tm *TokenManager) RevokeRefreshTokenFromRedis(tenantID string, userID string, tokenID string, revokedBy string) error {
// 	return tm.refreshTokenHandler.Revoke(tenantID, userID, tokenID, revokedBy)
// }

// RevokeAllAccessTokens revokes the access token for a user (but not refresh token)
// This is typically called during token refresh to prevent orphaned access tokens
func (tm *TokenManager) RevokeAllAccessTokens(tenantID string, userID string, revokedBy string) error {
	if err := tm.accessTokenHandler.Revoke(tenantID, userID, revokedBy); err != nil {
		tm.logger.Error("Failed to revoke access token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	tm.logger.Debug("Access token revoked", "tenantID", tenantID, "userID", userID, "revokedBy", revokedBy)
	return nil
}

// RevokeAllTokens revokes all tokens (both access and refresh) for a user
// This is typically called on logout or security incidents
func (tm *TokenManager) RevokeAllTokens(tenantID string, userID string, revokedBy string) error {
	// Revoke access token
	if err := tm.accessTokenHandler.Revoke(tenantID, userID, revokedBy); err != nil {
		tm.logger.Error("Failed to revoke access token", "error", err, "tenantID", tenantID, "userID", userID)
		// Continue with refresh token even if access token fails
	}

	// Revoke refresh token
	if err := tm.refreshTokenHandler.Revoke(tenantID, userID, revokedBy); err != nil {
		tm.logger.Error("Failed to revoke refresh token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	tm.logger.Debug("All tokens revoked", "tenantID", tenantID, "userID", userID, "revokedBy", revokedBy)
	return nil
}

// // GetAccessTokenFromRedis retrieves access token metadata from Redis
// func (tm *TokenManager) GetAccessTokenFromRedis(tenantID string, userID string, tokenID string) (*model_auth_cache.TokenMetadata, error) {
// 	return tm.accessTokenHandler.GetOne(tenantID, userID, tokenID)
// }

// // GetAllAccessTokensFromRedis retrieves all access tokens from Redis
// func (tm *TokenManager) GetAllAccessTokensFromRedis(tenantID string, userID string) ([]model_auth_cache.TokenMetadata, error) {
// 	return tm.accessTokenHandler.GetAll(tenantID, userID)
// }

// // GetRefreshTokenFromRedis retrieves refresh token from Redis
// func (tm *TokenManager) GetRefreshTokenFromRedis(tenantID string, userID string, tokenID string) (*model_auth.RefreshToken, error) {
// 	return tm.refreshTokenHandler.GetOne(tenantID, userID, tokenID)
// }

// // GetAllRefreshTokensFromRedis retrieves all refresh tokens from Redis
// func (tm *TokenManager) GetAllRefreshTokensFromRedis(tenantID string, userID string) ([]model_auth.RefreshToken, error) {
// 	return tm.refreshTokenHandler.GetAll(tenantID, userID)
// }

// UpdateRefreshTokenLastUsed updates the last used timestamp for a refresh token
func (tm *TokenManager) UpdateRefreshTokenLastUsed(tenantID string, userID string, tokenString string) error {
	if refreshTokenHandler, ok := tm.refreshTokenHandler.(*RefreshTokenHandler); ok {
		return refreshTokenHandler.UpdateLastUsed(tenantID, userID, tokenString)
	}
	tm.logger.Debug("UpdateLastUsed not available for this token handler implementation")
	return nil
}

// DeleteAccessTokenFromRedis permanently deletes an access token from Redis
func (tm *TokenManager) DeleteAccessTokenFromRedis(tenantID string, userID string) error {
	return tm.accessTokenHandler.Delete(tenantID, userID)
}

// DeleteRefreshTokenFromRedis permanently deletes a refresh token from Redis
func (tm *TokenManager) DeleteRefreshTokenFromRedis(tenantID string, userID string) error {
	return tm.refreshTokenHandler.Delete(tenantID, userID)
}

// RevokeAccessToken revokes a JWT access token (legacy method for compatibility)
func (tm *TokenManager) RevokeAccessToken(tokenString string, revokedBy string) error {
	if tokenString == "" {
		return infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("token is required"))
	}
	metadata, err := tm.GetTokenMetadata(tokenString)
	if err != nil {
		return err
	}
	if metadata == nil {
		return infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("access token not found"))
	}
	if metadata.Revoked {
		return infra_error.Auth(infra_error.AuthTokenRevoked).WithError(errors.New("access token has been revoked"))
	}
	if metadata.RevokedAt != nil && metadata.RevokedAt.Before(time.Now()) {
		return infra_error.Auth(infra_error.AuthTokenRevoked).WithError(errors.New("access token has been revoked"))
	}
	if err := tm.accessTokenHandler.Revoke(metadata.TenantID, metadata.UserID, revokedBy); err != nil {
		return err
	}
	return nil
}

// RevokeRefreshToken revokes a refresh token (legacy method for compatibility)
func (tm *TokenManager) RevokeRefreshToken(tenantID string, userID string, tokenString string, revokedBy string, skipVerification bool) error {
	if tokenString == "" || tenantID == "" || userID == "" {
		return infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("token, tenantID, and userID are required"))
	}
	if !skipVerification {
		// Verify token exists and is valid
		_, err := tm.VerifyRefreshToken(tenantID, userID, tokenString)
		if err != nil {
			tm.logger.Error("Failed to verify refresh token", "error", err, "tenantID", tenantID, "userID", userID, "token", tokenString)
			return err
		}
	}
	// Revoke the token
	if err := tm.refreshTokenHandler.Revoke(tenantID, userID, revokedBy); err != nil {
		tm.logger.Error("Failed to revoke refresh token", "error", err, "tenantID", tenantID, "userID", userID, "token", tokenString, "requestBy", revokedBy)
		return infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}
	tm.logger.Info("Refresh token revoked", "tenantID", tenantID, "userID", userID, "token", tokenString, "requestBy", revokedBy)
	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user (legacy method for compatibility)
func (tm *TokenManager) RevokeAllUserRefreshTokens(tenantID string, userID string, requestBy string) error {
	if userID == "" || tenantID == "" {
		return errors.New("user_id and tenant_id are required")
	}

	if err := tm.refreshTokenHandler.Revoke(tenantID, userID, requestBy); err != nil {
		return err
	}

	return nil
}

// RevokeAllTenantTokens revokes all tokens for ALL users in a tenant
// This is used for tenant suspension or security incidents
// Returns the number of access and refresh tokens revoked
func (tm *TokenManager) RevokeAllTenantTokens(tenantID string, revokedBy string) (int, int, error) {
	if tenantID == "" {
		return 0, 0, infra_error.Validation(infra_error.ValidationRequiredFields, "tenantID")
	}
	if revokedBy == "" {
		return 0, 0, infra_error.Validation(infra_error.ValidationRequiredFields, "revokedBy")
	}

	tm.logger.Warn("Revoking ALL tokens for entire tenant", "tenantID", tenantID, "revokedBy", revokedBy)

	var accessTokensRevoked, refreshTokensRevoked int

	// Type assert to get concrete handlers
	accessHandler, ok := tm.accessTokenHandler.(*AccessTokenHandler)
	if !ok {
		return 0, 0, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("accessTokenHandler is not *AccessTokenHandler"))
	}

	refreshHandler, ok := tm.refreshTokenHandler.(*RefreshTokenHandler)
	if !ok {
		return 0, 0, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("refreshTokenHandler is not *RefreshTokenHandler"))
	}

	// Scan all access token keys for this tenant
	accessKeys, err := accessHandler.ScanKeys(tenantID)
	if err != nil {
		tm.logger.Error("Failed to scan access tokens", "error", err, "tenantID", tenantID)
		// Continue with refresh tokens even if this fails
	} else {
		// Revoke each access token
		for _, key := range accessKeys {
			// Extract userID from key pattern: prefix:tokens:tenantID:userID
			// Split and get the last part
			parts := parseRedisKey(key)
			if len(parts) >= 2 {
				userID := parts[len(parts)-1]
				if err := accessHandler.Revoke(tenantID, userID, revokedBy); err != nil {
					tm.logger.Warn("Failed to revoke access token", "error", err, "tenantID", tenantID, "userID", userID)
				} else {
					accessTokensRevoked++
				}
			}
		}
	}

	// Scan all refresh token keys for this tenant
	refreshKeys, err := refreshHandler.ScanKeys(tenantID)
	if err != nil {
		tm.logger.Error("Failed to scan refresh tokens", "error", err, "tenantID", tenantID)
		return accessTokensRevoked, refreshTokensRevoked, err
	}

	// Revoke each refresh token
	for _, key := range refreshKeys {
		// Extract userID from key pattern: prefix:refresh_tokens:tenantID:userID
		parts := parseRedisKey(key)
		if len(parts) >= 2 {
			userID := parts[len(parts)-1]
			if err := refreshHandler.Revoke(tenantID, userID, revokedBy); err != nil {
				tm.logger.Warn("Failed to revoke refresh token", "error", err, "tenantID", tenantID, "userID", userID)
			} else {
				refreshTokensRevoked++
			}
		}
	}

	tm.logger.Info("All tenant tokens revoked", "tenantID", tenantID, "accessTokensRevoked", accessTokensRevoked, "refreshTokensRevoked", refreshTokensRevoked)
	return accessTokensRevoked, refreshTokensRevoked, nil
}

// DeleteAllTenantTokens permanently deletes all tokens for ALL users in a tenant
// This is used for tenant deletion (cascade cleanup)
// Returns the number of access and refresh tokens deleted
func (tm *TokenManager) DeleteAllTenantTokens(tenantID string) (int, int, error) {
	if tenantID == "" {
		return 0, 0, infra_error.Validation(infra_error.ValidationRequiredFields, "tenantID")
	}

	tm.logger.Warn("Deleting ALL tokens for entire tenant (hard delete)", "tenantID", tenantID)

	// Type assert to get concrete handlers
	accessHandler, ok := tm.accessTokenHandler.(*AccessTokenHandler)
	if !ok {
		return 0, 0, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("accessTokenHandler is not *AccessTokenHandler"))
	}

	refreshHandler, ok := tm.refreshTokenHandler.(*RefreshTokenHandler)
	if !ok {
		return 0, 0, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("refreshTokenHandler is not *RefreshTokenHandler"))
	}

	// Delete all access tokens using pattern
	accessCount, err := accessHandler.DeleteByPattern(tenantID)
	if err != nil {
		tm.logger.Error("Failed to delete access tokens by pattern", "error", err, "tenantID", tenantID)
		// Continue with refresh tokens
	}

	// Delete all refresh tokens using pattern
	refreshCount, err := refreshHandler.DeleteByPattern(tenantID)
	if err != nil {
		tm.logger.Error("Failed to delete refresh tokens by pattern", "error", err, "tenantID", tenantID)
		return accessCount, refreshCount, err
	}

	tm.logger.Info("All tenant tokens deleted", "tenantID", tenantID, "accessTokensDeleted", accessCount, "refreshTokensDeleted", refreshCount)
	return accessCount, refreshCount, nil
}

// parseRedisKey extracts parts from a Redis key
// Example: "tokens:tenant-123:user-456" -> ["tokens", "tenant-123", "user-456"]
func parseRedisKey(key string) []string {
	// Simple split by colon
	result := []string{}
	current := ""
	for _, char := range key {
		if char == ':' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func (tm *TokenManager) GetTokenMetadata(accessTokenString string) (*model_auth_cache.TokenMetadata, error) {
	if accessTokenString == "" {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("empty access token"))
	}
	claims := &model_auth.AccessTokenClaims{}

	token, err := jwt.Parse(accessTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("invalid signing method"))
		}
		return []byte(tm.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("invalid token"))
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
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("user_id is required"))
	}
	if claims.TenantID == "" {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("tenant_id is required"))
	}
	// Get the single access token for this user
	accessTokenMetadata, err := tm.accessTokenHandler.GetOne(claims.TenantID, claims.UserID)
	if err != nil {
		return nil, err
	}

	if accessTokenMetadata == nil {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("access token not found"))
	}

	// Verify the stored token matches the provided token
	hashedAccessToken := sha256.Sum256([]byte(accessTokenString))
	accessTokenID := hex.EncodeToString(hashedAccessToken[:])
	if accessTokenMetadata.TokenID != accessTokenID {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("access token ID mismatch - possible token theft"))
	}
	return accessTokenMetadata, nil
}
