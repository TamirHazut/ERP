package service

import (
	"context"

	repository "erp.localhost/internal/auth/collections"
	"erp.localhost/internal/auth/models"
	auth_proto "erp.localhost/internal/auth/proto/v1"
	token "erp.localhost/internal/auth/token"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	authTypeEmail    string = "email"
	authTypeUsername string = "username"
)

type AuthService struct {
	logger         *logging.Logger
	userRepository *repository.UserRepository
	tokenManager   *token.TokenManager
	auth_proto.UnimplementedAuthServiceServer
}

func NewAuthService() *AuthService {
	logger := logging.NewLogger(logging.ModuleAuth)
	return &AuthService{
		logger: logger,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, req *auth_proto.AuthenticateRequest) (*auth_proto.AuthenticateResponse, error) {
	email := req.User.Email
	username := req.User.Username

	// Validate email or username are provided
	authEmail := true
	if email == "" && username == "" {
		return nil, erp_errors.Validation(erp_errors.ValidationRequiredFields, "email", "username")
	} else if username != "" {
		authEmail = false
	}

	// Get user by email or username
	tenantID := req.User.TenantId
	if tenantID == "" {
		return nil, erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenant_id")
	}
	var user *models.User
	var err error
	if authEmail {
		s.logger.Info("Authenticating user", "email", email)
		user, err = s.userRepository.GetUserByEmail(tenantID, email)
		if err != nil {
			return nil, err
		}
	} else {
		s.logger.Info("Authenticating user", "username", username)
		user, err = s.userRepository.GetUserByUsername(tenantID, username)
		if err != nil {
			return nil, err
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
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.tokenManager.GenerateRefreshToken(ctx, token.GenerateRefreshTokenInput{
		UserID:   user.ID.String(),
		TenantID: user.TenantID,
	})
	if err != nil {
		return nil, err
	}

	// Return response
	return &auth_proto.AuthenticateResponse{
		User: &auth_proto.UserIdentifier{
			TenantId: user.TenantID,
			UserId:   user.ID.String(),
		},
		Tokens: &auth_proto.Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

func (s *AuthService) VerifyToken(context.Context, *auth_proto.VerifyTokenRequest) (*auth_proto.VerifyTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method VerifyToken not implemented")
}
func (s *AuthService) RefreshToken(context.Context, *auth_proto.RefreshTokenRequest) (*auth_proto.RefreshTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RefreshToken not implemented")
}
func (s *AuthService) RevokeToken(context.Context, *auth_proto.RevokeTokenRequest) (*auth_proto.RevokeTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RevokeToken not implemented")
}
func (s *AuthService) Logout(context.Context, *auth_proto.LogoutRequest) (*auth_proto.LogoutResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Logout not implemented")
}
func (s *AuthService) CheckPermissions(context.Context, *auth_proto.CheckPermissionsRequest) (*auth_proto.CheckPermissionsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CheckPermissions not implemented")
}
