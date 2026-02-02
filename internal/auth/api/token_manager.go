package api

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"erp.localhost/internal/auth/handler"
	"erp.localhost/internal/auth/hash"
	"erp.localhost/internal/auth/token"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	authv1_cache "erp.localhost/internal/infra/model/auth/v1/cache"
	validator_auth_cache "erp.localhost/internal/infra/model/auth/validator/cache"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	Issuer = "erp.localhost"
)

// TokenConfig holds configuration for token management
type TokenConfig struct {
	SecretKey            string
	TokenDuration        time.Duration
	RefreshTokenDuration time.Duration
}

// LoadTokenConfig loads token configuration from environment variables with defaults
func LoadTokenConfig() *TokenConfig {
	return &TokenConfig{
		SecretKey:            getEnv("JWT_SECRET_KEY", "secret"),
		TokenDuration:        parseDuration(getEnv("ACCESS_TOKEN_DURATION", "1h"), 1*time.Hour),
		RefreshTokenDuration: parseDuration(getEnv("REFRESH_TOKEN_DURATION", "168h"), 7*24*time.Hour),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseDuration parses a duration string or returns a default value
func parseDuration(value string, defaultDuration time.Duration) time.Duration {
	if value == "" {
		return defaultDuration
	}

	// Try parsing as duration string (e.g., "1h", "24h")
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}

	// Try parsing as seconds (e.g., "3600" for 1 hour)
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return defaultDuration
}

// type NewTokenResponse struct {
// 	UserId                string `json:"user_id"`
// 	TenantId              string `json:"tenant_id"`
// 	Token                 string `json:"token"`
// 	TokenExpiresAt        int64  `json:"token_expires_at"`
// 	RefreshToken          string `json:"refresh_token"`
// 	RefreshTokenExpiresAt int64  `json:"refresh_token_expires_at"`
// }

// TokenAPI coordinates all token operations including JWT generation/verification and Redis storage
type TokenAPI struct {
	secretKey            string
	tokenDuration        time.Duration
	refreshTokenDuration time.Duration
	accessTokenHandler   handler.TokenHandler[authv1_cache.TokenMetadata]
	refreshTokenHandler  handler.TokenHandler[authv1_cache.RefreshToken]
	logger               logger.Logger
}

// GenerateAccessTokenInput input for generating access tokens
type GenerateAccessTokenInput struct {
	UserId   string
	TenantId string
	Email    string
	Username string
	Roles    []string
}

// GenerateRefreshTokenInput input for generating refresh tokens
type GenerateRefreshTokenInput struct {
	UserId    string
	TenantId  string
	IPAddress string
	UserAgent string
	CreatedAt time.Time
}

func (i *GenerateAccessTokenInput) Validate() *infra_error.AppError {
	missingFields := []string{}
	if i.UserId == "" {
		missingFields = append(missingFields, "UserId")
	}
	if i.TenantId == "" {
		missingFields = append(missingFields, "TenantId")
	}
	if i.Email == "" || i.Username == "" {
		missingFields = append(missingFields, "Email", "Username")
	}
	if len(i.Roles) == 0 {
		missingFields = append(missingFields, "Roles")
	}
	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}
	return nil
}

// NewTokenAPI creates a new TokenManager
func NewTokenAPI(logger logger.Logger) (*TokenAPI, *infra_error.AppError) {
	// Load configuration from environment variables
	config := LoadTokenConfig()
	if config.SecretKey == "" || config.TokenDuration <= 0 || config.RefreshTokenDuration <= 0 {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: secret_key, token_duration, refresh_token_duration"))
		logger.Fatal("failed to create token manager", "error", err)
		return nil, err
	}
	logger.Info("Token configuration loaded",
		"access_token_duration", config.TokenDuration.String(),
		"refresh_token_duration", config.RefreshTokenDuration.String())

	accessTokenHandler, err := handler.NewAccessTokenHandler(logger)
	if err != nil {
		logger.Fatal("failed to create access token handler")
		return nil, err
	}

	refreshTokenHandler, err := handler.NewRefreshTokenHandler(logger)
	if err != nil {
		logger.Fatal("failed to create refresh token handler")
		return nil, err
	}

	return &TokenAPI{
		secretKey:            config.SecretKey,
		tokenDuration:        config.TokenDuration,
		refreshTokenDuration: config.RefreshTokenDuration,
		accessTokenHandler:   accessTokenHandler,
		refreshTokenHandler:  refreshTokenHandler,
		logger:               logger,
	}, nil
}

// ============================================================================
// TOKEN GENERATION
// ============================================================================

// GenerateAccessToken generates a new JWT access token
func (tm *TokenAPI) GenerateAccessToken(input *GenerateAccessTokenInput) (string, *authv1.AccessTokenClaims, *infra_error.AppError) {
	if err := input.Validate(); err != nil {
		return "", nil, err
	}

	now := time.Now()
	expiresAt := now.Add(tm.tokenDuration)

	// Create JWT claims with generated jti
	jwtClaims := &token.JWTAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // Generate jti (not persisted)
			Issuer:    Issuer,
			Subject:   input.UserId,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   input.UserId,
		TenantID: input.TenantId,
		Email:    input.Email,
		Roles:    input.Roles,
	}

	// Sign the JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenString, err := token.SignedString([]byte(tm.secretKey))
	if err != nil {
		return "", nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}

	// Convert to proto claims (jti not included)
	protoClaims := jwtClaims.ToProtoClaims()

	return tokenString, protoClaims, nil
}

// GenerateRefreshToken generates a new refresh token for the given user
func (tm *TokenAPI) GenerateRefreshToken(input GenerateRefreshTokenInput) (string, *authv1_cache.RefreshToken, *infra_error.AppError) {
	if input.UserId == "" {
		return "", nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("user_id is required"))
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
		return "", nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}

	// Encode to base64 URL-safe string (no padding)
	tokenString := base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash, err := hash.Hash(tokenString)
	if err != nil {
		return "", nil, infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}
	// Create refresh token storage model with metadata
	refreshToken := &authv1_cache.RefreshToken{
		TokenHash: tokenHash,
		UserId:    input.UserId,
		TenantId:  input.TenantId,
		ExpiresAt: timestamppb.New(expiresAt),
		CreatedAt: timestamppb.New(now),
	}

	// Validate before storing
	if err := validator_auth_cache.ValidateRefreshToken(refreshToken); err != nil {
		return "", nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}

	// Store refresh token in Redis (use tokenString as tokenID)
	if err := tm.refreshTokenHandler.Store(input.TenantId, input.UserId, refreshToken); err != nil {
		return "", nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}
	return tokenString, refreshToken, nil
}

// ============================================================================
// TOKEN VERIFICATION
// ============================================================================

// Full verification flow
func (tm *TokenAPI) VerifyAccessToken(tokenString string) (*authv1.AccessTokenClaims, *infra_error.AppError) {
	// 1. Parse and verify JWT signature
	jwtToken, parseErr := jwt.ParseWithClaims(tokenString, &token.JWTAccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tm.secretKey), nil
	})

	if parseErr != nil {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(parseErr)
	}

	if !jwtToken.Valid {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid)
	}

	// 2. Extract claims
	jwtClaims, ok := jwtToken.Claims.(*token.JWTAccessClaims)
	if !ok {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid)
	}

	// 3. Verify against Redis storage (CRITICAL!)
	storedMetadata, err := tm.accessTokenHandler.Validate(jwtClaims.TenantID, jwtClaims.UserID)
	if err != nil {
		tm.logger.Warn("Access token validation failed",
			"tenantID", jwtClaims.TenantID,
			"userID", jwtClaims.UserID,
			"error", err)
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}

	// 5. Verify token hasn't expired (double-check against Redis)
	if time.Now().After(storedMetadata.ExpiresAt.AsTime()) {
		tm.logger.Info("Access token has expired",
			"tenantID", jwtClaims.TenantID,
			"userID", jwtClaims.UserID)
		return nil, infra_error.Auth(infra_error.AuthTokenExpired)
	}

	// 6. All checks passed - return the claims
	tm.logger.Debug("Access token verified successfully",
		"tenantID", jwtClaims.TenantID,
		"userID", jwtClaims.UserID)

	return jwtClaims.ToProtoClaims(), nil
}

// VerifyRefreshToken verifies if the given refresh token is valid
func (tm *TokenAPI) VerifyRefreshToken(tenantID string, userID string, tokenString string) (*authv1_cache.RefreshToken, *infra_error.AppError) {
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

	if valid := hash.VerifyHash(tokenString, refreshToken.TokenHash); !valid {
		tm.logger.Warn("Attempted use of invalid refresh token", "tenantID", tenantID, "userID", userID)
		// Revoke the current valid token (security measure)
		_ = tm.RevokeTokens(tenantID, userID)
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("token mismatch - possible theft detected"))
	}

	// Basic validation + expired
	if err := validator_auth_cache.ValidateRefreshToken(refreshToken); err != nil {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(err)
	}

	// SECURITY: Check for suspicious activity
	// 1. Check if token is being reused (already used recently)
	if !refreshToken.LastUsedAt.AsTime().IsZero() {
		timeSinceLastUse := time.Since(refreshToken.LastUsedAt.AsTime())
		if timeSinceLastUse < 1*time.Minute {
			// Token used twice within 1 minute - possible token theft
			// Revoke all user tokens as security measure
			tm.logger.Warn("Suspicious: Token reused within 1 minute", "tenantID", tenantID, "userID", userID)
			if err := tm.RevokeTokens(tenantID, refreshToken.UserId); err != nil {
				return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
			}
			return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("suspicious activity detected - all sessions terminated"))
		}
	}

	// Update last used timestamp with safe type assertion
	if refreshTokenHandler, ok := tm.refreshTokenHandler.(*handler.RefreshTokenHandler); ok {
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
func (tm *TokenAPI) StoreTokens(tenantID string, userID string, accessTokenMetadata *authv1_cache.TokenMetadata, refreshToken *authv1_cache.RefreshToken) *infra_error.AppError {
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

// RevokeTokens revokes all tokens (both access and refresh) for a user
// This is typically called on logout or security incidents
func (tm *TokenAPI) RevokeTokens(tenantID string, userID string) *infra_error.AppError {
	// Revoke access token
	if err := tm.RevokeAccessToken(tenantID, userID); err != nil {
		tm.logger.Error("Failed to revoke access token", "error", err, "tenantID", tenantID, "userID", userID)
		// Continue with refresh token even if access token fails
	}

	// Revoke refresh token
	if err := tm.RevokeRefreshToken(tenantID, userID); err != nil {
		tm.logger.Error("Failed to revoke refresh token", "error", err, "tenantID", tenantID, "userID", userID)
		return err
	}

	tm.logger.Debug("All tokens revoked", "tenantID", tenantID, "userID", userID)
	return nil
}

// RevokeAccessToken revokes a JWT access token
func (tm *TokenAPI) RevokeAccessToken(tenantID, userID string) *infra_error.AppError {
	return tm.accessTokenHandler.Revoke(tenantID, userID)
}

// RevokeRefreshToken revokes a refresh token
func (tm *TokenAPI) RevokeRefreshToken(tenantID string, userID string) *infra_error.AppError {
	return tm.refreshTokenHandler.Revoke(tenantID, userID)
}

// RevokeAllTenantTokens permanently deletes all tokens for ALL users in a tenant
// This is used for tenant deletion (cascade cleanup)
// Returns the number of access and refresh tokens deleted
func (tm *TokenAPI) RevokeAllTenantTokens(tenantID string) (int, int, *infra_error.AppError) {
	if tenantID == "" {
		return 0, 0, infra_error.Validation(infra_error.ValidationRequiredFields, "tenantID")
	}

	tm.logger.Warn("Deleting ALL tokens for entire tenant (hard delete)", "tenantID", tenantID)

	// Type assert to get concrete handlers
	accessHandler, ok := tm.accessTokenHandler.(*handler.AccessTokenHandler)
	if !ok {
		tm.logger.Error("accessTokenHandler is not *AccessTokenHandler")
		return 0, 0, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("Internal server error"))
	}

	refreshHandler, ok := tm.refreshTokenHandler.(*handler.RefreshTokenHandler)
	if !ok {
		tm.logger.Error("refreshTokenHandler is not *RefreshTokenHandler")
		return 0, 0, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("Internal server error"))
	}

	// Delete all access tokens using pattern
	accessCount, err := accessHandler.DeleteByPattern(tenantID, "*")
	if err != nil {
		tm.logger.Error("Failed to delete access tokens by pattern", "error", err, "tenantID", tenantID)
		// Continue with refresh tokens
	}

	// Delete all refresh tokens using pattern
	refreshCount, err := refreshHandler.DeleteByPattern(tenantID, "*")
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

func (tm *TokenAPI) GetTokenMetadata(accessTokenString string) (*authv1_cache.TokenMetadata, *infra_error.AppError) {
	if accessTokenString == "" {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("empty access token"))
	}
	claims := &authv1.AccessTokenClaims{}

	token, parseErr := jwt.Parse(accessTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(tm.secretKey), nil
	})
	if parseErr != nil {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(parseErr)
	}
	if !token.Valid {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("invalid token"))
	}
	if claimsMap, ok := token.Claims.(jwt.MapClaims); ok {
		if sub, ok := claimsMap["sub"].(string); ok {
			claims.UserId = sub
		}
		if tenantID, ok := claimsMap["tenant_id"].(string); ok {
			claims.TenantId = tenantID
		}
	}
	if claims.UserId == "" {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("user_id is required"))
	}
	if claims.TenantId == "" {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("tenant_id is required"))
	}
	// Get the single access token for this user
	accessTokenMetadata, err := tm.accessTokenHandler.GetOne(claims.TenantId, claims.UserId)
	if err != nil {
		return nil, err
	}

	if accessTokenMetadata == nil {
		return nil, infra_error.Auth(infra_error.AuthTokenInvalid).WithError(errors.New("access token not found"))
	}

	return accessTokenMetadata, nil
}
