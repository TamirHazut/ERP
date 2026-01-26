package service

import (
	"context"

	"erp.localhost/internal/auth/api"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"

	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_infra "erp.localhost/internal/infra/model/infra/validator"
)

type AuthService struct {
	logger  logger.Logger
	authAPI *api.AuthAPI
	authv1.UnimplementedAuthServiceServer
}

func NewAuthService(authAPI *api.AuthAPI, logger logger.Logger) *AuthService {
	return &AuthService{
		logger:  logger,
		authAPI: authAPI,
	}
}

func (a *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.TokensResponse, error) {
	tenantID := req.GetTenantId()
	userPassword := req.GetPassword()
	email := req.GetEmail()
	username := req.GetUsername()

	newTokenResponse, err := a.authAPI.Login(tenantID, email, username, userPassword)
	if err != nil {
		a.logger.Error("failed to authenticate", "error", err.Error())
		return nil, infra_error.ToGRPCError(err)
	}

	return &authv1.TokensResponse{
		Tokens: &authv1.Tokens{
			Token:        newTokenResponse.Token,
			RefreshToken: newTokenResponse.RefreshToken,
		},
		ExpiresIn: &authv1.ExpiresIn{
			Token:        newTokenResponse.TokenExpiresAt,
			RefreshToken: newTokenResponse.RefreshTokenExpiresAt,
		},
	}, nil
}

func (a *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		a.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	tokens := req.GetTokens()

	message, err := a.authAPI.Logout(tenantID, userID, tokens.GetToken(), tokens.GetRefreshToken(), userID)
	if err != nil {
		a.logger.Error("failed to logout", "tenantID", tenantID, "userID", userID, "error", err.Error())
	} else {
		a.logger.Info("logout successful", "tenantID", tenantID, "userID", userID)
	}

	return &authv1.LogoutResponse{
		Message: message,
	}, infra_error.ToGRPCError(err)
}

func (a *AuthService) VerifyToken(ctx context.Context, req *authv1.VerifyTokenRequest) (*authv1.VerifyTokenResponse, error) {
	err := a.authAPI.VerifyToken(req.GetToken())
	if err != nil {
		a.logger.Error("failed to verify token", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	a.logger.Debug("token verified")
	return &authv1.VerifyTokenResponse{
		Valid: true,
	}, nil
}

func (a *AuthService) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.TokensResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
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
	return &authv1.TokensResponse{
		Tokens: &authv1.Tokens{
			Token:        newTokenResponse.Token,
			RefreshToken: newTokenResponse.RefreshToken,
		},
		ExpiresIn: &authv1.ExpiresIn{
			Token:        newTokenResponse.TokenExpiresAt,
			RefreshToken: newTokenResponse.RefreshTokenExpiresAt,
		},
	}, nil
}

func (a *AuthService) RevokeToken(ctx context.Context, req *authv1.RevokeTokenRequest) (*authv1.RevokeTokenResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		a.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := req.GetIdentifier().GetTenantId()
	userID := req.GetIdentifier().GetUserId()
	revokedBy := req.GetRevokedBy()
	token := req.GetTokens().GetToken()
	refreshToken := req.GetTokens().GetRefreshToken()

	if err := a.authAPI.RevokeTokens(tenantID, userID, token, refreshToken, revokedBy); err != nil {
		a.logger.Error("failed to revoke token", "tenantID", tenantID, "userID", userID, "revokedBy", revokedBy, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	a.logger.Debug("token revoked successfuly", "tenantID", tenantID, "userID", userID, "revokedBy", revokedBy)
	return &authv1.RevokeTokenResponse{
		Revoked: true,
	}, nil
}

func (a *AuthService) RevokeAllTenantTokens(ctx context.Context, req *authv1.RevokeAllTenantTokensRequest) (*authv1.RevokeAllTenantTokensResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
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

	return &authv1.RevokeAllTenantTokensResponse{
		Revoked:              true,
		AccessTokensRevoked:  int32(accessCount),
		RefreshTokensRevoked: int32(refreshCount),
	}, nil
}
