package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

const (

	// TODO: Get secret key and durations from environment variable
	secretKey            = "secret"
	tokenDuration        = 1 * time.Hour
	refreshTokenDuration = 7 * 24 * time.Hour
)

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
	proto_auth.UnimplementedAuthServiceServer
}

func NewAuthService() *AuthService {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)
	tokenManager := token.NewTokenManager(secretKey, tokenDuration, refreshTokenDuration)
	if tokenManager == nil {
		logger.Fatal("failed to create token manager")
		return nil
	}
	return &AuthService{
		logger:       logger,
		tokenManager: tokenManager,
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
	accessTokenExpiresAt := issuedAt.Add(tokenDuration)

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
	err = s.tokenManager.StoreTokens(tenantID, userID, accessTokenMetadata.TokenID, refreshToken.Token, accessTokenMetadata, refreshToken)
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
