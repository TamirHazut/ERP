package service

import (
	"context"
	"errors"

	collection "erp.localhost/internal/core/collection"
	"erp.localhost/internal/infra/convertor"
	mongo_collection "erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_core "erp.localhost/internal/infra/model/core"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
	proto_core "erp.localhost/internal/infra/proto/core/v1"
	proto_infra "erp.localhost/internal/infra/proto/infra/v1"
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
	rbacGRPCClient proto_auth.RBACServiceClient
	authGRPCClient proto_auth.AuthServiceClient
	proto_core.UnimplementedUserServiceServer
}

// TODO: allow system admin create user in other tenants
func NewUserService() *UserService {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)

	uc := mongo_collection.NewBaseCollectionHandler[model_core.User](model_mongo.CoreDB, model_mongo.UsersCollection, logger)
	if uc == nil {
		logger.Fatal("failed to create users collection handler")
		return nil
	}

	userCollection := collection.NewUserCollection(uc, logger)
	if userCollection == nil {
		logger.Fatal("failed to create users collection")
		return nil
	}

	return &UserService{
		logger:         logger,
		userCollection: userCollection,
	}
}

func (u *UserService) CreateUser(ctx context.Context, req *proto_core.CreateUserRequest) (*proto_core.CreateUserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()

	newUserData := req.GetUser()
	if err := validator.ValidateCreateUserData(newUserData); err != nil {
		errMsg := infra_error.Validation(infra_error.ValidationRequiredFields, "user").Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	if hasPermission, err := u.verifyPermission(ctx, identifier, model_auth.PermissionActionCreate); err != nil || !hasPermission {
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

	return &proto_core.CreateUserResponse{
		UserId: id,
	}, nil
}

func (u *UserService) GetUser(ctx context.Context, req *proto_core.GetUserRequest) (*proto_core.UserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}
	accountID := req.GetAccountId()
	if accountID == "" {
		errMsg := infra_error.Validation(infra_error.ValidationRequiredFields, "id").Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()

	if hasPermission, err := u.verifyPermission(ctx, identifier, model_auth.PermissionActionRead); err != nil || !hasPermission {
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

	return &proto_core.UserResponse{
		User: userProto,
	}, nil
}

func (u *UserService) GetUsers(ctx context.Context, req *proto_core.GetUsersRequest) (*proto_core.UsersResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()

	if hasPermission, err := u.verifyPermission(ctx, identifier, model_auth.PermissionActionRead); err != nil || !hasPermission {
		errMsg := infra_error.Auth(infra_error.AuthPermissionDenied).WithError(err).Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.PermissionDenied, errMsg)
	}

	var users []*model_core.User
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

	return &proto_core.UsersResponse{
		Users: usersProto,
	}, nil
}

func (u *UserService) UpdateUser(ctx context.Context, req *proto_core.UpdateUserRequest) (*proto_core.UpdateUserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
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

	if hasPermission, err := u.verifyPermission(ctx, identifier, model_auth.PermissionActionUpdate); err != nil || !hasPermission {
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
	return &proto_core.UpdateUserResponse{
		Updated: err != nil,
	}, err
}

func (u *UserService) DeleteUser(ctx context.Context, req *proto_core.DeleteUserRequest) (*proto_core.DeleteUserResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()

	accountID := req.GetAccountId()
	if accountID == "" {
		errMsg := infra_error.Validation(infra_error.ValidationRequiredFields, "id").Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	if hasPermission, err := u.verifyPermission(ctx, identifier, model_auth.PermissionActionDelete); err != nil || !hasPermission {
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
	return &proto_core.DeleteUserResponse{
		Deleted: err == nil,
	}, err
}

func (u *UserService) Login(ctx context.Context, req *proto_core.LoginRequest) (*proto_auth.TokensResponse, error) {
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
	case *proto_core.LoginRequest_Email:
		userIdentifier = v.Email
		filterType = filterTypeEmail
	case *proto_core.LoginRequest_Username:
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

	identifier := &proto_infra.UserIdentifier{
		TenantId: tenantID,
		UserId:   user.ID.Hex(),
	}
	res, err := u.authenticate(ctx, identifier, userPassword, user.PasswordHash)
	if err != nil {
		u.logger.Error("failed to authenticate", "error", err.Error())
	}
	return res, err
}

func (u *UserService) Logout(ctx context.Context, req *proto_core.LogoutRequest) (*proto_core.LogoutResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		u.logger.Error(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	tokens := req.GetTokens()

	revoked, err := u.revokeTokens(ctx, identifier, tokens, userID)
	if err != nil {
		u.logger.Error("failed to logout", "tenantID", tenantID, "userID", userID, "error", err.Error())
	}

	var message string
	if !revoked {
		message = "logout failed"
	} else {
		message = "logout successful"
	}
	return &proto_core.LogoutResponse{
		Message: message,
	}, err
}

func (u *UserService) authenticate(ctx context.Context, identifier *proto_infra.UserIdentifier, userPassword string, userHash string) (*proto_auth.TokensResponse, error) {
	req := &proto_auth.AuthenticateRequest{
		Identifier:   identifier,
		UserPassword: userPassword,
		UserHash:     userHash,
	}
	return u.authGRPCClient.Authenticate(ctx, req)
}

func (u *UserService) revokeTokens(ctx context.Context, identifier *proto_infra.UserIdentifier, tokens *proto_auth.Tokens, revokedBy string) (bool, error) {
	req := &proto_auth.RevokeTokenRequest{
		Identifier: identifier,
		Tokens:     tokens,
		RevokedBy:  revokedBy,
	}
	res, err := u.authGRPCClient.RevokeToken(ctx, req)
	return res.GetRevoked(), err
}

func (u *UserService) verifyPermission(ctx context.Context, identifier *proto_infra.UserIdentifier, permissionAction string) (bool, error) {
	req := &proto_auth.VerifyUserResourceRequest{
		Identifier:   identifier,
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
		Resources: []*proto_auth.VerifyResource{
			{
				Resource: &proto_auth.VerifyResource_Permission{
					Permission: &proto_auth.Permission{
						Identifier: &proto_auth.Permission_Permission{
							Permission: &proto_auth.PermissionIdentifier{
								Resource: model_auth.ResourceTypeUser,
								Action:   permissionAction,
							},
						},
					},
				},
			},
		},
	}
	res, err := u.rbacGRPCClient.VerifyUserResource(ctx, req)
	if err != nil {
		return false, err
	}
	resources := res.GetResources()
	if len(resources) == 0 {
		return false, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("unexpeted response error"))
	}
	return resources[0].GetPermission().GetHasPermission().GetValue(), nil
}

func (u *UserService) getUser(tenantID string, accountID string, filterType filterType) (*model_core.User, error) {
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
