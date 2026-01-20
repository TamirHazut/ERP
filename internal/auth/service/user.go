package service

import (
	"context"

	"erp.localhost/internal/auth/api"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_infra "erp.localhost/internal/infra/model/infra/validator"
)

type UserService struct {
	logger  logger.Logger
	userAPI *api.UserAPI
	authv1.UnimplementedUserServiceServer
}

func NewUserService(userAPI *api.UserAPI, logger logger.Logger) *UserService {
	return &UserService{
		logger:  logger,
		userAPI: userAPI,
	}
}

func (u *UserService) CreateUser(ctx context.Context, req *authv1.CreateUserRequest) (*authv1.CreateUserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		u.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	newUser := req.GetUser()

	// convert from proto user to model user
	id, err := u.userAPI.CreateUser(tenantID, identifier.GetUserId(), newUser)
	if err != nil {
		u.logger.Error("failed to create user", "tenant_id", tenantID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &authv1.CreateUserResponse{
		UserId: id,
	}, nil
}

func (u *UserService) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.User, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
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

	return user, nil
}

func (u *UserService) ListUsers(ctx context.Context, req *authv1.ListUsersRequest) (*authv1.ListUsersResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
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

	return &authv1.ListUsersResponse{
		Users: users,
	}, nil
}

func (u *UserService) UpdateUser(ctx context.Context, req *authv1.UpdateUserRequest) (*authv1.UpdateUserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		u.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()

	newUser := req.GetUser()

	// Add logic to verify only non important fields are updated
	res, err := u.userAPI.UpdateUser(tenantID, userID, newUser)
	if err != nil {
		u.logger.Error("failed to update account", "tenantID", tenantID, "error", err)
		err = infra_error.ToGRPCError(err)
	}
	return &authv1.UpdateUserResponse{
		Updated: res,
	}, err
}

func (u *UserService) DeleteUser(ctx context.Context, req *authv1.DeleteUserRequest) (*authv1.DeleteUserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		u.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	targetTenantID := req.GetTargetTenantId()
	accountID := req.GetAccountId()

	err := u.userAPI.DeleteUser(tenantID, userID, targetTenantID, accountID)
	if err != nil {
		u.logger.Error("failed to delete account", "tenantID", tenantID, "error", err)
		err = infra_error.ToGRPCError(err)
	}
	return &authv1.DeleteUserResponse{
		Deleted: err == nil,
	}, err
}
