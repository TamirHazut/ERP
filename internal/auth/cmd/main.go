package cmd

import (
	authv1 "erp.localhost/internal/auth/proto/auth/v1"
	"erp.localhost/internal/auth/service"
	infra_grpc "erp.localhost/internal/infra/grpc"
	shared_models "erp.localhost/internal/infra/models/shared"
	"google.golang.org/grpc"
)

// TODO: when breaking to microservices, this will be the entry point for the auth service
func Main() {
	certs := &shared_models.Certs{}
	services := map[*grpc.ServiceDesc]any{
		&authv1.AuthService_ServiceDesc: service.NewAuthService(),
	}
	grpcServer := infra_grpc.NewGRPCServer(certs, shared_models.ModuleAuth, 5000, services)
	if grpcServer == nil {
		return
	}
	grpcServer.ListenAndServe()
}
