package service

import (
	"context"

	collection "erp.localhost/internal/core/collections"
	user_proto "erp.localhost/internal/core/proto/user/v1"
	mongo "erp.localhost/internal/infra/db/mongo"
	"erp.localhost/internal/infra/logging"
	core_models "erp.localhost/internal/infra/models/core"
	mongo_models "erp.localhost/internal/infra/models/db/mongo"
	shared_models "erp.localhost/internal/infra/models/shared"
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
	user_proto.UnimplementedUserServiceServer
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

func (s *UserService) CreateUser(ctx context.Context, req *user_proto.CreateUserRequest) (*user_proto.CreateUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateUser not implemented")
}

func (s *UserService) GetUser(ctx context.Context, req *user_proto.GetUserRequest) (*user_proto.GetUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetUser not implemented")
}

func (s *UserService) GetUsers(ctx context.Context, req *user_proto.GetUsersRequest) (*user_proto.GetUsersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetUsers not implemented")
}

func (s *UserService) UpdateUser(ctx context.Context, req *user_proto.UpdateUserRequest) (*user_proto.UpdateUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateUser not implemented")
}

func (s *UserService) DeleteUser(ctx context.Context, req *user_proto.DeleteUserRequest) (*user_proto.DeleteUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteUser not implemented")
}
