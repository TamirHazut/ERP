package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	collection "erp.localhost/internal/auth/collections"
	"erp.localhost/internal/auth/models"
	auth_models "erp.localhost/internal/auth/models/cache"
	auth_proto "erp.localhost/internal/auth/proto/auth/v1"
	"erp.localhost/internal/auth/rbac"
	token "erp.localhost/internal/auth/token/manager"
	token_manager "erp.localhost/internal/auth/token/manager"
	common_models "erp.localhost/internal/common/models"
	mongo "erp.localhost/internal/db/mongo"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
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

type AuthService struct {
	logger              *logging.Logger
	userCollection      *collection.UserCollection
	tokenManager        *token_manager.TokenManager
	rbacManager         *rbac.RBACManager
	auditLogsCollection *collection.AuditLogsCollection
	auth_proto.UnimplementedAuthServiceServer
}

func newCollectionHandler[T any](collection string) *mongo.BaseCollectionHandler[T] {
	logger := logging.NewLogger(common_models.ModuleAuth)
	return mongo.NewBaseCollectionHandler[T](string(collection), logger)
}

func NewAuthService() *AuthService {
	logger := logging.NewLogger(common_models.ModuleAuth)
	userCollectionHandler := newCollectionHandler[models.User](string(mongo.UsersCollection))
	if userCollectionHandler == nil {
		logger.Fatal("failed to create users collection handler")
		return nil
	}
	userCollection := collection.NewUserCollection(userCollectionHandler)
	if userCollection == nil {
		logger.Fatal("failed to create users collection handler")
		return nil
	}
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
	auditLogsCollectionHandler := newCollectionHandler[models.AuditLog](string(mongo.AuditLogsCollection))
	if auditLogsCollectionHandler == nil {
		logger.Fatal("failed to create audit logs collection handler")
		return nil
	}
	auditLogsCollection := collection.NewAuditLogsCollection(auditLogsCollectionHandler)
	return &AuthService{
		logger:              logger,
		userCollection:      userCollection,
		tokenManager:        tokenManager,
		rbacManager:         rbacManager,
		auditLogsCollection: auditLogsCollection,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, req *auth_proto.AuthenticateRequest) (*auth_proto.AuthenticateResponse, error) {
	email := req.User.Email
	username := req.User.Username

	// Validate email or username are provided
	authEmail := true
	if email == "" && username == "" {
		err := erp_errors.Validation(erp_errors.ValidationRequiredFields, "email", "username").Error()
		s.logger.Error(err, "function", "Authenticate")
		return nil, status.Error(codes.InvalidArgument, err)
	} else if username != "" {
		authEmail = false
	}

	// Get user by email or username
	tenantID := req.User.TenantId
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenant_id").Error())
	}
	var user *models.User
	var err error
	if authEmail {
		s.logger.Info("Authenticating user", "email", email)
		user, err = s.userCollection.GetUserByEmail(tenantID, email)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		s.logger.Info("Authenticating user", "username", username)
		user, err = s.userCollection.GetUserByUsername(tenantID, username)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	newTokenResponse, err := s.generateAndStoreTokens(user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Return response
	return &auth_proto.AuthenticateResponse{
		User: &auth_proto.UserIdentifier{
			TenantId: newTokenResponse.TenantID,
			UserId:   newTokenResponse.UserID,
		},
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

func (s *AuthService) RefreshToken(ctx context.Context, req *auth_proto.RefreshTokenRequest) (*auth_proto.RefreshTokenResponse, error) {
	user := req.User
	if user == nil || user.TenantId == "" || user.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "user").Error())
	}

	refreshToken, err := s.tokenManager.VerifyRefreshToken(user.TenantId, user.UserId, req.RefreshToken)
	if err != nil {
		s.logger.Error("Failed to verify refresh token", "error", err, "tenant_id", user.TenantId, "user_id", user.UserId, "refresh_token", req.RefreshToken)
		return nil, status.Error(codes.Internal, err.Error())
	}

	userData, err := s.userCollection.GetUserByID(user.TenantId, user.UserId)
	if err != nil {
		s.logger.Error("Failed to get user by id", "error", err, "tenant_id", user.TenantId, "user_id", user.UserId)
		return nil, status.Error(codes.Internal, err.Error())
	}

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

	return &auth_proto.RefreshTokenResponse{
		User: &auth_proto.UserIdentifier{
			TenantId: newTokenResponse.TenantID,
			UserId:   newTokenResponse.UserID,
		},
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
		User:    req.User,
		Revoked: true,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *auth_proto.LogoutRequest) (*auth_proto.LogoutResponse, error) {
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
		User:    user,
		Message: message,
	}, nil
}

func (s *AuthService) CheckPermissions(ctx context.Context, req *auth_proto.CheckPermissionsRequest) (*auth_proto.CheckPermissionsResponse, error) {
	tenantID := req.User.TenantId
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenant_id").Error())
	}
	userID := req.User.UserId
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "user_id").Error())
	}
	// Validate permissions
	permissions := req.Permissions
	if len(permissions) == 0 {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "permissions").Error())
	}

	permissionsCheckResponse, err := s.rbacManager.CheckUserPermissions(tenantID, userID, permissions)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	permissionsResponses := make([]*auth_proto.PermissionResponse, 0)
	for permission, hasPermission := range permissionsCheckResponse {
		permissionsResponses = append(permissionsResponses, &auth_proto.PermissionResponse{
			Permission:    permission,
			HasPermission: hasPermission,
		})
	}
	return &auth_proto.CheckPermissionsResponse{
		User: &auth_proto.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		Permissions: permissionsResponses,
	}, err
}

func (s *AuthService) revokeTokens(user *auth_proto.UserIdentifier, tokens *auth_proto.Tokens, requestBy string) error {
	if user == nil || user.TenantId == "" || user.UserId == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "user")
	}
	if requestBy == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "requestedBy")
	}

	var requestor *models.User
	var err error
	if tokens.AccessToken != "" || tokens.RefreshToken != "" {
		requestor, err = s.userCollection.GetUserByUsername(user.TenantId, requestBy)
		if err != nil {
			return erp_errors.Validation(erp_errors.ValidationRequiredFields, "permissions")
		}
	}
	if tokens.AccessToken != "" {
		// Validate user permission to revoke access_token
		permission, _ := models.CreatePermissionString(models.ResourceAccessToken, models.PermissionActionDelete)
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
		permission, _ := models.CreatePermissionString(models.ResourceRefreshToken, models.PermissionActionDelete)
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

func (s *AuthService) generateAccessToken(user *models.User) (auth_models.TokenMetadata, error) {

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
		return auth_models.TokenMetadata{}, status.Error(codes.Internal, err.Error())
	}

	hashedAccessToken := sha256.Sum256([]byte(accessToken))
	accessTokenID := hex.EncodeToString(hashedAccessToken[:])

	issuedAt := time.Now()
	accessTokenExpiresAt := issuedAt.Add(tokenDuration)

	accessTokenMetadata := auth_models.TokenMetadata{
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

func (s *AuthService) generateRefreshToken(user *models.User) (models.RefreshToken, error) {
	issuedAt := time.Now()
	// Generate refresh token
	refreshToken, err := s.tokenManager.GenerateRefreshToken(token.GenerateRefreshTokenInput{
		UserID:    user.ID.String(),
		TenantID:  user.TenantID,
		CreatedAt: issuedAt,
	})
	if err != nil {
		return models.RefreshToken{}, status.Error(codes.Internal, err.Error())
	}
	return refreshToken, nil
}

func (s *AuthService) generateAndStoreTokens(user *models.User) (*NewTokenResponse, error) {
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
