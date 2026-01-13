package service

import (
	"context"
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
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
	proto_infra "erp.localhost/internal/infra/proto/infra/v1"
	"erp.localhost/internal/infra/proto/validator"
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

// TODO: Add logs to all functions
// TODO: Remove Login + Logout from here - needs to be on UserService
// TODO: Instead of Login get Authenticate that recieves password and passwordHash and if match create tokens - needs user: tenantID, userID, passwordHash
type AuthService struct {
	logger       logger.Logger
	tokenManager *token.TokenManager
	config       *TokenConfig
	proto_auth.UnimplementedAuthServiceServer
}

func NewAuthService() *AuthService {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)

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
	return &AuthService{
		logger:       logger,
		tokenManager: tokenManager,
		config:       config,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, req *proto_auth.AuthenticateRequest) (*proto_auth.TokensResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		s.logger.Error("Failed to authenticate user", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}
	userPassword := req.GetUserPassword()
	userHash := req.GetUserHash()
	if userPassword == "" || userHash == "" {
		errMsg := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: user_password, user_hash")).Error()
		s.logger.Error("Failed to authenticate user", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}
	// input validations
	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()

	// Verify password
	hashedPassword, err := password.HashPassword(userPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !password.VerifyPassword(hashedPassword, userHash) {
		return nil, status.Error(codes.Unauthenticated, "invalid login credentials")
	}

	// Generate tokens
	newTokenResponse, err := s.generateAndStoreTokens(tenantID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Return response
	return &proto_auth.TokensResponse{
		Tokens: &proto_auth.Tokens{
			AccessToken:  newTokenResponse.AccessToken,
			RefreshToken: newTokenResponse.RefreshToken,
		},
		ExpiresIn: &proto_auth.ExpiresIn{
			AccessToken:  newTokenResponse.AccessTokenExpiresAt,
			RefreshToken: newTokenResponse.RefreshTokenExpiresAt,
		},
	}, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, req *proto_auth.VerifyTokenRequest) (*proto_auth.VerifyTokenResponse, error) {
	token := req.GetAccessToken()
	if token == "" {
		return nil, status.Error(codes.InvalidArgument, infra_error.Validation(infra_error.ValidationRequiredFields, "access_token").Error())
	}
	_, err := s.tokenManager.VerifyAccessToken(token)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto_auth.VerifyTokenResponse{
		Valid: true,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *proto_auth.RefreshTokenRequest) (*proto_auth.TokensResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	err := validator.ValidateUserIdentifier(identifier)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	token := req.GetRefreshToken()
	if token == "" {
		return nil, status.Error(codes.InvalidArgument, infra_error.Validation(infra_error.ValidationRequiredFields, "refresh_token").Error())
	}

	refreshToken, err := s.tokenManager.VerifyRefreshToken(tenantID, userID, token)
	if err != nil {
		s.logger.Error("Failed to verify refresh token", "error", err, "tenant_id", tenantID, "user_id", userID, "refresh_token", token)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Revoke old access tokens to prevent orphaned tokens
	// Note: We only revoke access tokens, not refresh tokens, since the refresh token
	// is still valid and will be revoked explicitly below
	if err := s.tokenManager.RevokeAllAccessTokens(tenantID, userID, "system"); err != nil {
		s.logger.Warn("Failed to revoke old access tokens before refresh", "error", err, "tenant_id", tenantID, "user_id", userID)
		// Continue anyway - non-critical failure
	}

	newTokenResponse, err := s.generateAndStoreTokens(tenantID, userID)
	if err != nil {
		s.logger.Error("Failed to generate and store tokens", "error", err, "tenant_id", tenantID, "user_id", userID)
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.tokenManager.RevokeRefreshToken(tenantID, userID, refreshToken.Token, "system", true)
	if err != nil {
		s.logger.Error("Failed to revoke refresh token", "error", err, "tenant_id", tenantID, "user_id", userID, "refresh_token", req.RefreshToken)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto_auth.TokensResponse{
		Tokens: &proto_auth.Tokens{
			AccessToken:  newTokenResponse.AccessToken,
			RefreshToken: newTokenResponse.RefreshToken,
		},
		ExpiresIn: &proto_auth.ExpiresIn{
			AccessToken:  newTokenResponse.AccessTokenExpiresAt,
			RefreshToken: newTokenResponse.RefreshTokenExpiresAt,
		},
	}, nil
}

func (s *AuthService) RevokeToken(ctx context.Context, req *proto_auth.RevokeTokenRequest) (*proto_auth.RevokeTokenResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	err := validator.ValidateUserIdentifier(identifier)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.revokeTokens(identifier, req.GetTokens(), req.GetRevokedBy()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto_auth.RevokeTokenResponse{
		Revoked: true,
	}, nil
}

func (s *AuthService) RevokeAllTenantTokens(ctx context.Context, req *proto_auth.RevokeAllTenantTokensRequest) (*proto_auth.RevokeAllTenantTokensResponse, error) {
	// Validate input
	tenantID := req.GetTenantId()
	revokedBy := req.GetRevokedBy()

	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	if revokedBy == "" {
		return nil, status.Error(codes.InvalidArgument, "revoked_by is required")
	}

	s.logger.Warn("Revoking all tenant tokens", "tenant_id", tenantID, "revoked_by", revokedBy)

	// TODO: Add RBAC check - only system admins should be able to revoke all tenant tokens
	// This is a critical operation that should require elevated permissions

	// Revoke all tokens for this tenant
	accessCount, refreshCount, err := s.tokenManager.RevokeAllTenantTokens(tenantID, revokedBy)
	if err != nil {
		s.logger.Error("Failed to revoke tenant tokens", "error", err, "tenant_id", tenantID)
		return nil, status.Error(codes.Internal, "failed to revoke tenant tokens")
	}

	s.logger.Info("All tenant tokens revoked", "tenant_id", tenantID, "access_tokens_revoked", accessCount, "refresh_tokens_revoked", refreshCount)

	return &proto_auth.RevokeAllTenantTokensResponse{
		Revoked:              true,
		AccessTokensRevoked:  int32(accessCount),
		RefreshTokensRevoked: int32(refreshCount),
	}, nil
}

func (s *AuthService) revokeTokens(identifier *proto_infra.UserIdentifier, tokens *proto_auth.Tokens, revokedBy string) error {
	if identifier.GetTenantId() == "" || identifier.GetUserId() == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "user")
	}
	if revokedBy == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "requestorUserID")
	}

	if tokens.AccessToken != "" {
		err := s.tokenManager.RevokeAccessToken(tokens.AccessToken, revokedBy)
		if err != nil {
			return err
		}
	}
	if tokens.RefreshToken != "" {
		err := s.tokenManager.RevokeRefreshToken(identifier.GetTenantId(), identifier.GetUserId(), tokens.RefreshToken, revokedBy, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *AuthService) generateAccessToken(tenantID string, userID string) (model_auth_cache.TokenMetadata, error) {
	// Generate access token
	accessToken, err := s.tokenManager.GenerateAccessToken(&token.GenerateAccessTokenInput{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		return model_auth_cache.TokenMetadata{}, status.Error(codes.Internal, err.Error())
	}

	hashedAccessToken := sha256.Sum256([]byte(accessToken))
	accessTokenID := hex.EncodeToString(hashedAccessToken[:])

	issuedAt := time.Now()
	accessTokenExpiresAt := issuedAt.Add(s.config.TokenDuration)

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

func (s *AuthService) generateRefreshToken(tenantID string, userID string) (model_auth.RefreshToken, error) {
	issuedAt := time.Now()
	// Generate refresh token
	refreshToken, err := s.tokenManager.GenerateRefreshToken(token.GenerateRefreshTokenInput{
		UserID:    userID,
		TenantID:  tenantID,
		CreatedAt: issuedAt,
	})
	if err != nil {
		return model_auth.RefreshToken{}, status.Error(codes.Internal, err.Error())
	}
	return refreshToken, nil
}

func (s *AuthService) generateAndStoreTokens(tenantID string, userID string) (*NewTokenResponse, error) {
	accessTokenMetadata, err := s.generateAccessToken(tenantID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	refreshToken, err := s.generateRefreshToken(tenantID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Store tokens (single token per user - automatically replaces existing)
	err = s.tokenManager.StoreTokens(tenantID, userID, accessTokenMetadata, refreshToken)
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
