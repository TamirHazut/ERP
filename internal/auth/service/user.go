package service

import (
	"context"
	"errors"

	"erp.localhost/internal/auth/api"
	collection "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/infra/convertor"
	mongo_collection "erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/generated/auth/v1"
	proto_infra "erp.localhost/internal/infra/proto/generated/infra/v1"
	"erp.localhost/internal/infra/proto/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type filterType int

const (
	filterTypeUnsupported filterType = iota
	filterTypeID
	filterTypeEmail
	filterTypeUsername
)

type UserService struct {
	logger         logger.Logger
	userCollection *collection.UserCollection
	authAPI        *api.AuthAPI
	rbacAPI        *api.RBACAPI
	proto_auth.UnimplementedUserServiceServer
}

// TODO: allow system users with proper permissions to query
func NewUserService(authAPI *api.AuthAPI, rbacAPI *api.RBACAPI) *UserService {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)

	userHandler := mongo_collection.NewBaseCollectionHandler[model_auth.User](model_mongo.AuthDB, model_mongo.UsersCollection, logger)
	userCollection := collection.NewUserCollection(userHandler, logger)
	if userCollection == nil {
		logger.Fatal("failed to create users collection")
		return nil
	}

	return &UserService{
		logger:         logger,
		userCollection: userCollection,
		authAPI:        authAPI,
		rbacAPI:        rbacAPI,
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
		errMsg := infra_error.Validation(infra_error.ValidationRequiredFields, "user").Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	if newUserData.GetTenantId() != tenantID {
		errMsg := infra_error.Auth(infra_error.AuthPermissionDenied).Error()
		u.logger.Error("user data tenant_id does not match identifier tenant_id", "error", errMsg)
		return nil, status.Error(codes.PermissionDenied, errMsg)
	}

	if err := u.hasPermission(identifier, model_auth.PermissionActionRead, tenantID); err != nil {
		errMsg := infra_error.Auth(infra_error.AuthPermissionDenied).WithError(err).Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.PermissionDenied, errMsg)
	}

	// verify the new user tenantID is the same as the one of the requesting user
	if newUserData.GetTenantId() != tenantID {
		errMsg := infra_error.Auth(infra_error.AuthPermissionDenied).Error()
		u.logger.Error("user data tenant_id does not match identifier tenant_id", "error", errMsg)
		return nil, status.Error(codes.PermissionDenied, errMsg)
	}

	user, _ := u.getUser(tenantID, newUserData.GetEmail(), filterTypeEmail)
	if user != nil {
		err := infra_error.Validation(infra_error.ConflictDuplicateEmail)
		u.logger.Error("failed to create new account", "tenantID", tenantID, "error", err.Error())
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	// convert from proto user to model user
	newUser, err := convertor.CreateUserFromProto(newUserData)
	if err != nil {
		err := infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("failed to create user"))
		u.logger.Error(err.WithError(errors.New("failed to convert user to proto user")).Error())
		return nil, status.Error(codes.Aborted, err.Error())
	}
	id, err := u.userCollection.CreateUser(newUser)
	if err != nil {
		u.logger.Error("failed to create user", "tenant_id", tenantID, "error", err.Error())
		return nil, status.Error(codes.Aborted, err.Error())
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
	if accountID == "" {
		errMsg := infra_error.Validation(infra_error.ValidationRequiredFields, "id").Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()

	if err := u.hasPermission(identifier, model_auth.PermissionActionRead, tenantID); err != nil {
		errMsg := infra_error.Auth(infra_error.AuthPermissionDenied).WithError(err).Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.PermissionDenied, errMsg)
	}

	// get user
	user, err := u.getUser(tenantID, accountID, filterTypeID)
	if err != nil {
		errMsg := err.Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.Aborted, errMsg)
	}

	// convertor from model user to proto user
	userProto := convertor.UserToProto(user)
	if userProto == nil {
		err := infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("Failed to get user"))
		u.logger.Error(err.WithError(errors.New("failed to convert user to proto user")).Error())
		return nil, status.Error(codes.Aborted, err.Error())
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

	if err := u.hasPermission(identifier, model_auth.PermissionActionRead, tenantID); err != nil {
		errMsg := infra_error.Auth(infra_error.AuthPermissionDenied).WithError(err).Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.PermissionDenied, errMsg)
	}

	var users []*model_auth.User
	var err error
	roleID := req.GetRoleId()
	if roleID != "" {
		users, err = u.userCollection.GetUsersByRoleID(tenantID, roleID)
	} else {
		users, err = u.userCollection.GetUsersByTenantID(tenantID)
	}

	if err != nil {
		errMsg := err.Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.Internal, errMsg)
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

	newUserData := req.GetUser()
	if err := validator.ValidateUpdateUserData(newUserData); err != nil {
		errMsg := infra_error.Validation(infra_error.ValidationRequiredFields, "user").Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	// verify the new user tenantID is the same as the one of the requesting user
	if newUserData.GetTenantId() != tenantID {
		errMsg := infra_error.Auth(infra_error.AuthPermissionDenied).Error()
		u.logger.Error("user data tenant_id does not match identifier tenant_id", "error", errMsg)
		return nil, status.Error(codes.PermissionDenied, errMsg)
	}

	if err := u.hasPermission(identifier, model_auth.PermissionActionUpdate, tenantID); err != nil {
		errMsg := infra_error.Auth(infra_error.AuthPermissionDenied).WithError(err).Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.PermissionDenied, errMsg)
	}

	user, err := u.getUser(tenantID, newUserData.GetEmail(), filterTypeEmail)
	if err != nil {
		errMsg := infra_error.Internal(infra_error.InternalDatabaseError, err).Error()
		u.logger.Error("failed to update account", "tenantID", tenantID, "error", errMsg)
		return nil, status.Error(codes.Aborted, errMsg)
	}

	err = convertor.UpdateUserFromProto(user, newUserData)
	if err != nil {
		errMsg := infra_error.Internal(infra_error.InternalUnexpectedError, err).Error()
		u.logger.Error("failed to update account", "tenantID", tenantID, "error", errMsg)
		return nil, status.Error(codes.Aborted, errMsg)
	}
	// Add logic to verify only non important fields are updated
	err = u.userCollection.UpdateUser(user)
	if err != nil {
		errMsg := err.Error()
		u.logger.Error("failed to update account", "tenantID", tenantID, "error", errMsg)
		err = status.Error(codes.Aborted, errMsg)
	}
	return &proto_auth.UpdateUserResponse{
		Updated: err != nil,
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
	accountID := req.GetAccountId()
	if accountID == "" {
		tenantID = req.GetTenantId()
	}

	if err := u.hasPermission(identifier, model_auth.PermissionActionDelete, tenantID); err != nil {
		errMsg := infra_error.Auth(infra_error.AuthPermissionDenied).WithError(err).Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.PermissionDenied, errMsg)
	}

	err := u.userCollection.DeleteUser(tenantID, accountID)
	if err != nil {
		errMsg := err.Error()
		u.logger.Error("failed to delete account", "tenantID", tenantID, "error", errMsg)
		err = status.Error(codes.Aborted, errMsg)
	}
	return &proto_auth.DeleteUserResponse{
		Deleted: err == nil,
	}, err
}

func (u *UserService) Login(ctx context.Context, req *proto_auth.LoginRequest) (*proto_auth.TokensResponse, error) {
	tenantID := req.GetTenantId()
	if tenantID == "" {
		errMsg := infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id").Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}
	userPassword := req.GetPassword()
	if userPassword == "" {
		errMsg := infra_error.Validation(infra_error.ValidationRequiredFields, "password").Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}
	accountID := req.GetAccountId()
	if accountID == nil {
		errMsg := infra_error.Validation(infra_error.ValidationRequiredFields, "account_id").Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	var userIdentifier string
	var filterType filterType

	switch v := accountID.(type) {
	case *proto_auth.LoginRequest_Email:
		userIdentifier = v.Email
		filterType = filterTypeEmail
	case *proto_auth.LoginRequest_Username:
		userIdentifier = v.Username
		filterType = filterTypeUsername
	default:
		errMsg := infra_error.Validation(infra_error.ValidationRequiredFields, "account_id").Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}
	user, err := u.getUser(tenantID, userIdentifier, filterType)
	if err != nil {
		errMsg := err.Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	newTokenResponse, err := u.authAPI.Authenticate(tenantID, user.ID.Hex(), userPassword, user.PasswordHash)
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

	err := u.authAPI.RevokeTokens(tenantID, userID, tokens.GetAccessToken(), tokens.GetRefreshToken(), userID)

	var message string
	if err != nil {
		u.logger.Error("failed to logout", "tenantID", tenantID, "userID", userID, "error", err.Error())
		message = "logout failed"
	} else {
		u.logger.Info("logout successful", "tenantID", tenantID, "userID", userID)
		message = "logout successful"
	}
	return &proto_auth.LogoutResponse{
		Message: message,
	}, infra_error.ToGRPCError(err)
}

// func (u *UserService) authenticate(ctx context.Context, identifier *proto_infra.UserIdentifier, userPassword string, userHash string) (*proto_auth.TokensResponse, error) {
// 	req := &proto_auth.AuthenticateRequest{
// 		Identifier:   identifier,
// 		UserPassword: userPassword,
// 		UserHash:     userHash,
// 	}
// 	return u.authAPI.Authenticate(ctx, req)
// }

func (u *UserService) hasPermission(identifier *proto_infra.UserIdentifier, permissionAction string, targetTenant string) error {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeUser, permissionAction)
	if err != nil {
		return err
	}
	return u.rbacAPI.Verification.HasPermission(identifier.GetTenantId(), identifier.GetUserId(), permission, targetTenant)
}

func (u *UserService) getUser(tenantID string, accountID string, filterType filterType) (*model_auth.User, error) {
	switch filterType {
	case filterTypeID:
		return u.userCollection.GetUserByID(tenantID, accountID)
	case filterTypeEmail:
		return u.userCollection.GetUserByEmail(tenantID, accountID)
	case filterTypeUsername:
		return u.userCollection.GetUserByUsername(tenantID, accountID)
	default:
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "account identifier")
	}
}
