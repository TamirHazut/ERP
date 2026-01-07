package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"erp.localhost/internal/auth/service"
	grpc_server "erp.localhost/internal/infra/grpc/server"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
	"google.golang.org/grpc"
)

const (
	serverPort = 5000
)

// TODO: when breaking to microservices, this will be the entry point for the auth service
func Main() {
	// Channel to listen for OS signals for graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to signal the gRPC server goroutine to stop
	quit := make(chan struct{})

	certs := model_shared.NewCerts()
	if certs == nil {
		return
	}
	rbacService := service.NewRBACService()
	if rbacService == nil {
		return
	}
	authService := service.NewAuthService()
	if authService == nil {
		return
	}
	services := map[*grpc.ServiceDesc]any{
		&proto_auth.RBACService_ServiceDesc: rbacService,
		&proto_auth.AuthService_ServiceDesc: authService,
	}
	grpcServer := grpc_server.NewGRPCServer(certs, model_shared.ModuleAuth, serverPort, services)

	if grpcServer == nil {
		return
	}

	// WaitGroup to wait for the gRPC server goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Run gRPC Server
		if err := grpcServer.ListenAndServe(quit); err != nil {
			return
		}
	}()

	// Wait for OS signal
	<-stopChan

	// Signal the gRPC server to stop
	close(quit)

	// Wait for the gRPC server goroutine to finish
	wg.Wait()
}
