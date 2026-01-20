package client

import (
	"context"
	"time"

	infra_error "erp.localhost/internal/infra/error"
	// authv1 "erp.localhost/internal/infra/model/auth/v1"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	infrav1 "erp.localhost/internal/infra/model/infra/v1"

	// proto_infra "erp.localhost/internal/infra/proto/generated/infra/v1"
	"erp.localhost/internal/infra/logging/logger"
)

type TokensResponse struct {
	Token              string
	TokenExpiry        time.Time
	RefreshToken       string
	RefreshTokenExpiry time.Time
}

type RevokeResponse struct {
	Revoked              bool
	AccessTokensRevoked  int32
	RefrehsTokensRevoked int32
}

type AuthClient interface {
	// Authentication - Login + Logout
	Login(ctx context.Context, tenantID, email, username, password string) (*TokensResponse, error)
	Logout(ctx context.Context, tenantID, userID, accessToken, refreshToken string) (string, error)
	// Access + Refresh Tokens
	VerifyToken(ctx context.Context, accessToken string) (bool, error)
	RefreshToken(ctx context.Context, tenantID, userID, refreshToken string) (*TokensResponse, error)
	RevokeToken(ctx context.Context, tenantID, userID, accessToken, refreshToken string) (bool, error)
	// Tenant-level token management
	RevokeAllTenantTokens(ctx context.Context, tenantID, userID, targetTenantID string) (*RevokeResponse, error)

	Close() error
}

// rbacClient implements RBACClient
type authClient struct {
	grpcClient *GRPCClient
	logger     logger.Logger
	stub       authv1.AuthServiceClient
}

func NewAuthGRPCClient(ctx context.Context, config *Config, logger logger.Logger) (AuthClient, error) {
	grpcClient, err := NewGRPCClient(ctx, config, logger)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalGRPCError, err)
	}
	stub := authv1.NewAuthServiceClient(grpcClient.Conn())
	return &authClient{
		grpcClient: grpcClient,
		logger:     logger,
		stub:       stub,
	}, nil
}

func (a *authClient) Login(ctx context.Context, tenantID, email, username, password string) (*TokensResponse, error) {
	req := &authv1.LoginRequest{
		TenantId: tenantID,
		Password: password,
	}
	if email != "" {
		req.AccountId = &authv1.LoginRequest_Email{
			Email: email,
		}
	} else if username != "" {
		req.AccountId = &authv1.LoginRequest_Username{
			Username: username,
		}
	} else {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "Email", "Username")
	}
	res, err := a.stub.Login(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err)
	}
	return &TokensResponse{
		Token:              res.GetTokens().GetToken(),
		TokenExpiry:        time.Unix(res.GetExpiresIn().GetToken(), 0),
		RefreshToken:       res.GetTokens().GetRefreshToken(),
		RefreshTokenExpiry: time.Unix(res.GetExpiresIn().GetRefreshToken(), 0),
	}, nil
}

func (a *authClient) Logout(ctx context.Context, tenantID, userID, accessToken, refreshToken string) (string, error) {
	req := &authv1.LogoutRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		Tokens: &authv1.Tokens{
			Token:        accessToken,
			RefreshToken: refreshToken,
		},
	}
	res, err := a.stub.Logout(ctx, req)
	if err != nil {
		return "", mapGRPCError(err)
	}
	return res.GetMessage(), nil
}

func (a *authClient) VerifyToken(ctx context.Context, accessToken string) (bool, error) {
	req := &authv1.VerifyTokenRequest{
		Token: accessToken,
	}
	res, err := a.stub.VerifyToken(ctx, req)
	if err != nil {
		return false, mapGRPCError(err)
	}
	return res.GetValid(), nil
}

func (a *authClient) RefreshToken(ctx context.Context, tenantID, userID, refreshToken string) (*TokensResponse, error) {
	req := &authv1.RefreshTokenRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		RefreshToken: refreshToken,
	}
	res, err := a.stub.RefreshToken(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err)
	}
	return &TokensResponse{
		Token:              res.GetTokens().GetToken(),
		TokenExpiry:        time.Unix(res.GetExpiresIn().GetToken(), 0),
		RefreshToken:       res.GetTokens().GetRefreshToken(),
		RefreshTokenExpiry: time.Unix(res.GetExpiresIn().GetRefreshToken(), 0),
	}, nil
}

func (a *authClient) RevokeToken(ctx context.Context, tenantID, userID, accessToken, refreshToken string) (bool, error) {
	req := &authv1.RevokeTokenRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		Tokens: &authv1.Tokens{
			Token:        accessToken,
			RefreshToken: refreshToken,
		},
	}
	res, err := a.stub.RevokeToken(ctx, req)
	if err != nil {
		return false, mapGRPCError(err)
	}
	return res.GetRevoked(), nil
}

func (a *authClient) RevokeAllTenantTokens(ctx context.Context, tenantID, userID, targetTenantID string) (*RevokeResponse, error) {
	req := &authv1.RevokeAllTenantTokensRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		TargetTenantId: targetTenantID,
	}
	res, err := a.stub.RevokeAllTenantTokens(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err)
	}
	return &RevokeResponse{
		Revoked:              res.GetRevoked(),
		AccessTokensRevoked:  res.GetAccessTokensRevoked(),
		RefrehsTokensRevoked: res.GetRefreshTokensRevoked(),
	}, nil
}

func (a *authClient) Close() error {
	return a.grpcClient.Close()
}
