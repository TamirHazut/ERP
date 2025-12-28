package service

import (
	"context"

	collection "erp.localhost/internal/auth/collections"
	user_proto "erp.localhost/internal/auth/proto/user/v1"
	mongo "erp.localhost/internal/db/mongo"
	"erp.localhost/internal/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	logger         *logging.Logger
	userCollection *collection.UserCollection
	user_proto.UnimplementedUserServiceServer
}

func NewUserService() *UserService {
	logger := logging.NewLogger(logging.ModuleAuth)
	dbHandler := mongo.NewMongoDBManager(mongo.AuthDB)
	if dbHandler == nil {
		logger.Fatal("failed to create db handler")
		return nil
	}
	userCollection := collection.NewUserCollection(dbHandler)
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

func (s *UserService) mustEmbedUnimplementedUserServiceServer() {
	panic("unimplemented")
}
