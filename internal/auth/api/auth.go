package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"time"

	"erp.localhost/internal/auth/password"
	token "erp.localhost/internal/auth/token"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_auth_cache "erp.localhost/internal/infra/model/auth/cache"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

type NewTokenResponse struct {
	UserID                string `json:"user_id"`
	TenantID              string `json:"tenant_id"`
	AccessToken           string `json:"access_token"`
	AccessTokenExpiresAt  int64  `json:"access_token_expires_at"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresAt int64  `json:"refresh_token_expires_at"`
}

// RBACAPI combines all RBAC APIs for easy initialization
type AuthAPI struct {
	logger       logger.Logger
	tokenManager *token.TokenManager
	config       *TokenConfig
}

func NewAuthAPI(logger logger.Logger) *AuthAPI {
	// Load configuration from environment variables
	config := LoadTokenConfig()
	logger.Info("Token configuration loaded",
		"access_token_duration", config.TokenDuration.String(),
		"refresh_token_duration", config.RefreshTokenDuration.String())

	tokenManager := token.NewTokenManager(config.SecretKey, config.TokenDuration, config.RefreshTokenDuration)
	if tokenManager == nil {
		logger.Fatal("failed to create token manager")
		return nil
	}
	return &AuthAPI{
		logger:       logger,
		tokenManager: tokenManager,
		config:       config,
	}
}

func (a *AuthAPI) Authenticate(tenantID, userID, userPassword, userHash string) (*NewTokenResponse, error) {
	if tenantID == "" || userID == "" || userPassword == "" || userHash == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, user_password, user_hash"))
		a.logger.Error("Failed to authenticate user", "error", err)
		return nil, err
	}
	// Verify password
	hashedPassword, err := password.HashPassword(userPassword)
	if err != nil {
		return nil, err
	}

	if !password.VerifyPassword(hashedPassword, userHash) {
		return nil, infra_error.Auth(infra_error.AuthInvalidCredentials)
	}

	// Generate tokens
	return a.generateAndStoreTokens(tenantID, userID)
}

func (a *AuthAPI) VerifyToken(token string) error {
	if token == "" {
		return status.Error(codes.InvalidArgument, infra_error.Validation(infra_error.ValidationRequiredFields, "access_token").Error())
	}
	_, err := a.tokenManager.VerifyAccessToken(token)
	return err
}

func (a *AuthAPI) RefreshToken(tenantID, userID, token string) (*NewTokenResponse, error) {
	if tenantID == "" || userID == "" || token == "" {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, refresh_token"))
	}

	refreshToken, err := a.tokenManager.VerifyRefreshToken(tenantID, userID, token)
	if err != nil {
		a.logger.Error("Failed to verify refresh token", "error", err, "tenant_id", tenantID, "user_id", userID, "refresh_token", token)
		return nil, err
	}

	// Revoke old access tokens to prevent orphaned tokens
	// Note: We only revoke access tokens, not refresh tokens, since the refresh token
	// is still valid and will be revoked explicitly below
	if err := a.tokenManager.RevokeAllAccessTokens(tenantID, userID, "system"); err != nil {
		a.logger.Warn("Failed to revoke old access tokens before refresh", "error", err, "tenant_id", tenantID, "user_id", userID)
		// Continue anyway - non-critical failure
	}

	newTokenResponse, err := a.generateAndStoreTokens(tenantID, userID)
	if err != nil {
		a.logger.Error("Failed to generate and store tokens", "error", err, "tenant_id", tenantID, "user_id", userID)
		return nil, err
	}

	err = a.tokenManager.RevokeRefreshToken(tenantID, userID, refreshToken.Token, "system", true)
	if err != nil {
		a.logger.Error("Failed to revoke refresh token", "error", err, "tenant_id", tenantID, "user_id", userID)
		return nil, err
	}
	return newTokenResponse, nil
}

func (a *AuthAPI) RevokeTokens(tenantID, userID, accessToken, refreshToken, revokedBy string) error {
	if tenantID == "" || userID == "" || accessToken == "" || refreshToken == "" || revokedBy == "" {
		return infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, access_token, refresh_token, revoked_by"))
	}

	if accessToken != "" {
		err := a.tokenManager.RevokeAccessToken(accessToken, revokedBy)
		if err != nil {
			return err
		}
	}
	if refreshToken != "" {
		err := a.tokenManager.RevokeRefreshToken(tenantID, userID, refreshToken, revokedBy, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AuthAPI) RevokeAllTenantTokens(tenantID, revokedBy string) (int, int, error) {
	if tenantID == "" || revokedBy == "" {
		return 0, 0, infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, access_token, refresh_token, revoked_by"))
	}

	a.logger.Warn("Revoking all tenant tokens", "tenant_id", tenantID, "revoked_by", revokedBy)

	// TODO: Add RBAC check - only system admins should be able to revoke all tenant tokens
	// This is a critical operation that should require elevated permissions

	// Revoke all tokens for this tenant
	return a.tokenManager.RevokeAllTenantTokens(tenantID, revokedBy)
}

func (a *AuthAPI) generateAccessToken(tenantID string, userID string) (model_auth_cache.TokenMetadata, error) {
	// Generate access token
	accessToken, err := a.tokenManager.GenerateAccessToken(&token.GenerateAccessTokenInput{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		return model_auth_cache.TokenMetadata{}, status.Error(codes.Internal, err.Error())
	}

	hashedAccessToken := sha256.Sum256([]byte(accessToken))
	accessTokenID := hex.EncodeToString(hashedAccessToken[:])

	issuedAt := time.Now()
	accessTokenExpiresAt := issuedAt.Add(a.config.TokenDuration)

	accessTokenMetadata := model_auth_cache.TokenMetadata{
		TokenID:   accessTokenID,
		JTI:       accessToken,
		UserID:    userID,
		TenantID:  tenantID,
		TokenType: "access",
		IssuedAt:  issuedAt,
		ExpiresAt: accessTokenExpiresAt,
		Revoked:   false,
		RevokedAt: nil,
		RevokedBy: "",
		IPAddress: "",
		UserAgent: "",
		Scopes:    []string{},
	}

	return accessTokenMetadata, nil
}

func (a *AuthAPI) generateRefreshToken(tenantID string, userID string) (model_auth.RefreshToken, error) {
	issuedAt := time.Now()
	// Generate refresh token
	refreshToken, err := a.tokenManager.GenerateRefreshToken(token.GenerateRefreshTokenInput{
		UserID:    userID,
		TenantID:  tenantID,
		CreatedAt: issuedAt,
	})
	if err != nil {
		return model_auth.RefreshToken{}, status.Error(codes.Internal, err.Error())
	}
	return refreshToken, nil
}

func (a *AuthAPI) generateAndStoreTokens(tenantID string, userID string) (*NewTokenResponse, error) {
	accessTokenMetadata, err := a.generateAccessToken(tenantID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	refreshToken, err := a.generateRefreshToken(tenantID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Store tokens (single token per user - automatically replaces existing)
	err = a.tokenManager.StoreTokens(tenantID, userID, accessTokenMetadata, refreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &NewTokenResponse{
		UserID:                userID,
		TenantID:              tenantID,
		AccessToken:           accessTokenMetadata.JTI,
		AccessTokenExpiresAt:  accessTokenMetadata.ExpiresAt.Unix(),
		RefreshToken:          refreshToken.Token,
		RefreshTokenExpiresAt: refreshToken.ExpiresAt.Unix(),
	}, nil
}
