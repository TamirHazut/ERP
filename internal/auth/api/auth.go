package api

import (
	"errors"
	"time"

	"erp.localhost/internal/auth/hash"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	authv1_cache "erp.localhost/internal/infra/model/auth/v1/cache"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthAPI struct {
	logger       logger.Logger
	rbacAPI      *RBACAPI
	userAPI      *UserAPI
	tokenManager *TokenAPI
}

func NewAuthAPI(rbacAPI *RBACAPI, userAPI *UserAPI, logger logger.Logger) (*AuthAPI, error) {

	tokenManager, err := NewTokenAPI(logger)
	if err != nil {
		logger.Fatal("failed to create token manager", "error", err)
		return nil, err
	}
	return &AuthAPI{
		logger:       logger,
		rbacAPI:      rbacAPI,
		userAPI:      userAPI,
		tokenManager: tokenManager,
	}, nil
}

func (a *AuthAPI) Login(tenantID, email, username, password string) (*NewTokenResponse, error) {
	if tenantID == "" || password == "" || (email == "" && username == "") {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, email/username, password"))
		a.logger.Error("failed to login", "error", err)
		return nil, err
	}

	var filterType FilterType
	if email != "" {
		filterType = filterTypeEmail
	} else if username != "" {
		filterType = filterTypeUsername
	} else {
		filterType = filterTypeUnsupported
	}
	user, err := a.userAPI.getUser(tenantID, email, filterType)
	if err != nil {
		a.logger.Error("failed to find user", "error", err)
		return nil, err
	}

	tokens, err := a.Authenticate(user, password)
	if user.LoginHistory == nil {
		user.LoginHistory = make([]*authv1.LoginRecord, 0)
	}
	user.LoginHistory = append(user.LoginHistory, &authv1.LoginRecord{
		Timestamp: timestamppb.Now(),
		Success:   tokens != nil,
	})
	if updateErr := a.userAPI.userHandler.UpdateUser(user); updateErr != nil {
		a.logger.Error("failed to update user login history", "error", err)
	}
	return tokens, err
}

func (a *AuthAPI) Logout(tenantID, userID, accessToken, refreshToken, revokedBy string) (string, error) {
	err := a.RevokeTokens(tenantID, userID, accessToken, refreshToken, revokedBy)
	if err != nil {
		return "logout failed", err
	}
	return "logout successful", err
}

func (a *AuthAPI) Authenticate(user *authv1.User, password string) (*NewTokenResponse, error) {
	if password == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, user_password, user_hash"))
		a.logger.Error("Failed to authenticate user", "error", err)
		return nil, err
	}

	if !hash.VerifyHash(password, user.GetPasswordHash()) {
		return nil, infra_error.Auth(infra_error.AuthInvalidCredentials)
	}

	// Generate tokens
	return a.generateAndStoreTokens(user)
}

func (a *AuthAPI) VerifyToken(token string) error {
	if token == "" {
		return status.Error(codes.InvalidArgument, infra_error.Validation(infra_error.ValidationRequiredFields, "access_token").Error())
	}
	_, err := a.tokenManager.VerifyAccessToken(token)
	return err
}

func (a *AuthAPI) RefreshToken(tenantID, userID, token string) (*NewTokenResponse, error) {
	if tenantID == "" || userID == "" || token == "" {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, refresh_token"))
	}

	// Verify the refresh token is valid
	_, err := a.tokenManager.VerifyRefreshToken(tenantID, userID, token)
	if err != nil {
		a.logger.Error("Failed to verify refresh token", "error", err, "tenant_id", tenantID, "user_id", userID, "refresh_token", token)
		return nil, err
	}

	// Revoke old access tokens to prevent orphaned tokens
	// Note: We only revoke access tokens, not refresh tokens, since the refresh token
	// is still valid and will be revoked explicitly below
	if err := a.tokenManager.RevokeAllAccessTokens(tenantID, userID, "system"); err != nil {
		a.logger.Warn("Failed to revoke old access tokens before refresh", "error", err, "tenant_id", tenantID, "user_id", userID)
		// Continue anyway - non-critical failure
	}
	user, err := a.userAPI.getUser(tenantID, userID, filterTypeID)
	if err != nil {
		a.logger.Error("failed to find user", "error", err)
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}

	newTokenResponse, err := a.generateAndStoreTokens(user)
	if err != nil {
		a.logger.Error("Failed to generate and store tokens", "error", err, "tenant_id", tenantID, "user_id", userID)
		return nil, err
	}

	// Revoke the old refresh token after successfully creating new tokens
	err = a.tokenManager.RevokeRefreshToken(tenantID, userID, token, "system", true)
	if err != nil {
		a.logger.Error("Failed to revoke old refresh token", "error", err, "tenant_id", tenantID, "user_id", userID)
		return nil, err
	}
	return newTokenResponse, nil
}

func (a *AuthAPI) RevokeTokens(tenantID, userID, accessToken, refreshToken, revokedBy string) error {
	if tenantID == "" || userID == "" || accessToken == "" || refreshToken == "" || revokedBy == "" {
		return infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, access_token, refresh_token, revoked_by"))
	}

	if accessToken != "" {
		err := a.tokenManager.RevokeAccessToken(accessToken, revokedBy)
		if err != nil {
			return err
		}
	}
	if refreshToken != "" {
		err := a.tokenManager.RevokeRefreshToken(tenantID, userID, refreshToken, revokedBy, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AuthAPI) RevokeAllTenantTokens(tenantID, revokedBy, targetTenantID string) (int, int, error) {
	if tenantID == "" || revokedBy == "" || targetTenantID == "" {
		return 0, 0, infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id"))
	}

	a.logger.Warn("Revoking all tenant tokens", "tenant_id", targetTenantID, "revoked_by", revokedBy)

	// This is a critical operation that should require elevated permissions
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeToken, model_auth.PermissionActionDelete)
	if err != nil {
		return 0, 0, err
	}
	err = a.rbacAPI.Verification.HasPermission(tenantID, revokedBy, permission, targetTenantID)
	if err != nil {
		return 0, 0, err
	}

	// Revoke all tokens for this tenant
	return a.tokenManager.RevokeAllTenantTokens(targetTenantID, revokedBy)
}

func (a *AuthAPI) generateAccessToken(user *authv1.User) (string, *authv1_cache.TokenMetadata, error) {
	// Generate access token
	userRoles := make([]string, len(user.GetRoles()))
	for i, role := range user.GetRoles() {
		userRoles[i] = role.RoleId
	}
	accessToken, claims, err := a.tokenManager.GenerateAccessToken(&GenerateAccessTokenInput{
		UserId:   user.GetId(),
		TenantId: user.GetTenantId(),
		Username: user.GetUsername(),
		Email:    user.GetEmail(),
		Roles:    userRoles,
	})
	if err != nil {
		return "", nil, status.Error(codes.Internal, err.Error())
	}

	accessTokenMetadata := &authv1_cache.TokenMetadata{
		Jti:       accessToken,
		UserId:    claims.GetUserId(),
		TenantId:  claims.GetTenantId(),
		IssuedAt:  claims.GetIssuedAt(),
		ExpiresAt: claims.GetExpiresAt(),
		Revoked:   false,
		RevokedAt: nil,
		RevokedBy: "",
		IpAddress: "",
		UserAgent: "",
		Scopes:    []string{},
	}

	return accessToken, accessTokenMetadata, nil
}

func (a *AuthAPI) generateRefreshToken(tenantID string, userID string) (string, *authv1_cache.RefreshToken, error) {
	issuedAt := time.Now()
	// Generate refresh token
	tokenString, refreshToken, err := a.tokenManager.GenerateRefreshToken(GenerateRefreshTokenInput{
		UserId:    userID,
		TenantId:  tenantID,
		CreatedAt: issuedAt,
	})
	if err != nil {
		return "", nil, status.Error(codes.Internal, err.Error())
	}
	return tokenString, refreshToken, nil
}

func (a *AuthAPI) generateAndStoreTokens(user *authv1.User) (*NewTokenResponse, error) {
	accessToken, accessTokenMetadata, err := a.generateAccessToken(user)
	if err != nil {
		return nil, err
	}
	refreshTokenString, refreshTokenModel, err := a.generateRefreshToken(user.GetTenantId(), user.GetId())
	if err != nil {
		return nil, err
	}

	// Store tokens (single token per user - automatically replaces existing)
	err = a.tokenManager.StoreTokens(user.GetTenantId(), user.GetId(), accessTokenMetadata, refreshTokenModel)
	if err != nil {
		return nil, err
	}

	return &NewTokenResponse{
		UserId:                user.GetId(),
		TenantId:              user.GetTenantId(),
		Token:                 accessToken,
		TokenExpiresAt:        accessTokenMetadata.ExpiresAt.AsTime().Unix(),
		RefreshToken:          refreshTokenString,
		RefreshTokenExpiresAt: refreshTokenModel.ExpiresAt.AsTime().Unix(),
	}, nil
}
