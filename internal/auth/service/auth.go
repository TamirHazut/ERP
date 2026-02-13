package service

import (
	"context"

	"erp.localhost/internal/auth/api"
	"google.golang.org/protobuf/types/known/timestamppb"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/event/producer"
	"erp.localhost/internal/infra/logging/logger"

	authv1 "erp.localhost/internal/infra/model/auth/v1"
	"erp.localhost/internal/infra/model/event"
	eventv1 "erp.localhost/internal/infra/model/event/v1"
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

func (a *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.Tokens, error) {
	tenantID := req.GetTenantId()
	userPassword := req.GetPassword()
	email := req.GetEmail()
	username := req.GetUsername()

	newTokenResponse, err := a.authAPI.Login(tenantID, email, username, userPassword)
	if err != nil {
		a.logger.Error("failed to authenticate", "error", err.Error())
		// Fire LOGIN_FAILED event
		failedEvent := &eventv1.LoginFailedEvent{
			Email:    email,
			Username: username,
			Reason:   err.Error(),
			Metadata: &eventv1.AuditMetadata{
				OccurredAt: timestamppb.Now(),
			},
		}
		if err := producer.Send(event.AuthUserLogin, tenantID, eventv1.EventType_EVENT_TYPE_LOGIN_FAILED, failedEvent); err != nil {
			a.logger.Error("failed to send event", "error", err)
		}
	}

	a.logger.Debug("login successfuly", "tenantID", tenantID, "email", email, "username", username)
	return newTokenResponse, infra_error.ToGRPCError(err)
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

	message, err := a.authAPI.Logout(tenantID, userID)
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
	token := req.GetToken()
	err := a.authAPI.VerifyToken(token)
	if err != nil {
		a.logger.Error("failed to verify token", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	a.logger.Debug("token verified")
	return &authv1.VerifyTokenResponse{
		Valid: true,
	}, nil
}

func (a *AuthService) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.Tokens, error) {
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
	}
	a.logger.Debug("tokens refreshed successfuly", "tenantID", tenantID, "userID", userID)
	return newTokenResponse, infra_error.ToGRPCError(err)
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

	if err := a.authAPI.RevokeTokens(tenantID, userID); err != nil {
		a.logger.Error("failed to revoke token", "tenantID", tenantID, "userID", userID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	a.logger.Debug("token revoked successfuly", "tenantID", tenantID, "userID", userID)
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
