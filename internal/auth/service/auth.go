package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"erp.localhost/internal/auth/password"
	"erp.localhost/internal/auth/rbac"
	token "erp.localhost/internal/auth/token/manager"
	token_manager "erp.localhost/internal/auth/token/manager"
	erp_errors "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging"
	auth_models "erp.localhost/internal/infra/model/auth"
	auth_cache_models "erp.localhost/internal/infra/model/auth/cache"
	core_models "erp.localhost/internal/infra/model/core"
	shared_models "erp.localhost/internal/infra/model/shared"
	auth_proto "erp.localhost/internal/infra/proto/auth/v1"
	"google.golang.org/grpc"
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
	logger       *logging.Logger
	tokenManager *token_manager.TokenManager
	rbacManager  *rbac.RBACManager
	auth_proto.UnimplementedAuthServiceServer
}

func NewAuthService() *AuthService {
	logger := logging.NewLogger(shared_models.ModuleAuth)
	tokenManager := token.NewTokenManager(secretKey, tokenDuration, refreshTokenDuration)
	if tokenManager == nil {
		logger.Fatal("failed to create token manager")
		return nil
	}
	rbacManager := rbac.NewRBACManager()
	if rbacManager == nil {
		logger.Fatal("failed to create rbac manager")
		return nil
	}
	return &AuthService{
		logger: logger,
		// userCollection:      userCollection,
		tokenManager: tokenManager,
		rbacManager:  rbacManager,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, req *auth_proto.AuthenticateRequest) (*auth_proto.TokensResponse, error) {
	// input validations
	tenantID := req.GetUser().GetTenantId()
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenant_id").Error())
	}
	account_id := req.GetUser().GetAccountId()
	if account_id == nil {
		err := erp_errors.Validation(erp_errors.ValidationRequiredFields).WithError(errors.New("missing account identifienr 'email' or 'username'")).Error()
		s.logger.Error(err, "function", "Authenticate")
		return nil, status.Error(codes.InvalidArgument, err)
	}

	// Proccess request
	var user *core_models.User
	var err error
	// TODO: uncomment when user grpc service is ready
	// switch account_id.(type) {
	// case *auth_proto.UserAuth_Email:
	// 	email := req.GetUser().GetEmail()
	// 	s.logger.Info("Authenticating user", "email", email)
	// 	user, err = s.userCollection.GetUserByEmail(tenantID, email)
	// 	if err != nil {
	// 		return nil, status.Error(codes.Internal, err.Error())
	// 	}
	// case *auth_proto.UserAuth_Username:
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
	return &auth_proto.TokensResponse{
		Tokens: &auth_proto.Tokens{
			AccessToken:  newTokenResponse.AccessToken,
			RefreshToken: newTokenResponse.RefreshToken,
		},
		ExpiresIn: &auth_proto.ExpiresIn{
			AccessToken:  newTokenResponse.AccessTokenExpiresAt,
			RefreshToken: newTokenResponse.RefreshTokenExpiresAt,
		},
	}, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, req *auth_proto.VerifyTokenRequest) (*auth_proto.VerifyTokenResponse, error) {
	token := req.AccessToken
	if token == "" {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "access_token").Error())
	}
	_, err := s.tokenManager.VerifyAccessToken(token)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auth_proto.VerifyTokenResponse{
		Valid: true,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *auth_proto.RefreshTokenRequest) (*auth_proto.TokensResponse, error) {
	user := req.User
	if user == nil || user.TenantId == "" || user.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "user").Error())
	}

	refreshToken, err := s.tokenManager.VerifyRefreshToken(user.TenantId, user.UserId, req.RefreshToken)
	if err != nil {
		s.logger.Error("Failed to verify refresh token", "error", err, "tenant_id", user.TenantId, "user_id", user.UserId, "refresh_token", req.RefreshToken)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: uncomment when user grpc service is ready
	// userData, err := s.userCollection.GetUserByID(user.TenantId, user.UserId)
	// if err != nil {
	// 	s.logger.Error("Failed to get user by id", "error", err, "tenant_id", user.TenantId, "user_id", user.UserId)
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }
	var userData *core_models.User
	newTokenResponse, err := s.generateAndStoreTokens(userData)
	if err != nil {
		s.logger.Error("Failed to generate and store tokens", "error", err, "tenant_id", user.TenantId, "user_id", user.UserId)
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.tokenManager.RevokeRefreshToken(user.TenantId, user.UserId, refreshToken.Token, "system", true)
	if err != nil {
		s.logger.Error("Failed to revoke refresh token", "error", err, "tenant_id", user.TenantId, "user_id", user.UserId, "refresh_token", req.RefreshToken)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &auth_proto.TokensResponse{
		Tokens: &auth_proto.Tokens{
			AccessToken:  newTokenResponse.AccessToken,
			RefreshToken: newTokenResponse.RefreshToken,
		},
		ExpiresIn: &auth_proto.ExpiresIn{
			AccessToken:  newTokenResponse.AccessTokenExpiresAt,
			RefreshToken: newTokenResponse.RefreshTokenExpiresAt,
		},
	}, nil
}

func (s *AuthService) RevokeToken(ctx context.Context, req *auth_proto.RevokeTokenRequest) (*auth_proto.RevokeTokenResponse, error) {
	err := s.revokeTokens(req.User, req.Tokens, req.RequestBy)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auth_proto.RevokeTokenResponse{
		Revoked: true,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *auth_proto.LogoutRequest) (*auth_proto.LogoutResponse, error) {
	// input validation
	tenantID := req.GetTenantId()
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenant_id").Error())
	}
	accessToken := req.GetTokens().GetAccessToken()
	if accessToken == "" {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "access_token").Error())
	}

	// proccess request
	metadata, err := s.tokenManager.GetTokenMetadata(req.Tokens.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if metadata == nil {
		return nil, status.Error(codes.Internal, erp_errors.Auth(erp_errors.AuthTokenInvalid).Error())
	}

	user := &auth_proto.UserIdentifier{
		TenantId: metadata.TenantID,
		UserId:   metadata.UserID,
	}
	revokeError := s.revokeTokens(user, req.Tokens, metadata.UserID)

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
	return &auth_proto.LogoutResponse{
		Message: message,
	}, nil
}

func (s *AuthService) CheckPermissions(req *auth_proto.CheckPermissionsRequest, stream grpc.ServerStreamingServer[auth_proto.PermissionResponse]) error {
	// Validate input
	tenantID := req.User.GetTenantId()
	if tenantID == "" {
		return status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenant_id").Error())
	}
	userID := req.User.GetUserId()
	if userID == "" {
		return status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "user_id").Error())
	}
	// Validate permissions
	permissions := req.GetPermissions()
	if len(permissions) == 0 {
		return status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "permissions").Error())
	}

	// proccess request
	permissionsCheckResponse, err := s.rbacManager.CheckUserPermissions(tenantID, userID, permissions)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	for permission, hasPermission := range permissionsCheckResponse {
		permissionRes := &auth_proto.PermissionResponse{
			Permission:    permission,
			HasPermission: hasPermission,
		}
		if err := stream.Send(permissionRes); err != nil {
			return err
		}
	}
	return nil
}

func (s *AuthService) revokeTokens(user *auth_proto.UserIdentifier, tokens *auth_proto.Tokens, requestBy string) error {
	if user.GetTenantId() == "" || user.GetUserId() == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "user")
	}
	if requestBy == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "requestedBy")
	}

	var requestor core_models.User
	var err error
	// TODO: uncomment when user grpc service is ready
	// if tokens.AccessToken != "" || tokens.RefreshToken != "" {
	// 	requestor, err = s.userCollection.GetUserByID(user.TenantId, requestBy)
	// 	if err != nil {
	// 		return erp_errors.Internal(erp_errors.InternalDatabaseError, errors.New("requestedBy id was not found"))
	// 	}
	// }
	if tokens.AccessToken != "" {
		// Validate user permission to revoke access_token
		permission, _ := auth_models.CreatePermissionString(auth_models.ResourceAccessToken, auth_models.PermissionActionDelete)
		if err = s.rbacManager.HasPermission(requestor.TenantID, requestor.ID.String(), permission); err != nil {
			return err
		}

		err = s.tokenManager.RevokeAccessToken(tokens.AccessToken, requestBy)
		if err != nil {
			return err
		}
	}
	if tokens.RefreshToken != "" {
		// Validate user permission to revoke refresh_token
		permission, _ := auth_models.CreatePermissionString(auth_models.ResourceRefreshToken, auth_models.PermissionActionDelete)
		if err = s.rbacManager.HasPermission(requestor.TenantID, requestor.ID.String(), permission); err != nil {
			return err
		}

		err := s.tokenManager.RevokeRefreshToken(user.TenantId, user.UserId, tokens.RefreshToken, requestBy, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *AuthService) generateAccessToken(user *core_models.User) (auth_cache_models.TokenMetadata, error) {

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
		return auth_cache_models.TokenMetadata{}, status.Error(codes.Internal, err.Error())
	}

	hashedAccessToken := sha256.Sum256([]byte(accessToken))
	accessTokenID := hex.EncodeToString(hashedAccessToken[:])

	issuedAt := time.Now()
	accessTokenExpiresAt := issuedAt.Add(tokenDuration)

	accessTokenMetadata := auth_cache_models.TokenMetadata{
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

func (s *AuthService) generateRefreshToken(user *core_models.User) (auth_models.RefreshToken, error) {
	issuedAt := time.Now()
	// Generate refresh token
	refreshToken, err := s.tokenManager.GenerateRefreshToken(token.GenerateRefreshTokenInput{
		UserID:    user.ID.String(),
		TenantID:  user.TenantID,
		CreatedAt: issuedAt,
	})
	if err != nil {
		return auth_models.RefreshToken{}, status.Error(codes.Internal, err.Error())
	}
	return refreshToken, nil
}

func (s *AuthService) generateAndStoreTokens(user *core_models.User) (*NewTokenResponse, error) {
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
