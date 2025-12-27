package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	collection "erp.localhost/internal/auth/collections"
	"erp.localhost/internal/auth/models"
	auth_proto "erp.localhost/internal/auth/proto/v1"
	"erp.localhost/internal/auth/rbac"
	token "erp.localhost/internal/auth/token"
	mongo "erp.localhost/internal/db/mongo"
	redis_models "erp.localhost/internal/db/redis/models"
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

type AuthService struct {
	logger         *logging.Logger
	userCollection *collection.UserCollection
	tokenManager   *token.TokenManager
	rbacManager    *rbac.RBACManager
	auth_proto.UnimplementedAuthServiceServer
}

func NewAuthService() *AuthService {
	logger := logging.NewLogger(logging.ModuleAuth)
	dbHandler := mongo.NewMongoDBManager(mongo.AuthDB)
	userCollection := collection.NewUserCollection(dbHandler)
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
		logger:         logger,
		userCollection: userCollection,
		tokenManager:   tokenManager,
		rbacManager:    rbacManager,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, req *auth_proto.AuthenticateRequest) (*auth_proto.AuthenticateResponse, error) {
	email := req.User.Email
	username := req.User.Username

	// Validate email or username are provided
	authEmail := true
	if email == "" && username == "" {
		return nil, status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationRequiredFields, "email", "username").Error())
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

	roles := []string{}
	for _, role := range user.Roles {
		roles = append(roles, role.RoleID)
	}
	permissions := []string{}
	for _, permission := range user.AdditionalPermissions {
		permissions = append(permissions, permission)
	}

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
		return nil, status.Error(codes.Internal, err.Error())
	}

	hashedAccessToken := sha256.Sum256([]byte(accessToken))
	accessTokenID := hex.EncodeToString(hashedAccessToken[:])

	accessTokenMetadata := redis_models.TokenMetadata{
		TokenID:   accessTokenID,
		JTI:       accessToken,
		UserID:    user.ID.String(),
		TenantID:  user.TenantID,
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(tokenDuration),
		Revoked:   false,
		RevokedAt: nil,
		RevokedBy: "",
		IPAddress: "",
		UserAgent: "",
		Scopes:    []string{},
	}

	// Generate refresh token
	refreshToken, err := s.tokenManager.GenerateRefreshToken(ctx, token.GenerateRefreshTokenInput{
		UserID:   user.ID.String(),
		TenantID: user.TenantID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Store tokens in Redis
	err = s.tokenManager.StoreTokens(user.TenantID, user.ID.String(), accessTokenID, refreshToken.Token, accessTokenMetadata, refreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Return response
	return &auth_proto.AuthenticateResponse{
		User: &auth_proto.UserIdentifier{
			TenantId: user.TenantID,
			UserId:   user.ID.String(),
		},
		Tokens: &auth_proto.Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken.Token,
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
	return nil, status.Error(codes.Unimplemented, "method RefreshToken not implemented")
}

func (s *AuthService) RevokeToken(ctx context.Context, req *auth_proto.RevokeTokenRequest) (*auth_proto.RevokeTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RevokeToken not implemented")
}

func (s *AuthService) Logout(ctx context.Context, req *auth_proto.LogoutRequest) (*auth_proto.LogoutResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Logout not implemented")
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
