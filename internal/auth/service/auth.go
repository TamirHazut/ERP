package service

import (
	"context"

	"erp.localhost/internal/auth/api"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"

	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/generated/auth/v1"
	"erp.localhost/internal/infra/proto/validator"
)

// TODO: Add logs to all functions
// TODO: Remove Login + Logout from here - needs to be on UserService
// TODO: Instead of Login get Authenticate that recieves password and passwordHash and if match create tokens - needs user: tenantID, userID, passwordHash
type AuthService struct {
	logger  logger.Logger
	authAPI *api.AuthAPI
	proto_auth.UnimplementedAuthServiceServer
}

func NewAuthService(authAPI *api.AuthAPI) *AuthService {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)

	return &AuthService{
		logger:  logger,
		authAPI: authAPI,
	}
}

func (a *AuthService) Authenticate(ctx context.Context, req *proto_auth.AuthenticateRequest) (*proto_auth.TokensResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		a.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()

	newTokenResponse, err := a.authAPI.Authenticate(tenantID, userID, req.GetUserPassword(), req.GetUserHash())
	if err != nil {
		return nil, infra_error.ToGRPCError(err)
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

func (a *AuthService) VerifyToken(ctx context.Context, req *proto_auth.VerifyTokenRequest) (*proto_auth.VerifyTokenResponse, error) {
	err := a.authAPI.VerifyToken(req.GetAccessToken())
	if err != nil {
		return nil, infra_error.ToGRPCError(err)
	}
	return &proto_auth.VerifyTokenResponse{
		Valid: true,
	}, nil
}

func (a *AuthService) RefreshToken(ctx context.Context, req *proto_auth.RefreshTokenRequest) (*proto_auth.TokensResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		a.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	token := req.GetRefreshToken()

	newTokenResponse, err := a.authAPI.RefreshToken(tenantID, userID, token)
	if err != nil {
		return nil, infra_error.ToGRPCError(err)
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

func (a *AuthService) RevokeToken(ctx context.Context, req *proto_auth.RevokeTokenRequest) (*proto_auth.RevokeTokenResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		a.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	if err := a.authAPI.RevokeTokens(req.GetIdentifier().GetTenantId(), req.GetIdentifier().GetUserId(), req.GetTokens().GetAccessToken(), req.GetTokens().GetRefreshToken(), req.GetRevokedBy()); err != nil {
		return nil, infra_error.ToGRPCError(err)
	}
	return &proto_auth.RevokeTokenResponse{
		Revoked: true,
	}, nil
}

func (a *AuthService) RevokeAllTenantTokens(ctx context.Context, req *proto_auth.RevokeAllTenantTokensRequest) (*proto_auth.RevokeAllTenantTokensResponse, error) {
	// Validate input
	tenantID := req.GetTenantId()
	revokedBy := req.GetRevokedBy()

	accessCount, refreshCount, err := a.authAPI.RevokeAllTenantTokens(tenantID, revokedBy)
	if err != nil {
		a.logger.Error("Failed to revoke tenant tokens", "error", err, "tenant_id", tenantID)
		return nil, infra_error.ToGRPCError(err)
	}

	a.logger.Info("All tenant tokens revoked", "tenant_id", tenantID, "access_tokens_revoked", accessCount, "refresh_tokens_revoked", refreshCount)

	return &proto_auth.RevokeAllTenantTokensResponse{
		Revoked:              true,
		AccessTokensRevoked:  int32(accessCount),
		RefreshTokensRevoked: int32(refreshCount),
	}, nil
}
