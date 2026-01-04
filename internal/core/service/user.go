package service

import (
	"context"

	collection "erp.localhost/internal/core/collections"
	mongo "erp.localhost/internal/infra/db/mongo"
	"erp.localhost/internal/infra/logging"
	core_models "erp.localhost/internal/infra/models/core"
	mongo_models "erp.localhost/internal/infra/models/db/mongo"
	shared_models "erp.localhost/internal/infra/models/shared"
	core_proto "erp.localhost/internal/infra/proto/core/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newCollectionHandler[T any](collection string) *mongo.BaseCollectionHandler[T] {
	logger := logging.NewLogger(shared_models.ModuleCore)
	return mongo.NewBaseCollectionHandler[T](string(collection), logger)
}

type UserService struct {
	logger         *logging.Logger
	userCollection *collection.UserCollection
	core_proto.UnimplementedUserServiceServer
}

func NewUserService() *UserService {
	logger := logging.NewLogger(shared_models.ModuleAuth)
	userCollectionHandler := newCollectionHandler[core_models.User](string(mongo_models.UsersCollection))
	if userCollectionHandler == nil {
		logger.Fatal("failed to create users collection handler")
		return nil
	}
	userCollection := collection.NewUserCollection(userCollectionHandler)
	if userCollection == nil {
		logger.Fatal("failed to create users collection")
		return nil
	}
	return &UserService{
		logger:         logger,
		userCollection: userCollection,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *core_proto.CreateUserRequest) (*core_proto.CreateUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateUser not implemented")
}

func (s *UserService) GetUser(ctx context.Context, req *core_proto.GetUserRequest) (*core_proto.UserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetUser not implemented")
}

func (s *UserService) GetUsers(ctx context.Context, req *core_proto.GetUsersRequest) (grpc.ServerStreamingClient[core_proto.UserResponse], error) {
	return nil, status.Error(codes.Unimplemented, "method GetUsers not implemented")
}

func (s *UserService) UpdateUser(ctx context.Context, req *core_proto.UpdateUserRequest) (*core_proto.UpdateUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateUser not implemented")
}

func (s *UserService) DeleteUser(ctx context.Context, req *core_proto.DeleteUserRequest) (*core_proto.DeleteUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteUser not implemented")
}
