package client

import (
	"context"
	"time"

	infra_error "erp.localhost/internal/infra/error"
	// model_auth "erp.localhost/internal/infra/model/auth"
	proto_auth "erp.localhost/internal/infra/proto/generated/auth/v1"
	// proto_infra "erp.localhost/internal/infra/proto/generated/infra/v1"
	"erp.localhost/internal/infra/logging/logger"
)

type TokensResponse struct {
	AccessToken        string
	AccessTokenExpiry  time.Time
	RefreshToken       string
	RefreshTokenExpiry time.Time
}

type RevokeResponse struct {
	Revoked              bool
	AccessTokensRevoked  int
	RefrehsTokensRevoked int
}

type AuthClient interface {
	// Authentication
	Authenticate(ctx context.Context, tenantID, userID, userPassword, userHash string) (*TokensResponse, error)
	// Access + Refresh Tokens
	VerifyToken(ctx context.Context, accessToken string) (bool, error)
	RefreshToken(ctx context.Context, tenantID, userID, refreshToken string) (*TokensResponse, error)
	RevokeToken(ctx context.Context, tenantID, userID, accessToken, refreshToken string) (bool, error)
	// Tenant-level token management
	RevokeAllTenantTokens(ctx context.Context, tenantID, userID string) (*RevokeResponse, error)

	Close() error
}

// rbacClient implements RBACClient
type authClient struct {
	grpcClient *GRPCClient
	logger     logger.Logger
	stub       proto_auth.AuthServiceClient
}

func NewAuthGRPCClient(ctx context.Context, config *Config, logger logger.Logger) (AuthClient, error) {
	grpcClient, err := NewGRPCClient(ctx, config, logger)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalGRPCError, err)
	}
	stub := proto_auth.NewAuthServiceClient(grpcClient.Conn())
	return &authClient{
		grpcClient: grpcClient,
		logger:     logger,
		stub:       stub,
	}, nil
}

func (a *authClient) Authenticate(ctx context.Context, tenantID, userID, userPassword, userHash string) (*TokensResponse, error) {
	req := &proto_auth.AuthenticateRequest{}
	res, err := a.stub.Authenticate(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err)
	}
	return &TokensResponse{
		AccessToken:        res.GetTokens().GetAccessToken(),
		AccessTokenExpiry:  time.Unix(res.GetExpiresIn().GetAccessToken(), 0),
		RefreshToken:       res.GetTokens().GetRefreshToken(),
		RefreshTokenExpiry: time.Unix(res.GetExpiresIn().GetRefreshToken(), 0),
	}, nil
}

// Access + Refresh Tokens
func (a *authClient) VerifyToken(ctx context.Context, accessToken string) (bool, error) {
	return false, nil
}
func (a *authClient) RefreshToken(ctx context.Context, tenantID, userID, refreshToken string) (*TokensResponse, error) {
	return nil, nil
}
func (a *authClient) RevokeToken(ctx context.Context, tenantID, userID, accessToken, refreshToken string) (bool, error) {
	return false, nil
}

// Tenant-level token management
func (a *authClient) RevokeAllTenantTokens(ctx context.Context, tenantID, userID string) (*RevokeResponse, error) {
	return nil, nil
}

func (a *authClient) Close() error {
	return a.grpcClient.Close()
}
