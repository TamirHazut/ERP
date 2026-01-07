package service

import (
	"context"

	collection "erp.localhost/internal/core/collection"
	mongo_collection "erp.localhost/internal/infra/db/mongo/collection"
	"erp.localhost/internal/infra/logging/logger"
	model_core "erp.localhost/internal/infra/model/core"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_core "erp.localhost/internal/infra/proto/core/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	logger         logger.Logger
	userCollection *collection.UserCollection
	// rbacGRPCClient proto_auth.RBACServiceClient
	proto_core.UnimplementedUserServiceServer
}

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

func (s *UserService) CreateUser(ctx context.Context, req *proto_core.CreateUserRequest) (*proto_core.UserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateUser not implemented")
}

func (s *UserService) GetUser(ctx context.Context, req *proto_core.GetUserRequest) (*proto_core.UserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetUser not implemented")
}

func (s *UserService) GetUsers(ctx context.Context, req *proto_core.GetUsersRequest) (grpc.ServerStreamingClient[proto_core.UserResponse], error) {
	return nil, status.Error(codes.Unimplemented, "method GetUsers not implemented")
}

func (s *UserService) UpdateUser(ctx context.Context, req *proto_core.UpdateUserRequest) (*proto_core.UpdateUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateUser not implemented")
}

func (s *UserService) DeleteUser(ctx context.Context, req *proto_core.DeleteUserRequest) (*proto_core.DeleteUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteUser not implemented")
}
