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
		a.logger.Warn("faild attemt to authenticate", "tenantID", tenantID, "userID", userID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	a.logger.Debug("user authenticated successfuly", "tenantID", tenantID, "userID", userID)
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
		a.logger.Error("failed to verify token", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	a.logger.Debug("token verified")
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
		a.logger.Error("failed to refresh token", "tenantID", tenantID, "userID", userID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	a.logger.Debug("tokens refreshed successfuly", "tenantID", tenantID, "userID", userID)
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

	tenantID := req.GetIdentifier().GetTenantId()
	userID := req.GetIdentifier().GetUserId()
	revokedBy := req.GetRevokedBy()

	if err := a.authAPI.RevokeTokens(tenantID, userID, req.GetTokens().GetAccessToken(), req.GetTokens().GetRefreshToken(), revokedBy); err != nil {
		a.logger.Error("failed to revoke token", "tenantID", tenantID, "userID", userID, "revokedBy", revokedBy, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	a.logger.Debug("token revoked successfuly", "tenantID", tenantID, "userID", userID, "revokedBy", revokedBy)
	return &proto_auth.RevokeTokenResponse{
		Revoked: true,
	}, nil
}

func (a *AuthService) RevokeAllTenantTokens(ctx context.Context, req *proto_auth.RevokeAllTenantTokensRequest) (*proto_auth.RevokeAllTenantTokensResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		a.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	// Validate input
	tenantID := req.GetIdentifier().GetTenantId()
	userID := req.GetIdentifier().GetUserId()
	targetTenantID := req.GetTargetTenantId()

	accessCount, refreshCount, err := a.authAPI.RevokeAllTenantTokens(tenantID, userID, targetTenantID)
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
