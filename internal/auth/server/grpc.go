package server

import (
	"net"

	auth_proto "erp.localhost/internal/auth/proto/v1"
	auth_service "erp.localhost/internal/auth/service"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartAuthServicegRPCServer(grpcServer *grpc.Server) {
	logger := logging.NewLogger(logging.ModuleAuth)
	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		logger.Fatal("failed to listen", "error", erp_errors.Internal(erp_errors.InternalGRPCError, err))
	}
	auth_proto.RegisterAuthServiceServer(grpcServer, &auth_service.AuthService{})
	reflection.Register(grpcServer)
	if err = grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve", "error", erp_errors.Internal(erp_errors.InternalGRPCError, err))
	}
}
