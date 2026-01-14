package service

import (
	"context"
	"errors"

	"erp.localhost/internal/auth/api"
	"erp.localhost/internal/infra/convertor"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/generated/auth/v1"
	"erp.localhost/internal/infra/proto/validator"
)

type UserService struct {
	logger  logger.Logger
	userAPI *api.UserAPI
	proto_auth.UnimplementedUserServiceServer
}

// TODO: allow system users with proper permissions to query
func NewUserService(userAPI *api.UserAPI) *UserService {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)
	return &UserService{
		logger:  logger,
		userAPI: userAPI,
	}
}

func (u *UserService) CreateUser(ctx context.Context, req *proto_auth.CreateUserRequest) (*proto_auth.CreateUserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		u.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()

	newUserData := req.GetUser()
	if err := validator.ValidateCreateUserData(newUserData); err != nil {
		err := infra_error.Validation(infra_error.ValidationRequiredFields, "user")
		u.logger.Error("failed to create user", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	newUser, err := convertor.CreateUserFromProto(newUserData)
	if err != nil {
		err := infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("failed to create user"))
		u.logger.Error("failed to convert proto model to system model", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// convert from proto user to model user
	id, err := u.userAPI.CreateUser(tenantID, identifier.GetUserId(), newUser)
	if err != nil {
		u.logger.Error("failed to create user", "tenant_id", tenantID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &proto_auth.CreateUserResponse{
		UserId: id,
	}, nil
}

func (u *UserService) GetUser(ctx context.Context, req *proto_auth.GetUserRequest) (*proto_auth.UserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		u.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	accountID := req.GetAccountId()
	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	targetTenantID := req.GetTargetTenantId()

	// get user
	user, err := u.userAPI.GetUser(tenantID, userID, targetTenantID, accountID)
	if err != nil {
		u.logger.Error("failed to get user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// convertor from model user to proto user
	userProto := convertor.UserToProto(user)
	if userProto == nil {
		err := infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("Failed to get user"))
		u.logger.Error("failed to convert proto model to system model", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &proto_auth.UserResponse{
		User: userProto,
	}, nil
}

func (u *UserService) GetUsers(ctx context.Context, req *proto_auth.GetUsersRequest) (*proto_auth.UsersResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		u.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	targetTenantID := req.GetTargetTenantId()

	users, err := u.userAPI.GetUsers(tenantID, userID, targetTenantID, req.GetRoleId())
	if err != nil {
		u.logger.Error("failed to get users", "tenant_id", tenantID, "user_id", userID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// convertor from model user to proto user
	usersProto := convertor.UsersToProto(users)

	return &proto_auth.UsersResponse{
		Users: usersProto,
	}, nil
}

func (u *UserService) UpdateUser(ctx context.Context, req *proto_auth.UpdateUserRequest) (*proto_auth.UpdateUserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		u.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()

	newUserData := req.GetUser()
	if err := validator.ValidateUpdateUserData(newUserData); err != nil {
		u.logger.Error("failed to update account", "tenantID", tenantID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	newUser, err := convertor.UserFromUpdateProto(newUserData)
	if err != nil {
		u.logger.Error("failed to update account", "tenantID", tenantID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// Add logic to verify only non important fields are updated
	res, err := u.userAPI.UpdateUser(tenantID, userID, newUser)
	if err != nil {
		u.logger.Error("failed to update account", "tenantID", tenantID, "error", err)
		err = infra_error.ToGRPCError(err)
	}
	return &proto_auth.UpdateUserResponse{
		Updated: res,
	}, err
}

func (u *UserService) DeleteUser(ctx context.Context, req *proto_auth.DeleteUserRequest) (*proto_auth.DeleteUserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		u.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	targetTenantID := req.GetTargetTenantId()
	accountID := req.GetAccountId()

	deleted, err := u.userAPI.DeleteUser(tenantID, userID, targetTenantID, accountID)
	if err != nil {
		u.logger.Error("failed to delete account", "tenantID", tenantID, "error", err)
		err = infra_error.ToGRPCError(err)
	}
	return &proto_auth.DeleteUserResponse{
		Deleted: deleted,
	}, err
}

func (u *UserService) Login(ctx context.Context, req *proto_auth.LoginRequest) (*proto_auth.TokensResponse, error) {
	tenantID := req.GetTenantId()
	userPassword := req.GetPassword()

	newTokenResponse, err := u.userAPI.Login(tenantID, req.GetEmail(), req.GetUsername(), userPassword)
	if err != nil {
		u.logger.Error("failed to authenticate", "error", err.Error())
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

func (u *UserService) Logout(ctx context.Context, req *proto_auth.LogoutRequest) (*proto_auth.LogoutResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		u.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	tokens := req.GetTokens()

	message, err := u.userAPI.Logout(tenantID, userID, tokens.GetAccessToken(), tokens.GetRefreshToken(), userID)
	if err != nil {
		u.logger.Error("failed to logout", "tenantID", tenantID, "userID", userID, "error", err.Error())
	} else {
		u.logger.Info("logout successful", "tenantID", tenantID, "userID", userID)
	}

	return &proto_auth.LogoutResponse{
		Message: message,
	}, infra_error.ToGRPCError(err)
}
