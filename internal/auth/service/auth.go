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
	model_core "erp.localhost/internal/infra/model/core"
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

func (s *AuthService) Login(ctx context.Context, req *proto_auth.LoginRequest) (*proto_auth.TokensResponse, error) {
	// input validations
	tenantID := req.GetTenantId()
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id").Error())
	}
	account_id := req.GetAccountId()
	if account_id == nil {
		err := infra_error.Validation(infra_error.ValidationRequiredFields).WithError(errors.New("missing account identifienr 'email' or 'username'")).Error()
		s.logger.Error(err, "function", "Authenticate")
		return nil, status.Error(codes.InvalidArgument, err)
	}

	// Proccess request
	var user *model_core.User
	var err error
	// TODO: uncomment when user grpc service is ready
	// switch account_id.(type) {
	// case *proto_auth.UserAuth_Email:
	// 	email := req.GetUser().GetEmail()
	// 	s.logger.Info("Authenticating user", "email", email)
	// 	user, err = s.userCollection.GetUserByEmail(tenantID, email)
	// 	if err != nil {
	// 		return nil, status.Error(codes.Internal, err.Error())
	// 	}
	// case *proto_auth.UserAuth_Username:
	// 	username := req.GetUser().GetUsername()
	// 	s.logger.Info("Authenticating user", "username", username)
	// 	user, err = s.userCollection.GetUserByUsername(tenantID, username)
	// 	if err != nil {
	// 		return nil, status.Error(codes.Internal, err.Error())
	// 	}
	// default:
	// 	return nil, status.Error(codes.InvalidArgument, errors.New("unknown account identifier type").Error())
	// }

	// Verify password
	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !password.VerifyPassword(hashedPassword, user.PasswordHash) {
		return nil, status.Error(codes.Unauthenticated, "invalid login credentials")
	}

	// Generate tokens
	newTokenResponse, err := s.generateAndStoreTokens(user)
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

func (s *AuthService) Logout(ctx context.Context, req *proto_auth.LogoutRequest) (*proto_auth.LogoutResponse, error) {
	// input validation
	tenantID := req.GetTenantId()
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id").Error())
	}
	accessToken := req.GetTokens().GetAccessToken()
	if accessToken == "" {
		return nil, status.Error(codes.InvalidArgument, infra_error.Validation(infra_error.ValidationRequiredFields, "access_token").Error())
	}

	// proccess request
	metadata, err := s.tokenManager.GetTokenMetadata(req.Tokens.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if metadata == nil {
		return nil, status.Error(codes.Internal, infra_error.Auth(infra_error.AuthTokenInvalid).Error())
	}

	user := &proto_infra.UserIdentifier{
		TenantId: metadata.TenantID,
		UserId:   metadata.UserID,
	}
	revokeError := s.revokeTokens(ctx, user, req.GetTokens(), metadata.UserID)

	// TODO: uncomment this when audit logs are implemented
	// errMsg := ""
	var message string
	if revokeError != nil {
		message = "logout failed"
		// 	errMsg = revokeError.Error()
	} else {
		message = "logout successful"
	}

	// auditLog := models.AuditLog{
	// 	TenantID:   metadata.TenantID,
	// 	Category:   models.CategoryAuth,
	// 	Action:     models.ActionLogout,
	// 	TargetID:   metadata.UserID,
	// 	TargetType: models.TargetTypeUser,
	// 	Result:     models.ResultSuccess,
	// 	Message:    message,
	// 	Error:      errMsg,
	// 	ActorID:    metadata.UserID,
	// 	ActorType:  models.ActorTypeUser,
	// }
	// err = s.auditLogsCollection.CreateAuditLog(metadata.TenantID, auditLog)
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }
	if revokeError != nil {
		return nil, status.Error(codes.Internal, revokeError.Error())
	}
	return &proto_auth.LogoutResponse{
		Message: message,
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

	// TODO: uncomment when user grpc service is ready
	// userData, err := s.userCollection.GetUserByID(user.TenantId, user.UserId)
	// if err != nil {
	// 	s.logger.Error("Failed to get user by id", "error", err, "tenant_id", user.TenantId, "user_id", user.UserId)
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }
	var userData *model_core.User
	newTokenResponse, err := s.generateAndStoreTokens(userData)
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

	if err := s.revokeTokens(ctx, identifier, req.GetTokens(), req.GetRevokedBy()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto_auth.RevokeTokenResponse{
		Revoked: true,
	}, nil
}

func (s *AuthService) revokeTokens(ctx context.Context, identifier *proto_infra.UserIdentifier, tokens *proto_auth.Tokens, revokedBy string) error {
	if identifier.GetTenantId() == "" || identifier.GetUserId() == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "user")
	}
	if revokedBy == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "requestorUserID")
	}

	// var requestor model_core.User
	// var err error
	// TODO: uncomment when user grpc service is ready
	// if tokens.AccessToken != "" || tokens.RefreshToken != "" {
	// 	requestor, err = s.userCollection.GetUserByID(user.TenantId, requestBy)
	// 	if err != nil {
	// 		return infra_error.Internal(infra_error.InternalDatabaseError, errors.New("requestedBy id was not found"))
	// 	}
	// }
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

func (s *AuthService) generateAccessToken(user *model_core.User) (model_auth_cache.TokenMetadata, error) {

	roles := []string{}
	for _, role := range user.Roles {
		roles = append(roles, role.RoleID)
	}
	permissions := []string{}
	permissions = append(permissions, user.AdditionalPermissions...)

	// Generate access token
	accessToken, err := s.tokenManager.GenerateAccessToken(&token.GenerateAccessTokenInput{
		UserID:      user.ID.String(),
		TenantID:    user.TenantID,
		Username:    user.Username,
		Email:       user.Email,
		Roles:       roles,
		Permissions: permissions,
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
		UserID:    user.ID.String(),
		TenantID:  user.TenantID,
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

func (s *AuthService) generateRefreshToken(user *model_core.User) (model_auth.RefreshToken, error) {
	issuedAt := time.Now()
	// Generate refresh token
	refreshToken, err := s.tokenManager.GenerateRefreshToken(token.GenerateRefreshTokenInput{
		UserID:    user.ID.String(),
		TenantID:  user.TenantID,
		CreatedAt: issuedAt,
	})
	if err != nil {
		return model_auth.RefreshToken{}, status.Error(codes.Internal, err.Error())
	}
	return refreshToken, nil
}

func (s *AuthService) generateAndStoreTokens(user *model_core.User) (*NewTokenResponse, error) {
	accessTokenMetadata, err := s.generateAccessToken(user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = s.tokenManager.StoreTokens(user.TenantID, user.ID.String(), accessTokenMetadata.TokenID, refreshToken.Token, accessTokenMetadata, refreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &NewTokenResponse{
		UserID:                user.ID.String(),
		TenantID:              user.TenantID,
		AccessToken:           accessTokenMetadata.JTI,
		AccessTokenExpiresAt:  accessTokenMetadata.ExpiresAt.Unix(),
		RefreshToken:          refreshToken.Token,
		RefreshTokenExpiresAt: refreshToken.ExpiresAt.Unix(),
	}, nil
}
